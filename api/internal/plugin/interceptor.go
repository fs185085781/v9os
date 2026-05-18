package plugin

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fs185085781/v9os/internal/ioc"
	"github.com/fs185085781/v9os/internal/ioc/uioc"
	"github.com/fs185085781/v9os/internal/model/plugin"
	"github.com/fs185085781/v9os/internal/plugin/manager"
	"github.com/fs185085781/v9os/pkg/util"
	"gorm.io/gorm"
)

// CallInterceptor looks up an enabled interceptor and invokes the backing plugin.
// It returns found=false when no plugin has registered the interceptor.
func CallInterceptor(action string, params map[string]interface{}, result interface{}) (bool, error) {
	feature, found, err := enabledInterceptorFeature(action)
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}
	pm := uioc.Get[*manager.AllPluginManage](ioc.KeyPluginManage).Switch(1)
	host, err := pm.PluginHost(feature.PluginCode, nil)
	if err != nil {
		return true, fmt.Errorf("plugin unavailable: %w", err)
	}
	reqData, _ := json.Marshal(params)
	body, err := util.Post(nil, host+"/plugin/"+action, reqData, nil)
	if err != nil {
		return true, fmt.Errorf("plugin call failed: %w", err)
	}
	var resp struct {
		Code int             `json:"code"`
		Msg  string          `json:"msg"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return true, fmt.Errorf("invalid plugin response: %w", err)
	}
	if resp.Code != 0 {
		return true, fmt.Errorf("%s", resp.Msg)
	}
	if result != nil && len(resp.Data) > 0 {
		return true, json.Unmarshal(resp.Data, result)
	}
	return true, nil
}

// GetInterceptorPluginCode returns the plugin code for an enabled interceptor.
func GetInterceptorPluginCode(action string) string {
	feature, found, err := enabledInterceptorFeature(action)
	if err != nil || !found {
		return ""
	}
	return feature.PluginCode
}

func enabledInterceptorFeature(action string) (plugin.PluginFeature, bool, error) {
	var feature plugin.PluginFeature
	err := uioc.Cache().MemCacheObject("plugin:interceptor:"+action, &feature, func() interface{} {
		var item plugin.PluginFeature
		err := uioc.Database().Read().Where("content = ? AND enabled = 1", action).First(&item).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return item
		}
		return item
	}, 5*time.Minute)
	if err != nil {
		err = uioc.Database().Read().Where("content = ? AND enabled = 1", action).First(&feature).Error
	}
	if err == gorm.ErrRecordNotFound || feature.ID == 0 {
		return feature, false, nil
	}
	return feature, true, err
}
