package middleware

import (
	"net/http"

	"github.com/fs185085781/v9os/internal/inface/user"
	"github.com/fs185085781/v9os/internal/ioc"
	"github.com/fs185085781/v9os/internal/logger"
	"github.com/gin-gonic/gin"
)

// Auth JWT认证中间件
type Auth struct {
}

// NewAuth 构造函数
func NewAuth(log logger.Logger) *Auth {
	res := &Auth{}
	log.Println("[认证中间件]已初始化")
	return res
}

// Middleware 生成Gin中间件
func (a *Auth) Middleware() (string, gin.HandlerFunc) {
	if a == nil {
		return "auth", func(c *gin.Context) { c.Next() }
	}
	return "auth", func(c *gin.Context) {
		userID, ok := c.Get("userID")
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": -1,
				"msg":  "Authorization header required",
			})
			return
		}
		// 权限检查
		uid, _ := userID.(string)
		userProvider := ioc.Ioc().Get(ioc.KeyUserProvider).(user.UserProvider)
		if uid != "" && !userProvider.CheckActionAuth(uid, c.Request.URL.Path) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code": -1,
				"msg":  "无权限访问该功能",
			})
			return
		}
		c.Next()
	}
}
