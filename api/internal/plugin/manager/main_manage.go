package manager

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fs185085781/v9os/internal/cache"
	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/internal/inface/distributed"
	"github.com/fs185085781/v9os/internal/ioc"
	"github.com/fs185085781/v9os/internal/ioc/uioc"
	"github.com/fs185085781/v9os/internal/logger"
	"github.com/fs185085781/v9os/internal/model/plugin"
	"github.com/fs185085781/v9os/internal/store"
	"github.com/fs185085781/v9os/pkg/locales"
	"github.com/fs185085781/v9os/pkg/util"
	"github.com/gin-gonic/gin"
)

type pluginInfo struct {
	port        int
	mu          sync.Mutex
	cmd         *exec.Cmd
	runKey      string
	closeDelay  int
	closeTime   time.Time
	needAddTime bool
}
type mainPluginManage struct {
	pluginMap     sync.Map
	serverPort    int
	runPluginMap  sync.Map
	healthyClient *http.Client
	cfg           config.Config
	cache         cache.Cache
	log           logger.Logger
	common        *commonPluginManage
}

func (p *mainPluginManage) pluginExecPath(code string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(p.PluginDir(code), code+".exe")
	}
	return filepath.Join(p.PluginDir(code), code)
}

func (p *mainPluginManage) PluginDir(code string) string {
	return filepath.Join(p.common.pluginRootDir(), "main", code)
}

func (o *commonPluginManage) newMainManage(serverPort int, cfg config.Config, c cache.Cache, log logger.Logger) IPluginManage {
	a := &mainPluginManage{
		serverPort: serverPort,
		healthyClient: &http.Client{
			Timeout: 2 * time.Second,
		},
		cfg:    cfg,
		cache:  c,
		log:    log,
		common: o,
	}
	a.syncClearCloudPluginInfo()
	util.Go(a.checkPluginClose)
	a.log.Println("[main plugin manager] initialized")
	return a
}
func (p *mainPluginManage) checkPluginClose() {
	for {
		time.Sleep(time.Minute)
		p.pluginMap.Range(func(key, value any) bool {
			pluginInfo := value.(*pluginInfo)
			if pluginInfo.closeDelay > 0 {
				if pluginInfo.needAddTime {
					pluginInfo.closeTime = time.Now().Add(time.Duration(pluginInfo.closeDelay) * time.Minute)
					pluginInfo.needAddTime = false
				} else if time.Now().After(pluginInfo.closeTime) {
					p.closePlugin(key.(string), pluginInfo)
				}
			}
			return true
		})
		p.syncPluginInfoCloud()
	}
}

func (p *mainPluginManage) GetPluginName(code, key string) (string, error) {
	if v, ok := p.runPluginMap.Load(key); ok {
		return v.(string), nil
	}
	if code != "" {
		db := uioc.Database()
		var pluginModel plugin.Plugin
		err := db.Read().Where("code = ?", code).First(&pluginModel).Error
		if err == nil && pluginModel.Status == 1 && pluginModel.PluginType == 1 && pluginModel.DebugPort > 0 {
			return code, nil
		}
	}
	return "", nil
}

func (p *mainPluginManage) isPortAvailable(port int) bool {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return false
	}
	listener.Close()
	return true
}
func (p *mainPluginManage) allocatePort() int {
	const maxRetries = 20
	for range maxRetries {
		tmpPort := rand.Intn(10000) + 10000
		if !p.isPortAvailable(tmpPort) {
			continue
		}
		var conflict = false
		p.pluginMap.Range(func(key, value any) bool {
			tmpInfo := value.(*pluginInfo)
			if tmpInfo.port == tmpPort {
				conflict = true
				return false
			}
			return true
		})

		if !conflict {
			return tmpPort
		}
	}
	return 0
}

func mainPluginText(ctx *gin.Context, key string) string {
	if ctx == nil {
		return locales.GetText("en", key)
	}
	return locales.GetTextCtx(ctx, key)
}

