package manager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/internal/inface/distributed"
	infaceUser "github.com/fs185085781/v9os/internal/inface/user"
	"github.com/fs185085781/v9os/internal/ioc"
	"github.com/fs185085781/v9os/internal/ioc/uioc"
	"github.com/fs185085781/v9os/internal/logger"
	"github.com/fs185085781/v9os/internal/model/plugin"
	"github.com/fs185085781/v9os/internal/store"
	"github.com/gin-gonic/gin"
)

type webPluginManage struct {
	cfg    config.Config
	log    logger.Logger
	common *commonPluginManage
}

func (m *webPluginManage) pluginEntryPath(code string) string {
	return filepath.ToSlash(filepath.Join("/api/webplugin", code, "index.html"))
}

func (o *commonPluginManage) newWebManage(cfg config.Config, log logger.Logger) IPluginManage {
	res := &webPluginManage{common: o, cfg: cfg, log: log}
	res.syncPluginFeatures("", "")
	return res
}

func (m *webPluginManage) PluginDir(code string) string {
	return filepath.Join(m.common.pluginRootDir(), "web", code)
}

func (m *webPluginManage) installPackage(appInfo *store.AppInfo, progress func(int, string) bool) (*packageManifest, error) {
	result, err := m.common.installPluginPackage(m.common.resolvePackageURL(appInfo), m.PluginDir(appInfo.Code), 2, appInfo.Code, m.cfg, m.log, progress)
	if err != nil {
		return nil, err
	}
	return result.Manifest, nil
}

func (m *webPluginManage) Install(appInfo *store.AppInfo, opts InstallOptions) (*plugin.Plugin, error) {
	manifest, err := m.installPackage(appInfo, opts.Progress)
	if err != nil {
		return nil, err
	}
	pluginModel := m.common.buildPluginModel(appInfo, manifest)
	m.common.snapshotPluginIcon(&pluginModel, m.PluginDir(pluginModel.Code))
	if err := m.common.upsertPluginModel(&pluginModel); err != nil {
		return nil, err
	}
	m.syncPluginFeatures(pluginModel.Code, pluginModel.Name)
	return &pluginModel, nil
}

func (m *webPluginManage) InstallLocalPackage(zipPath string, opts InstallOptions) (*plugin.Plugin, error) {
	result, err := m.common.installLocalPluginPackage(zipPath, m.PluginDir, 2, "", m.cfg)
	if err != nil {
		return nil, err
	}
	pluginModel := result.Manifest.toPluginModel()
	m.common.snapshotPluginIcon(&pluginModel, m.PluginDir(pluginModel.Code))
	if err := m.common.upsertPluginModel(&pluginModel); err != nil {
		return nil, err
	}
	m.syncPluginFeatures(pluginModel.Code, pluginModel.Name)
	return &pluginModel, nil
}

func (m *webPluginManage) Uninstall(pluginModel plugin.Plugin) error {
	if err := os.RemoveAll(m.PluginDir(pluginModel.Code)); err != nil {
		return err
	}
	return m.common.deletePluginModel(pluginModel.Code)
}

func (m *webPluginManage) syncPluginFeatures(code, name string) {
	list := []interface{}{}
	list = append(list, map[string]interface{}{
		"feature": "访问",
		"label":   "访问首页",
		"method":  "access",
	})
	list = append(list, map[string]interface{}{
		"feature": "访问",
		"label":   "写数据",
		"method":  "set",
	})
	list = append(list, map[string]interface{}{
		"feature": "访问",
		"label":   "读数据",
		"method":  "get",
	})
	list = append(list, map[string]interface{}{
		"feature": "访问",
		"label":   "删数据",
		"method":  "del",
	})
	alls := []string{}
	if code != "" {
		alls = append(alls, code)
	} else {
		dbPool := uioc.Database()
		var plugins []plugin.Plugin
		dbPool.Read().Where("status = ? and plugin_type = ?", 1, 2).Find(&plugins)
		for _, p := range plugins {
			alls = append(alls, p.Code)
		}
	}
	if len(alls) > 0 {
		userProvider := ioc.Ioc().Get(ioc.KeyUserProvider).(infaceUser.UserProvider)
		for _, itemCode := range alls {
			userProvider.SyncPluginAuths(itemCode, name, list)
		}
	}
}

func (m *webPluginManage) GetPluginName(code, key string) (string, error) {
	return "", m.common.unsupportedPluginAction("GetPluginName")
}

func (m *webPluginManage) PluginHost(pluginName string, ctx *gin.Context) (string, error) {
	db := uioc.Database()
	var pluginModel plugin.Plugin
	err := db.Read().Where("code = ?", pluginName).First(&pluginModel).Error
	if err != nil {
		return "", err
	}
	if pluginModel.Status != 1 {
		return "", fmt.Errorf("web plugin %s is disabled", pluginName)
	}
	if pluginModel.PluginType != 2 {
		return "", fmt.Errorf("plugin %s type is invalid", pluginName)
	}
	indexJSON := filepath.Join(m.PluginDir(pluginName), "index.json")
	indexHTML := filepath.Join(m.PluginDir(pluginName), "index.html")
	if !m.common.fileExists(indexJSON) || !m.common.fileExists(indexHTML) {
		_, err = m.installPackage(&store.AppInfo{
			Code:       pluginName,
			Version:    pluginModel.Version,
			PluginType: 2,
		}, nil)
		if err != nil {
			return "", err
		}
		if !m.common.fileExists(indexJSON) || !m.common.fileExists(indexHTML) {
			return "", fmt.Errorf("web plugin %s not installed", pluginName)
		}
	}
	m.syncLocalPackage(pluginName)
	return m.pluginEntryPath(pluginName), nil
}

func (m *webPluginManage) syncLocalPackage(pluginName string) {
	provider := ioc.Ioc().Get(ioc.KeyDistributedProvider)
	if provider == nil {
		return
	}
	plugins := provider.(distributed.DistributedProvider).Plugins()
	plugins.SyncLocalPluginPackage(pluginName, 2)
}

func (m *webPluginManage) Close(pluginName string) error {
	return nil
}

func (m *webPluginManage) CloseAll() error {
	return nil
}

func (m *webPluginManage) Stop(pluginName string) {

}
