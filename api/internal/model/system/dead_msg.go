package system

import (
	"github.com/fs185085781/v9os/internal/model/base"
)

// @model name=死信消息
type DeadMsg struct {
	base.BaseModel
	// @field name=插件名称
	Plugin string `gorm:"column:plugin;size:50"`
	// @field name=消息类型
	// @select 1=插件订阅 2=Websocket 3=相对地址订阅 4=绝对地址订阅
	Stype int `gorm:"column:stype" json:"Stype,string"`
	// @field name=回调地址
	Url string `gorm:"column:url;size:255"`
	// @field name=消息数据
	Data string `gorm:"column:data;type:text"`
	// @field name=消息ID
	MsgId string `gorm:"column:msg_id;size:40"`
}

func (s *DeadMsg) TableName() string {
	return "dead_msg"
}

func init() {
	base.RegisterMigrate(&DeadMsg{})
}
