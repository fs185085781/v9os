package distributed

import (
	"fmt"
	"runtime"

	"github.com/fs185085781/v9os/internal/ioc"
)

type DefaultProvider struct {
	ctx RuntimeContext
}

func NewDefaultProvider() *DefaultProvider {
	return &DefaultProvider{}
}

func (d *DefaultProvider) Init(ctx RuntimeContext) error {
	d.ctx = ctx
	return nil
}

func (d *DefaultProvider) Enabled() bool {
	return false
}

func (d *DefaultProvider) SupportDistributed() bool {
	return false
}

func (d *DefaultProvider) ValidateRuntime() error {
	return nil
}

func (d *DefaultProvider) Start() error {
	return nil
}

func (d *DefaultProvider) Close() error {
	return nil
}

func (d *DefaultProvider) Nodes() NodeRegistry {
	return d
}

func (d *DefaultProvider) Plugins() PluginRegistry {
	return d
}

func (d *DefaultProvider) Websockets() WebsocketRegistry {
	return d
}

func (d *DefaultProvider) Affinity() AffinityRouter {
	return d
}

func (d *DefaultProvider) LocalMachineID() string {
	if d.ctx.Config == nil {
		return ""
	}
	return d.ctx.Config.Machine().MachineId
}

func (d *DefaultProvider) LocalIp() string {
	if d.ctx.Config == nil {
		return ""
	}
	loadIp := d.ctx.Config.Machine().LoadIp
	if loadIp != "" {
		return loadIp
	}
	return "127.0.0.1"
}

func (d *DefaultProvider) Resolve(machineID string) (string, bool) {
	if machineID == "" || machineID == d.LocalMachineID() {
		return fmt.Sprintf("%s:%d", d.LocalIp(), d.ctx.Config.Machine().Port), true
	}
	return "", false
}

func (d *DefaultProvider) All() map[string]string {
	machineID := d.LocalMachineID()
	if machineID == "" {
		return map[string]string{}
	}
	return map[string]string{machineID: fmt.Sprintf("%s:%d", d.LocalIp(), d.ctx.Config.Machine().Port)}
}

func (d *DefaultProvider) Info(machineID string) (MachineInfo, bool) {
	localMachineID := d.LocalMachineID()
	if machineID != "" && machineID != localMachineID {
		return MachineInfo{}, false
	}
	return MachineInfo{
		MachineID: localMachineID,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}, true
}

func (d *DefaultProvider) SyncLocalMainPlugins(pluginCodes []string) {
}

func (d *DefaultProvider) ResolveMainPlugin(pluginCode string, broadcast bool) []string {
	return []string{fmt.Sprintf("%s:%d", d.LocalIp(), d.ctx.Config.Machine().Port)}
}

func (d *DefaultProvider) SyncLocalPluginPackage(code string, pluginType int) {
}

func (d *DefaultProvider) SyncLocalUsers(userIDs []string) {
}

func (d *DefaultProvider) ResolveUser(userID string) []string {
	if userID == "" {
		return nil
	}
	if resolver, ok := ioc.Ioc().Get(ioc.KeyWebsocketUserResolver).(func(string) bool); ok && !resolver(userID) {
		return nil
	}
	return []string{fmt.Sprintf("%s:%d", d.LocalIp(), d.ctx.Config.Machine().Port)}
}

func (d *DefaultProvider) ResolveAffinity(distributedID string) (string, bool, error) {
	return "", true, nil
}

func initProvider() {
	if ioc.Ioc().Get(ioc.KeyDistributedProvider) != nil {
		return
	}
	ioc.Ioc().Register(ioc.KeyDistributedProvider, NewDefaultProvider())
}

func init() {
	initProvider()
}

var _ DistributedProvider = (*DefaultProvider)(nil)
var _ NodeRegistry = (*DefaultProvider)(nil)
var _ PluginRegistry = (*DefaultProvider)(nil)
var _ WebsocketRegistry = (*DefaultProvider)(nil)
var _ AffinityRouter = (*DefaultProvider)(nil)
