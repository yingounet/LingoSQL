package mcp

import "context"

type contextKey string

const userIDKey contextKey = "user_id"

// WithUserID 将 user_id 注入 context
func WithUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserID 从 context 获取 user_id
func GetUserID(ctx context.Context) (int, bool) {
	v := ctx.Value(userIDKey)
	if v == nil {
		return 0, false
	}
	uid, ok := v.(int)
	return uid, ok
}
