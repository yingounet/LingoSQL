package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"lingosql/internal/utils"
)

type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    int
	window   time.Duration
}

type visitor struct {
	count    int
	resetAt  time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}
}

func (r *rateLimiter) allow(ip string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	v, ok := r.visitors[ip]
	if !ok || now.After(v.resetAt) {
		r.visitors[ip] = &visitor{count: 1, resetAt: now.Add(r.window)}
		return true
	}
	if v.count >= r.limit {
		return false
	}
	v.count++
	return true
}

// RateLimitMiddleware 简单限流中间件（按 IP）
func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	limiter := newRateLimiter(limit, window)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.allow(ip) {
			utils.Error(c, http.StatusTooManyRequests, "请求过于频繁")
			c.Abort()
			return
		}
		c.Next()
	}
}
