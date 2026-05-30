package database

import (
	"sync"

	"github.com/fs185085781/v9os/internal/ioc"
	"github.com/fs185085781/v9os/internal/model/base"
	"github.com/fs185085781/v9os/internal/model/user"
	"github.com/fs185085781/v9os/pkg/util"
	"gorm.io/gorm"
)

func AutoMigrate() error {
	db := ioc.Ioc().Get(ioc.KeyDatabase).(Database).Write()
	registerMap := ioc.Ioc().GetOrRegister(ioc.KeyPluginDataMap, &sync.Map{}).(*sync.Map)
	ioc.Ioc().Unregister(ioc.KeyPluginDataMap)
	registerMap.Range(func(key, value interface{}) bool {
		plugin, ok := key.(base.PluginDataTable)
		if ok {
			db.Session(&gorm.Session{}).Table(plugin.PluginTable).AutoMigrate(plugin.PlugDataStruct)
		}
		return true
	})
	registerMap = ioc.Ioc().GetOrRegister(ioc.KeyModelMap, &sync.Map{}).(*sync.Map)
	ioc.Ioc().Unregister(ioc.KeyModelMap)
	objs := []interface{}{}
	registerMap.Range(func(key, value interface{}) bool {
		objs = append(objs, key)
		return true
	})
	err := db.AutoMigrate(objs...)
	if err != nil {
		return err
	}
	//查询是否有管理员,没有就创建一个
	checkCreateAdmin()
	return nil
}

func checkCreateAdmin() error {
	dbPool := ioc.Ioc().Get(ioc.KeyDatabase).(Database)
	var u user.User
	dbPool.GetByID(1, &u)
	if u.ID > 0 {
		return nil
	}
	return dbPool.Create(&user.User{
		Username: "admin",
		Name:     "Admin",
		Password: util.EncodePassword("123456"),
		Enabled:  1,
		Avatar:   "/assets/images/logo.png",
	})
}