func (p *mainPluginManage) PluginHost(pluginName string, ctx *gin.Context) (string, error) {
	info, _ := p.pluginMap.LoadOrStore(pluginName, &pluginInfo{})
	pluginInfo := info.(*pluginInfo)
	if pluginInfo.port != 0 {
		p.syncLocalPackage(pluginName)
		if pluginInfo.closeDelay > 0 {
			pluginInfo.needAddTime = true
		}
		return fmt.Sprintf("http://127.0.0.1:%d", pluginInfo.port), nil
	}
	pluginInfo.mu.Lock()
	defer pluginInfo.mu.Unlock()
	if pluginInfo.port != 0 {
		p.syncLocalPackage(pluginName)
		if pluginInfo.closeDelay > 0 {
			pluginInfo.needAddTime = true
		}
		return fmt.Sprintf("http://127.0.0.1:%d", pluginInfo.port), nil
	}
	db := uioc.Database().Write()
	var pluginModel plugin.Plugin
	err := db.Where("code = ?", pluginName).First(&pluginModel).Error
	if err != nil {
		return "", err
	}
	if pluginModel.Status != 1 {
		return "", fmt.Errorf(mainPluginText(ctx, "common.plugin.disabled"), pluginName)
	}
	if pluginModel.PluginType != 1 {
		return "", fmt.Errorf(mainPluginText(ctx, "common.plugin.typeinvalid"), pluginName)
	}
	if pluginModel.DebugPort > 0 {
		pluginInfo.port = pluginModel.DebugPort
		pluginInfo.closeDelay = 0
		pluginInfo.runKey = pluginName
		p.runPluginMap.Store(pluginInfo.runKey, pluginName)
		return fmt.Sprintf("http://127.0.0.1:%d", pluginModel.DebugPort), nil
	}
	rdir := p.PluginDir(pluginName)
	runPath := p.pluginExecPath(pluginName)
	if _, err = os.Stat(runPath); os.IsNotExist(err) {
		_, err = p.installPackage(&store.AppInfo{
			Code:       pluginName,
			Version:    pluginModel.Version,
			PluginType: 1,
		}, nil)
		if err != nil {
			return "", err
		}
		runPath = p.pluginExecPath(pluginName)
	} else if err != nil {
		return "", err
	}
	if runtime.GOOS != "windows" {
		if err2 := os.Chmod(runPath, 0755); err2 != nil {
			p.log.Error("set plugin executable permission failed", logger.NewField("runPath", runPath), logger.NewField("err", err2))
			return "", err2
		}
	}
	tmpPort := p.allocatePort()
	if tmpPort == 0 {
		return "", fmt.Errorf("%s", mainPluginText(ctx, "common.plugin.allocateportfailed"))
	}
	pluginInfo.runKey = util.UUID()
	p.runPluginMap.Store(pluginInfo.runKey, pluginName)
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("./"+pluginName+".exe", strconv.Itoa(tmpPort), strconv.Itoa(p.serverPort), pluginInfo.runKey)
		fn := ioc.Ioc().Get(ioc.KeyHideCmdFunc)
		if fn != nil {
			fn.(func(*exec.Cmd))(cmd)
		}
	default: // linux,macos,android
		cmd = exec.Command("./"+pluginName, strconv.Itoa(tmpPort), strconv.Itoa(p.serverPort), pluginInfo.runKey)
	}
	cmd.Dir = rdir
	err = cmd.Start()
	if err != nil {
		p.log.Error("start plugin failed", logger.NewField("pluginName", pluginName), logger.NewField("err", err))
		return "", err
	}
	var status int32 = 0
	util.Go(func() {
		cmd.Wait()
		atomic.StoreInt32(&status, 1)
		p.Close(pluginName)
	})
	pluginInfo.cmd = cmd
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-timeout:
			cmd.Process.Kill()
			return "", fmt.Errorf(mainPluginText(ctx, "common.plugin.starttimeout"), pluginName)
		case <-ticker.C:
			if atomic.LoadInt32(&status) == 1 {
				return "", fmt.Errorf(mainPluginText(ctx, "common.plugin.startfailedornotfound"), pluginName)
			}
			if p.healthy(tmpPort, false) {
				pluginInfo.port = tmpPort
				pluginInfo.closeDelay = pluginModel.CloseDelay
				if pluginInfo.closeDelay > 0 {
					pluginInfo.needAddTime = true
				}
				p.syncPluginInfoCloud()
				return fmt.Sprintf("http://127.0.0.1:%d", tmpPort), nil
			}
		}
	}
}

func (p *mainPluginManage) syncClearCloudPluginInfo() {
	ioc.Ioc().Get(ioc.KeyDistributedProvider).(distributed.DistributedProvider).Plugins().SyncLocalMainPlugins(nil)
}

