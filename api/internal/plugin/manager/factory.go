package manager

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fs185085781/v9os/internal/cache"
	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/internal/inface/distributed"
	"github.com/fs185085781/v9os/internal/ioc"
	"github.com/fs185085781/v9os/internal/ioc/uioc"
	"github.com/fs185085781/v9os/internal/logger"
	"github.com/fs185085781/v9os/internal/model/plugin"
	"github.com/fs185085781/v9os/internal/store"
	"github.com/fs185085781/v9os/pkg/util"
	"gorm.io/gorm"
)

type commonPluginManage struct {
}
type AllPluginManage struct {
	main  IPluginManage
	web   IPluginManage
	third IPluginManage
	frame IPluginManage
}

func (o *commonPluginManage) pluginRootDir() string {
	return filepath.Join(util.RunDir(), "plugins")
}

func NewPluginManage(serverPort int, cfg config.Config, c cache.Cache, log logger.Logger) *AllPluginManage {
	common := &commonPluginManage{}
	all := &AllPluginManage{
		main:  common.newMainManage(serverPort, cfg, c, log),
		web:   common.newWebManage(cfg, log),
		third: common.newThirdManage(cfg, c, log),
		frame: common.newFrameManage(),
	}
	common.cleanupOrphanPluginDirs(all, cfg, log)
	return all
}

func (o *AllPluginManage) Switch(ptype int) IPluginManage {
	switch ptype {
	case 1:
		return o.main
	case 2:
		return o.web
	case 3:
		return o.third
	case 4:
		return o.frame
	default:
		return nil
	}
}

func (o *commonPluginManage) cleanupOrphanPluginDirs(all *AllPluginManage, cfg config.Config, log logger.Logger) {
	if all == nil {
		return
	}
	db := uioc.Database()
	var plugins []plugin.Plugin
	if err := db.Read().Where("plugin_type IN ?", []int{1, 2, 3}).Find(&plugins).Error; err != nil {
		if log != nil {
			log.Warn("cleanup orphan plugin dirs skipped", logger.NewField("err", err))
		}
		return
	}
	localAllowedPlugins := map[string]struct{}{}
	useMachineWhitelist := false
	if provider, ok := ioc.Ioc().Get(ioc.KeyDistributedProvider).(distributed.DistributedProvider); ok && provider.Enabled() && cfg != nil {
		useMachineWhitelist = true
		var machine struct {
			AllowedPluginCodes string `gorm:"column:allowed_plugin_codes"`
		}
		if err := db.Read().Table("machine_ee").Where("machine_id = ?", cfg.Machine().MachineId).First(&machine).Error; err == nil {
			var codes []string
			if err := json.Unmarshal([]byte(machine.AllowedPluginCodes), &codes); err == nil {
				for _, code := range codes {
					code = strings.TrimSpace(code)
					if code != "" {
						localAllowedPlugins[code] = struct{}{}
					}
				}
			}
		}
	}
	installed := map[int]map[string]struct{}{
		1: {},
		2: {},
		3: {},
	}
	for _, item := range plugins {
		code := strings.TrimSpace(item.Code)
		if code == "" {
			continue
		}
		if useMachineWhitelist {
			if _, ok := localAllowedPlugins[code]; !ok {
				continue
			}
		} else if item.PluginType == 3 && cfg != nil && strings.TrimSpace(item.FirstMachine) != "" && item.FirstMachine != cfg.Machine().MachineId {
			continue
		}
		if _, ok := installed[item.PluginType]; ok {
			installed[item.PluginType][code] = struct{}{}
		}
	}
	o.cleanupOrphanPluginDirByType(1, filepath.Join(o.pluginRootDir(), "main"), all.main, installed[1], log)
	o.cleanupOrphanPluginDirByType(2, filepath.Join(o.pluginRootDir(), "web"), all.web, installed[2], log)
	o.cleanupOrphanPluginDirByType(3, filepath.Join(o.pluginRootDir(), "third"), all.third, installed[3], log)
}

