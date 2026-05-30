package cache

import (
	"errors"
	"time"
)

var CacheIsNil = errors.New("Cache is nil")

type Lock interface {
	Lock()
	UnLock()
	TryLock() bool
	WithLock(fn func())
	GetVal() string
	IsAlive() bool
}

type Cache interface {
	SetValue(key string, val []byte, timeout time.Duration) error
	GetValue(key string) ([]byte, error)
	RemoveValue(key string) error
	RemovePrefix(prefix string) error
	Close() error
	CreateLock(key string) Lock
	GetLock(key, val string) Lock
	GetObject(key string, result any) error
	SetObject(key string, value any, timeout time.Duration) error
	GetObjectRetry(key string, result any) (bool, error)
	SetObjectRetry(key string, value any, timeout time.Duration) error
	MemCacheObject(key string, result any, fn func() interface{}, timeout time.Duration) error
	SupportDistributed() bool
}
