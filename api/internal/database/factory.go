package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/fs185085781/v9os/internal/cache"
	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/internal/logger"
	"github.com/fs185085781/v9os/internal/model/system"
	"gorm.io/gorm"
	gormLog "gorm.io/gorm/logger"
)

type dbLog struct {
	log     logger.Logger
	showSql bool
}

func (l *dbLog) LogMode(level gormLog.LogLevel) gormLog.Interface {
	return l
}
func (l *dbLog) Info(ctx context.Context, msg string, data ...interface{}) {
	l.log.Info(fmt.Sprintf(msg, data...))
}
func (l *dbLog) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.log.Warn(fmt.Sprintf(msg, data...))
}
func (l *dbLog) Error(ctx context.Context, msg string, data ...interface{}) {
	l.log.Error(fmt.Sprintf(msg, data...))
}
func (l *dbLog) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, rowsAffected := fc()
	l.log.Debug(sql, logger.NewField("rowsAffected", rowsAffected))
	if l.showSql {
		fmt.Printf("[%s] %s -> %d \n", time.Now().Format("2006-01-02 15:04:05"), sql, rowsAffected)
	}
}
func NewDatabase(cfg *config.DatabaseConfig, c cache.Cache, tmpLog logger.Logger) (Database, error) {
	log := &dbLog{
		log:     tmpLog,
		showSql: cfg.ShowSql,
	}
	if len(cfg.DSN) < 1 {
		return nil, errors.New("配置中暂未发现数据库连接串")
	}
	c2 := &gromCache{
		cache: c,
	}
	db, err := newOneDb(cfg, c2, cfg.DSN[0], log)
	if err != nil {
		return nil, err
	}
	dbs := make([]*dbStatus, 0)
	dbs = append(dbs, &dbStatus{
		db:     db,
		dsn:    cfg.DSN[0],
		status: true,
	})
	for index, dsn := range cfg.DSN {
		if index == 0 {
			continue
		}
		dbs = append(dbs, &dbStatus{
			db:     nil,
			dsn:    dsn,
			status: false,
		})
	}
	log.log.Println("[" + cfg.Driver + "数据库]已初始化")
	return &dbPool{
		cfg:   cfg,
		dbs:   dbs,
		log:   log,
		cache: c2,
	}, nil
}

func newOneDb(cfg *config.DatabaseConfig, c *gromCache, dsn string, log *dbLog) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	switch cfg.Driver {
	case "sqlite":
		db, err = newSqliteDb(cfg, dsn, log)
	case "mysql":
		db, err = newMysqlDb(cfg, dsn, log)
	case "postgres":
		db, err = newPostgresDb(cfg, dsn, log)
	case "gaussdb":
		db, err = newGaussdbDb(cfg, dsn, log)
	case "sqlserver":
		db, err = newSqlserverDb(cfg, dsn, log)
	case "clickhouse":
		db, err = newClickhouseDb(cfg, dsn, log)
	}
	if err != nil {
		return nil, err
	}
	if db == nil {
		return nil, fmt.Errorf("驱动%s暂未支持", cfg.Driver)
	}
	if !cfg.SoftDelete {
		db = db.Unscoped()
	}
	if cfg.Cache && !cfg.ShowSql {
		c.registerCallback(db)
	}
	return db, nil
}

type dbPool struct {
	cfg      *config.DatabaseConfig
	dbs      []*dbStatus
	dbPoolMu sync.Mutex
	cache    *gromCache
	log      *dbLog
}

func (d *dbPool) GetObject(key string, entitie interface{}) (bool, error) {
	val, err := d.GetValue(key)
	if err != nil {
		return false, err
	}
	if val == "" {
		return false, nil
	}
	err = json.Unmarshal([]byte(val), entitie)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *dbPool) SetObject(key string, entitie interface{}) error {
	res, err := json.Marshal(entitie)
	if err != nil {
		return err
	}
	return d.SetValue(key, string(res))
}

func (d *dbPool) GetValue(key string) (string, error) {
	var data system.Data
	err := d.Write().Where("data_key = ?", key).First(&data).Error
	if !d.IsOk(err) {
		return "", err
	}
	if data.ID < 1 {
		return "", nil
	}
	return data.DataValue, nil
}

func (d *dbPool) SetValue(key string, val string) error {
	var data system.Data
	err := d.Write().Where("data_key = ?", key).First(&data).Error
	if !d.IsOk(err) {
		return err
	}
	data.DataValue = val
	data.DataKey = key
	if data.ID > 0 {
		err = d.Write().Updates(&data).Error
	} else {
		err = d.Write().Create(&data).Error
	}
	return err
}

