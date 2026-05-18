package plugin

import (
	"strconv"

	"github.com/fs185085781/v9os/internal/model/base"
)

// @model name=插件数据
type PluginData struct {
	base.BaseModel
	// @field name=插件名称
	PluginName string `gorm:"index;column:plugin_name;size:50"`
	// @field name=插件表名
	PluginTable string `gorm:"index;column:plugin_table;size:50"`
	// @field name=插件数据ID
	DataId string `gorm:"index;column:data_id;size:40"`
	// @field name=用户ID
	UserId string `gorm:"index;column:user_id;size:40"`
	// @field name=部门ID
	DeptId string `gorm:"index;column:dept_id;size:40"`
	// @field name=字段1
	Field1 string `gorm:"column:field1;size:255"`
	// @field name=字段2
	Field2 string `gorm:"column:field2;size:255"`
	// @field name=字段3
	Field3 string `gorm:"column:field3;size:255"`
	// @field name=字段4
	Field4 string `gorm:"column:field4;size:255"`
	// @field name=字段5
	Field5 string `gorm:"column:field5;size:255"`
	// @field name=字段6
	Field6 string `gorm:"column:field6;size:255"`
	// @field name=字段7
	Field7 string `gorm:"column:field7;size:255"`
	// @field name=字段8
	Field8 string `gorm:"column:field8;size:255"`
	// @field name=字段9
	Field9 string `gorm:"column:field9;size:255"`
	// @field name=字段10
	Field10 string `gorm:"column:field10;size:255"`
	// @field name=字段11
	Field11 string `gorm:"column:field11;size:255"`
	// @field name=字段12
	Field12 string `gorm:"column:field12;size:255"`
	// @field name=字段13
	Field13 string `gorm:"column:field13;size:255"`
	// @field name=字段14
	Field14 string `gorm:"column:field14;size:255"`
	// @field name=字段15
	Field15 string `gorm:"column:field15;size:255"`
	// @field name=字段16
	Field16 string `gorm:"column:field16;size:255"`
	// @field name=字段17
	Field17 string `gorm:"column:field17;size:255"`
	// @field name=字段18
	Field18 string `gorm:"column:field18;size:255"`
	// @field name=字段19
	Field19 string `gorm:"column:field19;size:255"`
	// @field name=字段20
	Field20 string `gorm:"column:field20;size:255"`
	// @field name=文本字段1
	TextField1 string `gorm:"column:text_field1;type:text"`
	// @field name=文本字段2
	TextField2 string `gorm:"column:text_field2;type:text"`
	// @field name=文本字段3
	TextField3 string `gorm:"column:text_field3;type:text"`
	// @field name=文本字段4
	TextField4 string `gorm:"column:text_field4;type:text"`
	// @field name=文本字段5
	TextField5 string `gorm:"column:text_field5;type:text"`
	// @field name=索引字段1
	IndexField1 string `gorm:"index;column:index_field1;size:190"`
	// @field name=索引字段2
	IndexField2 string `gorm:"index;column:index_field2;size:190"`
	// @field name=索引字段3
	IndexField3 string `gorm:"index;column:index_field3;size:190"`
	// @field name=索引字段4
	IndexField4 string `gorm:"index;column:index_field4;size:190"`
	// @field name=索引字段5
	IndexField5 string `gorm:"index;column:index_field5;size:190"`
}

func init() {
	base.RegisterPluginData(base.PluginDataTable{
		PlugDataStruct: &PluginData{},
		PluginTable:    "plugin_data_" + strconv.Itoa(0),
	})
}
