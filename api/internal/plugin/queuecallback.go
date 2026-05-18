package plugin

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/fs185085781/v9os/internal/cache"
	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/internal/inface/distributed"
	"github.com/fs185085781/v9os/internal/ioc"
	"github.com/fs185085781/v9os/internal/ioc/uioc"
	"github.com/fs185085781/v9os/internal/logger"
	"github.com/fs185085781/v9os/internal/model/system"
	"github.com/fs185085781/v9os/internal/queue"
	"github.com/fs185085781/v9os/pkg/util"
	"github.com/spf13/cast"
)

type UserQueueCallback struct {
	log logger.Logger
}

type websocketPushResp struct {
	Code int `json:"code"`
	Data struct {
		Delivered bool `json:"delivered"`
	} `json:"data"`
}

func NewUserQueueCallback(log logger.Logger) *UserQueueCallback {
	return &UserQueueCallback{
		log: log,
	}
}
func (q *UserQueueCallback) DeadMsgCheck(info *queue.PushInfo) bool {
	c := uioc.Cache()
	lock := c.CreateLock("mq:dealcheck:" + info.MsgId)
	lock.Lock()
	defer lock.UnLock()
	b, err := c.GetValue("mq:dealcheck-v:" + info.MsgId)
	var current int
	if err != nil {
		if err == cache.CacheIsNil {
			current = 0
		} else {
			current = 5
		}
	} else if len(b) == 8 {
		current = int(binary.LittleEndian.Uint64(b))
	} else {
		current, _ = strconv.Atoi(string(b))
	}
	if current >= 10 {
		//超过10次(5次重试 + 5次存库失败) 硬性丢弃
		jsonMsg, err := json.Marshal(info.Data)
		if err == nil {
			q.log.Error("消息已超过10次处理,硬性丢弃", logger.NewField("plugin", info.Plugin), logger.NewField("url", info.Url), logger.NewField("stype", info.Stype), logger.NewField("msgId", info.MsgId), logger.NewField("data", string(jsonMsg)))
		}
		return true
	}
	if current >= 5 {
		//超过5次,记录到数据库,并返回true永久删除
		jsonMsg, err := json.Marshal(info.Data)
		if err != nil {
			return true
		}
		db := uioc.Database()
		err = db.Create(&system.DeadMsg{
			Plugin: info.Plugin,
			MsgId:  info.MsgId,
			Url:    info.Url,
			Stype:  info.Stype,
			Data:   string(jsonMsg),
		})
		if err == nil {
			return true
		}
	}
	current++
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(current))
	c.SetValue("mq:dealcheck-v:"+info.MsgId, buf, 72*time.Hour)
	return false
}

// pluginCode 代表插件订阅者(用来判断最终由哪台机器执行callback回调),在stype为2,3,4的情况下可空
func (q *UserQueueCallback) queueCallbackCommon(uri, pluginCode string, stype int, data []byte) bool {
	urlPrefixes := make([]string, 0)
	cfg := uioc.Config()
	distributedProvider := uioc.Get[distributed.DistributedProvider](ioc.KeyDistributedProvider)
	switch stype {
	case queue.StypePluginUnicast, queue.StypePluginBroadcast:
		hosts := distributedProvider.Plugins().ResolveMainPlugin(pluginCode, stype == queue.StypePluginBroadcast)
		for _, host := range hosts {
			urlPrefixes = append(urlPrefixes, fmt.Sprintf("http://%s", host))
		}
	case queue.StypeWebsocket:
		var msg WebsocketMessage
		err := json.Unmarshal(data, &msg)
		if err != nil {
			return true
		}
		if msg.To == "" {
			if handlerAny := ioc.Ioc().Get(ioc.KeyChatWebsocketHandler); handlerAny != nil {
				if handler, ok := handlerAny.(WebsocketMessageHandler); ok {
					return handler(msg)
				}
			}
			return true
		}
		hosts := websocketBroadcastHosts(distributedProvider, cfg)
		for _, host := range hosts {
			urlPrefixes = append(urlPrefixes, fmt.Sprintf("http://%s", host))
		}
		if len(urlPrefixes) == 0 {
			return q.saveOfflineWebsocketMessage(msg)
		}
	case queue.StypeRelativeURL:
		urlPrefixes = append(urlPrefixes, fmt.Sprintf("http://127.0.0.1:%d", cfg.Machine().Port))
	case queue.StypeAbsoluteURL:
		urlPrefixes = append(urlPrefixes, "")
	default:
		return true
	}
	header := make(map[string][]string)
	if stype == queue.StypePluginUnicast || stype == queue.StypeWebsocket || stype == queue.StypeRelativeURL || stype == queue.StypePluginBroadcast {
		timeCheck, err := util.EncryptGCM(cast.ToString(util.UnixSeconds()), util.AdjustKey([]byte(cfg.Server().CommunicationKey)))
		if err == nil {
			header["Timecheck"] = []string{timeCheck}
		}
	}
	if stype == queue.StypeWebsocket {
		return q.broadcastWebsocket(uri, urlPrefixes, data, header)
	}
	isOk := true
	for _, urlPrefix := range urlPrefixes {
		url := fmt.Sprintf("%s%s", urlPrefix, uri)
		_, err := util.Post(nil, url, data, header)
		if err != nil {
			isOk = false
		}
	}
	return isOk
}

func websocketBroadcastHosts(distributedProvider distributed.DistributedProvider, cfg config.Config) []string {
	if distributedProvider == nil {
		return []string{fmt.Sprintf("127.0.0.1:%d", cfg.Machine().Port)}
	}
	nodes := distributedProvider.Nodes().All()
	if len(nodes) == 0 {
		if host, ok := distributedProvider.Nodes().Resolve(""); ok && host != "" {
			return []string{host}
		}
		return []string{fmt.Sprintf("127.0.0.1:%d", cfg.Machine().Port)}
	}
	hosts := make([]string, 0, len(nodes))
	for _, host := range nodes {
		if host != "" {
			hosts = append(hosts, host)
		}
	}
	return hosts
}

func (q *UserQueueCallback) broadcastWebsocket(uri string, urlPrefixes []string, data []byte, header map[string][]string) bool {
	var msg WebsocketMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return true
	}
	hasDelivered := false
	hasPostError := false
	for _, urlPrefix := range urlPrefixes {
		url := fmt.Sprintf("%s%s", urlPrefix, uri)
		body, err := util.Post(nil, url, data, header)
		if err != nil {
			hasPostError = true
			continue
		}
		var resp websocketPushResp
		if err := json.Unmarshal(body, &resp); err != nil || resp.Code != 0 {
			hasPostError = true
			continue
		}
		if resp.Data.Delivered {
			hasDelivered = true
		}
	}
	if hasDelivered {
		return true
	}
	if hasPostError {
		return false
	}
	return q.saveOfflineWebsocketMessage(msg)
}

func (q *UserQueueCallback) saveOfflineWebsocketMessage(msg WebsocketMessage) bool {
	db := uioc.Database()
	err := db.Create(&system.OfflineChatMsg{
		From:     msg.From,
		To:       msg.To,
		Msg:      msg.Msg,
		Type:     msg.Type,
		DateTime: msg.DateTime.UnixMilli(),
	})
	return err == nil
}
func (q *UserQueueCallback) CallBack(info *queue.PushInfo) bool {
	str, err := json.Marshal(info.Data)
	if err != nil {
		return true
	}
	return q.queueCallbackCommon(info.Url, info.Plugin, info.Stype, str)
}
