package manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fs185085781/v9os/internal/cache"
	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/internal/inface/distributed"
	infaceUser "github.com/fs185085781/v9os/internal/inface/user"
	"github.com/fs185085781/v9os/internal/ioc"
	"github.com/fs185085781/v9os/internal/ioc/uioc"
	"github.com/fs185085781/v9os/internal/logger"
	"github.com/fs185085781/v9os/internal/model/plugin"
	"github.com/fs185085781/v9os/internal/store"
	"github.com/fs185085781/v9os/pkg/util"
	"github.com/gin-gonic/gin"
)

type thirdPluginManage struct {
	cfg       config.Config
	cache     cache.Cache
	log       logger.Logger
	pluginMap sync.Map
	common    *commonPluginManage
}

type thirdPluginInfo struct {
	mu          sync.Mutex
	cmd         *exec.Cmd
	closeDelay  int
	closeTime   time.Time
	needAddTime bool
}

func (m *thirdPluginManage) pluginEntryPath(code string) string {
	return filepath.ToSlash(filepath.Join("/api/thirdplugin", code))
}

func (m *thirdPluginManage) scriptPaths(code string) (restartPath, stopPath string) {
	base := m.PluginDir(code)
	if runtime.GOOS == "windows" {
		return filepath.Join(base, "restart.bat"), filepath.Join(base, "stop.bat")
	}
	return filepath.Join(base, "restart.sh"), filepath.Join(base, "stop.sh")
}

func (m *thirdPluginManage) validateScripts(code string) error {
	restartPath, stopPath := m.scriptPaths(code)
	if !m.common.fileExists(restartPath) {
		return fmt.Errorf("third plugin restart script not found: %s", restartPath)
	}
	if !m.common.fileExists(stopPath) {
		return fmt.Errorf("third plugin stop script not found: %s", stopPath)
	}
	return nil
}

func (o *commonPluginManage) newThirdManage(cfg config.Config, c cache.Cache, log logger.Logger) IPluginManage {
	m := &thirdPluginManage{
		common: o,
		cfg:    cfg,
		cache:  c,
		log:    log,
	}
	util.Go(m.checkPluginClose)
	m.syncPluginFeatures("", "")
	return m
}

func (m *thirdPluginManage) PluginDir(code string) string {
	return filepath.Join(m.common.pluginRootDir(), "third", code)
}

func (m *thirdPluginManage) installPackage(appInfo *store.AppInfo, progress func(int, string) bool) (*packageManifest, error) {
	result, err := m.common.installPluginPackage(m.common.resolvePackageURL(appInfo), m.PluginDir(appInfo.Code), 3, appInfo.Code, m.cfg, m.log, progress)
	if err != nil {
		return nil, err
	}
	if err := m.validateScripts(result.Manifest.Code); err != nil {
		return nil, err
	}
	return result.Manifest, nil
}

func (m *thirdPluginManage) syncPluginFeatures(code, name string) {
	list := []interface{}{}
	list = append(list, map[string]interface{}{
		"feature": "访问",
		"label":   "访问首页",
		"method":  "access",
	})
	alls := []string{}
	if code != "" {
		alls = append(alls, code)
	} else {
		dbPool := uioc.Database()
		var plugins []plugin.Plugin
		dbPool.Read().Where("status = ? and plugin_type = ?", 1, 3).Find(&plugins)
		for _, p := range plugins {
			alls = append(alls, p.Code)
		}
	}
	if len(alls) > 0 {
		userProvider := ioc.Ioc().Get(ioc.KeyUserProvider).(infaceUser.UserProvider)
		for _, itemCode := range alls {
			userProvider.SyncPluginAuths(itemCode, name, list)
		}
	}
}

