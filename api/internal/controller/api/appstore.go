package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fs185085781/v9os/internal/ioc"
	"github.com/fs185085781/v9os/internal/ioc/uioc"
	"github.com/fs185085781/v9os/internal/logger"
	"github.com/fs185085781/v9os/internal/model/plugin"
	"github.com/fs185085781/v9os/internal/plugin/manager"
	"github.com/fs185085781/v9os/pkg/util"

	"github.com/fs185085781/v9os/internal/controller"
	"github.com/fs185085781/v9os/internal/store"
	"github.com/gin-gonic/gin"
)

type AppStoreController struct {
	*controller.BaseController
	fontLocks sync.Map
}

type appStoreInstallRequest struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	HostURL string `json:"hostUrl"`
}

const appStoreInstallMachineProxyHeader = "X-AppStore-Install-Machine-Proxy"

func init() {
	c := &AppStoreController{
		BaseController: controller.GetBaseController(),
	}
	//获取并持久化应用图标
	c.RegisterPublic("api", "GET", "/appstore/img/:code", c.Img)
	//获取并持久化字体
	c.RegisterPublic("api", "POST", "/appstore/read_fonts", c.ReadFonts)
	//获取字体列表
	c.RegisterPublic("api", "POST", "/appstore/fonts", c.Fonts)
	c.RegisterApi("POST", "/appstore/categories", c.Categories, "应用管理", "应用商店", "分类")
	c.RegisterApi("POST", "/appstore/apps", c.Apps, "应用管理", "应用商店", "列表")
	c.RegisterApi("POST", "/appstore/search", c.Search, "应用管理", "应用商店", "搜索")
	c.RegisterApi("POST", "/appstore/detail", c.Detail, "应用管理", "应用商店", "详情")
	c.RegisterApi("POST", "/appstore/install", c.Install, "应用管理", "应用商店", "安装/升级")
	c.RegisterApi("POST", "/appstore/add", c.Add, "应用管理", "应用商店", "添加应用")
	c.RegisterApi("POST", "/appstore/uninstall", c.Uninstall, "应用管理", "应用商店", "卸载")
	c.RegisterApi("POST", "/appstore/installed", c.Installed, "应用管理", "应用商店", "已安装列表")
}

func (c *AppStoreController) Categories(ctx *gin.Context) {
	data, err := c.Store().GetCategories()
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.OkData(ctx, data)
}

func (c *AppStoreController) Fonts(ctx *gin.Context) {
	data, err := c.Store().GetFontOptions()
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.OkData(ctx, data)
}

func (c *AppStoreController) readFontsByUrl(url string, name string) io.ReadCloser {
	lockAny, _ := c.fontLocks.LoadOrStore(name, &sync.Mutex{})
	lock := lockAny.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()
	fontDir := filepath.Join(util.RunDir(), "fonts")
	fp := filepath.Join(fontDir, name)
	if f, err := os.Open(fp); err == nil {
		return f
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, util.HttpAutoRedirectKey, true)
	resp, err := util.GetResp(ctx, url, nil)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}
	if err := os.MkdirAll(fontDir, 0755); err != nil {
		return nil
	}
	file, err := os.Create(fp)
	if err != nil {
		return nil
	}
	if _, err := io.Copy(file, resp.Body); err != nil {
		file.Close()
		os.Remove(fp)
		return nil
	}
	if err := file.Close(); err != nil {
		return nil
	}
	f, err := os.Open(fp)
	if err != nil {
		return nil
	}
	return f
}
func (c *AppStoreController) ReadFonts(ctx *gin.Context) {
	param := c.Param(ctx)
	font := param.ParamString("font")
	ui := param.ParamString("ui")
	name := font + ".font"
	if font == "default" {
		name = ui + "_default.font"
	}
	url := c.Store().GetFontUrl(font, ui)
	if url == "" {
		c.ErrCode(ctx, http.StatusNotFound, "font not found")
		return
	}
	f := c.readFontsByUrl(url, name)
	if f == nil {
		c.ErrCode(ctx, http.StatusNotFound, "font not found")
		return
	}
	defer f.Close()
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Cache-Control", "public, max-age=31536000")
	ctx.Header("Content-Disposition", "inline; filename="+param.ParamString("font")+".font")
	ctx.Status(http.StatusOK)
	io.Copy(ctx.Writer, f)
}

