package api

import (
	"fmt"
	"strings"
	"time"

	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/internal/controller"
	infaceUser "github.com/fs185085781/v9os/internal/inface/user"
	"github.com/fs185085781/v9os/internal/ioc"
	"github.com/fs185085781/v9os/internal/ioc/uioc"
	"github.com/fs185085781/v9os/internal/logger"
	"github.com/fs185085781/v9os/internal/model/plugin"
	"github.com/fs185085781/v9os/internal/model/user"
	"github.com/fs185085781/v9os/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/cast"
)

type UserController struct {
	*controller.BaseController
}

func init() {
	c := &UserController{
		BaseController: controller.GetBaseController(),
	}
	//登陆
	c.RegisterPublic("api", "POST", "/user/login", c.Login)
	//用户个性化获取
	c.RegisterPublic("api", "POST", "/user/settings", c.Settings)
	//用户token刷新
	c.RegisterPublic("api", "POST", "/user/token", c.Token)
	//用户权限列表
	c.RegisterApi("POST", "/user/auths", c.Auths)
	//用户信息
	c.RegisterApi("POST", "/user/info", c.Info)
	//保存用户个性化
	c.RegisterApi("POST", "/user/saveSettings", c.SaveSettings)
	//更新用户资料
	c.RegisterApi("POST", "/user/updateProfile", c.UpdateProfile)
	//修改用户密码
	c.RegisterApi("POST", "/user/changePasswordByToken", c.ChangePasswordByToken)
	//验证用户密码
	c.RegisterApi("POST", "/user/verifyPassword", c.VerifyPassword)
	//获取用户授权模块
	c.RegisterApi("POST", "/user/auth-modules", c.AuthModules)
	//获取用户桌面应用
	c.RegisterApi("POST", "/user/desktop_apps", c.DesktopApps)
	//保存用户桌面应用
	c.RegisterApi("POST", "/user/saveDesktopApp", c.SaveDesktopApp)
	//删除用户桌面应用
	c.RegisterApi("POST", "/user/deleteDesktopApp", c.DeleteDesktopApp)
	//获取系统代理(高风险)
	c.RegisterAdminApi("POST", "/user/proxyToken", c.ProxyToken)
}

type moduleInfo struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	Icon      string `json:"icon"`
	Type      int    `json:"type"`
	AccessUrl string `json:"accessUrl"`
}

func (c *UserController) AuthModules(ctx *gin.Context) {
	uid, ok := ctx.Get("userID")
	userID := ""
	var ps []plugin.Plugin
	if ok {
		userID = uid.(string)
		userProvider := uioc.Get[infaceUser.UserProvider](ioc.KeyUserProvider)
		ps = userProvider.UserAuthModules(cast.ToUint(userID))
	}
	modMap := make(map[string]moduleInfo)
	for _, item := range ps {
		if _, exists := modMap[item.Code]; !exists {
			modMap[item.Code] = moduleInfo{
				Code:      item.Code,
				Name:      item.Name,
				Type:      item.PluginType,
				Icon:      item.IconUrl,
				AccessUrl: item.AccessUrl,
			}
		}
	}
	modules := make([]moduleInfo, 0)
	for _, mod := range modMap {
		modules = append(modules, mod)
	}
	c.OkData(ctx, modules)
}

func (c *UserController) DesktopApps(ctx *gin.Context) {
	userID := cast.ToUint(ctx.GetString("userID"))
	var data []user.DesktopApp
	if err := c.Database().Read().Where("user_id = ?", userID).Order("sort asc,id asc").Find(&data).Error; err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.OkData(ctx, data)
}

func (c *UserController) SaveDesktopApp(ctx *gin.Context) {
	var data user.DesktopApp
	if err := ctx.ShouldBindBodyWithJSON(&data); err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	userID := cast.ToUint(ctx.GetString("userID"))
	if userID == 0 {
		c.FailMsg(ctx, "用户不能为空")
		return
	}
	data.UserID = userID
	data.AppType = strings.TrimSpace(data.AppType)
	if data.Title == "" || data.AppType == "" {
		c.FailMsg(ctx, "标题和类型不能为空")
		return
	}
	if data.AppType != "system" && data.AppType != "plugin" && data.AppType != "iframe" {
		c.FailMsg(ctx, "类型不支持")
		return
	}
	db := c.Database().Write()
	if data.ID > 0 {
		if err := db.Model(&user.DesktopApp{}).Where("id = ? AND user_id = ?", data.ID, userID).Updates(map[string]interface{}{
			"icon":     data.Icon,
			"title":    data.Title,
			"app_type": data.AppType,
			"code":     data.Code,
			"url":      data.Url,
			"sort":     data.Sort,
		}).Error; err != nil {
			c.ErrMsg(ctx, err)
			return
		}
		c.Ok(ctx)
		return
	}
	if err := db.Create(&data).Error; err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.OkData(ctx, data)
}

