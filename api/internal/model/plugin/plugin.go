package plugin

import "github.com/fs185085781/v9os/internal/model/base"

// @model name=插件
type Plugin struct {
	base.BaseModel
	// @field name=安装机器
	FirstMachine string `gorm:"column:first_machine;size:255"`
	// @field name=运行错误
	RuntimeError string `gorm:"column:runtime_error;type:text"`
	// @field name=插件名称
	Name string `gorm:"column:name;size:100"`
	// @field name=插件描述
	// @textarea
	Description string `gorm:"column:description;type:text"`
	// @field name=关闭延迟
	CloseDelay int `gorm:"column:close_delay" json:"CloseDelay,string"` // 0 常驻，其他为多少分钟后关闭
	// @field name=插件编码
	Code string `gorm:"column:code;size:40"` // 插件编码
	// @field name=状态
	// @select 0=禁用 1=启用
	Status int `gorm:"column:status" json:"Status,string"` // 0 禁用 1 启用
	// @field name=备注
	Remark string `gorm:"column:remark;size:255"` // 备注
	// @field name=版本
	Version string `gorm:"column:version;size:20"` // 版本
	// 三方插件在分布式下只安装到一台机器上
	// @field name=插件类型
	// @select 1=主程序插件 2=前端插件 3=三方插件 4=云应用
	PluginType int `gorm:"column:plugin_type" json:"PluginType,string"` // 1:主程序插件 2:前端插件 3:三方插件 4:远程iframe
	// 安装后前端会引用该 js，用于控制主程序前端功能
	// @field name=前端js路径
	WebHook string `gorm:"column:web_hook;size:255"` // 前端 js 路径
	// @field name=限制最低主程序版本
	LimitVersion string `gorm:"column:limit_version;size:255"` // 限制最低主程序版本
	// @field name=图标url
	IconUrl string `gorm:"column:icon_url;size:255"` // 图标 url
	// @field name=访问地址
	AccessUrl string `gorm:"column:access_url;size:255"` // 三方插件访问地址
	// @field name=调试端口
	DebugPort int `gorm:"column:debug_port" json:"DebugPort,string"` // 大于 0 时为调试模式
	// @field name=打开格式
	OpenExts string `gorm:"column:open_exts;type:text"`
	// @field name=编辑格式
	EditExts string `gorm:"column:edit_exts;type:text"`
	// @field name=扩展格式
	ExpandExts string `gorm:"column:expand_exts;size:100"`
}

func (a *Plugin) TableName() string {
	return "plugin"
}

func init() {
	base.RegisterMigrate(&Plugin{})
}
