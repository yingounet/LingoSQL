package models

// ViewInfo 视图信息
type ViewInfo struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
	CreatedAt  string `json:"created_at,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
}

// ProcedureInfo 存储过程信息
type ProcedureInfo struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
	Parameters string `json:"parameters,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
}

// FunctionInfo 函数信息
type FunctionInfo struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
	ReturnType string `json:"return_type"`
	Parameters string `json:"parameters,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
}

// TriggerInfo 触发器信息
type TriggerInfo struct {
	Name      string `json:"name"`
	Event     string `json:"event"`
	Table     string `json:"table"`
	Timing    string `json:"timing"`
	Definition string `json:"definition"`
	CreatedAt string `json:"created_at,omitempty"`
}

// EventInfo 事件信息（MySQL）
type EventInfo struct {
	Name         string `json:"name"`
	Definition   string `json:"definition"`
	Status       string `json:"status"`
	OnCompletion string `json:"on_completion"`
	CreatedAt    string `json:"created_at,omitempty"`
}

// CreateViewRequest 创建视图请求
type CreateViewRequest struct {
	Database   string `json:"database" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Definition string `json:"definition" binding:"required"`
}

// CreateProcedureRequest 创建存储过程请求
type CreateProcedureRequest struct {
	Database   string `json:"database" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Definition string `json:"definition" binding:"required"`
}

// ExecuteProcedureRequest 执行存储过程请求
type ExecuteProcedureRequest struct {
	Database   string        `json:"database" binding:"required"`
	Parameters []interface{} `json:"parameters,omitempty"`
}

// CreateFunctionRequest 创建函数请求
type CreateFunctionRequest struct {
	Database   string `json:"database" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Definition string `json:"definition" binding:"required"`
	ReturnType string `json:"return_type" binding:"required"`
}

// CreateTriggerRequest 创建触发器请求
type CreateTriggerRequest struct {
	Database   string `json:"database" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Table      string `json:"table" binding:"required"`
	Event      string `json:"event" binding:"required"`
	Timing     string `json:"timing" binding:"required"`
	Definition string `json:"definition" binding:"required"`
}

// CreateEventRequest 创建事件请求
type CreateEventRequest struct {
	Database   string `json:"database" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Definition string `json:"definition" binding:"required"`
}
