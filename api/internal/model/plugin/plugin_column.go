package plugin

import "github.com/fs185085781/v9os/internal/model/base"

// @model name=插件字段
type PluginColumn struct {
	base.BaseModel
	// @field name=插件名称
	PluginName string `gorm:"index;column:plugin_name;size:50"`
	// @field name=插件表名
	PluginTable string `gorm:"index;column:plugin_table;size:50"`
	//主程序结构体的字段名称(用于映射到插件的结构体字段)
	// @field name=主字段名称
	MainFieldName string `gorm:"column:main_field_name;size:255"`
	//主程序表的字段名称(用于真实数据库查询或者写入的替换)
	// @field name=主列名称
	MainColumnName string `gorm:"column:main_column_name;size:255"`
	//插件结构体的字段名称(用于和主程序结构体字段形成映射)
	// @field name=字段名称
	FieldName string `gorm:"column:field_name;size:255"`
	//插件的字段名称(用于真实数据库查询或者写入的字段)
	// @field name=列名称
	ColumnName string `gorm:"column:column_name;size:255"`
	//插件结构体的GO类型(用于json序列化正确,float64,string,int,int64仅支持该4种类型)
	// @field name=字段类型
	// @select float64=浮点型 string=字符串 int=整型 int64=长整型
	FieldType string `gorm:"column:field_type;size:100"`
	// @field name=是否索引
	// @select 1=是 2=否
	IsIndex uint `gorm:"column:is_index" json:"IsIndex,string"`
	// @field name=是否文本
	// @select 1=是 2=否
	IsText uint `gorm:"column:is_text" json:"IsText,string"`
}

func (r *PluginColumn) TableName() string {
	return "plugin_column"
}

func init() {
	base.RegisterMigrate(&PluginColumn{})
}
