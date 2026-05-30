package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/internal/logger"
	"github.com/fs185085781/v9os/pkg/util"
	"github.com/redis/go-redis/v9"
)

// 缓存实现 ---- 开始
func newRedisCache(cfg *config.RedisCacheConfig, tmpLog logger.Logger) (Cache, error) {
	log := &redisLog{
		log: tmpLog,
	}
	redis.SetLogger(log)
	var rsClient redisInterface
	switch cfg.Mode {
	case "standalone":
		rsClient = redis.NewClient(&redis.Options{
			Addr:         cfg.Addrs[0],
			Password:     cfg.Password,
			DB:           cfg.DB,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.MinIdleConns,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			DialTimeout:  cfg.DialTimeout,
		})
	case "sentinel":
		rsClient = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    cfg.MasterName,
			SentinelAddrs: cfg.Addrs,
			Password:      cfg.Password,
			DB:            cfg.DB,
			PoolSize:      cfg.PoolSize,
			MinIdleConns:  cfg.MinIdleConns,
			ReadTimeout:   cfg.ReadTimeout,
			WriteTimeout:  cfg.WriteTimeout,
			DialTimeout:   cfg.DialTimeout,
			MaxRetries:    3,
		})
	case "cluster":
		rsClient = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        cfg.Addrs,
			Password:     cfg.Password,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.MinIdleConns,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			DialTimeout:  cfg.DialTimeout,
			MaxRetries:   3,
		})
	default:
		return nil, fmt.Errorf("Redis模式%s暂不支持", cfg.Mode)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := rsClient.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}
	store := &redisCache{
		rsClient: rsClient,
		clientID: util.UUID(),
	}
	store.Cache = store
	store.log = log
	store.log.log.Println("[Redis缓存]已初始化")
	return store, nil
}

type redisCache struct {
	ObjectSupport
	rsClient redisInterface
	clientID string
	log      *redisLog
}
type redisInterface interface {
	Close() error
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
	Ping(ctx context.Context) *redis.StatusCmd
}
type redisLog struct {
	log logger.Logger
}

func (r *redisLog) Printf(ctx context.Context, format string, v ...interface{}) {
	r.log.Info(fmt.Sprintf(format, v...))
}

// Close implements Cache.
func (r *redisCache) Close() error {
	err := r.rsClient.Close()
	r.log.log.Println("[Redis缓存]已关闭", logger.NewField("error", err))
	return err
}

// GetValue implements Cache.
func (r *redisCache) GetValue(key string) ([]byte, error) {
	result := r.rsClient.Get(context.Background(), key)
	if err := result.Err(); err != nil {
		if err == redis.Nil {
			return nil, CacheIsNil
		}
		return nil, err
	}
	r.log.log.Debug("Redis缓存GetValue成功", logger.NewField("key", key), logger.NewField("value", result.String()))
	return result.Bytes()
}

// RemovePrefix implements Cache.
func (r *redisCache) RemovePrefix(prefix string) error {
	ctx := context.Background()
	iter := r.rsClient.Scan(ctx, 0, prefix+":*", 0).Iterator()
	for iter.Next(ctx) {
		if err := r.rsClient.Del(ctx, iter.Val()).Err(); err != nil {
			r.log.log.Error("Redis缓存RemovePrefix失败", logger.NewField("key", iter.Val()), logger.NewField("error", err))
			return err
		}
	}
	err := iter.Err()
	if err != nil {
		r.log.log.Error("Redis缓存RemovePrefix失败", logger.NewField("prefix", prefix), logger.NewField("error", err))
		return err
	}
	r.log.log.Debug("Redis缓存RemovePrefix成功", logger.NewField("prefix", prefix))
	return nil
}

// RemoveValue implements Cache.
func (r *redisCache) RemoveValue(key string) error {
	err := r.rsClient.Del(context.Background(), key).Err()
	if err != nil {
		r.log.log.Error("Redis缓存RemoveValue失败", logger.NewField("key", key), logger.NewField("error", err))
		return err
	}
	r.log.log.Debug("Redis缓存RemoveValue成功", logger.NewField("key", key))
	return nil
}

// SetValue implements Cache.
func (r *redisCache) SetValue(key string, val []byte, timeout time.Duration) error {
	err := r.rsClient.Set(context.Background(), key, val, timeout).Err()
	if err != nil {
		r.log.log.Error("Redis缓存SetValue失败", logger.NewField("key", key), logger.NewField("error", err))
		return err
	}
	r.log.log.Debug("Redis缓存SetValue成功", logger.NewField("key", key), logger.NewField("value", val), logger.NewField("timeout", timeout))
	return nil
}

