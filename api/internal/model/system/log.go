package system

import (
	"encoding/json"

	"github.com/fs185085781/v9os/internal/logger"
	"github.com/fs185085781/v9os/internal/model/base"
	"github.com/fs185085781/v9os/pkg/util"
	"gorm.io/gorm"
)

// @model name=日志
type Log struct {
	base.BaseModel
	// @field name=日志等级
	// @select debug=调试 info=信息 warn=告警 error=错误
	Level logger.Level `gorm:"column:level;size:10"`
	// @field name=日志消息
	Msg string `gorm:"column:msg;size:512"`
	// @field name=日志时间
	// @datetime
	Time uint64 `gorm:"column:time"`
	// @field name=日志详情
	// @textarea
	Text string `gorm:"column:text;type:text"`
}

func (l *Log) TableName() string {
	return "log"
}

func init() {
	base.RegisterMigrate(&Log{})
}

func WriteLog(db *gorm.DB, lvl logger.Level, msg string, fields ...logger.Field) {
	if len(msg) >= 500 {
		msg = msg[:500]
	}
	log := Log{
		Level: lvl,
		Msg:   msg,
		Time:  uint64(util.UnixMilliseconds()),
	}
	data := make(map[string]interface{})
	for _, field := range fields {
		data[field.Key] = field.Value
	}
	jsonData, _ := json.Marshal(data)
	log.Text = string(jsonData)
	db.Create(&log)
}
