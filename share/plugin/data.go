package plugin

import (
	"errors"

	"github.com/spf13/cast"
)

func GetData(key string) (string, error) {
	data := map[string]interface{}{"key": key}
	resultMap, err := httpPost("/data/get", data)
	if err != nil {
		return "", err
	}
	if resultMap != nil && cast.ToInt(resultMap["code"]) == 0 {
		return cast.ToString(resultMap["data"]), nil
	}
	if resultMap == nil {
		return "", errors.New("get data failed")
	}
	return "", errors.New(cast.ToString(resultMap["msg"]))
}

func SetData(key, val string) error {
	data := map[string]interface{}{"key": key, "val": val}
	resultMap, err := httpPost("/data/set", data)
	if err != nil {
		return err
	}
	if resultMap != nil && cast.ToInt(resultMap["code"]) == 0 {
		return nil
	}
	if resultMap == nil {
		return errors.New("set data failed")
	}
	return errors.New(cast.ToString(resultMap["msg"]))
}
