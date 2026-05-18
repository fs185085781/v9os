package plugin

import "github.com/fs185085781/v9os/internal/model/base"

// @model name=插件数据迁移任务
type PluginTableExpandTask struct {
	base.BaseModel
	// @field name=插件映射ID
	PluginTableID uint `gorm:"index;column:plugin_table_id"`
	// @field name=插件名称
	PluginName string `gorm:"index;column:plugin_name;size:100"`
	// @field name=插件表名
	PluginTable string `gorm:"index;column:plugin_table;size:100"`
	// @field name=原物理表
	OldRealTable string `gorm:"index;column:old_real_table;size:100"`
	// @field name=新物理表
	NewRealTable string `gorm:"index;column:new_real_table;size:100"`
	// @field name=数据量
	DataLength uint64 `gorm:"column:data_length" json:"DataLength,string"`
	// @field name=已迁移行数
	MigratedRows uint64 `gorm:"column:migrated_rows" json:"MigratedRows,string"`
	// @field name=重试次数
	RetryCount uint `gorm:"column:retry_count" json:"RetryCount,string"`
	// @field name=状态
	// @select 1=需要迁移 2=迁移中 3=落定
	Status uint `gorm:"index;column:status;default:1" json:"Status,string"`
	// @field name=错误信息
	ErrorMsg string `gorm:"column:error_msg;size:500"`
}

func init() {
	base.RegisterMigrate(&PluginTableExpandTask{})
}

func (r *PluginTableExpandTask) TableName() string {
	return "plugin_table_expand_task"
}
