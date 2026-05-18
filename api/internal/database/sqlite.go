package database

import (
	"path/filepath"
	"time"

	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/pkg/util"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newSqliteDb(cnf *config.DatabaseConfig, dsn string, log *dbLog) (*gorm.DB, error) {
	dsn = filepath.Join(util.RunDir(), dsn)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger:      log, // 生产环境可改为Warn
		PrepareStmt: true,
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	// 连接池配置
	sqlDB.SetMaxIdleConns(cnf.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cnf.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	return db, nil
}