func (c *AppStoreController) Img(ctx *gin.Context) {
	code := strings.TrimSpace(ctx.Param("code"))
	if code == "" {
		c.ErrCode(ctx, http.StatusBadRequest, "code is required")
		return
	}
	var pluginModel plugin.Plugin
	if err := c.Database().Read().Where("code = ?", code).First(&pluginModel).Error; err != nil {
		c.ErrCode(ctx, http.StatusNotFound, err.Error())
		return
	}
	pluginManage := c.PluginManage(pluginModel.PluginType)
	if pluginModel.PluginType == 4 {
		if pluginModel.IconUrl == "" {
			c.ErrCode(ctx, http.StatusNotFound, "plugin icon not found")
			return
		}
		ctx.Redirect(http.StatusTemporaryRedirect, pluginModel.IconUrl)
		return
	}
	if pluginManage == nil {
		c.ErrCode(ctx, http.StatusNotFound, "plugin manager not found")
		return
	}
	logoPath := filepath.Join(pluginManage.PluginDir(code), "logo.png")
	if _, err := os.Stat(logoPath); err != nil {
		c.ErrCode(ctx, http.StatusNotFound, err.Error())
		return
	}
	ctx.Header("Cache-Control", "public, max-age=31536000")
	ctx.File(logoPath)
}

func (c *AppStoreController) Apps(ctx *gin.Context) {
	param := c.Param(ctx)
	category := param.ParamString("category")
	page := param.ParamInt("page")
	if page < 1 {
		page = 1
	}
	pageSize := param.ParamInt("pageSize")
	if pageSize < 1 {
		pageSize = 20
	}
	result, err := c.Store().GetAppsByCategory(category, page, pageSize)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	for i := range result.Data {
		normalizeStoreAppInfo(result.Data[i].Code, &result.Data[i])
	}
	c.markInstalled(result.Data)
	c.OkData(ctx, result)
}

func (c *AppStoreController) Search(ctx *gin.Context) {
	param := c.Param(ctx)
	keyword := param.ParamString("keyword")
	page := param.ParamInt("page")
	if page < 1 {
		page = 1
	}
	pageSize := param.ParamInt("pageSize")
	if pageSize < 1 {
		pageSize = 20
	}
	result, err := c.Store().SearchApps(keyword, page, pageSize)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	for i := range result.Data {
		normalizeStoreAppInfo(result.Data[i].Code, &result.Data[i])
	}
	c.markInstalled(result.Data)
	c.OkData(ctx, result)
}

func (c *AppStoreController) Detail(ctx *gin.Context) {
	param := c.Param(ctx)
	code := param.ParamString("code")
	appInfo, storeErr := c.Store().GetAppDetail(code)
	if storeErr != nil {
		appInfo = &store.AppInfo{
			Code: code,
		}
	} else if appInfo != nil {
		appInfo.StoreVersion = appInfo.Version
		appInfo.Version = ""
	}

	versions := make([]store.AppVersion, 0)
	if storeErr == nil {
		var versionErr error
		versions, versionErr = c.Store().GetAppVersions(code)
		if versionErr != nil {
			versions = make([]store.AppVersion, 0)
		}
	}

	var pluginModel plugin.Plugin
	c.Database().Read().Where("code = ?", code).First(&pluginModel)
	if pluginModel.ID > 0 {
		mergeLocalPluginAppInfo(appInfo, pluginModel)
		appInfo.Installed = true
		appInfo.InstalledVersion = pluginModel.Version
	} else if storeErr != nil {
		c.ErrMsg(ctx, storeErr)
		return
	}
	normalizeStoreAppInfo(code, appInfo)
	c.OkData(ctx, gin.H{
		"app":      appInfo,
		"versions": versions,
	})
}

