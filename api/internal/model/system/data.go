package system

import "github.com/fs185085781/v9os/internal/model/base"

// @model name=数据
type Data struct {
	base.BaseModel
	// @field name=数据键
	DataKey string `gorm:"column:data_key;index;size:128"`
	// @field name=数据值
	// @textarea
	DataValue string `gorm:"column:data_value;type:text"`
}

func (l *Data) TableName() string {
	return "data"
}

func init() {
	base.RegisterMigrate(&Data{})
}
