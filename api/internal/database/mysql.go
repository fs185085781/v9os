package database

import (
	"time"

	"github.com/fs185085781/v9os/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func newMysqlDb(cnf *config.DatabaseConfig, dsn string, log *dbLog) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:      log,
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
