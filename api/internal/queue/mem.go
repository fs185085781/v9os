package queue

import (
	"errors"
	"fmt"
	"sync"
	"time"

	cache2 "github.com/fs185085781/v9os/internal/cache"
	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/internal/logger"
	"github.com/fs185085781/v9os/pkg/util"
)

type memQueue struct {
	cache         cache2.Cache
	queueLock     sync.Mutex                           // 队列专用锁
	eventTypeLock sync.RWMutex                         // 订阅关系专用读写锁
	notEmpty      *sync.Cond                           // 队列不为空的条件
	notFull       *sync.Cond                           // 队列未满的条件
	items         []*Message                           // 存储数据的切片
	capacity      int                                  // 队列容量
	size          int                                  // 当前队列元素数量
	eventMap      map[string]map[string]*SubscribeInfo // 订阅关系映射
	consuming     bool                                 // 消费状态标志
	stopChan      chan struct{}                        // 停止信号通道
	log           logger.Logger
}

// Close implements Queue.
func (s *memQueue) Close() error {
	s.queueLock.Lock()
	defer s.queueLock.Unlock()
	// 停止消费循环
	if s.consuming {
		close(s.stopChan)
		s.consuming = false
	}
	// 唤醒所有等待的协程
	s.notEmpty.Broadcast()
	s.notFull.Broadcast()
	// 清理资源
	s.items = nil
	s.size = 0
	s.eventMap = nil
	s.log.Println("[文件队列]已关闭")
	return nil
}

// Publish implements Queue.
func (q *memQueue) Publish(msg *Message) error {
	if msg == nil {
		return errors.New("消息不可为空")
	}
	if msg.MsgId == "" {
		msg.MsgId = util.UUID()
	}
	q.queueLock.Lock()
	defer q.queueLock.Unlock()
	// 队列满时阻塞等待
	for q.size == q.capacity {
		q.notFull.Wait()
	}
	q.items = append(q.items, msg)
	q.size++
	q.notEmpty.Signal()
	q.log.Debug("文件队列已发布消息", logger.NewField("eventType", msg.EventType), logger.NewField("data", msg.Data))
	util.Go(func() {
		q.cache.SetObjectRetry("mq:mem:messages", q.items, 9999*time.Hour)
	})
	return nil
}

// Subscribe implements Queue.
func (s *memQueue) Subscribe(eventType, callback, plugin string, stype int) error {
	s.log.Debug("文件队列订阅", logger.NewField("eventType", eventType), logger.NewField("callback", callback))
	s.eventTypeLock.Lock()
	defer s.eventTypeLock.Unlock()

	// 创建主题映射如果不存在
	if _, exists := s.eventMap[eventType]; !exists {
		s.eventMap[eventType] = make(map[string]*SubscribeInfo)
	}
	// 添加订阅
	s.eventMap[eventType][callback] = &SubscribeInfo{
		Plugin: plugin,
		Stype:  stype,
		Url:    callback,
	}
	util.Go(func() {
		s.cache.SetObjectRetry("mq:mem:event-map", s.eventMap, 9999*time.Hour)
	})
	return nil
}

// Unsubscribe implements Queue.
func (s *memQueue) Unsubscribe(eventType, callback string) error {
	s.log.Debug("文件队列取消订阅", logger.NewField("eventType", eventType), logger.NewField("callback", callback))
	s.eventTypeLock.Lock()
	defer s.eventTypeLock.Unlock()

	urls, exists := s.eventMap[eventType]
	if !exists {
		return nil
	}
	// 移除订阅
	if _, exists := urls[callback]; !exists {
		return nil
	}
	delete(urls, callback)
	// 清理空主题
	if len(urls) == 0 {
		delete(s.eventMap, eventType)
	}
	util.Go(func() {
		s.cache.SetObjectRetry("mq:mem:event-map", s.eventMap, 9999*time.Hour)
	})
	return nil
}

func newMemQueue(cfg *config.MemQueueConfig, callback QueueCallback, c cache2.Cache, tmpLog logger.Logger) (Queue, error) {
	if cfg.Capacity <= 0 {
		return nil, errors.New("文件队列容量不可为空")
	}
	q := &memQueue{
		cache:    c,
		items:    make([]*Message, 0, cfg.Capacity),
		capacity: cfg.Capacity,
		eventMap: make(map[string]map[string]*SubscribeInfo),
		stopChan: make(chan struct{}),
		log:      tmpLog,
	}
	var fileMsgs []*Message
	if has, _ := c.GetObjectRetry("mq:mem:messages", &fileMsgs); has {
		q.items = fileMsgs
		q.size = len(fileMsgs)
	}
	var eventMapData map[string]map[string]*SubscribeInfo
	if has, _ := c.GetObjectRetry("mq:mem:event-map", &eventMapData); has {
		q.eventMap = eventMapData
	}
	q.notEmpty = sync.NewCond(&q.queueLock)
	q.notFull = sync.NewCond(&q.queueLock)
	q.consuming = true
	util.Go(func() { q.consumeLoop(callback) })
	q.log.Println("[文件队列]已初始化")
	return q, nil
}

func (q *memQueue) consumeLoop(callback QueueCallback) {
	for {
		select {
		case <-q.stopChan:
			return
		default:
			msg := q.take()
			if msg == nil {
				return
			}
			// 获取订阅者列表（分离锁）
			q.eventTypeLock.RLock()
			urls, exists := q.eventMap[msg.EventType]
			if !exists {
				q.eventTypeLock.RUnlock()
				continue
			}
			// 复制URL避免长时间持有锁
			urlList := make([]*SubscribeInfo, 0, len(urls))
			for _, info := range urls {
				urlList = append(urlList, info)
			}
			q.eventTypeLock.RUnlock()
			// 发送给所有订阅者,重试时只投递上次失败的
			isOk := true
			hasOk := false
			failUrls := make([]*PushInfo, 0)
			var lastFailUrls []*PushInfo
			partiallyId := fmt.Sprintf("mq:partially:%s", msg.MsgId)
			q.cache.GetObject(partiallyId, &lastFailUrls)
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
				delCheck := callback.DeadMsgCheck(failUrls[0])
				if !delCheck {
					q.Publish(msg)
				}
			}
		}
	}
}

func (q *memQueue) take() *Message {
	q.queueLock.Lock()
	defer q.queueLock.Unlock()

	// 队列空时等待
	for q.size == 0 {
		q.notEmpty.Wait()
	}
	if q.items == nil {
		return nil
	}

	item := q.items[0]
	q.items[0] = nil // 帮助GC回收
	q.items = q.items[1:]
	q.size--
	q.notFull.Signal()
	util.Go(func() {
		q.cache.SetObjectRetry("mq:mem:messages", q.items, 9999*time.Hour)
	})
	return item
}

// UnsubscribeByPlugin 按插件名清理所有订阅
func (s *memQueue) UnsubscribeByPlugin(pluginName string) error {
	s.eventTypeLock.Lock()
	defer s.eventTypeLock.Unlock()

	for eventType, urls := range s.eventMap {
		for key, info := range urls {
			if info.Plugin == pluginName {
				delete(urls, key)
			}
		}
		if len(urls) == 0 {
			delete(s.eventMap, eventType)
		}
	}
	util.Go(func() {
		s.cache.SetObjectRetry("mq:mem:event-map", s.eventMap, 9999*time.Hour)
	})
	return nil
}

func (s *memQueue) SupportDistributed() bool {
	return false
}
