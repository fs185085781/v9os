package user

import (
	"github.com/fs185085781/v9os/internal/ioc"
	"github.com/fs185085781/v9os/internal/ioc/uioc"
	"github.com/fs185085781/v9os/internal/model/plugin"
)

type DefaultProvider struct {
}

func (d *DefaultProvider) UserAuth(userID uint) []string {
	if userID == 1 {
		return []string{"all"}
	}
	return []string{}
}
func (d *DefaultProvider) UserDatascope(userID, deptID uint) *DataScope {
	return &DataScope{
		Scope: 1,
	}
}
func (d *DefaultProvider) CheckActionAuth(userID, path string) bool {
	return true
}
func (d *DefaultProvider) SyncKernelAuths(routes []*ioc.RouterStruct) {
}

func (d *DefaultProvider) SyncPluginAuths(pluginCode, pluginName string, actionsSlice []interface{}) {
}
func (d *DefaultProvider) UserAuthModules(userID uint) []plugin.Plugin {
	var ps []plugin.Plugin
	if userID == 1 {
		dbPool := uioc.Database()
		dbPool.Read().Where("status = ?", 1).Find(&ps)
		return ps
	}
	return ps
}

func initProvider() {
	if ioc.Ioc().Get(ioc.KeyUserProvider) != nil {
		return
	}
	ioc.Ioc().Register(ioc.KeyUserProvider, &DefaultProvider{})
}

func init() {
	initProvider()
}

var _ UserProvider = (*DefaultProvider)(nil)
