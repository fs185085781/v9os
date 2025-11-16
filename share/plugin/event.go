package plugin

import (
	"errors"
	"time"

	"github.com/spf13/cast"
)

// 实现事件发布与订阅
// 发送event事件数据data
func PushEvent(event string, data interface{}) error {
	param := make(map[string]interface{})
	param["event"] = event
	param["data"] = data
	resultMap, err := httpPost("/event/push", param)
	if err != nil {
		return err
	}
	if resultMap != nil && cast.ToInt(resultMap["code"]) == 0 {
		return nil
	}
	if resultMap == nil {
		return errors.New("push event failed")
	}
	return errors.New(cast.ToString(resultMap["msg"]))
}

// 订阅event事件,以method方法承接,以action方法处理事件数据
func SubscribeEvent(event, method string) {
	param := make(map[string]interface{})
	param["event"] = event
	param["method"] = method
	go func() {
		for {
			if runKey == "" || serverPort == "" {
				time.Sleep(time.Second)
				continue
			}
			break
		}
		httpPost("/event/subscribe", param)
	}()
}

// 取消订阅event事件,取消method的回调
func UnsubscribeEvent(event, method string) error {
	param := make(map[string]interface{})
	param["event"] = event
	param["method"] = method
	resultMap, err := httpPost("/event/unsubscribe", param)
	if err != nil {
		return err
	}
	if resultMap != nil && cast.ToInt(resultMap["code"]) == 0 {
		return nil
	}
	if resultMap == nil {
		return errors.New("unsubscribe event failed")
	}
	return errors.New(cast.ToString(resultMap["msg"]))
}
