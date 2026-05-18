package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fs185085781/v9os/internal/cache"
	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/internal/logger"
	"github.com/fs185085781/v9os/pkg/util"
	"github.com/redis/go-redis/v9"
)

type redisQueue struct {
	client     *redis.Client
	cache      cache.Cache
	cfg        *config.RedisQueueConfig
	log        logger.Logger
	ctx        context.Context
	cancel     context.CancelFunc
	groupName  string
	consumerID string
	subs       *subscribesCache
}

// newRedisQueue 创建Redis队列实例
func newRedisQueue(cfg *config.RedisQueueConfig, callback QueueCallback, c cache.Cache, tmpLog logger.Logger) (Queue, error) {
	q := &redisQueue{
		cfg:        cfg,
		cache:      c,
		log:        tmpLog,
		groupName:  "v9os",
		consumerID: util.UUID(),
		subs:       newSubscribesCache(c, "mq:subscribes:redis"),
	}
	isError := false
	defer func() {
		if isError {
			q.Close()
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	q.ctx = ctx
	q.cancel = cancel
	// 初始化Redis客户端
	if err := q.initRedisClient(); err != nil {
		isError = true
		return nil, err
	}

	// 创建消费者组
	if err := q.createConsumerGroup(); err != nil {
		isError = true
		return nil, err
	}
	// 启动消费循环
	util.Go(func() {
		q.consumeLoop(callback)
	})
	// 定期同步订阅关系(分布式场景下其他机器的变更)
	util.Go(func() {
		q.syncLoop()
	})
	q.log.Println("[Redis队列]已初始化")
	return q, nil
}

// syncLoop 定期检查订阅关系版本号,有变化则重新加载
func (q *redisQueue) syncLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-q.ctx.Done():
			return
		case <-ticker.C:
			q.subs.SyncIfNeeded()
		}
	}
}

// initRedisClient 初始化Redis客户端
func (q *redisQueue) initRedisClient() error {
	options := &redis.Options{
		Addr:     q.cfg.Addr,
		Password: q.cfg.Password,
		DB:       q.cfg.DB,
	}

	if q.cfg.PoolSize > 0 {
		options.PoolSize = q.cfg.PoolSize
	}
	q.client = redis.NewClient(options)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return q.client.Ping(ctx).Err()
}

// createConsumerGroup 创建消费者组
func (q *redisQueue) createConsumerGroup() error {
	topic := q.cfg.Topic

	// 尝试创建消费者组，如果已存在则忽略错误
	err := q.client.XGroupCreateMkStream(q.ctx, topic, q.groupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("创建消费者组失败: %w", err)
	}
	return nil
}

// Publish 发布消息到Redis Stream
func (q *redisQueue) Publish(msg *Message) error {
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
	topic := q.cfg.Topic
	// 使用Redis Streams发布消息
	_, err = q.client.XAdd(q.ctx, &redis.XAddArgs{
		Stream: topic,
		Values: map[string]interface{}{
			"data": string(data),
		},
	}).Result()
	if err != nil {
		return err
	}
	q.log.Debug("Redis队列已发布消息",
		logger.NewField("eventType", msg.EventType),
		logger.NewField("topic", topic))
	return nil
}

// Subscribe 订阅消息
func (q *redisQueue) Subscribe(eventType, callback, plugin string, stype int) error {
	q.log.Debug("Redis队列订阅",
		logger.NewField("eventType", eventType),
		logger.NewField("callback", callback))

	lock := q.cache.CreateLock("mq:subscribe:unsubscribe")
	lock.Lock()
	defer lock.UnLock()

	return q.subs.Update(func(mapData map[string]map[string]*SubscribeInfo) {
		if _, exists := mapData[eventType]; !exists {
			mapData[eventType] = make(map[string]*SubscribeInfo)
		}
		mapData[eventType][callback] = &SubscribeInfo{
			Plugin: plugin,
			Stype:  stype,
			Url:    callback,
		}
	})
}

