package uioc

import (
	"os/exec"
	"sync"

	"github.com/fs185085781/v9os/internal/cache"
	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/internal/database"
	"github.com/fs185085781/v9os/internal/ioc"
	"github.com/fs185085781/v9os/internal/logger"
	"github.com/fs185085781/v9os/internal/queue"
)

func Has(key string) bool {
	return ioc.Ioc().Get(key) != nil
}

func Get[T any](key string) T {
	return ioc.Ioc().Get(key).(T)
}

func Config() config.Config {
	return ioc.Ioc().Get(ioc.KeyConfig).(config.Config)
}
func Log() logger.Logger {
	return ioc.Ioc().Get(ioc.KeyLog).(logger.Logger)
}
func Database() database.Database {
	return ioc.Ioc().Get(ioc.KeyDatabase).(database.Database)
}
func Cache() cache.Cache {
	return ioc.Ioc().Get(ioc.KeyCache).(cache.Cache)
}
func Queue() queue.Queue {
	return ioc.Ioc().Get(ioc.KeyQueue).(queue.Queue)
}
func ControllerMap() *sync.Map {
	return ioc.Ioc().Get(ioc.KeyControllerMap).(*sync.Map)
}
func AfterFuncs() []interface{} {
	v := ioc.Ioc().Get(ioc.KeyAfterFunc)
	if v == nil {
		return nil
	}
	return v.([]interface{})
}
func RestartFunc() func(bool) {
	v := ioc.Ioc().Get(ioc.KeyRestartFunc)
	if v == nil {
		return nil
	}
	return v.(func(bool))
}
func HideCmdFunc() func(*exec.Cmd) {
	v := ioc.Ioc().Get(ioc.KeyHideCmdFunc)
	if v == nil {
		return nil
	}
	return v.(func(*exec.Cmd))
}
func SystemCloseFunc() func() {
	v := ioc.Ioc().Get(ioc.KeySystemCloseFunc)
	if v == nil {
		return nil
	}
	return v.(func())
}
func LBMap() map[string]string {
	v := ioc.Ioc().Get(ioc.KeyLBMap)
	if v == nil {
		return nil
	}
	return v.(map[string]string)
}
func PluginMap() map[string]map[string]bool {
	v := ioc.Ioc().Get(ioc.KeyPluginMap)
	if v == nil {
		return nil
	}
	return v.(map[string]map[string]bool)
}
func MachineAllowedPluginMap() map[string]map[string]bool {
	v := ioc.Ioc().Get(ioc.KeyMachineAllowedPluginMap)
	if v == nil {
		return nil
	}
	return v.(map[string]map[string]bool)
}
func PluginAllowedMachineMap() map[string]map[string]bool {
	v := ioc.Ioc().Get(ioc.KeyPluginAllowedMachineMap)
	if v == nil {
		return nil
	}
	return v.(map[string]map[string]bool)
}
func WebsocketMap() map[string]map[string]bool {
	v := ioc.Ioc().Get(ioc.KeyWebsocketMap)
	if v == nil {
		return nil
	}
	return v.(map[string]map[string]bool)
}
