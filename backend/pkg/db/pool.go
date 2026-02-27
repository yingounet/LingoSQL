package db

import (
	"context"
	"fmt"
	"sync"
	"time"

	"lingosql/internal/dao/mysql"
	"lingosql/internal/dao/postgresql"
	"lingosql/pkg/types"
)

// 重新导出类型，方便其他包使用
type RowFilter = types.RowFilter
type ColumnDef = types.ColumnDef
type TableRowsResult = types.TableRowsResult

// Executor 数据库执行器接口
type Executor interface {
	Execute(sql string) ([]string, [][]interface{}, int, error)
	ExecuteUpdate(sql string) (int, int, error)
	GetDatabases() ([]string, error)
	GetTables(database string) ([]map[string]interface{}, error)
	GetTableInfo(database, table string) (map[string]interface{}, error)
	GetTableColumns(database, table string) ([]map[string]interface{}, error)
	GetTableIndexes(database, table string) ([]map[string]interface{}, error)
	GetTableRows(database, table string, filters []types.RowFilter, page, pageSize int) (*types.TableRowsResult, error)
	UpdateTableRow(database, table string, primaryKey map[string]interface{}, data map[string]interface{}) (int, error)
	GetVersion() (string, error)
	Close() error

	// 权限检查
	CheckAdminPermissions() (map[string]bool, error)

	// 数据库管理
	CreateDatabase(name string, charset string, collation string) error
	DropDatabase(name string) error
	GetDatabaseInfo(name string) (map[string]interface{}, error)
	RenameDatabase(oldName, newName string) error

	// 用户管理
	GetUsers() ([]map[string]interface{}, error)
	CreateUser(username, host, password string) error
	DropUser(username, host string) error
	AlterUserPassword(username, host, newPassword string) error
	GetUserGrants(username, host string) ([]string, error)

	// 权限管理
	GrantDatabasePrivileges(username, host, database string, privileges []string) error
	RevokeDatabasePrivileges(username, host, database string, privileges []string) error
	GrantTablePrivileges(username, host, database, table string, privileges []string) error
	RevokeTablePrivileges(username, host, database, table string, privileges []string) error
	GrantColumnPrivileges(username, host, database, table string, columnPrivileges map[string][]string) error
	RevokeColumnPrivileges(username, host, database, table string, columnPrivileges map[string][]string) error
	GetPermissionTree(username, host string) ([]map[string]interface{}, error)

	// 表管理（在指定数据库内）
	CreateTable(database string, tableName string, createDDL string) error
	DropTable(database string, tableName string) error

	// SQL 执行计划
	Explain(sql string) ([]map[string]interface{}, int, error)

	// 事务控制
	BeginTransaction() error
	CommitTransaction() error
	RollbackTransaction() error
}

// ConnectionPool 连接池管理器
type ConnectionPool struct {
	executors     map[string]Executor  // key: "connectionID:database"
	mu            sync.RWMutex
	timeouts      map[string]time.Time // key: "connectionID:database"
	cleanupCancel context.CancelFunc   // 用于停止清理 goroutine
	cleanupMu     sync.Mutex           // 保护 cleanup 生命周期
}

var pool *ConnectionPool
var once sync.Once

// GetPool 获取连接池实例
func GetPool() *ConnectionPool {
	once.Do(func() {
		pool = &ConnectionPool{
			executors: make(map[string]Executor),
			timeouts:  make(map[string]time.Time),
		}
	})
	return pool
}

// GetExecutor 获取执行器
// connectionID=0 表示临时连接，不会被缓存，调用方需要自行关闭
func (p *ConnectionPool) GetExecutor(connectionID int, dbType, host string, port int, database, username, password string) (Executor, error) {
	// connectionID=0 是临时连接，不缓存
	if connectionID == 0 {
		return p.createExecutor(dbType, host, port, database, username, password)
	}

	// 使用 connectionID:database 作为缓存 key
	cacheKey := fmt.Sprintf("%d:%s", connectionID, database)

	p.mu.RLock()
	executor, exists := p.executors[cacheKey]
	p.mu.RUnlock()

	if exists {
		// 更新超时时间
		p.mu.Lock()
		p.timeouts[cacheKey] = time.Now().Add(30 * time.Minute)
		p.mu.Unlock()
		return executor, nil
	}

	// 创建新连接
	newExecutor, err := p.createExecutor(dbType, host, port, database, username, password)
	if err != nil {
		return nil, err
	}

	// 存储连接
	p.mu.Lock()
	p.executors[cacheKey] = newExecutor
	p.timeouts[cacheKey] = time.Now().Add(30 * time.Minute)
	p.mu.Unlock()

	return newExecutor, nil
}

// createExecutor 创建新的执行器
func (p *ConnectionPool) createExecutor(dbType, host string, port int, database, username, password string) (Executor, error) {
	switch dbType {
	case "mysql", "mariadb":
		return mysql.NewMySQLExecutor(host, port, database, username, password)
	case "postgresql":
		return postgresql.NewPostgreSQLExecutor(host, port, database, username, password)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", dbType)
	}
}

// CloseConnection 关闭指定连接 ID 的所有执行器
func (p *ConnectionPool) CloseConnection(connectionID int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	prefix := fmt.Sprintf("%d:", connectionID)
	var lastErr error

	for key, executor := range p.executors {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			if err := executor.Close(); err != nil {
				lastErr = err
			}
			delete(p.executors, key)
			delete(p.timeouts, key)
		}
	}

	return lastErr
}

// CleanupExpired 清理过期连接
func (p *ConnectionPool) CleanupExpired() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	for key, timeout := range p.timeouts {
		if now.After(timeout) {
			if executor, exists := p.executors[key]; exists {
				executor.Close()
				delete(p.executors, key)
				delete(p.timeouts, key)
			}
		}
	}
}

// StartCleanup 启动定期清理任务
func (p *ConnectionPool) StartCleanup() {
	p.cleanupMu.Lock()
	defer p.cleanupMu.Unlock()
	if p.cleanupCancel != nil {
		return // 已启动
	}
	ctx, cancel := context.WithCancel(context.Background())
	p.cleanupCancel = cancel
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				p.CleanupExpired()
			}
		}
	}()
}

// StopCleanup 停止定期清理任务，用于优雅退出
func (p *ConnectionPool) StopCleanup() {
	p.cleanupMu.Lock()
	defer p.cleanupMu.Unlock()
	if p.cleanupCancel != nil {
		p.cleanupCancel()
		p.cleanupCancel = nil
	}
}
