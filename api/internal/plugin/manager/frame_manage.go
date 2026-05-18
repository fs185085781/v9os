package manager

import (
	"crypto/md5"
	"fmt"
	"strings"

	infaceUser "github.com/fs185085781/v9os/internal/inface/user"
	"github.com/fs185085781/v9os/internal/ioc"
	"github.com/fs185085781/v9os/internal/ioc/uioc"
	"github.com/fs185085781/v9os/internal/model/plugin"
	"github.com/fs185085781/v9os/internal/store"
	"github.com/gin-gonic/gin"
)

type framePluginManage struct {
	common *commonPluginManage
}

func (o *commonPluginManage) newFrameManage() IPluginManage {
	m := &framePluginManage{common: o}
	m.syncPluginFeatures("", "")
	return m
}

func (m *framePluginManage) GetPluginName(code, key string) (string, error) {
	return "", m.common.unsupportedPluginAction("GetPluginName")
}

func (m *framePluginManage) PluginHost(pluginName string, ctx *gin.Context) (string, error) {
	var pluginModel plugin.Plugin
	err := uioc.Database().Read().Where("code = ?", pluginName).First(&pluginModel).Error
	if err != nil {
		return "", err
	}
	if pluginModel.Status != 1 {
		return "", fmt.Errorf("frame plugin %s is disabled", pluginName)
	}
	if pluginModel.PluginType != 4 {
		return "", fmt.Errorf("plugin %s type is invalid", pluginName)
	}
	accessURL := strings.TrimSpace(pluginModel.AccessUrl)
	if accessURL == "" {
		return "", fmt.Errorf("frame plugin access url not found")
	}
	return accessURL, nil
}

func (m *framePluginManage) Close(pluginName string) error {
	return nil
}

func (m *framePluginManage) CloseAll() error {
	return nil
}

func (m *framePluginManage) Stop(pluginName string) {
}

func (m *framePluginManage) Install(appInfo *store.AppInfo, opts InstallOptions) (*plugin.Plugin, error) {
	if strings.TrimSpace(appInfo.AccessUrl) == "" {
		return nil, fmt.Errorf("frame plugin access url not found")
	}
	code := framePluginCode(appInfo.AccessUrl)
	pluginModel := plugin.Plugin{
		Name:        appInfo.Name,
		Description: appInfo.Description,
		Code:        code,
		Status:      appInfo.Status,
		Remark:      appInfo.Remark,
		Version:     appInfo.Version,
		PluginType:  4,
		IconUrl:     appInfo.IconUrl,
		NeedLogin:   appInfo.NeedLogin,
		AccessUrl:   strings.TrimSpace(appInfo.AccessUrl),
	}
	if pluginModel.Status == 0 {
		pluginModel.Status = 1
	}
	if pluginModel.Version == "" {
		pluginModel.Version = "0.0.0"
	}
	if err := m.common.upsertPluginModel(&pluginModel); err != nil {
		return nil, err
	}
	m.syncPluginFeatures(pluginModel.Code, pluginModel.Name)
	return &pluginModel, nil
}

func (m *framePluginManage) Uninstall(pluginModel plugin.Plugin) error {
	code := strings.TrimSpace(pluginModel.Code)
	if code == "" {
		code = framePluginCode(pluginModel.AccessUrl)
	}
	m.deletePluginAuths(code)
	return m.common.deletePluginModel(code)
}

func (m *framePluginManage) PluginDir(code string) string {
	return ""
}

func (m *framePluginManage) syncPluginFeatures(code, name string) {
	list := []interface{}{
		map[string]interface{}{
			"feature": "访问",
			"label":   "访问首页",
			"method":  "access",
		},
	}
	alls := []string{}
	if code != "" {
		alls = append(alls, code)
	} else {
		var plugins []plugin.Plugin
		uioc.Database().Read().Where("status = ? and plugin_type = ?", 1, 4).Find(&plugins)
		for _, p := range plugins {
			if p.Code != "" {
				alls = append(alls, p.Code)
			}
		}
	}
	if len(alls) == 0 {
		return
	}
	userProvider := ioc.Ioc().Get(ioc.KeyUserProvider)
	if userProvider == nil {
		return
	}
	for _, itemCode := range alls {
		userProvider.(infaceUser.UserProvider).SyncPluginAuths(itemCode, name, list)
	}
}

func (m *framePluginManage) deletePluginAuths(code string) {
	if code == "" {
		return
	}
	uioc.Database().Write().Where("plugin_code = ?", code).Delete(&plugin.PluginFeature{})
	userProvider := ioc.Ioc().Get(ioc.KeyUserProvider)
	if userProvider != nil {
		userProvider.(infaceUser.UserProvider).SyncPluginAuths(code, "", []interface{}{})
	}
}

func framePluginCode(accessURL string) string {
	return fmt.Sprintf("frame_%x", md5.Sum([]byte(strings.TrimSpace(accessURL))))
}