func (d *dbPool) IsOk(err error) bool {
	return err == nil || errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, gorm.ErrDuplicatedKey) || errors.Is(err, gorm.ErrForeignKeyViolated) || errors.Is(err, gorm.ErrCheckConstraintViolated)
}

// Close implements Database.
func (d *dbPool) Close() error {
	d.dbPoolMu.Lock()
	defer d.dbPoolMu.Unlock()
	var lastErr error
	for _, db := range d.dbs {
		if db.db == nil {
			continue
		}
		db, _ := db.db.DB()
		err := db.Close()
		if err != nil {
			lastErr = err
		}
	}
	d.log.log.Println("[" + d.cfg.Driver + "数据库]已关闭")
	return lastErr
}

func (d *dbPool) ClearCache(table string) {
	d.cache.cache.RemovePrefix(d.cache.getDeletePrefixKey(table))
}

func (d *dbPool) SupportDistributed() bool {
	return d.cfg.Driver != "sqlite"
}

func session(db *gorm.DB) *gorm.DB {
	return db.Session(&gorm.Session{})
}

// Read implements Database.
func (d *dbPool) Read() *gorm.DB {
	d.dbPoolMu.Lock()
	defer d.dbPoolMu.Unlock()
	if len(d.dbs) <= 1 {
		return d.writeCore()
	}
	i := rand.Intn(len(d.dbs)-1) + 1
	if d.dbs[i].status {
		return session(d.dbs[i].db)
	}
	if d.dbs[i].db == nil {
		db, err := newOneDb(d.cfg, d.cache, d.dbs[i].dsn, d.log)
		if err != nil {
			return nil
		}
		d.dbs[i].db = db
	}
	err := d.healthCheck(d.dbs[i].db)
	if err != nil {
		return nil
	}
	d.dbs[i].status = true
	return session(d.dbs[i].db)
}

func (d *dbPool) Write() *gorm.DB {
	d.dbPoolMu.Lock()
	defer d.dbPoolMu.Unlock()
	return d.writeCore()
}

func (d *dbPool) writeCore() *gorm.DB {
	if !d.dbs[0].status {
		err := d.healthCheck(d.dbs[0].db)
		if err != nil {
			return nil
		}
		d.dbs[0].status = true
	}
	return session(d.dbs[0].db)
}

// Transaction implements Database.
func (d *dbPool) Transaction(fn func(tx *gorm.DB) error) error {
	db := d.Write()
	if db == nil {
		return errors.New("Database connection failed")
	}
	return db.Transaction(fn)
}

func (d *dbPool) GetByID(id uint, entity interface{}) error {
	db := d.Read()
	if db == nil {
		return errors.New("Database connection failed")
	}
	if err := db.First(entity, id).Error; err != nil {
		return err
	}
	return nil
}

func (d *dbPool) Create(entity interface{}) error {
	db := d.Write()
	if db == nil {
		return errors.New("Database connection failed")
	}
	return db.Create(entity).Error
}

func (d *dbPool) Update(entity interface{}) error {
	db := d.Write()
	if db == nil {
		return errors.New("Database connection failed")
	}
	return db.Select("*").Updates(entity).Error
}

func (d *dbPool) Delete(id uint, entity interface{}) error {
	db := d.Write()
	if db == nil {
		return errors.New("Database connection failed")
	}
	return db.Delete(entity, id).Error
}

func (d *dbPool) Find(entities interface{}, conditions interface{}, args ...interface{}) error {
	db := d.Read()
	if db == nil {
		return errors.New("Database connection failed")
	}
	return db.Where(conditions, args...).Find(entities).Error
}

func (d *dbPool) First(entitie interface{}, conditions interface{}, args ...interface{}) error {
	db := d.Read()
	if db == nil {
		return errors.New("Database connection failed")
	}
	return db.Where(conditions, args...).First(entitie).Error
}

func (d *dbPool) Count(entity interface{}, count *int64, conditions interface{}, args ...interface{}) error {
	db := d.Read()
	if db == nil {
		return errors.New("Database connection failed")
	}
	return db.Model(entity).Where(conditions, args...).Count(count).Error
}

func (d *dbPool) healthCheck(db *gorm.DB) error {
	_, err := db.DB()
	if err != nil {
		return err
	}
	return nil
}

type dbStatus struct {
	dsn    string
	db     *gorm.DB
	status bool //true正常 false异常
}