func (m *thirdPluginManage) Install(appInfo *store.AppInfo, opts InstallOptions) (*plugin.Plugin, error) {
	if opts.Upgrade {
		if err := m.Close(appInfo.Code); err != nil && !errors.Is(err, os.ErrNotExist) {
			m.log.Warn("third plugin stop before upgrade failed", logger.NewField("code", appInfo.Code), logger.NewField("err", err))
		}
	}
	manifest, err := m.installPackage(appInfo, opts.Progress)
	if err != nil {
		return nil, err
	}
	pluginModel := m.common.buildPluginModel(appInfo, manifest)
	if manifest.ThirdPort > 0 && strings.TrimSpace(opts.AccessOrigin) != "" {
		pluginModel.AccessUrl = strings.TrimRight(opts.AccessOrigin, "/") + ":" + strconv.Itoa(manifest.ThirdPort)
	}
	m.common.snapshotPluginIcon(&pluginModel, m.PluginDir(pluginModel.Code))
	if err := m.common.upsertPluginModel(&pluginModel); err != nil {
		return nil, err
	}
	_ = m.Close(pluginModel.Code)
	m.syncPluginFeatures(pluginModel.Code, pluginModel.Name)
	return &pluginModel, nil
}

func (m *thirdPluginManage) InstallLocalPackage(zipPath string, opts InstallOptions) (*plugin.Plugin, error) {
	result, err := m.common.installLocalPluginPackage(zipPath, m.PluginDir, 3, "", m.cfg)
	if err != nil {
		return nil, err
	}
	if err := m.validateScripts(result.Manifest.Code); err != nil {
		return nil, err
	}
	pluginModel := result.Manifest.toPluginModel()
	if result.Manifest.ThirdPort > 0 && strings.TrimSpace(opts.AccessOrigin) != "" {
		pluginModel.AccessUrl = strings.TrimRight(opts.AccessOrigin, "/") + ":" + strconv.Itoa(result.Manifest.ThirdPort)
	}
	m.common.snapshotPluginIcon(&pluginModel, m.PluginDir(pluginModel.Code))
	if err := m.common.upsertPluginModel(&pluginModel); err != nil {
		return nil, err
	}
	_ = m.Close(pluginModel.Code)
	m.syncPluginFeatures(pluginModel.Code, pluginModel.Name)
	return &pluginModel, nil
}

func (m *thirdPluginManage) Uninstall(pluginModel plugin.Plugin) error {
	if err := m.Close(pluginModel.Code); err != nil && !errors.Is(err, os.ErrNotExist) {
		m.log.Warn("third plugin stop before uninstall failed", logger.NewField("code", pluginModel.Code), logger.NewField("err", err))
	}
	if err := os.RemoveAll(m.PluginDir(pluginModel.Code)); err != nil {
		return err
	}
	return m.common.deletePluginModel(pluginModel.Code)
}

func (m *thirdPluginManage) hasRunHere(pluginModel plugin.Plugin) bool {
	return pluginModel.FirstMachine == m.cfg.Machine().MachineId
}

func (m *thirdPluginManage) checkPluginClose() {
	for {
		time.Sleep(time.Minute)
		m.pluginMap.Range(func(key, value any) bool {
			code := key.(string)
			info := value.(*thirdPluginInfo)
			if !m.isProcessRunning(info.cmd) {
				m.pluginMap.Delete(code)
				return true
			}
			if info.closeDelay <= 0 {
				return true
			}
			if info.needAddTime {
				info.closeTime = time.Now().Add(time.Duration(info.closeDelay) * time.Minute)
				info.needAddTime = false
				return true
			}
			if time.Now().After(info.closeTime) {
				_ = m.Close(code)
			}
			return true
		})
	}
}

func (m *thirdPluginManage) getPluginModel(code string) (*plugin.Plugin, error) {
	db := uioc.Database()
	var pluginModel plugin.Plugin
	err := db.Read().Where("code = ?", code).First(&pluginModel).Error
	if err != nil {
		return nil, err
	}
	if pluginModel.Status != 1 {
		return nil, fmt.Errorf("plugin %s is disabled", code)
	}
	if strings.TrimSpace(pluginModel.RuntimeError) != "" {
		return nil, errors.New(pluginModel.RuntimeError)
	}
	if pluginModel.PluginType != 3 {
		return nil, fmt.Errorf("plugin %s type is invalid", code)
	}
	return &pluginModel, nil
}

func (m *thirdPluginManage) setRuntimeError(code string, err error) {
	if err == nil {
		return
	}
	m.log.Warn("third plugin runtime error", logger.NewField("code", code), logger.NewField("err", err))
	_ = uioc.Database().Write().Model(&plugin.Plugin{}).Where("code = ?", code).Update("runtime_error", err.Error()).Error
}

