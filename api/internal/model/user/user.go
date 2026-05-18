package user

import (
	"github.com/fs185085781/v9os/internal/model/base"
)

// @model name=用户
type User struct {
	base.BaseModel
	// @field name=用户名
	Username string `gorm:"index;column:username;size:40"`
	// @field name=姓名
	Name string `gorm:"column:name;size:100"`
	// @field name=密码
	Password string `gorm:"column:password;size:100"`
	// @field name=邮箱
	Email string `gorm:"index;column:email;size:40"`
	// @field name=手机号
	Phone string `gorm:"index;column:phone;size:40"`
	// @field name=OTP
	Otp string `gorm:"column:otp;size:100"`
	// @field name=备注
	// @textarea
	Remark string `gorm:"column:remark;size:255"`
	// @field name=启用
	// @select 1=启用 2=禁用
	Enabled int `gorm:"column:enabled"`
	// @field name=头像
	Avatar string `gorm:"column:avatar;size:255"`
	// @field name=QQ OpenId
	QqOpenId string `gorm:"index;column:qq_open_id;size:40"`
	// @field name=微信 OpenId
	WxOpenId string `gorm:"index;column:wx_open_id;size:40"`
	IsAdmin  int    `gorm"-"`
}

func (u *User) TableName() string {
	return "user"
}
func init() {
	base.RegisterMigrate(&User{})
}
