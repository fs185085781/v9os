package user

import (
	"github.com/fs185085781/v9os/internal/model/plugin"

	"github.com/fs185085781/v9os/internal/ioc"
)

type DataScope struct {
	Scope   int      `json:"scope"`   // 1=全部 2=本部门及下级 3=仅本部门 4=仅本人 5=自定义部门
	DeptIds []string `json:"deptIds"` // scope=2或5时的部门ID列表
}

type UserProvider interface {
	UserAuth(userID uint) []string
	UserDatascope(userID, deptID uint) *DataScope
	CheckActionAuth(userID, path string) bool
	SyncKernelAuths(routes []*ioc.RouterStruct)
	SyncPluginAuths(pluginCode, pluginName string, actionsSlice []interface{})
	UserAuthModules(userID uint) []plugin.Plugin
}
