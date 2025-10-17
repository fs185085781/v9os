package plugin

import (
	"errors"

	"github.com/spf13/cast"
)

// 实现插件配置
func SetConfig(val string) error {
	data := make(map[string]interface{})
	data["val"] = val
	resultMap, err := httpPost("/config/set", data)
	if err != nil {
		return err
	}
	if resultMap != nil && cast.ToInt(resultMap["code"]) == 0 {
		return nil
	}
	if resultMap == nil {
		return errors.New("set config failed")
	}
	return errors.New(cast.ToString(resultMap["msg"]))
}
func GetConfig() (string, error) {
	data := make(map[string]interface{})
	resultMap, err := httpPost("/config/get", data)
	if err != nil {
		return "", err
	}
	if resultMap != nil && cast.ToInt(resultMap["code"]) == 0 {
		return cast.ToString(resultMap["data"]), nil
	}
	if resultMap == nil {
		return "", errors.New("get config failed")
	}
	return "", errors.New(cast.ToString(resultMap["msg"]))
}
