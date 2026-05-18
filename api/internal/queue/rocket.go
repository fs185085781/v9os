package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/apache/rocketmq-client-go/v2/rlog"
	cache2 "github.com/fs185085781/v9os/internal/cache"
	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/internal/logger"
	"github.com/fs185085781/v9os/pkg/util"
)

type rocketQueue struct {
	producer rocketmq.Producer
	consumer rocketmq.PushConsumer
	cache    cache2.Cache
	cfg      *config.RocketQueueConfig
	log      logger.Logger
	subs     *subscribesCache
	stopChan chan struct{}
}

type rocketLog struct {
	log logger.Logger
}

func (r *rocketLog) Debug(msg string, fields map[string]interface{}) {
	r.Println(logger.DebugLevel, msg, fields)
}
func (r *rocketLog) Error(msg string, fields map[string]interface{}) {
	r.Println(logger.ErrorLevel, msg, fields)
}
func (r *rocketLog) Fatal(msg string, fields map[string]interface{}) {
	r.Println(logger.ErrorLevel, msg, fields)
}
func (r *rocketLog) Info(msg string, fields map[string]interface{}) {
	r.Println(logger.InfoLevel, msg, fields)
}
func (r *rocketLog) Level(level string) {

}

func (r *rocketLog) OutputPath(path string) (err error) {
	return nil
}

// Warning implements rlog.Logger.
func (r *rocketLog) Warning(msg string, fields map[string]interface{}) {
	r.Println(logger.WarnLevel, msg, fields)
}

func (r *rocketLog) Println(lvl logger.Level, msg string, fields map[string]interface{}) {
	vs := make([]logger.Field, 0)
	for k, v := range fields {
		vs = append(vs, logger.Field{Key: k, Value: v})
	}
	r.log.Log(lvl, msg, vs...)
}

func newRocketQueue(cfg *config.RocketQueueConfig, callback QueueCallback, c cache2.Cache, tmpLog logger.Logger) (Queue, error) {
	rlog.SetLogger(&rocketLog{log: tmpLog})
	store := &rocketQueue{
		cache:    c,
		cfg:      cfg,
		log:      tmpLog,
		subs:     newSubscribesCache(c, "mq:subscribes:rocket"),
		stopChan: make(chan struct{}),
	}
	isError := false
	defer func() {
		if isError {
			store.Close()
		}
	}()
	// 初始化生产者
	if err := store.initProducer(); err != nil {
		isError = true
		return nil, fmt.Errorf("RocketMQ生产者初始化失败: %w", err)
	}
	// 初始化消费者
	if err := store.initConsumer(); err != nil {
		isError = true
		return nil, fmt.Errorf("RocketMQ消费者初始化失败: %w", err)
	}
	//启动消费
	util.Go(func() { store.consume(callback) })
	// 定期同步订阅关系
	util.Go(func() { store.syncLoop() })
	store.log.Println("[RocketMQ队列]已初始化")
	return store, nil
}

// syncLoop 定期检查订阅关系版本号,有变化则重新加载
func (s *rocketQueue) syncLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.subs.SyncIfNeeded()
		}
	}
}

// Close implements Queue.
func (s *rocketQueue) Close() error {
	select {
	case <-s.stopChan:
	default:
		close(s.stopChan)
	}
	if s.consumer != nil {
		if err := s.consumer.Shutdown(); err != nil {
		} else {
		}
		s.consumer = nil
	}
	if s.producer != nil {
		if err := s.producer.Shutdown(); err != nil {
		} else {
		}
		s.producer = nil
	}
	s.log.Println("[RocketMQ队列]已关闭")
	return nil
}

