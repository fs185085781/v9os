package store

import (
	"fmt"

	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/internal/logger"
)

func NewStore(cnf *config.ServerConfig, tmpLog logger.Logger) (Store, error) {
	if cnf.StoreType == "v9os" {
		return newV9osStore(cnf.StoreHost, tmpLog)
	}
	return nil, fmt.Errorf("暂不支持%s服务数据源", cnf.StoreType)
}
