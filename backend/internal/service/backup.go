package service

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"lingosql/internal/config"
	"lingosql/internal/models"
	"lingosql/internal/utils"
)

// BackupManifest 备份目录内的元数据
type BackupManifest struct {
	ConnectionID int       `json:"connection_id"`
	Database     string    `json:"database"`
	UserID       int       `json:"user_id"`
	CreatedAt    time.Time `json:"created_at"`
	Compress     bool      `json:"compress"`
	SchemaOnly   bool      `json:"schema_only"`
	Files        []struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	} `json:"files"`
	TotalSize int64 `json:"total_size"`
}

// RunBackup 执行备份，返回 backup_id、文件列表、总大小
func (s *MaintenanceService) RunBackup(connectionID, userID int, req *models.BackupRequest, taskID int, taskService *TaskService) (map[string]interface{}, error) {
	conn, err := s.connectionDAO.GetByID(connectionID)
	if err != nil {
		return nil, err
	}
	if conn.UserID != userID {
		return nil, fmt.Errorf("无权访问此连接")
	}

	var dbConfig models.DbConfig
	if err := json.Unmarshal([]byte(conn.DbConfigJSON), &dbConfig); err != nil {
		return nil, fmt.Errorf("配置解析失败")
	}
	password, err := utils.Decrypt(dbConfig.PasswordEncrypted)
	if err != nil {
		return nil, err
	}

	// 仅支持直连（SSH 隧道需要额外实现）
	if conn.ConnectionType == "ssh_tunnel" {
		return nil, fmt.Errorf("备份暂不支持 SSH 隧道连接，请使用直连模式")
	}

	cfg := config.GetConfig()
	if cfg == nil || cfg.Backup.RootPath == "" {
		return nil, fmt.Errorf("备份配置未就绪")
	}

	rootPath := cfg.Backup.RootPath
	maxFileSizeMB := cfg.Backup.MaxFileSizeMB
	if req.MaxFileSizeMB > 0 {
		maxFileSizeMB = req.MaxFileSizeMB
	}

	compress := req.Compress
	schemaOnly := req.SchemaOnly
	database := req.Database
	if database == "" {
		database = dbConfig.Database
	}
	if database == "" {
		return nil, fmt.Errorf("未指定备份数据库")
	}

	timestamp := time.Now().Format("20060102_150405")
	dirName := fmt.Sprintf("%d_%s_%s", connectionID, sanitizeDirName(database), timestamp)
	backupDir := filepath.Join(rootPath, dirName)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("创建备份目录失败: %w", err)
	}

	var files []struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	}
	var totalSize int64

	switch conn.DBType {
	case "mysql", "mariadb":
		files, totalSize, err = s.runMySQLBackup(&dbConfig, password, backupDir, database, req.Tables, compress, schemaOnly, maxFileSizeMB, taskID, taskService)
	case "postgresql":
		files, totalSize, err = s.runPostgreSQLBackup(&dbConfig, password, backupDir, database, req.Tables, compress, schemaOnly, maxFileSizeMB, taskID, taskService)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", conn.DBType)
	}
	if err != nil {
		os.RemoveAll(backupDir)
		return nil, err
	}

	manifest := BackupManifest{
		ConnectionID: connectionID,
		Database:     database,
		UserID:       userID,
		CreatedAt:    time.Now(),
		Compress:     compress,
		SchemaOnly:   schemaOnly,
		Files:        files,
		TotalSize:    totalSize,
	}
	manifestPath := filepath.Join(backupDir, "manifest.json")
	manifestData, _ := json.MarshalIndent(manifest, "", "  ")
	if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
		os.RemoveAll(backupDir)
		return nil, fmt.Errorf("写入 manifest 失败: %w", err)
	}

	downloadURL := fmt.Sprintf("/api/admin/backups/%s/download", dirName)
	return map[string]interface{}{
		"backup_id":     dirName,
		"download_url":  downloadURL,
		"file_size":     totalSize,
		"file_count":    len(files),
		"backup_dir":    backupDir,
	}, nil
}

