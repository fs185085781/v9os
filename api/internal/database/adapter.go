package database

import (
	"gorm.io/gorm"
)

type Database interface {
	Write() *gorm.DB
	Read() *gorm.DB
	Close() error
	Transaction(fn func(tx *gorm.DB) error) error
	Create(entity interface{}) error
	GetByID(id uint, entity interface{}) error
	Update(entity interface{}) error
	Delete(id uint, entity interface{}) error
	Find(entities interface{}, conditions interface{}, args ...interface{}) error
	First(entitie interface{}, conditions interface{}, args ...interface{}) error
	Count(entity interface{}, count *int64, conditions interface{}, args ...interface{}) error
	ClearCache(tableName string)
	SupportDistributed() bool
	IsOk(err error) bool
	GetObject(key string, entitie interface{}) (bool, error)
	SetObject(key string, entitie interface{}) error
	GetValue(key string) (string, error)
	SetValue(key string, val string) error
}
