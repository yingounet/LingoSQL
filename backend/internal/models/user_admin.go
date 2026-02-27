package models

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username   string `json:"username" binding:"required"`
	Host       string `json:"host,omitempty"`        // MySQL only
	Password   string `json:"password" binding:"required"`
	IsSuperuser bool `json:"is_superuser,omitempty"` // PostgreSQL only
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	Username    string `json:"username" binding:"required"`
	Host        string `json:"host,omitempty"` // MySQL only
	NewPassword string `json:"new_password" binding:"required"`
}

// DeleteUserRequest 删除用户请求
type DeleteUserRequest struct {
	Username string `json:"username" binding:"required"`
	Host     string `json:"host,omitempty"` // MySQL only
}
