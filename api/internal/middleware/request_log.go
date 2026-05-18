package middleware

import (
	"time"

	"github.com/fs185085781/v9os/internal/logger"
	"github.com/gin-gonic/gin"
)

// RequestLog 请求日志中间件
type RequestLog struct {
	log logger.Logger
}

// NewRequestLog 构造函数
func NewRequestLog(log logger.Logger) *RequestLog {
	res := &RequestLog{log: log}
	log.Println("[访问日志中间件]已初始化")
	return res
}

// Middleware 生成Gin中间件
func (r *RequestLog) Middleware() gin.HandlerFunc {
	if r == nil {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		query := c.Request.URL.RawQuery
		c.Next()
		end := time.Now()
		latency := end.Sub(start)
		r.log.Info("request log",
			logger.NewField("status", c.Writer.Status()),
			logger.NewField("method", c.Request.Method),
			logger.NewField("path", path),
			logger.NewField("query", query),
			logger.NewField("userID", c.GetString("userID")),
			logger.NewField("ip", c.ClientIP()),
			logger.NewField("user-agent", c.Request.UserAgent()),
			logger.NewField("latency", latency.String()),
			logger.NewField("time", end.Format(time.RFC3339)),
		)
	}
}
