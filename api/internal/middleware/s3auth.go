package middleware

import (
	"net/http"

	"github.com/fs185085781/v9os/internal/logger"
	"github.com/gin-gonic/gin"
)

// Auth JWT认证中间件
type S3Auth struct {
}

// NewAuth 构造函数
func NewS3Auth(log logger.Logger) *S3Auth {
	res := &S3Auth{}
	log.Println("[S3认证中间件]已初始化")
	return res
}

// Middleware 生成Gin中间件
func (a *S3Auth) Middleware() (string, gin.HandlerFunc) {
	if a == nil {
		return "s3auth", func(c *gin.Context) { c.Next() }
	}
	return "s3auth", func(c *gin.Context) {
		tokenString := ""
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": -1,
				"msg":  "S3未登录",
			})
			return
		}
		c.Next()
	}
}
