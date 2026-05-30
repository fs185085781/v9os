package uioc

import (
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
func RestartFunc() func(bool) {
	v := ioc.Ioc().Get(ioc.KeyRestartFunc)
	if v == nil {
		return nil
	}
	return v.(func(bool))
}