func (o *commonPluginManage) cleanupOrphanPluginDirByType(pluginType int, root string, manage IPluginManage, installed map[string]struct{}, log logger.Logger) {
	entries, err := os.ReadDir(root)
	if err != nil {
		if !os.IsNotExist(err) && log != nil {
			log.Warn("read plugin dir failed", logger.NewField("pluginType", pluginType), logger.NewField("root", root), logger.NewField("err", err))
		}
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		code := strings.TrimSpace(entry.Name())
		if code == "" || strings.HasPrefix(code, ".") {
			continue
		}
		if _, ok := installed[code]; ok {
			continue
		}
		dir := filepath.Join(root, code)
		o.stopManagedPluginProcess(manage, code)
		if err := os.RemoveAll(dir); err != nil {
			if log != nil {
				log.Warn("remove orphan plugin dir failed", logger.NewField("pluginType", pluginType), logger.NewField("code", code), logger.NewField("dir", dir), logger.NewField("err", err))
			}
			continue
		}
		if log != nil {
			log.Info("orphan plugin dir removed", logger.NewField("pluginType", pluginType), logger.NewField("code", code), logger.NewField("dir", dir))
		}
	}
}

func (o *commonPluginManage) stopManagedPluginProcess(manage IPluginManage, code string) {
	if manage == nil {
		return
	}
	switch m := manage.(type) {
	case *mainPluginManage:
		_ = m.Close(code)
	case *thirdPluginManage:
		m.closeManagedProcess(code)
	default:
		_ = manage.Close(code)
		manage.Stop(code)
	}
}

func (o *commonPluginManage) resolvePackageURL(appInfo *store.AppInfo) string {
	if appInfo == nil {
		return ""
	}
	downloadOS := runtime.GOOS
	downloadArch := runtime.GOARCH
	if appInfo.PluginType == 2 {
		downloadOS = "all"
		downloadArch = "all"
	}
	if len(appInfo.Packages) == 0 && strings.TrimSpace(appInfo.Code) != "" && strings.TrimSpace(appInfo.Version) != "" {
		if s := ioc.Ioc().Get(ioc.KeyStore); s != nil {
			return s.(store.Store).PluginDownloadUrl(appInfo.Code, appInfo.Version, downloadOS, downloadArch)
		}
		return ""
	}
	pkg := o.findPackageForTarget(appInfo, downloadOS, downloadArch)
	if pkg == nil {
		return ""
	}
	if s := ioc.Ioc().Get(ioc.KeyStore); s != nil {
		return s.(store.Store).PluginDownloadUrl(appInfo.Code, appInfo.Version, pkg.OS, pkg.Arch)
	}
	return ""
}

func (o *commonPluginManage) buildPluginModel(appInfo *store.AppInfo, manifest *packageManifest) plugin.Plugin {
	pluginModel := manifest.toPluginModel()
	if appInfo == nil {
		return pluginModel
	}
	if pluginModel.Name == "" {
		pluginModel.Name = appInfo.Name
	}
	if pluginModel.Description == "" {
		pluginModel.Description = appInfo.Description
	}
	if pluginModel.Code == "" {
		pluginModel.Code = appInfo.Code
	}
	if pluginModel.Version == "" {
		pluginModel.Version = appInfo.Version
	}
	if pluginModel.PluginType == 0 {
		pluginModel.PluginType = appInfo.PluginType
	}
	if strings.TrimSpace(appInfo.IconUrl) != "" {
		pluginModel.IconUrl = strings.TrimSpace(appInfo.IconUrl)
	}
	if pluginModel.LimitVersion == "" {
		pluginModel.LimitVersion = appInfo.LimitVersion
	}
	return pluginModel
}

func (o *commonPluginManage) snapshotPluginIcon(pluginModel *plugin.Plugin, pluginDir string) {
	if pluginModel == nil {
		return
	}
	iconURL := strings.TrimSpace(pluginModel.IconUrl)
	if iconURL == "" || pluginDir == "" {
		return
	}
	if err := o.savePluginLogo(iconURL, filepath.Join(pluginDir, "logo.png")); err == nil {
		pluginModel.IconUrl = "/api/appstore/img/" + strings.TrimSpace(pluginModel.Code)
	}
}