func (m *thirdPluginManage) clearRuntimeError(code string) {
	_ = uioc.Database().Write().Model(&plugin.Plugin{}).Where("code = ?", code).Update("runtime_error", "").Error
}

func (m *thirdPluginManage) readThirdPluginPort(code string) (int, error) {
	data, err := os.ReadFile(filepath.Join(m.PluginDir(code), "index.json"))
	if err != nil {
		return 0, err
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return 0, err
	}
	switch value := raw["ThirdPort"].(type) {
	case string:
		port, convErr := strconv.Atoi(value)
		if convErr == nil && port > 0 {
			return port, nil
		}
	case float64:
		port := int(value)
		if port > 0 {
			return port, nil
		}
	}
	return 0, fmt.Errorf("third plugin port not found in index.json: %s", code)
}

func (m *thirdPluginManage) isProcessRunning(cmd *exec.Cmd) bool {
	return cmd != nil && cmd.Process != nil && cmd.ProcessState == nil
}

func (m *thirdPluginManage) closeManagedProcess(code string) {
	value, ok := m.pluginMap.Load(code)
	if !ok {
		return
	}
	info := value.(*thirdPluginInfo)
	info.mu.Lock()
	defer info.mu.Unlock()
	if m.isProcessRunning(info.cmd) {
		_ = info.cmd.Process.Kill()
	}
	m.pluginMap.Delete(code)
}

