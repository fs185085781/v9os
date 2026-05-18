package manager

import (
	"fmt"

	"github.com/fs185085781/v9os/internal/model/plugin"
	"github.com/fs185085781/v9os/internal/store"
	"github.com/gin-gonic/gin"
)

type InstallOptions struct {
	Upgrade      bool
	AccessOrigin string
	Progress     func(percent int, msg string) bool
}

type IPluginManage interface {
	// GetPluginName resolves a runtime key to a main plugin code.
	GetPluginName(code, key string) (string, error)
	PluginHost(pluginName string, ctx *gin.Context) (string, error)
	Close(pluginName string) error
	CloseAll() error
	Stop(pluginName string)
	Install(appInfo *store.AppInfo, opts InstallOptions) (*plugin.Plugin, error)
	Uninstall(pluginModel plugin.Plugin) error
	PluginDir(code string) string
}

func (o *commonPluginManage) unsupportedPluginAction(action string) error {
	return fmt.Errorf("%s is unsupported for this plugin manager", action)
}