func sanitizeDirName(s string) string {
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	return s
}

// BackupListItem 列表项（含 backup_id）
type BackupListItem struct {
	BackupID string
	BackupManifest
}

// ListBackups 列出备份
func (s *MaintenanceService) ListBackups(userID int, connectionID *int, page, pageSize int) ([]BackupListItem, int, error) {
	cfg := config.GetConfig()
	if cfg == nil || cfg.Backup.RootPath == "" {
		return nil, 0, fmt.Errorf("备份配置未就绪")
	}

	entries, err := os.ReadDir(cfg.Backup.RootPath)
	if err != nil {
		return nil, 0, err
	}

	var items []BackupListItem
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		manifestPath := filepath.Join(cfg.Backup.RootPath, e.Name(), "manifest.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}
		var m BackupManifest
		if json.Unmarshal(data, &m) != nil {
			continue
		}
		if m.UserID != userID {
			continue
		}
		if connectionID != nil && m.ConnectionID != *connectionID {
			continue
		}
		items = append(items, BackupListItem{BackupID: e.Name(), BackupManifest: m})
	}

	total := len(items)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	if offset >= total {
		return []BackupListItem{}, total, nil
	}
	end := offset + pageSize
	if end > total {
		end = total
	}
	return items[offset:end], total, nil
}

// GetBackupByID 根据 backup_id 获取备份信息
func (s *MaintenanceService) GetBackupByID(backupID string, userID int) (*BackupManifest, string, error) {
	cfg := config.GetConfig()
	if cfg == nil {
		return nil, "", fmt.Errorf("备份配置未就绪")
	}
	backupDir := filepath.Join(cfg.Backup.RootPath, backupID)
	manifestPath := filepath.Join(backupDir, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, "", fmt.Errorf("备份不存在")
	}
	var m BackupManifest
	if json.Unmarshal(data, &m) != nil {
		return nil, "", fmt.Errorf("备份元数据无效")
	}
	if m.UserID != userID {
		return nil, "", fmt.Errorf("无权访问此备份")
	}
	return &m, backupDir, nil
}

// DeleteBackup 删除备份
func (s *MaintenanceService) DeleteBackup(backupID string, userID int) error {
	_, backupDir, err := s.GetBackupByID(backupID, userID)
	if err != nil {
		return err
	}
	return os.RemoveAll(backupDir)
}

// ServeBackupDownload 提供备份下载（单文件直接返回，多文件打包 zip）
func (s *MaintenanceService) ServeBackupDownload(backupID string, userID int, singleFile string) (string, string, error) {
	m, backupDir, err := s.GetBackupByID(backupID, userID)
	if err != nil {
		return "", "", err
	}

	if singleFile != "" {
		// 下载单个文件
		for _, f := range m.Files {
			if f.Name == singleFile {
				p := filepath.Join(backupDir, f.Name)
				if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
					return p, f.Name, nil
				}
				return "", "", fmt.Errorf("文件不存在")
			}
		}
		return "", "", fmt.Errorf("文件不存在")
	}

	if len(m.Files) == 1 {
		p := filepath.Join(backupDir, m.Files[0].Name)
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p, m.Files[0].Name, nil
		}
	}

	// 多文件打包为 zip
	zipPath := filepath.Join(backupDir, "backup.zip")
	zf, err := os.Create(zipPath)
	if err != nil {
		return "", "", err
	}
	defer zf.Close()
	zw := zip.NewWriter(zf)
	for _, f := range m.Files {
		fp := filepath.Join(backupDir, f.Name)
		info, err := os.Stat(fp)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		w, err := zw.Create(f.Name)
		if err != nil {
			continue
		}
		r, _ := os.Open(fp)
		io.Copy(w, r)
		r.Close()
	}
	zw.Close()
	return zipPath, "backup.zip", nil
}