func (c *AppStoreController) Install(ctx *gin.Context) {
	var req appStoreInstallRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.FailMsg(ctx, c.GetText(ctx, "common.param.invalid"))
		return
	}
	if !c.StreamStart(ctx) {
		return
	}
	writeError := func(err error) {
		c.Log().Error("appstore install failed", logger.NewField("code", strings.TrimSpace(req.Code)), logger.NewField("err", err))
		c.StreamWrite(ctx, gin.H{
			"code": -1,
			"msg":  err.Error(),
			"done": true,
		})
	}
	writeFail := func(msg string) {
		c.Log().Warn("appstore install rejected", logger.NewField("code", strings.TrimSpace(req.Code)), logger.NewField("msg", msg))
		c.StreamWrite(ctx, gin.H{
			"code": -1,
			"msg":  msg,
			"done": true,
		})
	}
	writeProgress := func(percent int, msg string) bool {
		if msg == "downloading" {
			msg = c.GetText(ctx, "common.progress.downloading")
		}
		return c.StreamWrite(ctx, gin.H{
			"code":     0,
			"progress": percent,
			"msg":      msg,
		})
	}
	isUpgrade := req.Type == "2"
	code := strings.TrimSpace(req.Code)
	hostUrl := strings.TrimSpace(req.HostURL)
	if code == "" {
		writeFail(c.GetText(ctx, "common.param.invalid"))
		return
	}
	if !writeProgress(1, c.GetText(ctx, "common.progress.preparing")) {
		return
	}
	c.Log().Info("appstore install started",
		logger.NewField("code", code),
		logger.NewField("type", req.Type),
		logger.NewField("userID", ctx.GetString("userID")),
		logger.NewField("proxy", ctx.GetHeader(appStoreInstallMachineProxyHeader) != ""))
	appInfo, err := c.Store().GetAppDetail(code)
	if err != nil {
		writeError(err)
		return
	}
	normalizeStoreAppInfo(code, appInfo)
	targetHost, localInstall, err := c.resolveInstallMachine(ctx, appInfo)
	if err != nil {
		writeError(err)
		return
	}
	if !appInfo.Installable {
		writeFail(appInfo.InstallReason)
		return
	}
	if !localInstall {
		c.proxyAppStoreInstall(ctx, targetHost, req, writeError)
		return
	}
	lock := c.Cache().CreateLock("lock:plugin:install:" + code)
	if !lock.TryLock() {
		writeFail(c.GetText(ctx, "common.distributed.installrunning"))
		return
	}
	defer lock.UnLock()
	pluginManage := c.PluginManage(appInfo.PluginType)
	if pluginManage == nil {
		writeFail(c.GetText(ctx, "common.appstore.pluginmanagernotfound"))
		return
	}
	_, err = pluginManage.Install(appInfo, manager.InstallOptions{
		Upgrade:      isUpgrade,
		AccessOrigin: hostUrl,
		Progress:     writeProgress,
	})
	if err != nil {
		writeError(err)
		return
	}
	c.Log().Info("appstore install finished",
		logger.NewField("code", code),
		logger.NewField("pluginType", appInfo.PluginType),
		logger.NewField("userID", ctx.GetString("userID")))
	c.StreamWrite(ctx, gin.H{
		"code":     0,
		"progress": 100,
		"msg":      c.GetText(ctx, "common.progress.installed"),
		"done":     true,
	})
}

