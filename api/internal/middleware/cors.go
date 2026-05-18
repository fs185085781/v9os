package middleware

import (
	"net/http"

	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/internal/logger"
	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件
type CORS struct {
	cfg *config.CORSConfig
}

// NewCORS 构造函数
func NewCORS(cfg *config.CORSConfig, log logger.Logger) *CORS {
	if !cfg.Enabled {
		return nil
	}
	log.Println("[CORS中间件]已初始化")
	return &CORS{cfg: cfg}
}

// Middleware 生成Gin中间件
func (c *CORS) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Methods", "*")
		ctx.Header("Access-Control-Allow-Headers", "*")
		ctx.Header("Access-Control-Expose-Headers", "*")
		ctx.Header("Access-Control-Max-Age", "86400")
		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}
		ctx.Next()
	}
}
