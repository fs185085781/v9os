package cache

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/internal/logger"
	"github.com/fs185085781/v9os/pkg/util"
)

// 缓存实现 ---- 开始
func newFileCache(cnf *config.FileCacheConfig, tmpLog logger.Logger) (Cache, error) {
	log := &fileLog{
		log: tmpLog,
	}
	opts := badger.DefaultOptions(filepath.Join(util.RunDir(), cnf.Dir)).
		WithValueLogFileSize(100 << 20).
		WithNumGoroutines(16).
		WithLogger(log)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	store := &fileCache{
		db:  db,
		log: log,
	}
	store.Cache = store
	store.log.log.Println("[文件缓存]已初始化")
	return store, nil
}

type fileCache struct {
	ObjectSupport
	db    *badger.DB
	locks sync.Map
	mu    sync.Mutex
	log   *fileLog
}
type fileLog struct {
	log logger.Logger
}

func (f *fileLog) Debugf(format string, v ...interface{}) {
	f.log.Debug(fmt.Sprintf(format, v...))
}

func (f *fileLog) Errorf(format string, v ...interface{}) {
	f.log.Error(fmt.Sprintf(format, v...))
}

func (f *fileLog) Infof(format string, v ...interface{}) {
	f.log.Info(fmt.Sprintf(format, v...))
}

func (f *fileLog) Warningf(format string, v ...interface{}) {
	f.log.Warn(fmt.Sprintf(format, v...))
}

// Close implements Cache.
func (f *fileCache) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	time.Sleep(100 * time.Millisecond)
	err := f.db.Close()
	f.log.log.Println("[文件缓存]已关闭")
	return err
}

// GetValue implements Cache.
func (f *fileCache) GetValue(key string) ([]byte, error) {
	var valCopy []byte
	err := f.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		// 复制值（避免闭包结束数据失效）
		valCopy, err = item.ValueCopy(nil)
		return err
	})
	if err == badger.ErrKeyNotFound {
		return nil, CacheIsNil
	}
	f.log.log.Debug("获取缓存值", logger.NewField("key", key), logger.NewField("error", err))
	return valCopy, err
}

// RemovePrefix implements Cache.
func (f *fileCache) RemovePrefix(prefix string) error {
	err := f.db.Update(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefixBytes := []byte(prefix + ":")
		for it.Seek(prefixBytes); it.ValidForPrefix(prefixBytes); it.Next() {
			if err := txn.Delete(it.Item().KeyCopy(nil)); err != nil {
				return err
			}
		}
		return nil
	})
	f.log.log.Debug("删除缓存前缀", logger.NewField("key", prefix), logger.NewField("error", err))
	return err
}

// RemoveValue implements Cache.
func (f *fileCache) RemoveValue(key string) error {
	err := f.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
	f.log.log.Debug("删除缓存值", logger.NewField("key", key), logger.NewField("error", err))
	return err
}

// SetValue implements Cache.
func (f *fileCache) SetValue(key string, val []byte, timeout time.Duration) error {
	err := f.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(key), val)
		if timeout > 0 {
			e = e.WithTTL(timeout)
		}
		return txn.SetEntry(e)
	})
	f.log.log.Debug("设置缓存值", logger.NewField("key", key), logger.NewField("error", err))
	return err
}

// CreateLock implements Cache.
func (f *fileCache) CreateLock(key string) Lock {
	l := &fileLock{
		key:   key,
		val:   util.UUID(),
		cache: f,
	}
	return l
}

func (f *fileCache) GetLock(key, val string) Lock {
	l, exists := f.locks.Load(key)
	if !exists {
		return nil
	}
	cur := l.(*currentFileLock)
	if cur.lock.val != val {
		return nil
	}
	return cur.lock
}

// 缓存实现 ---- 结束
// 锁实现 ---- 开始
type fileLock struct {
	key   string
	val   string
	cache *fileCache
}
type currentFileLock struct {
	lock      *fileLock
	mu        sync.Mutex
	holdCount int
	nextChan  []*nextLockInfo
}

type nextLockInfo struct {
	lock *fileLock
	ch   chan struct{}
}

func (r *fileLock) GetVal() string {
	return r.val
}

func (f *fileLock) Lock() {
	mu, _ := f.cache.locks.LoadOrStore(f.key, &currentFileLock{
		lock:      f,
		nextChan:  make([]*nextLockInfo, 0),
		holdCount: 0,
	})
	cur := mu.(*currentFileLock)
	cur.mu.Lock()
	if cur.lock.val == f.val {
		cur.holdCount++
		cur.mu.Unlock()
		return
	}
	ch := make(chan struct{})
	cur.nextChan = append(cur.nextChan, &nextLockInfo{
		lock: f,
		ch:   ch,
	})
	cur.mu.Unlock()
	<-ch

}

func (f *fileLock) TryLock() bool {
	mu, _ := f.cache.locks.LoadOrStore(f.key, &currentFileLock{
		lock:      f,
		nextChan:  make([]*nextLockInfo, 0),
		holdCount: 0,
	})
	cur := mu.(*currentFileLock)
	cur.mu.Lock()
	defer cur.mu.Unlock()
	if cur.lock.val == f.val {
		//重入
		cur.holdCount++
		return true
	}
	return false
}

// UnLock implements Lock.
func (f *fileLock) UnLock() {
	mu, exists := f.cache.locks.Load(f.key)
	if !exists {
		//锁不存在,可能原因是已经释放过,这次属于二次释放
		return
	}
	cur := mu.(*currentFileLock)
	cur.mu.Lock()
	defer cur.mu.Unlock()
	if cur.lock.val != f.val {
		//不是当前锁,可能原因是已经释放过,这次属于二次释放
		return
	}
	cur.holdCount -= 1
	if cur.holdCount > 0 {
		//还有未重入释放的,不释放锁
		return
	}
	//释放锁
	if len(cur.nextChan) > 0 {
		next := cur.nextChan[0]
		cur.nextChan = cur.nextChan[1:]
		cur.lock = next.lock
		close(next.ch)
	} else {
		f.cache.locks.Delete(f.key)
	}

}

// WithLock implements Lock.
func (f *fileLock) WithLock(fn func()) {
	f.Lock()
	defer f.UnLock()
	fn()
}

// 锁实现 ---- 结束

func (f *fileCache) SupportDistributed() bool {
	return false
}
