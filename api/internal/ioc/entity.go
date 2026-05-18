package ioc

import (
	"sync"

	"github.com/gin-gonic/gin"
)

type RouterStruct struct {
	Method      string
	Path        string
	Handler     func(*gin.Context)
	Ware        string
	AuthName    string // 权限名称，如 "用户管理"
	AuthFeature string // 功能组，如 "插件管理"
	AuthLabel   string // 按钮组，如 "新增/编辑"
}
type GroupRoutes struct {
	Mu     sync.Mutex
	Routes []*RouterStruct
}