// CreateLock implements Cache.
func (r *redisCache) CreateLock(key string) Lock {
	return &redisLock{
		key: "lock:" + key,
		rs:  r,
		val: r.clientID + util.UUID(),
	}
}

func (r *redisCache) GetLock(key, val string) Lock {
	return &redisLock{
		key: "lock:" + key,
		rs:  r,
		val: val,
	}
}

// 缓存实现 ---- 结束
// 锁实现 ---- 开始
type redisLock struct {
	key string
	val string
	rs  *redisCache
}

func (r *redisLock) GetVal() string {
	return r.val
}

func (r *redisLock) IsAlive() bool {
	lockVal, err := r.rs.GetValue(r.key)
	if err != nil || lockVal == nil {
		return false
	}
	return string(lockVal) == r.val
}

func (r *redisLock) Lock() {
	for {
		ok, err := r.rs.rsClient.SetNX(context.Background(), r.key, r.val, 30*time.Second).Result()
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if ok {
			r.startWatchdog() // 启动看门狗续期
			return
		} else {
			lockVal, _ := r.rs.GetValue(r.key)
			if lockVal != nil && string(lockVal) == r.val {
				r.handlersPlus(1)
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}
func (r *redisLock) handlersPlus(i int) int {
	var handlers int
	handlersKey := "handlers:" + r.key + ":" + r.val
	for {
		if i == 0 {
			err := r.rs.RemoveValue(handlersKey)
			if err == nil {
				return 0
			}
		} else {
			err := r.rs.GetObject(handlersKey, &handlers)
			if err == nil || err == CacheIsNil {
				err = r.rs.SetObjectRetry(handlersKey, handlers+i, time.Hour*12)
				if err == nil {
					return handlers
				}
			}
		}
		time.Sleep(time.Second)
	}
}
func (r *redisLock) startWatchdog() {
	util.Go(func() {
		timerKey := "timer:" + r.key + ":" + r.val
		r.rs.log.log.Debug("[Redis缓存]看门狗启动成功", logger.NewField("timerKey", timerKey))
		for {
			err := r.rs.SetValue(timerKey, []byte("1"), 72*time.Hour)
			if err == nil {
				break
			}
			time.Sleep(3 * time.Second)
		}
		for {
			time.Sleep(10 * time.Second)
			timer, err := r.rs.GetValue(timerKey)
			if (err == nil || err == CacheIsNil) && !(timer != nil && string(timer) == "1") {
				break
			}
			lockVal, err := r.rs.GetValue(r.key)
			if (err == nil || err == CacheIsNil) && !(lockVal != nil && string(lockVal) == r.val) {
				r.rs.RemoveValue(timerKey)
				break
			}
			r.rs.rsClient.Expire(context.Background(), r.key, 30*time.Second)
		}
		r.rs.log.log.Debug("[Redis缓存]看门狗已关闭", logger.NewField("timerKey", timerKey))
	})
}

// TryLock implements Lock.
func (r *redisLock) TryLock() bool {
	ok, err := r.rs.rsClient.SetNX(context.Background(), r.key, r.val, 30*time.Second).Result()
	if err != nil {
		return false
	}
	if ok {
		r.startWatchdog() // 启动看门狗续期
		return true
	} else {
		lockVal, _ := r.rs.GetValue(r.key)
		if lockVal != nil && string(lockVal) == r.val {
			r.handlersPlus(1)
			return true
		}
	}
	return false
}

// UnLock implements Lock.
func (r *redisLock) UnLock() {
	mapKey := r.key + ":" + r.val
	// 获取当前重入计数
	handlers := r.handlersPlus(-1)
	if handlers > 0 {
		return
	}
	r.handlersPlus(0)
	timerKey := "timer:" + mapKey
	script := `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		redis.call("DEL", KEYS[2])
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end`
	isFirst := true
	for range 3 {
		_, err := r.rs.rsClient.Eval(context.Background(), script, []string{r.key, timerKey}, r.val).Result()
		if err == nil {
			break
		}
		if isFirst {
			isFirst = false
			r.rs.log.log.Error("Redis缓存UnLock失败", logger.NewField("key", r.key), logger.NewField("error", err))
		}
		time.Sleep(time.Second)
	}
}

// WithLock implements Lock.
func (r *redisLock) WithLock(fn func()) {
	r.Lock()
	defer r.UnLock()
	fn()
}

//锁实现 ---- 结束

func (r *redisCache) SupportDistributed() bool {
	return true
}