func (o *commonPluginManage) savePluginLogo(iconURL string, targetPath string) error {
	if !strings.HasPrefix(iconURL, "http") {
		return fmt.Errorf("icon url is not remote")
	}
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Get(iconURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download icon failed: %s", resp.Status)
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}
	tmpPath := targetPath + ".tmp"
	out, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, io.LimitReader(resp.Body, 10*1024*1024))
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(tmpPath)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		return closeErr
	}
	return os.Rename(tmpPath, targetPath)
}

func (o *commonPluginManage) upsertPluginModel(pluginModel *plugin.Plugin) error {
	db := uioc.Database().Write()
	var exists plugin.Plugin
	err := db.Where("code = ?", pluginModel.Code).First(&exists).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.Create(pluginModel).Error
		}
		return err
	}
	updates := map[string]interface{}{
		"first_machine": pluginModel.FirstMachine,
		"runtime_error": pluginModel.RuntimeError,
		"name":          pluginModel.Name,
		"description":   pluginModel.Description,
		"close_delay":   pluginModel.CloseDelay,
		"code":          pluginModel.Code,
		"status":        pluginModel.Status,
		"remark":        pluginModel.Remark,
		"version":       pluginModel.Version,
		"plugin_type":   pluginModel.PluginType,
		"interceptors":  pluginModel.Interceptors,
		"web_hook":      pluginModel.WebHook,
		"limit_version": pluginModel.LimitVersion,
		"need_login":    pluginModel.NeedLogin,
		"icon_url":      pluginModel.IconUrl,
		"access_url":    pluginModel.AccessUrl,
		"debug_port":    pluginModel.DebugPort,
	}
	return db.Model(&plugin.Plugin{}).Where("code = ?", pluginModel.Code).Updates(updates).Error
}

func (o *commonPluginManage) deletePluginModel(code string) error {
	db := uioc.Database()
	return db.Write().Where("code = ?", code).Delete(&plugin.Plugin{}).Error
}

func (o *commonPluginManage) findPackageForTarget(appInfo *store.AppInfo, targetOS string, targetArch string) *store.AppPackage {
	if appInfo == nil {
		return nil
	}
	targetOS = strings.ToLower(strings.TrimSpace(targetOS))
	targetArch = strings.ToLower(strings.TrimSpace(targetArch))
	var generic *store.AppPackage
	for i := range appInfo.Packages {
		pkg := &appInfo.Packages[i]
		pkgOS := strings.ToLower(strings.TrimSpace(pkg.OS))
		pkgArch := strings.ToLower(strings.TrimSpace(pkg.Arch))
		if pkgOS == targetOS && pkgArch == targetArch {
			return pkg
		}
		if pkgOS == "all" && pkgArch == "all" {
			generic = pkg
		}
	}
	return generic
}

type packageManifest struct {
	FirstMachine string `json:"FirstMachine"`
	RuntimeError string `json:"RuntimeError"`
	Name         string `json:"Name"`
	Description  string `json:"Description"`
	CloseDelay   int    `json:"CloseDelay,string"`
	Code         string `json:"Code"`
	Status       int    `json:"Status,string"`
	Remark       string `json:"Remark"`
	Version      string `json:"Version"`
	PluginType   int    `json:"PluginType,string"`
	Interceptors string `json:"Interceptors"`
	WebHook      string `json:"WebHook"`
	LimitVersion string `json:"LimitVersion"`
	NeedLogin    int    `json:"NeedLogin,string"`
	IconUrl      string `json:"IconUrl"`
	ThirdPort    int    `json:"ThirdPort,string"`
	DebugPort    int    `json:"DebugPort,string"`
}

func (m *packageManifest) normalize(expectedCode string, expectedType int, cfg config.Config) error {
	if m.Code == "" {
		m.Code = expectedCode
	}
	if m.Code == "" {
		return fmt.Errorf("plugin code not found in index.json")
	}
	if expectedCode != "" && m.Code != expectedCode {
		return fmt.Errorf("plugin code mismatch: expect %s got %s", expectedCode, m.Code)
	}
	if m.PluginType == 0 {
		m.PluginType = expectedType
	}
	if expectedType > 0 && m.PluginType != expectedType {
		return fmt.Errorf("plugin type mismatch: expect %d got %d", expectedType, m.PluginType)
	}
	if m.Status == 0 {
		m.Status = 1
	}
	if m.Version == "" {
		m.Version = "0.0.0"
	}
	if expectedType == 3 && m.FirstMachine == "" && cfg != nil {
		m.FirstMachine = cfg.Machine().MachineId
	}
	return nil
}

