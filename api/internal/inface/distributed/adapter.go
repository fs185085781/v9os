package distributed

import (
	"github.com/fs185085781/v9os/internal/cache"
	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/internal/database"
	"github.com/fs185085781/v9os/internal/logger"
	"github.com/fs185085781/v9os/internal/queue"
	"github.com/gin-gonic/gin"
)

type RuntimeContext struct {
	Config   config.Config
	Database database.Database
	Cache    cache.Cache
	Queue    queue.Queue
	Log      logger.Logger
}

type DistributedProvider interface {
	Init(ctx RuntimeContext) error
	Enabled() bool
	SupportDistributed() bool
	ValidateRuntime() error
	Start() error
	Close() error
	Nodes() NodeRegistry
	Plugins() PluginRegistry
	Websockets() WebsocketRegistry
	Affinity() AffinityRouter
}

type NodeRegistry interface {
	LocalMachineID() string
	LocalIp() string
	Resolve(machineID string) (string, bool)
	All() map[string]string
	Info(machineID string) (MachineInfo, bool)
	ProxyByMachineWhitelist(ctx *gin.Context, pluginCode string) bool
	GetNodesAuth() int64
	GetMachineIds(pluginCode string) []string
}

type MachineInfo struct {
	MachineID string
	OS        string
	Arch      string
}

type PluginRegistry interface {
	SyncLocalMainPlugins(pluginCodes []string)
	ResolveMainPlugin(pluginCode string, broadcast bool) []string
	SyncLocalPluginPackage(code string, pluginType int)
}

type WebsocketRegistry interface {
	SyncLocalUsers(userIDs []string)
	ResolveUser(userID string) []string
}

type AffinityRouter interface {
	ResolveAffinity(distributedID string) (targetHost string, local bool, err error)
}
