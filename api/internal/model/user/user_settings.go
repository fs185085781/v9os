package user

import "github.com/fs185085781/v9os/internal/model/base"

// @model name=用户设置
type UserSettings struct {
	base.BaseModel
	// @field name=颜色
	// @select green=绿色 blue=蓝色 orange=橙色 purple=紫色 red=红色 cyan=蓝绿色 pink=粉色 yellow=黄色 gray=灰色 deepBlue=深蓝色 deepPurple=深紫色 brown=褐色 diy=自定义
	Color string `gorm:"column:color;size:100"`
	// @field name=颜色描述
	ColorDesc string `gorm:"column:color_desc;size:255"`
	// @field name=语言
	// @select zh=中文 en=英文
	Lang string `gorm:"column:lang;size:100"`
	// @field name=主题
	// @select light=浅色 dark=暗黑
	Theme string `gorm:"column:theme;size:100"`
	// @field name=外观
	// @select macos=苹果系统 win10=微软系统 deepin=深度系统 pad=平板 backend=后台管理
	Mode string `gorm:"column:mode;size:100"`
	// @field name=字体
	Font string `gorm:"column:font;size:100"`
	// @field name=圆角
	// @select true=是 false=否
	Round string `gorm:"column:round;size:100"`
	// @field name=默认桌面壁纸
	DefaultWallpaper string `gorm:"column:default_wallpaper;size:255"`
	// @field name=壁纸类型
	// @select image=图片 video=视频
	DefaultWallpaperType string `gorm:"column:default_wallpaper_type;size:20"`
	// @field name=Dock应用
	DockApps string `gorm:"column:dock_apps;type:text"`
	// @field name=透明度
	Transparent int `gorm:"column:transparent"`
	// @field name=用户ID
	UserID string `gorm:"column:user_id;size:40"`
}

func (u *UserSettings) TableName() string {
	return "user_settings"
}

func init() {
	base.RegisterMigrate(&UserSettings{})
}