func (c *UserController) DeleteDesktopApp(ctx *gin.Context) {
	var param struct {
		ID uint `json:"ID"`
	}
	if err := ctx.ShouldBindBodyWithJSON(&param); err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	if param.ID == 0 {
		c.FailMsg(ctx, "ID不能为空")
		return
	}
	if err := c.Database().Write().Where("id = ? AND user_id = ?", param.ID, cast.ToUint(ctx.GetString("userID"))).Delete(&user.DesktopApp{}).Error; err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.Ok(ctx)
}

func (c *UserController) Login(ctx *gin.Context) {
	param := c.Param(ctx)
	username := param.ParamString("username")
	password := param.ParamString("password")
	if username == "" || password == "" {
		c.FailMsg(ctx, c.GetText(ctx, "common.user.usernamepwderr"))
		return
	}
	var u user.User
	c.Database().First(&u, "username = ?", username)
	if u.ID <= 0 {
		c.FailMsg(ctx, c.GetText(ctx, "common.user.usernamepwderr"))
		return
	}
	if !util.CheckPassword(password, u.Password) {
		c.FailMsg(ctx, c.GetText(ctx, "common.user.usernamepwderr"))
		return
	}
	if u.Enabled != 1 {
		c.FailMsg(ctx, c.GetText(ctx, "common.user.usernotenabled"))
		return
	}
	token, err := c.Auth().GenerateToken(cast.ToString(u.ID), "")
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.OkData(ctx, token)
}

func (c *UserController) Info(ctx *gin.Context) {
	u := c.UserInfo(ctx)
	if u == nil {
		return
	}
	// 敏感字段处理
	u.Password = ""
	u.Otp = maskOtp(u.Otp)
	u.Phone = maskPhone(u.Phone)
	u.Email = maskEmail(u.Email)
	u.WxOpenId = maskOther(u.WxOpenId)
	u.QqOpenId = maskOther(u.QqOpenId)
	c.OkData(ctx, u)
}

func (c *UserController) Auths(ctx *gin.Context) {
	userID, ok := ctx.Get("userID")
	if !ok {
		c.FailMsg(ctx, c.GetText(ctx, "common.user.usernoauths"))
		return
	}
	userProvider := uioc.Get[infaceUser.UserProvider](ioc.KeyUserProvider)
	auths := userProvider.UserAuth(cast.ToUint(userID))
	if auths == nil {
		auths = []string{}
	}
	c.OkData(ctx, auths)
}

func (c *UserController) Token(ctx *gin.Context) {
	var req struct {
		Token string `json:"token"`
	}
	err := ctx.ShouldBindBodyWithJSON(&req)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	if req.Token == "" {
		c.FailMsg(ctx, c.GetText(ctx, "common.user.usernoauths"))
		return
	}
	token, err := c.Auth().RefreshToken(ctx, req.Token)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.OkData(ctx, token)
}

func (c *UserController) Settings(ctx *gin.Context) {
	uid, ok := ctx.Get("userID")
	userID := ""
	if ok {
		userID = uid.(string)
		var settings user.UserSettings
		err := c.Database().First(&settings, "user_id = ?", userID)
		if err == nil {
			c.OkData(ctx, settings)
			return
		}
	}
	wset := WebSettingsGet()
	if wset == nil {
		c.FailMsg(ctx, c.GetText(ctx, "common.websettings.getfailed"))
		return
	}
	settings := user.UserSettings{
		UserID:               userID,
		Color:                wset.DefaultColor,
		Lang:                 wset.DefaultLang,
		Theme:                wset.DefaultTheme,
		Mode:                 wset.DefaultMode,
		Font:                 wset.DefaultFont,
		Round:                wset.DefaultRound,
		DefaultWallpaper:     wset.DefaultWallpaper,
		DefaultWallpaperType: wset.DefaultWallpaperType,
		Transparent:          0,
	}
	if settings.UserID != "" {
		err := c.Database().Create(&settings)
		c.Log().Error("user save settings failed", logger.NewField("userID", userID), logger.NewField("err", err))
	}
	c.OkData(ctx, settings)
}
func (c *UserController) ProxyToken(ctx *gin.Context) {
	param := c.Param(ctx)
	proxyHost := param.ParamString("host")
	mapData := make(map[string]string)
	mapData["proxy_token"] = c.Config().Server().ProxyToken
	mapData["proxy_host"] = c.Config().Server().ProxyHost
	if mapData["proxy_host"] == "" && proxyHost != "" {
		mapData["proxy_host"] = proxyHost
		cfg := c.Config()
		cur := cfg.ConfigAll()
		newCfg := *cur
		newCfg.Server.ProxyHost = proxyHost
		changed := config.ConfigChanged(cur, &newCfg)
		if changed {
			*cfg.ConfigAll() = newCfg
			cfg.Save()
		}
	}
	c.OkData(ctx, mapData)
}

