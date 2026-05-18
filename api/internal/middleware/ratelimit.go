package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/internal/logger"
	"github.com/fs185085781/v9os/pkg/util"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter 基于IP的限流器
type RateLimiter struct {
	limiters      map[string]*rate.Limiter
	lastAccess    map[string]time.Time // 记录每个IP最后访问时间
	mu            sync.Mutex
	rps           float64
	burst         int
	enabled       bool
	cleanupTicker *time.Ticker
}

// NewRateLimiter 构造函数
func NewRateLimiter(cfg *config.RateLimitConfig, log logger.Logger) *RateLimiter {
	if !cfg.Enabled {
		return nil
	}

	r := &RateLimiter{
		limiters:   make(map[string]*rate.Limiter),
		lastAccess: make(map[string]time.Time),
		rps:        cfg.RPS,
		burst:      cfg.Burst,
		enabled:    true,
	}
	r.startCleanupRoutine()
	log.Println("[限流器中间件]已初始化")
	return r
}

// Middleware 生成Gin中间件
func (r *RateLimiter) Middleware() gin.HandlerFunc {
	if r == nil || !r.enabled {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		limiter := r.getLimiter(c.ClientIP())
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"msg":  "too many requests",
				"code": -1,
			})
			return
		}
		c.Next()
	}
}

// getLimiter 获取或创建限流器
func (r *RateLimiter) getLimiter(ip string) *rate.Limiter {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 更新最后访问时间
	r.lastAccess[ip] = time.Now()

	if limiter, exists := r.limiters[ip]; exists {
		return limiter
	}

	limiter := rate.NewLimiter(rate.Limit(r.rps), r.burst)
	r.limiters[ip] = limiter
	return limiter
}

// startCleanupRoutine 启动定期清理未使用IP限流器的协程
func (r *RateLimiter) startCleanupRoutine() {
	// 每5分钟执行一次清理
	r.cleanupTicker = time.NewTicker(5 * time.Minute)
	util.Go(func() {
		for {
			select {
			case <-r.cleanupTicker.C:
				r.cleanupInactiveLimiters()
			}
		}
	})
}

// cleanupInactiveLimiters 清理长时间未使用的限流器
// 默认清理15分钟内未访问的IP限流器
func (r *RateLimiter) cleanupInactiveLimiters() {
	threshold := time.Now().Add(-15 * time.Minute)
	r.mu.Lock()
	defer r.mu.Unlock()

	for ip, lastTime := range r.lastAccess {
		if lastTime.Before(threshold) {
			delete(r.limiters, ip)
			delete(r.lastAccess, ip)
		}
	}
}

func (r *RateLimiter) StopCleanup() {
	if r.cleanupTicker != nil {
		r.cleanupTicker.Stop()
	}
}
