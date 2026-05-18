package queue

import (
	"fmt"

	"github.com/fs185085781/v9os/internal/cache"
	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/internal/logger"
)

func NewQueue(cfg *config.QueuebaseConfig, c cache.Cache, tmpLog logger.Logger, callback QueueCallback) (Queue, error) {
	switch cfg.Driver {
	case "mem":
		return newMemQueue(cfg.Mem, callback, c, tmpLog)
	case "rocket":
		return newRocketQueue(cfg.Rocket, callback, c, tmpLog)
	case "redis":
		return newRedisQueue(cfg.Redis, callback, c, tmpLog)
	}
	return nil, fmt.Errorf("驱动%s暂未支持", cfg.Driver)
}
