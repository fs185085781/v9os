package plugin

import (
	"errors"

	"github.com/spf13/cast"
)

// 实现分布式缓存
func SetCache(key, val string, minute int) error {
	data := make(map[string]interface{})
	data["key"] = key
	data["val"] = val
	data["time"] = minute
	resultMap, err := httpPost("/cache/set", data)
	if err != nil {
		return err
	}
	if resultMap != nil && cast.ToInt(resultMap["code"]) == 0 {
		return nil
	}
	if resultMap == nil {
		return errors.New("set cache failed")
	}
	return errors.New(cast.ToString(resultMap["msg"]))
}
func GetCache(key string) (string, error) {
	data := make(map[string]interface{})
	data["key"] = key
	resultMap, err := httpPost("/cache/get", data)
	if err != nil {
		return "", err
	}
	if resultMap != nil && cast.ToInt(resultMap["code"]) == 0 {
		return cast.ToString(resultMap["data"]), nil
	}
	if resultMap == nil {
		return "", errors.New("get cache failed")
	}
	return "", errors.New(cast.ToString(resultMap["msg"]))
}
