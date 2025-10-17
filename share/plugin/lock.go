package plugin

import (
	"github.com/spf13/cast"
)

// 实现分布式锁
func TryLock(key string) (string, bool) {
	data := make(map[string]interface{})
	data["key"] = key
	resultMap, err := httpPost("/lock/tryLock", data)
	if err != nil {
		return "", false
	}
	if resultMap != nil && cast.ToInt(resultMap["code"]) == 0 {
		return cast.ToString(resultMap["data"]), true
	}
	return "", false
}
func UnLock(key, val string) {
	data := make(map[string]interface{})
	data["key"] = key
	data["val"] = val
	for i := 0; i < 3; i++ {
		resultMap, err := httpPost("/lock/unLock", data)
		if err != nil {
			continue
		}
		if resultMap != nil && cast.ToInt(resultMap["code"]) == 0 {
			return
		}
	}
}
