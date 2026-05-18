package plugin

import (
	"encoding/json"
	"errors"
	"io"
)

type AppConfigData struct {
	IsRemote bool `json:"is_remote"`
}
type appConfigResp struct {
	Code int            `json:"code"`
	Data *AppConfigData `json:"data"`
	Msg  string         `json:"msg"`
}

func AppConfig() (*AppConfigData, error) {
	data := make(map[string]interface{})
	resp, err := httpPostResp("/app/config", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var resultMap *appConfigResp
	err = json.Unmarshal(body, &resultMap)
	if err != nil {
		return nil, err
	}
	if resultMap != nil && resultMap.Code == 0 && resultMap.Data != nil {
		return resultMap.Data, nil
	}
	if resultMap == nil {
		return nil, errors.New("app config failed")
	}
	return nil, errors.New(resultMap.Msg)
}
