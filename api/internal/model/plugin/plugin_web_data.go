package plugin

import "github.com/fs185085781/v9os/internal/model/base"

// @model name=插件扩展数据
type PluginWebData struct {
	base.BaseModel
	// @field name=用户ID
	UserID uint `gorm:"column:user_id;index:user_code_key_index"`
	// @field name=插件编码
	Code string `gorm:"column:code;size:20;index:user_code_key_index"`
	// @field name=数据键
	DataKey string `gorm:"column:data_key;size:128;index:user_code_key_index"`
	// @field name=数据值
	// @textarea
	DataValue string `gorm:"column:data_value;type:text"`
}

func (l *PluginWebData) TableName() string {
	return "plugin_web_data"
}

func init() {
	base.RegisterMigrate(&PluginWebData{})
}
