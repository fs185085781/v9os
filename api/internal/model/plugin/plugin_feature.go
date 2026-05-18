package plugin

import "github.com/fs185085781/v9os/internal/model/base"

// PluginFeature 插件功能二级表,用于精确查询插件的拦截器
// @model name=插件功能
type PluginFeature struct {
	base.BaseModel
	// @field name=插件编码
	PluginCode string `gorm:"column:plugin_code;size:20;index"`
	// @field name=状态
	// @select 0=禁用 1=启用
	Enabled int `gorm:"column:enabled"` // 0=禁用 1=启用
	// @field name=功能内容
	Content string `gorm:"column:content;size:100;index"` // user-login 等
}

func (a *PluginFeature) TableName() string {
	return "plugin_feature"
}

func init() {
	base.RegisterMigrate(&PluginFeature{})
}