func (m *thirdPluginManage) waitReady(code string, timeout time.Duration) error {
	port, err := m.readThirdPluginPort(code)
	if err != nil {
		return err
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, dialErr := net.DialTimeout("tcp", "127.0.0.1:"+strconv.Itoa(port), 500*time.Millisecond)
		if dialErr == nil {
			conn.Close()
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("third plugin %s start timeout", code)
}

func (m *thirdPluginManage) GetPluginName(code, key string) (string, error) {
	return "", m.common.unsupportedPluginAction("GetPluginName")
}

func (m *thirdPluginManage) PluginHost(pluginName string, ctx *gin.Context) (string, error) {
	pluginModel, err := m.getPluginModel(pluginName)
	if err != nil {
		return "", err
	}
	accessURL := strings.TrimRight(strings.TrimSpace(pluginModel.AccessUrl), "/")
	if accessURL == "" {
		return "", fmt.Errorf("third plugin access url not found")
	}
	if m.hasRunHere(*pluginModel) {
		infoAny, _ := m.pluginMap.LoadOrStore(pluginName, &thirdPluginInfo{})
		info := infoAny.(*thirdPluginInfo)
		if m.isProcessRunning(info.cmd) {
			if info.closeDelay > 0 {
				info.needAddTime = true
			}
			return accessURL, nil
		}
		info.mu.Lock()
		defer info.mu.Unlock()
		if m.isProcessRunning(info.cmd) {
			if info.closeDelay > 0 {
				info.needAddTime = true
			}
			return accessURL, nil
		}
		restartPath, stopPath := m.scriptPaths(pluginName)
		if !m.common.fileExists(restartPath) || !m.common.fileExists(stopPath) {
			_, err := m.installPackage(&store.AppInfo{
				Code:       pluginName,
				Version:    pluginModel.Version,
				PluginType: 3,
			}, nil)
			if err != nil {
				m.setRuntimeError(pluginName, err)
				m.pluginMap.Delete(pluginName)
				return "", err
			}
			restartPath, stopPath = m.scriptPaths(pluginName)
			if !m.common.fileExists(restartPath) || !m.common.fileExists(stopPath) {
				m.pluginMap.Delete(pluginName)
				return "", fmt.Errorf("third plugin %s not installed", pluginName)
			}
		}
		if runtime.GOOS != "windows" {
			if err := os.Chmod(restartPath, 0755); err != nil {
				m.pluginMap.Delete(pluginName)
				return "", err
			}
		}
		cmd := m.execScript(pluginName, restartPath)
		if err := cmd.Start(); err != nil {
			m.setRuntimeError(pluginName, err)
			m.pluginMap.Delete(pluginName)
			return "", err
		}
		info.cmd = cmd
		info.closeDelay = pluginModel.CloseDelay
		info.needAddTime = pluginModel.CloseDelay > 0
		info.closeTime = time.Time{}
		util.Go(func() {
			err := cmd.Wait()
			if err != nil {
				m.log.Warn("third plugin exited", logger.NewField("code", pluginName), logger.NewField("err", err))
			}
			current, ok := m.pluginMap.Load(pluginName)
			if !ok || current != info {
				return
			}
			info.mu.Lock()
			defer info.mu.Unlock()
			if info.cmd == cmd {
				m.pluginMap.Delete(pluginName)
			}
		})
		if err := m.waitReady(pluginName, 30*time.Second); err != nil {
			if cmd.Process != nil && cmd.ProcessState == nil {
				_ = cmd.Process.Kill()
			}
			m.setRuntimeError(pluginName, err)
			m.pluginMap.Delete(pluginName)
			return "", err
		}
		m.clearRuntimeError(pluginName)
		return accessURL, nil
	}
	distributedProvider := ioc.Ioc().Get(ioc.KeyDistributedProvider).(distributed.DistributedProvider)
	machineIDs := distributedProvider.Nodes().GetMachineIds(pluginName)
	if !slices.Contains(machineIDs, m.cfg.Machine().MachineId) {
		err := fmt.Errorf("current machine is not allowed to run third plugin: %s", pluginName)
		m.setRuntimeError(pluginName, err)
		return "", err
	}
	machineID := strings.TrimSpace(pluginModel.FirstMachine)
	if machineID == "" {
		return "", fmt.Errorf("third plugin install machine not found")
	}
	host, ok := distributedProvider.Nodes().Resolve(machineID)
	if !ok || strings.TrimSpace(host) == "" {
		return "", fmt.Errorf("machine host not found: %s", machineID)
	}
	req, err := http.NewRequest(http.MethodGet, "http://"+host+m.pluginEntryPath(pluginName), nil)
	if err != nil {
		return "", err
	}
	resp, err := (&http.Client{
		Timeout: 35 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		message := strings.TrimSpace(string(body))
		if message == "" {
			return "", fmt.Errorf("remote third plugin %s start failed, status: %s", pluginName, resp.Status)
		}
		return "", fmt.Errorf("remote third plugin %s start failed, status: %s, body: %s", pluginName, resp.Status, message)
	}
	return accessURL, nil
}

func (m *thirdPluginManage) CloseAll() error {
	var firstErr error
	m.pluginMap.Range(func(key, value any) bool {
		if err := m.Close(key.(string)); err != nil && firstErr == nil && !errors.Is(err, os.ErrNotExist) {
			firstErr = err
		}
		return true
	})
	return firstErr
}

func (m *thirdPluginManage) Stop(code string) {
	pluginModel := plugin.Plugin{Code: code}
	db := uioc.Database()
	if err := db.Read().Where("code = ?", code).First(&pluginModel).Error; err != nil {
		return
	}
	if !m.hasRunHere(pluginModel) {
		return
	}
	_, stopPath := m.scriptPaths(code)
	if !m.common.fileExists(stopPath) {
		m.pluginMap.Delete(code)
		return
	}
	m.execScriptWait(code, stopPath)
	m.pluginMap.Delete(code)
}

func (m *thirdPluginManage) Close(code string) error {
	pluginModel := plugin.Plugin{Code: code}
	db := uioc.Database()
	if err := db.Read().Where("code = ?", code).First(&pluginModel).Error; err != nil {
		return nil
	}
	canRun := m.hasRunHere(pluginModel)
	if !canRun {
		return nil
	}
	_, stopPath := m.scriptPaths(code)
	if !m.common.fileExists(stopPath) {
		m.pluginMap.Delete(code)
		return os.ErrNotExist
	}
	err := m.execScriptWait(code, stopPath)
	m.pluginMap.Delete(code)
	return err
}

func (m *thirdPluginManage) execScript(code, scriptPath string) *exec.Cmd {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", scriptPath)
		if fn := ioc.Ioc().Get(ioc.KeyHideCmdFunc); fn != nil {
			fn.(func(*exec.Cmd))(cmd)
		}
	} else {
		cmd = exec.Command("sh", scriptPath)
	}
	cmd.Dir = m.PluginDir(code)
	return cmd
}

func (m *thirdPluginManage) execScriptWait(code, scriptPath string) error {
	cmd := m.execScript(code, scriptPath)
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}
