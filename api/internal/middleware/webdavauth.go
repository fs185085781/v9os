package middleware

import (
	"net/http"

	"github.com/fs185085781/v9os/internal/logger"
	"github.com/gin-gonic/gin"
)

// Auth JWT认证中间件
type WebdavAuth struct {
}

// NewAuth 构造函数
func NewWebdavAuth(log logger.Logger) *WebdavAuth {
	res := &WebdavAuth{}
	log.Println("[Webdav认证中间件]已初始化")
	return res
}

// Middleware 生成Gin中间件
func (a *WebdavAuth) Middleware() (string, gin.HandlerFunc) {
	if a == nil {
		return "davauth", func(c *gin.Context) { c.Next() }
	}
	return "davauth", func(c *gin.Context) {
		tokenString := ""
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": -1,
				"msg":  "Webdav未登录",
			})
			return
		}
		c.Next()
	}
}
