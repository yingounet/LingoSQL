package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"lingosql/internal/config"
)

// CORSMiddleware 跨域中间件
func CORSMiddleware() gin.HandlerFunc {
	cfg := config.GetConfig()
	allowedOrigins := make(map[string]struct{}, len(cfg.CORS.AllowedOrigins))
	allowAll := false
	for _, origin := range cfg.CORS.AllowedOrigins {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}
		if origin == "*" {
			allowAll = true
			continue
		}
		allowedOrigins[origin] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if allowAll {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else if origin != "" {
			if _, ok := allowedOrigins[origin]; ok {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				if c.Request.Method == "OPTIONS" {
					c.AbortWithStatus(http.StatusForbidden)
					return
				}
			}
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
