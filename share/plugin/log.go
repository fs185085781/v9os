package plugin

import (
	"errors"

	"github.com/spf13/cast"
)

// 实现插件日志
func Log(level string, msg string, fields map[string]interface{}) error {
	data := make(map[string]interface{})
	data["level"] = level
	data["msg"] = msg
	data["fields"] = fields
	resultMap, err := httpPost("/log/set", data)
	if err != nil {
		return err
	}
	if resultMap != nil && cast.ToInt(resultMap["code"]) == 0 {
		return nil
	}
	if resultMap == nil {
		return errors.New("set log failed")
	}
	return errors.New(cast.ToString(resultMap["msg"]))
}