func (p *mainPluginManage) syncPluginInfoCloud() {
	pluginCodes := make([]string, 0)
	p.pluginMap.Range(func(key, value any) bool {
		pluginCodes = append(pluginCodes, key.(string))
		return true
	})
	ioc.Ioc().Get(ioc.KeyDistributedProvider).(distributed.DistributedProvider).Plugins().SyncLocalMainPlugins(pluginCodes)
}
func (p *mainPluginManage) healthy(port int, close bool) bool {
	url := fmt.Sprintf("http://127.0.0.1:%d/healthy?close=%t", port, close)
	return util.GetHealthy(url)
}

func (p *mainPluginManage) closeAll() {
	p.pluginMap.Range(func(key, value any) bool {
		pluginInfo := value.(*pluginInfo)
		p.closePlugin(key.(string), pluginInfo)
		return true
	})
	p.syncPluginInfoCloud()
}

func (p *mainPluginManage) CloseAll() error {
	p.closeAll()
	return nil
}

func (p *mainPluginManage) Close(pluginName string) error {
	v, ok := p.pluginMap.Load(pluginName)
	if !ok {
		return nil
	}
	pluginInfo := v.(*pluginInfo)
	p.closePlugin(pluginName, pluginInfo)
	p.syncPluginInfoCloud()
	return nil
}

func (p *mainPluginManage) Stop(pluginName string) {
	v, ok := p.pluginMap.Load(pluginName)
	if !ok {
		return
	}
	info := v.(*pluginInfo)
	p.closePluginByDel(pluginName, info, false)
}
func (p *mainPluginManage) closePluginByDel(key string, info *pluginInfo, delInfo bool) {
	info.mu.Lock()
	defer info.mu.Unlock()
	flag := p.healthy(info.port, true)
	if flag {
		time.Sleep(1 * time.Second)
	}
	if info.cmd != nil {
		info.cmd.Process.Kill()
	}
	if delInfo {
		p.runPluginMap.Delete(info.runKey)
		p.pluginMap.Delete(key)
	}
}
func (p *mainPluginManage) closePlugin(key string, info *pluginInfo) {
	p.closePluginByDel(key, info, true)
}

func (p *mainPluginManage) installPackage(appInfo *store.AppInfo, progress func(int, string) bool) (*packageManifest, error) {
	result, err := p.common.installPluginPackage(p.common.resolvePackageURL(appInfo), p.PluginDir(appInfo.Code), 1, appInfo.Code, p.cfg, p.log, progress)
	if err != nil {
		return nil, err
	}
	return result.Manifest, nil
}

func (p *mainPluginManage) syncLocalPackage(pluginName string) {
	provider := ioc.Ioc().Get(ioc.KeyDistributedProvider)
	if provider == nil {
		return
	}
	plugins := provider.(distributed.DistributedProvider).Plugins()
	plugins.SyncLocalPluginPackage(pluginName, 1)
}

func (p *mainPluginManage) Install(appInfo *store.AppInfo, opts InstallOptions) (*plugin.Plugin, error) {
	if opts.Upgrade {
		p.Close(appInfo.Code)
	}
	manifest, err := p.installPackage(appInfo, opts.Progress)
	if err != nil {
		return nil, err
	}
	pluginModel := p.common.buildPluginModel(appInfo, manifest)
	p.common.snapshotPluginIcon(&pluginModel, p.PluginDir(pluginModel.Code))
	if err := p.common.upsertPluginModel(&pluginModel); err != nil {
		return nil, err
	}
	if pluginModel.Status == 1 {
		if _, err := p.PluginHost(pluginModel.Code, nil); err != nil {
			return nil, err
		}
	}
	return &pluginModel, nil
}

func (p *mainPluginManage) InstallLocalPackage(zipPath string, opts InstallOptions) (*plugin.Plugin, error) {
	result, err := p.common.installLocalPluginPackage(zipPath, p.PluginDir, 1, "", p.cfg)
	if err != nil {
		return nil, err
	}
	pluginModel := result.Manifest.toPluginModel()
	p.common.snapshotPluginIcon(&pluginModel, p.PluginDir(pluginModel.Code))
	if err := p.common.upsertPluginModel(&pluginModel); err != nil {
		return nil, err
	}
	if pluginModel.Status == 1 {
		if _, err := p.PluginHost(pluginModel.Code, nil); err != nil {
			return nil, err
		}
	}
	return &pluginModel, nil
}

func (p *mainPluginManage) Uninstall(pluginModel plugin.Plugin) error {
	code := pluginModel.Code
	q := uioc.Queue()
	q.UnsubscribeByPlugin(code)
	p.Close(code)
	if err := os.RemoveAll(p.PluginDir(code)); err != nil {
		return err
	}
	return p.common.deletePluginModel(code)
}
