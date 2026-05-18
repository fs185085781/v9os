package system

import (
	"github.com/fs185085781/v9os/internal/model/base"
)

// @model name=离线聊天消息
type OfflineChatMsg struct {
	base.BaseModel
	// @field name=来自用户ID
	From string `gorm:"index;column:from;size:40"`
	// @field name=目标用户ID
	To string `gorm:"index;column:to;size:40"`
	// @field name=消息内容
	Msg string `gorm:"column:msg;size:255"`
	// @field name=消息类型
	Type string `gorm:"column:type;size:20"`
	// @field name=消息时间
	// @datetime
	DateTime int64 `gorm:"index;column:date_time"`
}

func (s *OfflineChatMsg) TableName() string {
	return "offline_chat_msg"
}

func init() {
	base.RegisterMigrate(&OfflineChatMsg{})
}