func (c *UserController) SaveSettings(ctx *gin.Context) {
	userID, ok := ctx.Get("userID")
	if !ok {
		c.FailMsg(ctx, c.GetText(ctx, "common.user.usernoauths"))
		return
	}
	var dbSettings user.UserSettings
	err := c.Database().First(&dbSettings, "user_id = ?", userID)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	var settings user.UserSettings
	err = ctx.ShouldBindBodyWithJSON(&settings)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	dbSettings.Color = settings.Color
	dbSettings.ColorDesc = settings.ColorDesc
	dbSettings.Lang = settings.Lang
	dbSettings.Theme = settings.Theme
	dbSettings.Mode = settings.Mode
	dbSettings.Font = settings.Font
	dbSettings.Round = settings.Round
	dbSettings.DefaultWallpaper = settings.DefaultWallpaper
	dbSettings.DefaultWallpaperType = settings.DefaultWallpaperType
	dbSettings.DockApps = settings.DockApps
	dbSettings.Transparent = settings.Transparent
	err = c.Database().Update(&dbSettings)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.Ok(ctx)
}

func (c *UserController) UpdateProfile(ctx *gin.Context) {
	u := c.UserInfo(ctx)
	if u == nil {
		return
	}
	param := c.Param(ctx)
	if v := param.ParamString("name"); v != "" {
		u.Name = v
	}
	if v := param.ParamString("avatar"); v != "" {
		u.Avatar = v
	}
	err := c.Database().Update(u)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.Ok(ctx)
}

func (c *UserController) ChangePasswordByToken(ctx *gin.Context) {
	u := c.UserInfo(ctx)
	if u == nil {
		return
	}
	param := c.Param(ctx)
	verifyToken := param.ParamString("verifyToken")
	newPassword := param.ParamString("newPassword")
	if verifyToken == "" || newPassword == "" {
		c.FailMsg(ctx, c.GetText(ctx, "common.user.parammissing"))
		return
	}
	level := "1"
	if u.Otp != "" || u.Phone != "" || u.Email != "" {
		level = "2"
	}
	if err := c.checkVerifyToken(verifyToken, u.ID, "changePassword", level); err != nil {
		c.FailMsg(ctx, err.Error())
		return
	}
	u.Password = util.EncodePassword(newPassword)
	err := c.Database().Update(u)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.consumeVerifyToken(verifyToken, level)
	c.Ok(ctx)
}

func (c *UserController) VerifyPassword(ctx *gin.Context) {
	u := c.UserInfo(ctx)
	if u == nil {
		return
	}
	param := c.Param(ctx)
	password := param.ParamString("password")
	action := param.ParamString("action")
	if password == "" || action == "" {
		c.FailMsg(ctx, c.GetText(ctx, "common.user.parammissing"))
		return
	}
	if !util.CheckPassword(password, u.Password) {
		c.FailMsg(ctx, c.GetText(ctx, "common.user.pwderror"))
		return
	}
	token := c.generateVerifyToken(u.ID, action, "1")
	c.OkData(ctx, map[string]string{"verifyToken": token})
}

func (c *UserController) generateVerifyToken(userID uint, action, level string) string {
	token := uuid.NewString()
	cacheKey := fmt.Sprintf("verify:token_%s:%s", level, token)
	payload := fmt.Sprintf("%d:%s", userID, action)
	_ = c.Cache().SetValue(cacheKey, []byte(payload), 5*time.Minute)
	return token
}

func (c *UserController) checkVerifyToken(token string, expectedUserID uint, action, level string) error {
	if token == "" {
		return fmt.Errorf("verify token is required")
	}
	cacheKey := fmt.Sprintf("verify:token_%s:%s", level, token)
	stored, err := c.Cache().GetValue(cacheKey)
	if err != nil {
		return fmt.Errorf("verify token invalid or expired")
	}
	expected := fmt.Sprintf("%d:%s", expectedUserID, action)
	if string(stored) != expected {
		return fmt.Errorf("verify token invalid or expired")
	}
	return nil
}

// consumeVerifyToken 消费（删除）token，在操作全部成功后调用
func (c *UserController) consumeVerifyToken(token, level string) {
	cacheKey := fmt.Sprintf("verify:token_%s:%s", level, token)
	_ = c.Cache().RemoveValue(cacheKey)
}

// maskOtp OTP打码，有值则首尾保留中间打码，无值为空
func maskOtp(otp string) string {
	if otp == "" {
		return ""
	}
	if len(otp) <= 2 {
		return otp[:1] + "**"
	}
	return otp[:1] + strings.Repeat("*", len(otp)-2) + otp[len(otp)-1:]
}

// maskPhone 手机号打码，如 13812345678 → 138****5678
func maskPhone(phone string) string {
	if len(phone) <= 4 {
		return phone
	}
	if len(phone) >= 7 {
		return phone[:3] + strings.Repeat("*", len(phone)-7) + phone[len(phone)-4:]
	}
	return phone[:1] + strings.Repeat("*", len(phone)-2) + phone[len(phone)-1:]
}

func maskOther(s string) string {
	if s == "" {
		return ""
	}
	return s[:1] + "****" + s[len(s)-1:]
}

// maskEmail 邮箱打码，如 test@example.com → t***@example.com
func maskEmail(email string) string {
	if email == "" {
		return ""
	}
	at := strings.Index(email, "@")
	if at <= 0 {
		return email
	}
	local := email[:at]
	domain := email[at:]
	if len(local) <= 1 {
		return local + "***" + domain
	}
	return local[:1] + strings.Repeat("*", len(local)-1) + domain
}
