package user

import "github.com/fs185085781/v9os/internal/model/base"

// @model name=桌面快捷方式
type DesktopApp struct {
	base.BaseModel
	// @field name=用户ID
	UserID uint `gorm:"index;column:user_id"`
	// @field name=图标
	Icon string `gorm:"column:icon;size:255"`
	// @field name=标题
	Title string `gorm:"column:title;size:100"`
	// @field name=应用类型
	// @select system=系统应用 plugin=插件应用 iframe=远程地址
	AppType string `gorm:"column:app_type;size:20"` // system | plugin | iframe
	// @field name=代码
	Code string `gorm:"column:code;size:80"`
	// @field name=地址
	Url string `gorm:"column:url;size:500"`
	// @field name=排序
	Sort int `gorm:"column:sort"`
}

func (d *DesktopApp) TableName() string {
	return "desktop_app"
}

func init() {
	base.RegisterMigrate(&DesktopApp{})
}
