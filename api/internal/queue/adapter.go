package queue

import (
	"encoding/binary"
	"sync"
	"time"

	"github.com/fs185085781/v9os/internal/cache"
)

// 订阅类型常量
const (
	StypePluginUnicast   = 1 // 插件订阅(单播,使用plugin:sync:map判断所在机器)
	StypeWebsocket       = 2 // WebSocket订阅
	StypeRelativeURL     = 3 // 相对地址订阅
	StypeAbsoluteURL     = 4 // 绝对地址订阅
	StypePluginBroadcast = 5 // 插件订阅(广播模式)
)

// pluginCode 代表插件订阅者(用来判断最终由哪台机器执行callback回调),在stype为2,3,4的情况下可空
type Queue interface {
	Publish(message *Message) error
	Subscribe(eventType, callback, plugin string, stype int) error
	Unsubscribe(eventType, callback string) error
	UnsubscribeByPlugin(plugin string) error
	Close() error
	SupportDistributed() bool
}

type QueueCallback interface {
	CallBack(info *PushInfo) bool
	DeadMsgCheck(info *PushInfo) bool
}

type Message struct {
	EventType string
	Data      interface{}
	MsgId     string
}

type SubscribeInfo struct {
	Plugin string
	Stype  int
	Url    string
}

type PushInfo struct {
	Plugin string
	Stype  int
	Url    string
	Data   interface{}
	MsgId  string
}

// subscribesCache 订阅关系本地缓存,避免每条消息都从分布式cache反序列化
type subscribesCache struct {
	mu       sync.RWMutex
	local    map[string]map[string]*SubscribeInfo // 本地副本
	version  int64                                // 本地版本号
	cache    cache.Cache
	cacheKey string // 订阅关系在cache中的key,如 mq:subscribes:redis
	verKey   string // 版本号在cache中的key,如 mq:subscribes:redis:ver
}

func newSubscribesCache(c cache.Cache, cacheKey string) *subscribesCache {
	sc := &subscribesCache{
		local:    make(map[string]map[string]*SubscribeInfo),
		cache:    c,
		cacheKey: cacheKey,
		verKey:   cacheKey + ":ver",
	}
	// 启动时从cache加载
	sc.forceReload()
	return sc
}

// Get 获取订阅关系(读本地缓存)
func (sc *subscribesCache) Get() map[string]map[string]*SubscribeInfo {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.local
}

// Update 更新订阅关系(同时写本地和cache),调用方需自行加分布式锁
func (sc *subscribesCache) Update(fn func(mapData map[string]map[string]*SubscribeInfo)) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	// 从cache读最新(分布式锁内,保证一致)
	var mapData map[string]map[string]*SubscribeInfo
	has, _ := sc.cache.GetObjectRetry(sc.cacheKey, &mapData)
	if !has {
		mapData = make(map[string]map[string]*SubscribeInfo)
	}
	fn(mapData)
	if err := sc.cache.SetObjectRetry(sc.cacheKey, mapData, time.Hour*9999); err != nil {
		return err
	}
	// 递增版本号
	sc.version++
	ver := make([]byte, 8)
	binary.LittleEndian.PutUint64(ver, uint64(sc.version))
	sc.cache.SetValue(sc.verKey, ver, time.Hour*9999)
	// 更新本地
	sc.local = mapData
	return nil
}

// SyncIfNeeded 检查版本号,有变化则重新加载(消费循环定期调用)
func (sc *subscribesCache) SyncIfNeeded() {
	remoteVer := sc.getRemoteVersion()
	sc.mu.RLock()
	localVer := sc.version
	sc.mu.RUnlock()
	if remoteVer != localVer {
		sc.forceReload()
	}
}

func (sc *subscribesCache) forceReload() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	var mapData map[string]map[string]*SubscribeInfo
	has, _ := sc.cache.GetObjectRetry(sc.cacheKey, &mapData)
	if !has {
		mapData = make(map[string]map[string]*SubscribeInfo)
	}
	sc.local = mapData
	sc.version = sc.getRemoteVersion()
}

func (sc *subscribesCache) getRemoteVersion() int64 {
	b, err := sc.cache.GetValue(sc.verKey)
	if err != nil || len(b) != 8 {
		return 0
	}
	return int64(binary.LittleEndian.Uint64(b))
}
