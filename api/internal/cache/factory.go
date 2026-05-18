package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/internal/logger"
	"github.com/fs185085781/v9os/pkg/util"
)

type ObjectSupport struct {
	Cache
}

func (o *ObjectSupport) GetObject(key string, result any) error {
	if reflect.ValueOf(result).Kind() != reflect.Ptr {
		return errors.New("请传入指针")
	}
	data, err := o.GetValue(key)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, result)
}

type memCacheItem struct {
	Data           interface{} `json:"data"`
	CreatedAt      int64       `json:"createdAt"`
	CleanupTimeout int64       `json:"cleanupTimeout"`
}

var memCache map[string]*memCacheItem
var memCacheMu sync.RWMutex
var memCacheOnce sync.Once

func cleanupMemCache() {
	for {
		time.Sleep(5 * time.Minute)
		memCacheMu.Lock()
		now := util.UnixSeconds()
		for k, v := range memCache {
			if v.CreatedAt+v.CleanupTimeout <= now {
				delete(memCache, k)
			}
		}
		memCacheMu.Unlock()
	}
}

func memCacheExpired(item *memCacheItem, now int64, timeout time.Duration) bool {
	if item == nil {
		return true
	}
	return item.CreatedAt+int64(timeout.Seconds()) <= now
}

func normalizeMemCacheTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return time.Minute
	}
	return timeout
}

func memCacheStorageTimeout(timeout time.Duration) time.Duration {
	return timeout * 2
}

func copyCacheData(src any, dst any) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}

func (o *ObjectSupport) MemCacheObject(key string, result any, fn func() interface{}, timeout time.Duration) error {
	timeout = normalizeMemCacheTimeout(timeout)
	memCacheOnce.Do(func() {
		memCache = map[string]*memCacheItem{}
		util.Go(cleanupMemCache)
	})
	now := util.UnixSeconds()
	memCacheMu.RLock()
	item := memCache[key]
	memCacheMu.RUnlock()
	if !memCacheExpired(item, now, timeout) {
		return copyCacheData(item.Data, result)
	}
	stored := &memCacheItem{}
	if err := o.GetObject(key, stored); err == nil && stored.CreatedAt > 0 && stored.Data != nil && !memCacheExpired(stored, now, timeout) {
		stored.CleanupTimeout = int64(timeout.Seconds())
		memCacheMu.Lock()
		memCache[key] = stored
		memCacheMu.Unlock()
		return copyCacheData(stored.Data, result)
	}

	item = &memCacheItem{
		Data:           fn(),
		CreatedAt:      now,
		CleanupTimeout: int64(timeout.Seconds()),
	}
	if item.Data == nil {
		return errors.New("fn result is nil")
	}
	memCacheMu.Lock()
	memCache[key] = item
	memCacheMu.Unlock()
	if err := o.SetObject(key, item, memCacheStorageTimeout(timeout)); err != nil {
		return err
	}
	return copyCacheData(item.Data, result)
}
func (o *ObjectSupport) needRetry(err error) bool {
	switch err {
	case badger.ErrBannedKey, badger.ErrTruncateNeeded, badger.ErrEmptyKey, badger.ErrDBClosed, badger.ErrDiscardedTxn, badger.ErrConflict,
		badger.ErrReadOnlyTxn, badger.ErrInvalidKey:
		return false
	}
	return true
}
func (o *ObjectSupport) GetObjectRetry(key string, result any) (bool, error) {
	if reflect.ValueOf(result).Kind() != reflect.Ptr {
		return false, errors.New("请传入指针")
	}
	var lastErr error
	for range 10 {
		data, err := o.GetValue(key)
		if err != nil && err != CacheIsNil {
			if !o.needRetry(err) {
				return false, err
			}
			lastErr = err
			time.Sleep(time.Second)
			continue
		}
		if err == CacheIsNil {
			return false, nil
		}
		err = json.Unmarshal(data, result)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, lastErr
}
func (o *ObjectSupport) SetObjectRetry(key string, value any, timeout time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	var lastErr error
	for range 10 {
		err = o.SetValue(key, data, timeout)
		if err != nil {
			if !o.needRetry(err) {
				return err
			}
			lastErr = err
			time.Sleep(time.Second)
			continue
		}
		return nil
	}
	return lastErr
}

func (o *ObjectSupport) SetObject(key string, value any, timeout time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return o.SetValue(key, data, timeout)
}

func NewCache(cnf *config.CachebaseConfig, tmpLog logger.Logger) (Cache, error) {
	if cnf.Driver == "file" {
		return newFileCache(cnf.File, tmpLog)
	}
	if cnf.Driver == "redis" {
		return newRedisCache(cnf.Redis, tmpLog)
	}
	return nil, fmt.Errorf("暂不支持%s缓存驱动", cnf.Driver)
}
