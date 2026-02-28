package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"lingosql/internal/config"
	"lingosql/internal/utils"
)

// HTTPSOnlyMiddleware 在生产环境强制 HTTPS
func HTTPSOnlyMiddleware() gin.HandlerFunc {
	cfg := config.GetConfig()
	return func(c *gin.Context) {
		if cfg.Server.Mode != "release" {
			c.Next()
			return
		}
		if c.Request.TLS != nil {
			c.Next()
			return
		}
		if strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
			c.Next()
			return
		}
		utils.Error(c, http.StatusForbidden, "仅支持 HTTPS 访问")
		c.Abort()
	}
}

// SecurityHeadersMiddleware 设置安全响应头
func SecurityHeadersMiddleware() gin.HandlerFunc {
	cfg := config.GetConfig()
	return func(c *gin.Context) {
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("Referrer-Policy", "no-referrer")
		c.Writer.Header().Set("X-XSS-Protection", "0")
		if cfg.Server.Mode == "release" && (c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https")) {
			c.Writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}