func (m *packageManifest) toPluginModel() plugin.Plugin {
	return plugin.Plugin{
		FirstMachine: m.FirstMachine,
		RuntimeError: m.RuntimeError,
		Name:         m.Name,
		Description:  m.Description,
		CloseDelay:   m.CloseDelay,
		Code:         m.Code,
		Status:       m.Status,
		Remark:       m.Remark,
		Version:      m.Version,
		PluginType:   m.PluginType,
		Interceptors: m.Interceptors,
		WebHook:      m.WebHook,
		LimitVersion: m.LimitVersion,
		NeedLogin:    m.NeedLogin,
		IconUrl:      m.IconUrl,
		DebugPort:    m.DebugPort,
	}
}

type installPackageResult struct {
	Manifest     *packageManifest
	TargetDir    string
	ManifestPath string
}

func (o *AllPluginManage) InstallLocalPackage(zipPath string, pluginType int, opts InstallOptions) (*plugin.Plugin, error) {
	switch pluginType {
	case 1:
		if m, ok := o.main.(*mainPluginManage); ok {
			return m.InstallLocalPackage(zipPath, opts)
		}
	case 2:
		if m, ok := o.web.(*webPluginManage); ok {
			return m.InstallLocalPackage(zipPath, opts)
		}
	case 3:
		if m, ok := o.third.(*thirdPluginManage); ok {
			return m.InstallLocalPackage(zipPath, opts)
		}
	}
	return nil, fmt.Errorf("plugin manager not found: %d", pluginType)
}

func (o *commonPluginManage) installLocalPluginPackage(zipPath string, targetDir func(string) string, expectedType int, expectedCode string, cfg config.Config) (*installPackageResult, error) {
	if strings.TrimSpace(zipPath) == "" {
		return nil, fmt.Errorf("package file is empty")
	}
	tmpRoot := filepath.Join(o.pluginRootDir(), ".tmp", util.UUID())
	extractDir := filepath.Join(tmpRoot, "extract")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpRoot)
	if err := o.unzipPackage(zipPath, extractDir); err != nil {
		return nil, err
	}
	manifestPath := filepath.Join(extractDir, "index.json")
	manifest, err := o.readManifest(manifestPath)
	if err != nil {
		return nil, err
	}
	if err := manifest.normalize(expectedCode, expectedType, cfg); err != nil {
		return nil, err
	}
	if targetDir == nil {
		return nil, fmt.Errorf("target dir resolver is nil")
	}
	finalTargetDir := targetDir(manifest.Code)
	if strings.TrimSpace(finalTargetDir) == "" {
		return nil, fmt.Errorf("target dir is empty")
	}
	sourceDir := filepath.Dir(manifestPath)
	if err := os.MkdirAll(filepath.Dir(finalTargetDir), 0755); err != nil {
		return nil, err
	}
	if err := o.replaceDir(sourceDir, finalTargetDir); err != nil {
		return nil, err
	}
	return &installPackageResult{
		Manifest:     manifest,
		TargetDir:    finalTargetDir,
		ManifestPath: filepath.Join(finalTargetDir, "index.json"),
	}, nil
}

