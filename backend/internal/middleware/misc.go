package middleware

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/emplyra/backend/internal/utils"
)

// RateLimiter implements a simple per-IP in-memory token bucket limiter.
type RateLimiter struct {
	mu        sync.Mutex
	limiters  map[string]*rate.Limiter
	perSecond rate.Limit
	burst     int
}

func NewRateLimiter(perSecond rate.Limit, burst int) *RateLimiter {
	return &RateLimiter{
		limiters:  map[string]*rate.Limiter{},
		perSecond: perSecond,
		burst:     burst,
	}
}

func (r *RateLimiter) get(ip string) *rate.Limiter {
	r.mu.Lock()
	defer r.mu.Unlock()
	if l, ok := r.limiters[ip]; ok {
		return l
	}
	l := rate.NewLimiter(r.perSecond, r.burst)
	r.limiters[ip] = l
	if len(r.limiters) > 10000 {
		r.limiters = map[string]*rate.Limiter{ip: l}
	}
	return l
}

func (r *RateLimiter) Middleware() gin.HandlerFunc {
	if r == nil {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !r.get(ip).Allow() {
			utils.Abort(c, http.StatusTooManyRequests, "rate limit exceeded", nil)
			return
		}
		c.Next()
	}
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		slog.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"ip", c.ClientIP(),
		)
	}
}

// RequestSizeLimit caps the size of the request body (with recovery of the error result).
func RequestSizeLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}

// SecurityHeaders adds sane HTTP security headers.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Next()
	}
}
