package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"lingosql/internal/config"
	"lingosql/internal/dao/sqlite"
	"lingosql/internal/handler"
	"lingosql/internal/mcp"
	"lingosql/internal/middleware"
	"lingosql/internal/service"
	"lingosql/internal/utils"
	migrationsqlite "lingosql/migrations/sqlite"
	"lingosql/pkg/db"
)

func main() {
	// 加载配置
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	if err := config.Load(configPath); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	cfg := config.GetConfig()

	// 初始化日志
	initLogger(cfg)

	// 初始化 JWT
	utils.InitJWT()

	// 初始化数据库
	dbPath := cfg.Database.Path
	if dbPath == "" {
		dbPath = "./data/lingosql.db"
	}

	// 确保数据目录存在
	dataDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("创建数据目录失败: %v", err)
	}

	// 执行数据库迁移
	if err := migrationsqlite.Migrate(dbPath); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 初始化 SQLite 数据库连接
	sqliteDB, err := migrationsqlite.InitDB(dbPath)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer sqliteDB.Close()

	// 初始化 DAO
	userDAO := sqlite.NewUserDAO(sqliteDB)
	connectionDAO := sqlite.NewConnectionDAO(sqliteDB)
	favoriteDAO := sqlite.NewFavoriteDAO(sqliteDB)
	historyDAO := sqlite.NewHistoryDAO(sqliteDB)
	systemHistoryDAO := sqlite.NewSystemHistoryDAO(sqliteDB)
	auditDAO := sqlite.NewAuditDAO(sqliteDB)
	taskDAO := sqlite.NewTaskDAO(sqliteDB)
	systemSettingsDAO := sqlite.NewSystemSettingsDAO(sqliteDB)

	// 初始化服务
	authService := service.NewAuthService(userDAO)
	connectionService := service.NewConnectionService(connectionDAO)
	databaseService := service.NewDatabaseService(connectionDAO, systemHistoryDAO)
	tableService := service.NewTableService(connectionDAO, systemHistoryDAO)
	queryService := service.NewQueryService(connectionDAO, historyDAO)
	favoriteService := service.NewFavoriteService(favoriteDAO, connectionDAO)
	historyService := service.NewHistoryService(historyDAO, connectionDAO)
	systemHistoryService := service.NewSystemHistoryService(systemHistoryDAO, connectionDAO)
	databaseAdminService := service.NewDatabaseAdminService(connectionDAO, systemHistoryDAO)
	userAdminService := service.NewUserAdminService(connectionDAO, systemHistoryDAO)
	permissionAdminService := service.NewPermissionAdminService(connectionDAO, systemHistoryDAO)
	tableAdminService := service.NewTableAdminService(connectionDAO, systemHistoryDAO)
	importService := service.NewImportService(connectionDAO, systemHistoryDAO)
	exportService := service.NewExportService(connectionDAO)
	rowDataService := service.NewRowDataService(connectionDAO, systemHistoryDAO)
	dbObjectsService := service.NewDbObjectsService(connectionDAO, systemHistoryDAO)
	maintenanceService := service.NewMaintenanceService(connectionDAO, systemHistoryDAO)
	auditService := service.NewAuditService(auditDAO)
	taskService := service.NewTaskService(taskDAO)
	systemSettingsService := service.NewSystemSettingsService(systemSettingsDAO)
	installService := service.NewInstallService(sqliteDB, userDAO, systemSettingsDAO)

	// 从 system_settings 覆盖配置（已安装时）
	if err := systemSettingsService.ApplyToConfig(); err != nil {
		log.Printf("警告: 加载系统配置失败: %v", err)
	}

	// 初始化处理器
	authHandler := handler.NewAuthHandler(authService, auditService, systemSettingsService)
	connectionHandler := handler.NewConnectionHandler(connectionService, auditService)
	databaseHandler := handler.NewDatabaseHandler(databaseService)
	tableHandler := handler.NewTableHandler(tableService)
	queryHandler := handler.NewQueryHandler(queryService)
	favoriteHandler := handler.NewFavoriteHandler(favoriteService)
	historyHandler := handler.NewHistoryHandler(historyService)
	systemHistoryHandler := handler.NewSystemHistoryHandler(systemHistoryService)
	databaseAdminHandler := handler.NewDatabaseAdminHandler(databaseAdminService)
	userAdminHandler := handler.NewUserAdminHandler(userAdminService)
	permissionAdminHandler := handler.NewPermissionAdminHandler(permissionAdminService)
	tableAdminHandler := handler.NewTableAdminHandler(tableAdminService)
	importHandler := handler.NewImportHandler(importService, taskService, auditService)
	exportHandler := handler.NewExportHandler(taskService, exportService, auditService)
	rowDataHandler := handler.NewRowDataHandler(rowDataService, tableService)
	dbObjectsHandler := handler.NewDbObjectsHandler(dbObjectsService)
	maintenanceHandler := handler.NewMaintenanceHandler(maintenanceService, taskService)
	auditHandler := handler.NewAuditHandler(auditService)
	taskHandler := handler.NewTaskHandler(taskService, importService, exportService, maintenanceService)
	installHandler := handler.NewInstallHandler(installService, auditService, systemSettingsService)

	// 启动连接池清理任务
	db.GetPool().StartCleanup()

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 创建路由
	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.HTTPSOnlyMiddleware())
	r.Use(middleware.SecurityHeadersMiddleware())
	if cfg.RateLimit.Enabled {
		r.Use(middleware.RateLimitMiddleware(cfg.RateLimit.DefaultRPM, time.Minute))
	}
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(gin.Recovery())

	// 健康检查（与业务接口统一使用 {code, message, data}）
	r.GET("/health", func(c *gin.Context) {
		utils.Success(c, gin.H{"status": "ok"})
	})

	// 认证路由（不需要认证）
	api := r.Group("/api")
	{
		// 安装引导（无需认证）
		install := api.Group("/install")
		{
			install.GET("/status", installHandler.GetStatus)
			install.POST("/setup", installHandler.Setup)
		}

		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/logout", middleware.AuthMiddleware(), authHandler.Logout)
			auth.GET("/me", middleware.AuthMiddleware(), authHandler.GetMe)
		}
	}

	// 需要认证的路由
	apiAuth := api.Group("")
	apiAuth.Use(middleware.AuthMiddleware())
	{
		// MCP HTTP 端点（可选，由配置控制）
		if cfg.MCP.Enabled {
			mcpHandler := mcp.NewServer(connectionService, databaseService, tableService, queryService)
			apiAuth.Any("/mcp", mcp.GinHandler(mcpHandler))
			apiAuth.Any("/mcp/*path", mcp.GinHandler(mcpHandler))
			logrus.Info("MCP HTTP 端点已启用: /api/mcp")
		}

		// 连接管理
		connections := apiAuth.Group("/connections")
		{
			connections.GET("", connectionHandler.GetConnections)
			connections.POST("", connectionHandler.CreateConnection)
			connections.POST("/test", connectionHandler.TestConnectionConfig) // 测试未保存的连接
			connections.GET("/:id", connectionHandler.GetConnection)
			connections.PUT("/:id/last-used", connectionHandler.UpdateLastUsed)
			connections.PUT("/:id", connectionHandler.UpdateConnection)
			connections.DELETE("/:id", connectionHandler.DeleteConnection)
			connections.POST("/:id/test", connectionHandler.TestConnection)   // 测试已保存的连接
			connections.PUT("/:id/default", connectionHandler.SetDefaultConnection)
		}

		// 数据库管理
		databases := apiAuth.Group("/databases")
		{
			databases.GET("", databaseHandler.GetDatabases)
		}

		// 表管理
		tables := apiAuth.Group("/tables")
		{
			tables.GET("", tableHandler.GetTables)
			tables.GET("/info", tableHandler.GetTableInfo)
			tables.GET("/columns", tableHandler.GetTableColumns)
			tables.GET("/indexes", tableHandler.GetTableIndexes)
			tables.GET("/rows", tableHandler.GetTableRows)
			tables.PUT("/rows", tableHandler.UpdateTableRow)
		}

		// 表数据操作
		tablesData := apiAuth.Group("/tables/data")
		{
			tablesData.POST("", rowDataHandler.GetTableData)
			tablesData.POST("/update", rowDataHandler.UpdateTableRowData)
			tablesData.POST("/batch-insert", rowDataHandler.BatchInsertData)
			tablesData.POST("/batch-update", rowDataHandler.BatchUpdateData)
			tablesData.POST("/batch-delete", rowDataHandler.BatchDeleteData)
			tablesData.POST("/delete-by-keys", rowDataHandler.DeleteByPrimaryKeys)
			tablesData.POST("/compare", rowDataHandler.CompareData)
			tablesData.POST("/find-replace", rowDataHandler.FindReplaceData)
		}

		// SQL 查询
		query := apiAuth.Group("/query")
		{
			query.POST("/execute", queryHandler.Execute)
			query.POST("/explain", queryHandler.Explain)
			
			// 事务控制
			transaction := query.Group("/transaction")
			{
				transaction.POST("/begin", queryHandler.BeginTransaction)
				transaction.POST("/commit", queryHandler.CommitTransaction)
				transaction.POST("/rollback", queryHandler.RollbackTransaction)
			}
		}

		// 收藏管理
		favorites := apiAuth.Group("/favorites")
		{
			favorites.GET("", favoriteHandler.GetFavorites)
			favorites.POST("", favoriteHandler.CreateFavorite)
			favorites.POST("/:id/use", favoriteHandler.RecordFavoriteUse)
			favorites.PUT("/:id", favoriteHandler.UpdateFavorite)
			favorites.DELETE("/:id", favoriteHandler.DeleteFavorite)
		}

		// 历史记录
		history := apiAuth.Group("/history")
		{
			history.GET("", historyHandler.GetHistory)
			history.GET("/:id", historyHandler.GetHistoryByID)
			history.DELETE("/:id", historyHandler.DeleteHistory)
		}

		// 系统历史记录
		systemHistory := apiAuth.Group("/system-history")
		{
			systemHistory.GET("", systemHistoryHandler.GetSystemHistory)
			systemHistory.DELETE("/:id", systemHistoryHandler.DeleteSystemHistory)
		}

		// 审计日志
		auditLogs := apiAuth.Group("/audit-logs")
		{
			auditLogs.GET("", auditHandler.GetAuditLogs)
		}

		// 数据库管理
		admin := apiAuth.Group("/admin")
		{
			// 权限检查
			admin.GET("/permissions/check", databaseAdminHandler.CheckAdminPermissions)

			// 数据库管理
			databasesAdmin := admin.Group("/databases")
			{
				databasesAdmin.GET("", databaseAdminHandler.GetDatabaseList)
				databasesAdmin.POST("", databaseAdminHandler.CreateDatabase)
				databasesAdmin.GET("/:name/info", databaseAdminHandler.GetDatabaseInfo)
				databasesAdmin.DELETE("/:name", databaseAdminHandler.DropDatabase)
				databasesAdmin.POST("/:name/rename", databaseAdminHandler.RenameDatabase)
			}

			// 用户管理
			usersAdmin := admin.Group("/users")
			{
				usersAdmin.GET("", userAdminHandler.GetUsers)
				usersAdmin.POST("", userAdminHandler.CreateUser)
				usersAdmin.DELETE("", userAdminHandler.DeleteUser)
				usersAdmin.PUT("/password", userAdminHandler.ChangeUserPassword)
				usersAdmin.GET("/grants", userAdminHandler.GetUserGrants)
			}

			// 权限管理
			permissionsAdmin := admin.Group("/permissions")
			{
				permissionsAdmin.GET("/tree", permissionAdminHandler.GetPermissionTree)
				permissionsAdmin.POST("/grant", permissionAdminHandler.GrantPermission)
				permissionsAdmin.POST("/revoke", permissionAdminHandler.RevokePermission)
			}

			// 表管理（创建/删除表）
			tablesAdmin := admin.Group("/tables")
			{
				tablesAdmin.POST("", tableAdminHandler.CreateTable)
				tablesAdmin.DELETE("/:name", tableAdminHandler.DropTable)
				tablesAdmin.POST("/alter", tableAdminHandler.AlterTable)
				tablesAdmin.POST("/indexes", tableAdminHandler.CreateIndex)
				tablesAdmin.DELETE("/indexes/:index_name", tableAdminHandler.DropIndex)
				tablesAdmin.POST("/rename", tableAdminHandler.RenameTable)
				tablesAdmin.POST("/optimize", maintenanceHandler.OptimizeTable)
				tablesAdmin.POST("/repair", maintenanceHandler.RepairTable)
				tablesAdmin.POST("/analyze", maintenanceHandler.AnalyzeTable)
			}

			// 数据库对象管理
			viewsAdmin := admin.Group("/views")
			{
				viewsAdmin.GET("", dbObjectsHandler.GetViews)
				viewsAdmin.POST("", dbObjectsHandler.CreateView)
				viewsAdmin.DELETE("/:name", dbObjectsHandler.DropView)
			}

			proceduresAdmin := admin.Group("/procedures")
			{
				proceduresAdmin.GET("", dbObjectsHandler.GetProcedures)
				proceduresAdmin.POST("", dbObjectsHandler.CreateProcedure)
				proceduresAdmin.DELETE("/:name", dbObjectsHandler.DropProcedure)
				proceduresAdmin.POST("/:name/execute", dbObjectsHandler.ExecuteProcedure)
			}

			functionsAdmin := admin.Group("/functions")
			{
				functionsAdmin.GET("", dbObjectsHandler.GetFunctions)
				functionsAdmin.POST("", dbObjectsHandler.CreateFunction)
				functionsAdmin.DELETE("/:name", dbObjectsHandler.DropFunction)
			}

			triggersAdmin := admin.Group("/triggers")
			{
				triggersAdmin.GET("", dbObjectsHandler.GetTriggers)
				triggersAdmin.POST("", dbObjectsHandler.CreateTrigger)
				triggersAdmin.DELETE("/:name", dbObjectsHandler.DropTrigger)
			}

			eventsAdmin := admin.Group("/events")
			{
				eventsAdmin.GET("", dbObjectsHandler.GetEvents)
				eventsAdmin.POST("", dbObjectsHandler.CreateEvent)
				eventsAdmin.DELETE("/:name", dbObjectsHandler.DropEvent)
			}

			// 数据库维护
			admin.POST("/backup", maintenanceHandler.BackupDatabase)
			admin.POST("/restore", maintenanceHandler.RestoreDatabase)

			// 备份管理
			admin.GET("/backups", maintenanceHandler.ListBackups)
			admin.GET("/backups/:id/download", maintenanceHandler.DownloadBackup)
			admin.DELETE("/backups/:id", maintenanceHandler.DeleteBackup)
		}

		// 数据导入
		importGroup := apiAuth.Group("/import")
		{
			importGroup.POST("/data", importHandler.ImportData)
			importGroup.POST("/sql", importHandler.ExecuteSQLFile)
			importGroup.POST("/data/async", importHandler.ImportDataAsync)
			importGroup.POST("/sql/async", importHandler.ExecuteSQLFileAsync)
		}

		// 数据导出
		exportGroup := apiAuth.Group("/export")
		{
			exportGroup.POST("/data", exportHandler.ExportDataAsync)
		}

		// 数据导入/导出（文档路径）
		dataGroup := apiAuth.Group("/data")
		{
			exportDoc := dataGroup.Group("/export")
			{
				exportDoc.POST("", exportHandler.ExportDataSync)
				exportDoc.POST("/async", exportHandler.ExportDataAsyncDoc)
				exportDoc.GET("/async", exportHandler.ListExportTasks)
				exportDoc.GET("/async/:id", exportHandler.GetExportTask)
				exportDoc.POST("/async/:id/cancel", exportHandler.CancelExportTask)
				exportDoc.POST("/async/:id/retry", exportHandler.RetryExportTask)
				exportDoc.GET("/async/:id/errors", exportHandler.GetExportErrors)
				exportDoc.GET("/:id", taskHandler.DownloadTaskResult)
			}

			importDoc := dataGroup.Group("/import")
			{
				importDoc.POST("", importHandler.ImportData)
				importDoc.POST("/preview", importHandler.ImportDataPreview)
				importDoc.POST("/async", importHandler.ImportDataAsyncDoc)
				importDoc.GET("", importHandler.ListImportTasks)
				importDoc.GET("/:id", importHandler.GetImportTask)
				importDoc.POST("/:id/cancel", importHandler.CancelImportTask)
				importDoc.POST("/:id/retry", importHandler.RetryImportTask)
				importDoc.GET("/:id/errors", importHandler.GetImportErrors)
			}
		}

		// 任务中心
		tasks := apiAuth.Group("/tasks")
		{
			if cfg.RateLimit.Enabled {
				tasks.Use(middleware.RateLimitMiddleware(cfg.RateLimit.PollingRPM, time.Minute))
			}
			tasks.GET("", taskHandler.GetTasks)
			tasks.GET("/:id", taskHandler.GetTask)
			tasks.POST("/:id/retry", taskHandler.RetryTask)
			tasks.POST("/:id/cancel", taskHandler.CancelTask)
			tasks.GET("/:id/download", taskHandler.DownloadTaskResult)
		}
	}

	// 启动服务器（支持优雅退出）
	port := cfg.Server.Port
	if port == 0 {
		port = 8080
	}
	addr := fmt.Sprintf(":%d", port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		logrus.Infof("服务器启动在端口 %d", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("正在关闭服务器...")

	db.GetPool().StopCleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("服务器优雅退出失败: %v", err)
	}
	logrus.Info("服务器已退出")
}

func initLogger(cfg *config.Config) {
	level, err := logrus.ParseLevel(cfg.Log.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	if cfg.Log.Format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{})
	}
}
