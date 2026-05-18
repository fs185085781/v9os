package api

import (
	"errors"
	"reflect"
	"time"

	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/internal/controller"
	"github.com/fs185085781/v9os/internal/ioc/uioc"
	"github.com/gin-gonic/gin"
)

type SystemController struct {
	*controller.BaseController
}

// @model name=网站设置
type webSettings struct {
	// @field name=网站标题
	Title string
	// @field name=网站副标题
	Subtitle string
	// @field name=网站logo
	Logo string
	// @field name=默认密码
	DefaultPwd string
	// @field name=默认颜色
	// @select green=绿色 blue=蓝色 orange=橙色 purple=紫色 red=红色 cyan=蓝绿色 pink=粉色 yellow=黄色 gray=灰色 deepBlue=深蓝色 deepPurple=深紫色 brown=褐色
	DefaultColor string
	// @field name=默认语言
	// @select zh=中文 en=英文
	DefaultLang string
	// @field name=默认主题
	// @select light=浅色 dark=暗黑
	DefaultTheme string
	// @field name=默认外观
	// @select macos=Macos win10=Win10 deepin=Deepin backend=Backend
	DefaultMode string
	// @field name=默认字体
	// @select default=默认
	DefaultFont string
	// @field name=默认圆角
	// @select true=是 false=否
	DefaultRound string
	// @field name=默认桌面壁纸
	DefaultWallpaper string
	// @field name=壁纸类型
	// @select image=图片 video=视频
	DefaultWallpaperType string
	// @field name=哀悼模式
	// @select true=是 false=否
	Mourning string
	// @field name=备案名称
	BeianName string
	// @field name=安全入口
	SafeEntry string
	// @field name=版本信息
	Version string
}

func init() {
	c := &SystemController{
		BaseController: controller.GetBaseController(),
	}
	//获取配置(高风险)
	c.RegisterAdminApi("POST", "/system/configGet", c.GetConfig)
	//保存配置(高风险)
	c.RegisterAdminApi("POST", "/system/configSave", c.SaveConfig)
	//获取站点设置
	c.RegisterPublic("api", "POST", "/system/settingsGet", c.SettingsGet)
	//保存站点设置
	c.RegisterAdminApi("POST", "/system/settingsSave", c.SettingsSave)
}

// GetConfig 获取系统配置（仅管理员）
func (c *SystemController) GetConfig(ctx *gin.Context) {
	c.OkData(ctx, c.Config().ConfigAll())
}

// SaveConfig 保存系统配置（仅管理员）
func (c *SystemController) SaveConfig(ctx *gin.Context) {
	var newCfg config.ConfigAll
	if err := ctx.ShouldBindBodyWithJSON(&newCfg); err != nil {
		c.FailMsg(ctx, c.GetText(ctx, "common.param.invalid"))
		return
	}
	cfg := c.Config()
	cur := cfg.ConfigAll()
	// 覆盖只读字段，防止通过API篡改
	if newCfg.Server != nil && cur.Server != nil {
		newCfg.Server.SystemId = cur.Server.SystemId
		newCfg.Server.PasswordKey = cur.Server.PasswordKey
		newCfg.Server.CommunicationKey = cur.Server.CommunicationKey
	}
	if newCfg.Auth != nil && cur.Auth != nil {
		newCfg.Auth.Secret = cur.Auth.Secret
		newCfg.Auth.SecretTime = cur.Auth.SecretTime
		newCfg.Auth.LastSecret = cur.Auth.LastSecret
	}
	if newCfg.Distributed == nil && cur.Distributed != nil {
		newCfg.Distributed = cur.Distributed
	}
	if err := c.validateDistributedConfig(ctx, cur, &newCfg); err != nil {
		c.FailMsg(ctx, err.Error())
		return
	}
	changed := config.ConfigChanged(cur, &newCfg)
	if changed {
		*cfg.ConfigAll() = newCfg
		if err := cfg.Save(); err != nil {
			c.FailMsg(ctx, err.Error())
			return
		}
	}
	c.OkData(ctx, gin.H{"restart": changed})
	if changed {
		go func() {
			time.Sleep(1 * time.Second)
			if fn := uioc.RestartFunc(); fn != nil {
				fn(true)
			}
		}()
	}
}

func (c *SystemController) validateDistributedConfig(ctx *gin.Context, cur, next *config.ConfigAll) error {
	if next == nil || next.Distributed == nil || !next.Distributed.Enabled {
		return nil
	}
	if !c.Distributed().SupportDistributed() {
		return errors.New(c.GetText(ctx, "common.distributed.editionunsupported"))
	}
	if cur == nil {
		return errors.New(c.GetText(ctx, "common.config.empty"))
	}
	if !reflect.DeepEqual(cur.Database, next.Database) || !reflect.DeepEqual(cur.Cachebase, next.Cachebase) || !reflect.DeepEqual(cur.Queuebase, next.Queuebase) {
		return errors.New(c.GetText(ctx, "common.distributed.savebaserequired"))
	}
	if !c.Database().SupportDistributed() {
		return errors.New(c.GetText(ctx, "common.distributed.databaseunsupported"))
	}
	if !c.Cache().SupportDistributed() {
		return errors.New(c.GetText(ctx, "common.distributed.cacheunsupported"))
	}
	if !c.Queue().SupportDistributed() {
		return errors.New(c.GetText(ctx, "common.distributed.queueunsupported"))
	}
	return nil
}

func WebSettingsGet() *webSettings {
	c := controller.GetBaseController()
	var wset webSettings
	flag, err := c.Database().GetObject("web-settings", &wset)
	if err != nil {
		return nil
	}
	if !flag {
		wset = webSettings{
			Title:                "V9OS",
			Subtitle:             "Powered by QQ185085781",
			Logo:                 "/assets/images/logo.png",
			DefaultPwd:           "123456",
			DefaultColor:         "green",
			DefaultLang:          "zh",
			DefaultTheme:         "light",
			DefaultMode:          "macos",
			DefaultFont:          "default",
			DefaultRound:         "true",
			DefaultWallpaper:     "default",
			DefaultWallpaperType: "image",
			BeianName:            "粤ICP备0000000000号",
			Mourning:             "false",
			SafeEntry:            "",
		}
		c.Database().SetObject("web-settings", &wset)
	}
	wset.Version = c.Config().Machine().Version
	return &wset
}
func (c *SystemController) SettingsGet(ctx *gin.Context) {
	wset := WebSettingsGet()
	if wset == nil {
		c.FailMsg(ctx, c.GetText(ctx, "common.websettings.getfailed"))
		return
	}
	c.OkData(ctx, wset)
}

func (c *SystemController) SettingsSave(ctx *gin.Context) {
	var wset webSettings
	if err := ctx.ShouldBindBodyWithJSON(&wset); err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	err := c.Database().SetObject("web-settings", &wset)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.Ok(ctx)
}