func (c *AppStoreController) Add(ctx *gin.Context) {
	pluginType := strings.TrimSpace(ctx.PostForm("pluginType"))
	if pluginType == "" {
		c.FailMsg(ctx, c.GetText(ctx, "common.param.invalid"))
		return
	}
	if pluginType == "4" {
		c.addRemoteApp(ctx)
		return
	}
	ptype, err := strconv.Atoi(pluginType)
	if err != nil || ptype < 1 || ptype > 3 {
		c.FailMsg(ctx, c.GetText(ctx, "common.param.invalid"))
		return
	}
	if err := validateLocalPackageTarget(ptype, strings.TrimSpace(ctx.PostForm("os")), strings.TrimSpace(ctx.PostForm("arch")), runtime.GOOS, runtime.GOARCH); err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	distributedProvider := c.Distributed()
	if distributedProvider != nil && distributedProvider.Enabled() {
		c.FailMsg(ctx, "分布式环境请前往分布式监控中心安装本地应用")
		return
	}
	file, err := ctx.FormFile("file")
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	tmpRoot := filepath.Join(util.RunDir(), "plugins", ".tmp", util.UUID())
	if err := os.MkdirAll(tmpRoot, 0755); err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	defer os.RemoveAll(tmpRoot)
	zipPath := filepath.Join(tmpRoot, "package.zip")
	if err := ctx.SaveUploadedFile(file, zipPath); err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	pluginManage := iocPluginManage(c)
	if pluginManage == nil {
		c.FailMsg(ctx, c.GetText(ctx, "common.appstore.pluginmanagernotfound"))
		return
	}
	accessOrigin := strings.TrimSpace(ctx.PostForm("hostUrl"))
	pluginModel, err := pluginManage.InstallLocalPackage(zipPath, ptype, manager.InstallOptions{
		AccessOrigin: accessOrigin,
	})
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.OkData(ctx, pluginModel)
}

