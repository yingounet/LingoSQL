package models

// BackupRequest 备份请求
type BackupRequest struct {
	ConnectionID int      `json:"connection_id" binding:"required"`
	Database     string   `json:"database,omitempty"`
	Tables       []string `json:"tables,omitempty"`
	Format       string   `json:"format,omitempty"` // sql, csv
}

// BackupResponse 备份响应
type BackupResponse struct {
	BackupID    string `json:"backup_id"`
	DownloadURL string `json:"download_url"`
	FileSize    int64  `json:"file_size"`
}

// RestoreRequest 恢复请求
type RestoreRequest struct {
	ConnectionID int    `json:"connection_id" binding:"required"`
	Database     string `json:"database,omitempty"`
	SQLFile      string `json:"sql_file" binding:"required"` // SQL 文件内容
}

// TableMaintenanceRequest 表维护操作请求
type TableMaintenanceRequest struct {
	ConnectionID int    `json:"connection_id" binding:"required"`
	Database     string `json:"database" binding:"required"`
	Table        string `json:"table" binding:"required"`
	Operation    string `json:"operation" binding:"required,oneof=optimize repair analyze"`
}