// Publish implements Queue.
func (s *rocketQueue) Publish(msg *Message) error {
	if msg == nil {
		return errors.New("消息不可为空")
	}
	if msg.MsgId == "" {
		msg.MsgId = util.UUID()
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	msg2 := primitive.NewMessage(s.cfg.Topic, data)
	_, err = s.producer.SendSync(context.Background(), msg2)
	if err != nil {
		return err
	}
	return nil
}

// Subscribe implements Queue.
func (s *rocketQueue) Subscribe(eventType, callback, plugin string, stype int) error {
	lock := s.cache.CreateLock("mq:subscribe:unsubscribe")
	lock.Lock()
	defer lock.UnLock()

	return s.subs.Update(func(mapData map[string]map[string]*SubscribeInfo) {
		if mapData[eventType] == nil {
			mapData[eventType] = make(map[string]*SubscribeInfo)
		}
		mapData[eventType][callback] = &SubscribeInfo{
			Plugin: plugin,
			Stype:  stype,
			Url:    callback,
		}
	})
}

// Unsubscribe implements Queue.
func (s *rocketQueue) Unsubscribe(eventType, callback string) error {
	lock := s.cache.CreateLock("mq:subscribe:unsubscribe")
	lock.Lock()
	defer lock.UnLock()

	return s.subs.Update(func(mapData map[string]map[string]*SubscribeInfo) {
		urls, exists := mapData[eventType]
		if !exists {
			return
		}
		delete(urls, callback)
		if len(urls) == 0 {
			delete(mapData, eventType)
		}
	})
}

func (s *rocketQueue) initProducer() error {
	opts := []producer.Option{
		producer.WithNameServer(s.cfg.Addrs),
		producer.WithGroupName("v9os"),
	}
	// 动态添加认证
	if s.cfg.AccessKey != "" && s.cfg.Secret != "" {
		opts = append(opts, producer.WithCredentials(primitive.Credentials{
			AccessKey: s.cfg.AccessKey,
			SecretKey: s.cfg.Secret,
		}))
	}
	p, err := rocketmq.NewProducer(opts...)
	if err != nil {
		return err
	}
	if err := p.Start(); err != nil {
		return err
	}
	if err := s.validateNameServer(s.cfg.Addrs); err != nil {
		return fmt.Errorf("地址无效:%w", err)
	}
	s.producer = p
	return nil
}

func (s *rocketQueue) initConsumer() error {
	opts := []consumer.Option{
		consumer.WithNameServer(s.cfg.Addrs),
		consumer.WithConsumerModel(consumer.Clustering),
		consumer.WithGroupName("v9os"),
		consumer.WithConsumeFromWhere(consumer.ConsumeFromFirstOffset),
		consumer.WithConsumeMessageBatchMaxSize(10),
	}
	// 动态添加认证
	if s.cfg.AccessKey != "" && s.cfg.Secret != "" {
		opts = append(opts, consumer.WithCredentials(primitive.Credentials{
			AccessKey: s.cfg.AccessKey,
			SecretKey: s.cfg.Secret,
		}))
	}
	c, err := rocketmq.NewPushConsumer(opts...)
	if err != nil {
		return err
	}
	if err := c.Start(); err != nil {
		return err
	}
	s.consumer = c
	return nil
}

func (s *rocketQueue) validateNameServer(addrs []string) error {
	var err2 error
	isOk := false
	for _, a := range addrs {
		conn, err := net.DialTimeout("tcp", a, 3*time.Second)
		if err == nil {
			conn.Close()
			isOk = true
		} else {
			err2 = err
		}
	}
	if isOk {
		return nil
	}
	return err2
}

func (s *rocketQueue) consume(callback QueueCallback) {
	s.consumer.Subscribe(s.cfg.Topic, consumer.MessageSelector{}, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		mapData := s.subs.Get()
		if mapData == nil {
			return consumer.ConsumeSuccess, nil
		}
		allOk := true
		// 按MsgId收集每条消息的首个失败info,用于逐一DeadMsgCheck
		failByMsg := make(map[string]*PushInfo)
		for _, msg := range msgs {
			var data Message
			if err := json.Unmarshal(msg.Body, &data); err != nil {
				continue
			}
			urls, exists := mapData[data.EventType]
			if !exists {
				continue
			}
			urlList := make([]*SubscribeInfo, 0, len(urls))
			for _, url := range urls {
				urlList = append(urlList, url)
			}
			// 每条消息独立处理 partially 逻辑
			var lastFailUrls []*PushInfo
			partiallyId := fmt.Sprintf("mq:partially:%s", data.MsgId)
			s.cache.GetObject(partiallyId, &lastFailUrls)
			msgOk := true
			hasOk := false
			failUrls := make([]*PushInfo, 0)
			for _, url := range urlList {
				needSend := true
				if len(lastFailUrls) > 0 {
					needSend = false
					for _, last := range lastFailUrls {
						if last.Url == url.Url && last.Stype == url.Stype && last.Plugin == url.Plugin {
							needSend = true
							break
						}
					}
				}
				if !needSend {
					continue
				}
				info := &PushInfo{
					Plugin: url.Plugin,
					Stype:  url.Stype,
					Url:    url.Url,
					Data:   data.Data,
					MsgId:  data.MsgId,
				}
				tmpOk := callback.CallBack(info)
				if !tmpOk {
					msgOk = false
					failUrls = append(failUrls, info)
				} else {
					hasOk = true
				}
			}
			if !msgOk {
				if hasOk {
					s.cache.SetObject(partiallyId, failUrls, 72*time.Hour)
				} else {
					s.cache.RemoveValue(partiallyId)
				}
				allOk = false
				if _, exists := failByMsg[data.MsgId]; !exists {
					failByMsg[data.MsgId] = failUrls[0]
				}
			}
		}
		if !allOk {
			allOk = true
			for _, info := range failByMsg {
				if !callback.DeadMsgCheck(info) {
					allOk = false
				}
			}
		}
		if allOk {
			return consumer.ConsumeSuccess, nil
		} else {
			return consumer.ConsumeRetryLater, nil
		}
	})
}

// UnsubscribeByPlugin 按插件名清理所有订阅
func (s *rocketQueue) UnsubscribeByPlugin(pluginName string) error {
	lock := s.cache.CreateLock("mq:subscribe:unsubscribe")
	lock.Lock()
	defer lock.UnLock()

	return s.subs.Update(func(mapData map[string]map[string]*SubscribeInfo) {
		for eventType, urls := range mapData {
			for key, info := range urls {
				if info.Plugin == pluginName {
					delete(urls, key)
				}
			}
			if len(urls) == 0 {
				delete(mapData, eventType)
			}
		}
	})
}

func (s *rocketQueue) SupportDistributed() bool {
	return true
}
