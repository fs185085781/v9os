package plugin

import "github.com/fs185085781/v9os/internal/model/base"

// @model name=插件映射
type PluginTable struct {
	base.BaseModel
	// @field name=插件名称
	PluginName string `gorm:"column:plugin_name;size:100"`
	// @field name=插件表名
	PluginTable string `gorm:"column:plugin_table;size:100"`
	// @field name=物理表
	RealTable string `gorm:"column:real_table;size:100"`
	// @field name=表数量
	DataLength uint64 `gorm:"column:data_length" json:"DataLength,string"`
	//定时任务检查,如果某个虚拟表(插件名称+插件表名)的数据较大(超过10w),则当前虚拟表独占物理表(NeedExpand=2),其物理表其他的较小的虚拟表通过任务表迁移到新的物理表上
	// @field name=需要扩容
	// @select 1=正常 2=独占物理表 3=迁移中
	NeedExpand uint `gorm:"index;column:need_expand;default:1" json:"NeedExpand,string"`
}

func init() {
	base.RegisterMigrate(&PluginTable{})
}

func (r *PluginTable) TableName() string {
	return "plugin_table"
}