// Unsubscribe 取消订阅
func (q *redisQueue) Unsubscribe(eventType, callback string) error {
	q.log.Debug("Redis队列取消订阅",
		logger.NewField("eventType", eventType),
		logger.NewField("callback", callback))

	lock := q.cache.CreateLock("mq:subscribe:unsubscribe")
	lock.Lock()
	defer lock.UnLock()

	return q.subs.Update(func(mapData map[string]map[string]*SubscribeInfo) {
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

// consumeLoop 消费循环 - 基于Redis Streams和消费者组
func (q *redisQueue) consumeLoop(callback QueueCallback) {
	topic := q.cfg.Topic
	q.log.Info("Redis队列开始消费",
		logger.NewField("topic", topic),
		logger.NewField("group", q.groupName))

	for {
		select {
		case <-q.ctx.Done():
			q.log.Info("Redis队列消费循环退出")
			return
		default:
			// 使用消费者组读取消息，确保消息只被一个消费者处理
			results, err := q.client.XReadGroup(q.ctx, &redis.XReadGroupArgs{
				Group:    q.groupName,
				Consumer: q.consumerID,
				Streams:  []string{topic, ">"},
				Count:    10,
				Block:    time.Second * 5,
				NoAck:    false,
			}).Result()

			if err != nil {
				if err == redis.Nil {
					// 超时，继续循环
					continue
				}
				if strings.Contains(err.Error(), "NOGROUP") {
					q.createConsumerGroup()
					time.Sleep(time.Second)
					continue
				}
				q.log.Error("读取消息失败", logger.NewField("error", err))
				time.Sleep(time.Second)
				continue
			}
			// 处理消息
			for _, stream := range results {
				for _, message := range stream.Messages {
					isOk := q.processMessage(message, callback)
					if isOk {
						q.client.XAck(q.ctx, topic, q.groupName, message.ID)
					}
				}
			}
		}
	}
}

// processMessage 处理接收到的消息
func (q *redisQueue) processMessage(message redis.XMessage, callback QueueCallback) bool {
	data, exists := message.Values["data"]
	if !exists {
		q.log.Error("消息格式错误，缺少data字段")
		return true
	}
	dataStr, ok := data.(string)
	if !ok {
		q.log.Error("消息data字段类型错误")
		return true
	}

	var msg Message
	if err := json.Unmarshal([]byte(dataStr), &msg); err != nil {
		q.log.Error("消息反序列化失败", logger.NewField("error", err))
		return true
	}
	// 从本地缓存获取订阅关系
	mapData := q.subs.Get()
	urls, exists := mapData[msg.EventType]
	if !exists {
		q.log.Debug("没有找到消息的订阅者", logger.NewField("eventType", msg.EventType))
		return true
	}
	// 发送给所有订阅者（基于共享的订阅关系）
	isOk := true
	hasOk := false
	failUrls := make([]*PushInfo, 0)
	var lastFailUrls []*PushInfo
	partiallyId := fmt.Sprintf("mq:partially:%s", msg.MsgId)
	q.cache.GetObject(partiallyId, &lastFailUrls)
	for _, url := range urls {
		needXf := true
		if len(lastFailUrls) > 0 {
			needXf = false
			for _, lastUrl := range lastFailUrls {
				if lastUrl.Url == url.Url && lastUrl.Stype == url.Stype && lastUrl.Plugin == url.Plugin {
					needXf = true
					break
				}
			}
		}
		if !needXf {
			continue
		}
		info := &PushInfo{
			Plugin: url.Plugin,
			Stype:  url.Stype,
			Url:    url.Url,
			Data:   msg.Data,
			MsgId:  msg.MsgId,
		}
		tmpOk := callback.CallBack(info)
		if !tmpOk {
			isOk = false
			failUrls = append(failUrls, info)
		} else {
			hasOk = true
		}
	}
	if !isOk {
		if hasOk {
			q.cache.SetObject(partiallyId, failUrls, 72*time.Hour)
		} else {
			q.cache.RemoveValue(partiallyId)
		}
		isOk = callback.DeadMsgCheck(failUrls[0])
	}
	return isOk
}

// Close 关闭Redis队列
func (q *redisQueue) Close() error {
	q.log.Info("正在关闭Redis队列...")

	if q.cancel != nil {
		q.cancel()
	}

	if q.client != nil {
		if err := q.client.Close(); err != nil {
			q.log.Error("关闭Redis客户端失败", logger.NewField("error", err))
		}
	}

	q.log.Println("[Redis队列]已关闭")
	return nil
}

// UnsubscribeByPlugin 按插件名清理所有订阅
func (q *redisQueue) UnsubscribeByPlugin(pluginName string) error {
	lock := q.cache.CreateLock("mq:subscribe:unsubscribe")
	lock.Lock()
	defer lock.UnLock()

	return q.subs.Update(func(mapData map[string]map[string]*SubscribeInfo) {
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

func (q *redisQueue) SupportDistributed() bool {
	return true
}