func (c *AppStoreController) addRemoteApp(ctx *gin.Context) {
	accessURL := strings.TrimSpace(ctx.PostForm("accessUrl"))
	if accessURL == "" {
		c.FailMsg(ctx, "远程应用访问地址不能为空")
		return
	}
	parsedURL, err := url.Parse(accessURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Host == "" {
		c.FailMsg(ctx, "远程应用访问地址无效")
		return
	}
	appInfo := &store.AppInfo{
		Name:        strings.TrimSpace(ctx.PostForm("name")),
		Description: strings.TrimSpace(ctx.PostForm("description")),
		Remark:      strings.TrimSpace(ctx.PostForm("remark")),
		IconUrl:     strings.TrimSpace(ctx.PostForm("iconUrl")),
		AccessUrl:   accessURL,
		Version:     "0.0.0",
		PluginType:  4,
		Status:      1,
	}
	pluginManage := c.PluginManage(4)
	if pluginManage == nil {
		c.FailMsg(ctx, c.GetText(ctx, "common.appstore.pluginmanagernotfound"))
		return
	}
	pluginModel, err := pluginManage.Install(appInfo, manager.InstallOptions{})
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.OkData(ctx, pluginModel)
}

func iocPluginManage(c *AppStoreController) *manager.AllPluginManage {
	v := ioc.Ioc().Get(ioc.KeyPluginManage)
	if v == nil {
		return nil
	}
	all, _ := v.(*manager.AllPluginManage)
	return all
}

func validateLocalPackageTarget(pluginType int, osName string, arch string, actualOS string, actualArch string) error {
	if pluginType == 2 {
		return nil
	}
	osName = strings.ToLower(strings.TrimSpace(osName))
	arch = strings.ToLower(strings.TrimSpace(arch))
	actualOS = strings.ToLower(strings.TrimSpace(actualOS))
	actualArch = strings.ToLower(strings.TrimSpace(actualArch))
	if !validLocalPackageOS(osName) || !validLocalPackageArch(arch) {
		return fmt.Errorf("请选择有效的系统和架构")
	}
	if osName != actualOS || arch != actualArch {
		return fmt.Errorf("当前机器系统架构不匹配: 当前 %s/%s, 选择 %s/%s", actualOS, actualArch, osName, arch)
	}
	return nil
}

func validLocalPackageOS(osName string) bool {
	switch osName {
	case "windows", "linux", "darwin", "android":
		return true
	default:
		return false
	}
}

func validLocalPackageArch(arch string) bool {
	switch arch {
	case "amd64", "arm64":
		return true
	default:
		return false
	}
}

func (c *AppStoreController) resolveInstallMachine(ctx *gin.Context, appInfo *store.AppInfo) (string, bool, error) {
	if appInfo == nil || appInfo.PluginType == 4 {
		return "", true, nil
	}
	distributedProvider := c.Distributed()
	if distributedProvider == nil || !distributedProvider.Enabled() {
		return "", true, nil
	}
	code := strings.TrimSpace(appInfo.Code)
	localMachineID := strings.TrimSpace(distributedProvider.Nodes().LocalMachineID())
	if appStoreMachineAllowsPluginInMemory(localMachineID, code) {
		appInfo.Installable, appInfo.InstallReason = storeAppInstallability(appInfo)
		return "", true, nil
	}
	if ctx.GetHeader(appStoreInstallMachineProxyHeader) != "" {
		return "", false, fmt.Errorf("current machine is not allowed to install plugin: %s", code)
	}
	reasons := make([]string, 0)
	pluginMap := uioc.PluginAllowedMachineMap()
	for machineID := range pluginMap[code] {
		if strings.TrimSpace(machineID) == "" || machineID == localMachineID {
			continue
		}
		host, ok := distributedProvider.Nodes().Resolve(machineID)
		if ok && strings.TrimSpace(host) != "" {
			if machine, ok := distributedProvider.Nodes().Info(machineID); ok {
				installable, reason := storeAppInstallabilityForTarget(appInfo, machine.OS, machine.Arch)
				if !installable {
					reasons = append(reasons, machineID+"("+machine.OS+"/"+machine.Arch+"): "+reason)
					continue
				}
				appInfo.Installable = true
				appInfo.InstallReason = ""
			}
			return host, false, nil
		}
	}
	if len(reasons) > 0 {
		err := fmt.Errorf("plugin %s has no supported online allowed install machine: %s", code, strings.Join(reasons, "; "))
		c.setPluginRuntimeError(code, err)
		return "", false, err
	}
	err := fmt.Errorf("plugin %s has no online allowed install machine", code)
	c.setPluginRuntimeError(code, err)
	return "", false, err
}

func (c *AppStoreController) setPluginRuntimeError(code string, err error) {
	if strings.TrimSpace(code) == "" || err == nil {
		return
	}
	_ = c.Database().Write().Model(&plugin.Plugin{}).Where("code = ?", code).Update("runtime_error", err.Error()).Error
}

func appStoreMachineAllowsPluginInMemory(machineID string, pluginCode string) bool {
	machineID = strings.TrimSpace(machineID)
	pluginCode = strings.TrimSpace(pluginCode)
	if machineID == "" || pluginCode == "" {
		return false
	}
	machineMap := uioc.MachineAllowedPluginMap()
	if machineMap == nil {
		return false
	}
	return machineMap[machineID][pluginCode]
}

func (c *AppStoreController) proxyAppStoreInstall(ctx *gin.Context, targetHost string, req appStoreInstallRequest, writeError func(error)) {
	body, err := json.Marshal(req)
	if err != nil {
		writeError(err)
		return
	}
	httpReq, err := http.NewRequestWithContext(ctx.Request.Context(), http.MethodPost, "http://"+targetHost+"/api/appstore/install", bytes.NewReader(body))
	if err != nil {
		writeError(err)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set(appStoreInstallMachineProxyHeader, "1")
	if auth := ctx.Request.Header.Get("Authorization"); auth != "" {
		httpReq.Header.Set("Authorization", auth)
	}
	if lang := ctx.Request.Header.Get("lang"); lang != "" {
		httpReq.Header.Set("lang", lang)
	}
	resp, err := (&http.Client{Timeout: 30 * time.Minute}).Do(httpReq)
	if err != nil {
		writeError(err)
		return
	}
	c.Log().Info("appstore install proxied", logger.NewField("code", req.Code), logger.NewField("targetHost", targetHost))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		msg := strings.TrimSpace(string(respBody))
		if msg == "" {
			msg = resp.Status
		}
		writeError(fmt.Errorf("%s", msg))
		return
	}
	buffer := make([]byte, 32*1024)
	for {
		n, readErr := resp.Body.Read(buffer)
		if n > 0 {
			if _, err := ctx.Writer.Write(buffer[:n]); err != nil {
				return
			}
			if flusher, ok := ctx.Writer.(http.Flusher); ok {
				flusher.Flush()
			}
		}
		if readErr == io.EOF {
			return
		}
		if readErr != nil {
			writeError(readErr)
			return
		}
	}
}

func (c *AppStoreController) Uninstall(ctx *gin.Context) {
	param := c.Param(ctx)
	code := param.ParamString("code")
	c.Log().Info("appstore uninstall started", logger.NewField("code", code), logger.NewField("userID", ctx.GetString("userID")))
	var pluginModel plugin.Plugin
	err := c.Database().Read().Where("code = ?", code).First(&pluginModel).Error
	if err != nil {
		c.Log().Error("appstore uninstall failed", logger.NewField("code", code), logger.NewField("err", err))
		c.ErrMsg(ctx, err)
		return
	}
	pluginManage := c.PluginManage(pluginModel.PluginType)
	if pluginManage == nil {
		c.FailMsg(ctx, c.GetText(ctx, "common.appstore.pluginmanagernotfound"))
		return
	}
	if err := pluginManage.Uninstall(pluginModel); err != nil {
		c.Log().Error("appstore uninstall failed", logger.NewField("code", code), logger.NewField("err", err))
		c.ErrMsg(ctx, err)
		return
	}
	c.Log().Info("appstore uninstall finished", logger.NewField("code", code), logger.NewField("pluginType", pluginModel.PluginType), logger.NewField("userID", ctx.GetString("userID")))
	c.Ok(ctx)
}

func (c *AppStoreController) Installed(ctx *gin.Context) {
	var plugins []plugin.Plugin
	err := c.Database().Read().Find(&plugins).Error
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.OkData(ctx, plugins)
}

func normalizeStoreAppInfo(code string, appInfo *store.AppInfo) {
	if appInfo == nil {
		return
	}
	if appInfo.Code == "" {
		appInfo.Code = code
	}
	appInfo.Installable, appInfo.InstallReason = storeAppInstallability(appInfo)
}

func mergeLocalPluginAppInfo(appInfo *store.AppInfo, pluginModel plugin.Plugin) {
	if pluginModel.Code != "" {
		appInfo.Code = pluginModel.Code
	}
	if pluginModel.Name != "" {
		appInfo.Name = pluginModel.Name
	}
	if pluginModel.Description != "" {
		appInfo.Description = pluginModel.Description
	}
	if pluginModel.IconUrl != "" {
		appInfo.IconUrl = pluginModel.IconUrl
	}
	if pluginModel.Remark != "" {
		appInfo.Remark = pluginModel.Remark
	}
	if pluginModel.Version != "" {
		appInfo.Version = pluginModel.Version
	}
	if pluginModel.PluginType != 0 {
		appInfo.PluginType = pluginModel.PluginType
	}
	if pluginModel.RuntimeError != "" {
		appInfo.RuntimeError = pluginModel.RuntimeError
	}
	if pluginModel.LimitVersion != "" {
		appInfo.LimitVersion = pluginModel.LimitVersion
	}
	if pluginModel.FirstMachine != "" {
		appInfo.FirstMachine = pluginModel.FirstMachine
	}
	if pluginModel.CloseDelay != 0 {
		appInfo.CloseDelay = pluginModel.CloseDelay
	}
	if pluginModel.Status != 0 {
		appInfo.Status = pluginModel.Status
	}
	if pluginModel.NeedLogin != 0 {
		appInfo.NeedLogin = pluginModel.NeedLogin
	}
	if pluginModel.Interceptors != "" {
		appInfo.Interceptors = pluginModel.Interceptors
	}
	if pluginModel.WebHook != "" {
		appInfo.WebHook = pluginModel.WebHook
	}
	if pluginModel.AccessUrl != "" {
		appInfo.AccessUrl = pluginModel.AccessUrl
	}
	if pluginModel.DebugPort != 0 {
		appInfo.DebugPort = pluginModel.DebugPort
	}
}

func storeCurrentTarget(pluginType int) (string, string) {
	if pluginType == 2 || pluginType == 4 {
		return "all", "all"
	}
	return runtime.GOOS, runtime.GOARCH
}

func storePackageMatches(pkg store.AppPackage, os string, arch string) bool {
	pkgOS := strings.ToLower(strings.TrimSpace(pkg.OS))
	pkgArch := strings.ToLower(strings.TrimSpace(pkg.Arch))
	return pkgOS == os && pkgArch == arch
}

func storePackageLabel(pkg store.AppPackage) string {
	pkgOS := strings.ToLower(strings.TrimSpace(pkg.OS))
	pkgArch := strings.ToLower(strings.TrimSpace(pkg.Arch))
	if pkgOS == "all" && pkgArch == "all" {
		return "all/all"
	}
	return pkgOS + "/" + pkgArch
}

func storeAppInstallability(appInfo *store.AppInfo) (bool, string) {
	if appInfo == nil {
		return false, "plugin not found"
	}
	if appInfo.PluginType == 4 {
		if strings.TrimSpace(appInfo.AccessUrl) == "" {
			return false, "远程应用访问地址不能为空"
		}
		return true, ""
	}
	os, arch := storeCurrentTarget(appInfo.PluginType)
	return storeAppInstallabilityForTarget(appInfo, os, arch)
}

func storeAppInstallabilityForTarget(appInfo *store.AppInfo, os string, arch string) (bool, string) {
	if appInfo == nil {
		return false, "plugin not found"
	}
	if appInfo.PluginType == 4 {
		if strings.TrimSpace(appInfo.AccessUrl) == "" {
			return false, "远程应用访问地址不能为空"
		}
		return true, ""
	}
	if appInfo.PluginType == 2 {
		os = "all"
		arch = "all"
	}
	os = strings.ToLower(strings.TrimSpace(os))
	arch = strings.ToLower(strings.TrimSpace(arch))
	if len(appInfo.Packages) == 0 {
		return false, "当前插件没有可安装的包"
	}
	available := make([]string, 0, len(appInfo.Packages))
	for _, pkg := range appInfo.Packages {
		pkgOS := strings.ToLower(strings.TrimSpace(pkg.OS))
		pkgArch := strings.ToLower(strings.TrimSpace(pkg.Arch))
		available = append(available, storePackageLabel(pkg))
		if storePackageMatches(pkg, os, arch) || (pkgOS == "all" && pkgArch == "all") {
			return true, ""
		}
	}
	return false, "当前内核环境不支持安装该插件，需要 " + os + "/" + arch + "，可用包: " + strings.Join(available, ", ")
}

func (c *AppStoreController) markInstalled(apps []store.AppInfo) {
	if len(apps) == 0 {
		return
	}
	var plugins []plugin.Plugin
	err := c.Database().Read().Find(&plugins).Error
	if err != nil {
		return
	}
	pluginMap := make(map[string]plugin.Plugin)
	for _, p := range plugins {
		pluginMap[p.Code] = p
	}
	for i := range apps {
		if p, ok := pluginMap[apps[i].Code]; ok {
			apps[i].Installed = true
			apps[i].InstalledVersion = p.Version
		}
	}
}
