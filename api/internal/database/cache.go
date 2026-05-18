package database

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/fs185085781/v9os/internal/cache"
	"github.com/fs185085781/v9os/pkg/util"
	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
)

type gromCache struct {
	cache cache.Cache
}

func (g *gromCache) registerCallback(db *gorm.DB) {
	db.Callback().Create().After("*").Register("after_create", g.initCache("delete"))
	db.Callback().Query().Before("*").Register("befor_query", g.initCache("befor_query"))
	db.Callback().Query().After("*").Register("after_query", g.initCache("after_query"))
	db.Callback().Update().After("*").Register("after_update", g.initCache("delete"))
	db.Callback().Delete().After("*").Register("after_delete", g.initCache("delete"))
	db.Callback().Raw().Before("*").Register("befor_raw", g.initCache("before_raw"))
	db.Callback().Raw().After("*").Register("after_raw", g.initCache("after_raw"))
}
func (g *gromCache) initCache(action string) func(*gorm.DB) {
	return func(db *gorm.DB) {
		table := ""
		if db.Statement.Schema != nil {
			table = strings.ToLower(db.Statement.Schema.Table)
		} else {
			table = strings.ToLower(db.Statement.Table)
		}
		table = strings.ReplaceAll(table, "`", "")
		switch action {
		case "befor_query":
			g.beforeQuery(db, table)
		case "after_query":
			g.afterQuery(db)
		case "delete":
			g.delete(table)
		case "before_raw", "after_raw":
			sql := strings.ToLower(db.Statement.SQL.String())
			isSelect := strings.HasPrefix(sql, "select")
			if table == "" {
				if strings.Contains(sql, " from ") {
					table = strings.Split(sql, " from ")[1]
					table = strings.Split(table, " ")[0]
				} else if strings.Contains(sql, "update ") {
					table = strings.Split(sql, "update ")[1]
					table = strings.Split(table, " ")[0]
					isSelect = false
				} else if strings.Contains(sql, " into ") {
					table = strings.Split(sql, " into ")[1]
					table = strings.Split(table, " ")[0]
					isSelect = false
				}
				table = strings.ReplaceAll(table, "`", "")
			}
			if table == "" {
				return
			}
			if isSelect {
				if action == "before_raw" {
					g.beforeQuery(db, table)
				} else {
					g.afterQuery(db)
				}
			} else if action == "after_raw" {
				g.delete(table)
			}
		}
	}
}

type cacheResult struct {
	Data         interface{} `json:"data"`
	RowsAffected *int64      `json:"rowsAffected"`
}

var errorHitCache = errors.New("hit cache")

func (g *gromCache) beforeQuery(db *gorm.DB, table string) {
	callbacks.BuildQuerySQL(db)
	sql := db.Statement.SQL.String()
	key := g.getQueryKey(table, sql, db.Statement.Vars...)
	db.InstanceSet("gorm:cache:key", key)
	res := cacheResult{
		Data:         db.Statement.Dest,
		RowsAffected: &db.RowsAffected,
	}
	has, _ := g.cache.GetObjectRetry(key, &res)
	if has {
		db.Error = errorHitCache
		return
	}
	lock := g.cache.CreateLock(key)
	lock.Lock()
	has, _ = g.cache.GetObjectRetry(key, &res)
	if has {
		db.Error = errorHitCache
		lock.UnLock()
		return
	}
	util.Go(func() {
		//防止GORM出现panic进行了死锁,如果1分钟没有解锁,则强制解锁
		defer lock.UnLock()
		time.Sleep(time.Minute)
	})
	db.InstanceSet("gorm:cache:lock", lock)
}

func (g *gromCache) afterQuery(db *gorm.DB) {
	lockObj, ok := db.InstanceGet("gorm:cache:lock")
	if ok && lockObj != nil {
		lock := lockObj.(cache.Lock)
		defer lock.UnLock()
	}
	if db.Error != nil {
		if errors.Is(db.Error, errorHitCache) {
			db.Error = nil
		}
		return
	}
	keyObj, _ := db.InstanceGet("gorm:cache:key")
	key := keyObj.(string)
	res := cacheResult{
		Data:         db.Statement.Dest,
		RowsAffected: &db.RowsAffected,
	}
	g.cache.SetObjectRetry(key, res, 30*time.Minute)
}
func (g *gromCache) delete(table string) {
	g.cache.RemovePrefix(g.getDeletePrefixKey(table))
}

func (g *gromCache) getQueryKey(tableName string, sql string, vars ...interface{}) string {
	buf := strings.Builder{}
	buf.WriteString(sql)
	for _, v := range vars {
		pv := reflect.ValueOf(v)
		if pv.Kind() == reflect.Ptr {
			buf.WriteString(fmt.Sprintf("%v", pv.Elem()))
		} else {
			buf.WriteString(fmt.Sprintf("%v", v))
		}
	}
	str := buf.String()
	replacer := strings.NewReplacer(
		"*", "_a_",
		" ", "_b_",
		"\n", "_c_",
		"\r", "_d_",
		"?", "_e_",
		"[", "_fl_",
		"]", "_fr_",
		"\\", "_g_",
	)
	str = replacer.Replace(str)
	return fmt.Sprintf("sql_cache:%s:%s", tableName, str)
}

func (g *gromCache) getDeletePrefixKey(tableName string) string {
	return fmt.Sprintf("sql_cache:%s", tableName)
}
