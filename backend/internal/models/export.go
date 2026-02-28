package models

// ExportDataRequest 导出请求
type ExportDataRequest struct {
	ConnectionID int    `json:"connection_id" binding:"required"`
	Database     string `json:"database" binding:"required"`
	Table        string `json:"table" binding:"required"`
	Format       string `json:"format" binding:"omitempty,oneof=csv json"`
	MaxRows      int    `json:"max_rows"`
}
