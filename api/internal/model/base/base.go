package base

import (
	"sync"

	"github.com/fs185085781/v9os/internal/ioc"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint           `gorm:"primarykey"`
	CreatedAt uint64         `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt uint64         `gorm:"column:updated_at;autoUpdateTime:milli"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func RegisterMigrate(obj interface{}) {
	registerMap := ioc.Ioc().GetOrRegister(ioc.KeyModelMap, &sync.Map{}).(*sync.Map)
	registerMap.Store(obj, struct{}{})
}

type PluginDataTable struct {
	PlugDataStruct interface{}
	PluginTable    string
}

func RegisterPluginData(obj PluginDataTable) {
	registerMap := ioc.Ioc().GetOrRegister(ioc.KeyPluginDataMap, &sync.Map{}).(*sync.Map)
	registerMap.Store(obj, struct{}{})
}