func (o *commonPluginManage) installPluginPackage(url, targetDir string, expectedType int, expectedCode string, cfg config.Config, log logger.Logger, progress func(int, string) bool) (*installPackageResult, error) {
	if strings.TrimSpace(url) == "" {
		return nil, fmt.Errorf("package url is empty")
	}
	tmpRoot := filepath.Join(o.pluginRootDir(), ".tmp", util.UUID())
	zipPath := filepath.Join(tmpRoot, "package.zip")
	extractDir := filepath.Join(tmpRoot, "extract")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpRoot)
	if err := o.downloadToFile(url, zipPath, log, progress); err != nil {
		return nil, err
	}
	if err := o.unzipPackage(zipPath, extractDir); err != nil {
		return nil, err
	}
	manifestPath := filepath.Join(extractDir, "index.json")
	manifest, err := o.readManifest(manifestPath)
	if err != nil {
		return nil, err
	}
	if err := manifest.normalize(expectedCode, expectedType, cfg); err != nil {
		return nil, err
	}
	sourceDir := filepath.Dir(manifestPath)
	if err := os.MkdirAll(filepath.Dir(targetDir), 0755); err != nil {
		return nil, err
	}
	if err := o.replaceDir(sourceDir, targetDir); err != nil {
		return nil, err
	}
	return &installPackageResult{
		Manifest:     manifest,
		TargetDir:    targetDir,
		ManifestPath: filepath.Join(targetDir, "index.json"),
	}, nil
}

func (o *commonPluginManage) downloadToFile(url, filePath string, log logger.Logger, progress func(int, string) bool) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()
	resp, err := (&http.Client{Timeout: 30 * time.Minute}).Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		var result struct {
			Msg string `json:"msg"`
		}
		if err := json.Unmarshal(body, &result); err == nil && strings.TrimSpace(result.Msg) != "" {
			return fmt.Errorf("%s", strings.TrimSpace(result.Msg))
		}
		message := strings.TrimSpace(string(body))
		if message == "" {
			return fmt.Errorf("download failed: %s", resp.Status)
		}
		return fmt.Errorf("download failed: %s, body: %s", resp.Status, message)
	}
	if progress != nil && !progress(2, "downloading") {
		return context.Canceled
	}
	buffer := make([]byte, 32*1024)
	var written int64
	lastPercent := 2
	for {
		n, readErr := resp.Body.Read(buffer)
		if n > 0 {
			if _, err := out.Write(buffer[:n]); err != nil {
				return err
			}
			written += int64(n)
			if progress != nil && resp.ContentLength > 0 {
				percent := 2 + int(written*97/resp.ContentLength)
				if percent > 99 {
					percent = 99
				}
				if percent > lastPercent {
					lastPercent = percent
					if !progress(percent, "downloading") {
						return context.Canceled
					}
				}
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	if progress != nil && lastPercent < 99 {
		if !progress(99, "downloading") {
			return context.Canceled
		}
	}
	if log != nil {
		log.Info("plugin package downloaded", logger.NewField("url", url), logger.NewField("filePath", filePath))
	}
	return nil
}

func (o *commonPluginManage) unzipPackage(zipPath, targetDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()
	for _, file := range reader.File {
		cleanName := filepath.Clean(file.Name)
		if cleanName == "." {
			continue
		}
		targetPath := filepath.Join(targetDir, cleanName)
		if !strings.HasPrefix(targetPath, filepath.Clean(targetDir)+string(os.PathSeparator)) && filepath.Clean(targetPath) != filepath.Clean(targetDir) {
			return fmt.Errorf("illegal zip path: %s", file.Name)
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}
		in, err := file.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, file.Mode())
		if err != nil {
			in.Close()
			return err
		}
		_, copyErr := io.Copy(out, in)
		in.Close()
		out.Close()
		if copyErr != nil {
			return copyErr
		}
	}
	return nil
}

func (o *commonPluginManage) readManifest(path string) (*packageManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	manifest := &packageManifest{}
	if err := json.Unmarshal(data, manifest); err != nil {
		return nil, err
	}
	return manifest, nil
}

func (o *commonPluginManage) replaceDir(sourceDir, targetDir string) error {
	backupDir := targetDir + ".bak." + util.UUID()
	if o.fileExists(targetDir) {
		if err := os.Rename(targetDir, backupDir); err != nil {
			return err
		}
	}
	restoreBackup := func() {
		if o.fileExists(backupDir) && !o.fileExists(targetDir) {
			_ = os.Rename(backupDir, targetDir)
		}
	}
	if err := os.Rename(sourceDir, targetDir); err != nil {
		if err := o.copyDir(sourceDir, targetDir); err != nil {
			restoreBackup()
			return err
		}
	}
	_ = os.RemoveAll(backupDir)
	return nil
}

func (o *commonPluginManage) copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, in)
		return err
	})
}

func (o *commonPluginManage) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
