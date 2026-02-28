package middleware

import (
	"github.com/gin-gonic/gin"
	"lingosql/internal/utils"
)

const requestIDHeader = "X-Request-Id"

// RequestIDMiddleware 请求 ID 中间件
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(requestIDHeader)
		if requestID == "" {
			requestID = utils.GenerateRequestID()
		}
		if requestID != "" {
			c.Set("request_id", requestID)
			c.Writer.Header().Set(requestIDHeader, requestID)
		}
		c.Next()
	}
}
