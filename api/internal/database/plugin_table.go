package database

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fs185085781/v9os/internal/cache"
	"github.com/fs185085781/v9os/internal/ioc"
	"github.com/fs185085781/v9os/internal/model/plugin"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

func PluginTableName(pluginName, table string) (string, error) {
	db := ioc.Ioc().Get(ioc.KeyDatabase).(Database)
	var tbs []plugin.PluginTable
	if err := db.Write().Where("plugin_name = ? and plugin_table = ?", pluginName, table).Find(&tbs).Error; err != nil {
		return "", err
	}
	if len(tbs) > 0 {
		tb := tbs[0]
		if tb.NeedExpand == 3 {
			return "", fmt.Errorf("plugin table expanding")
		}
		return tb.RealTable, nil
	}
	c := ioc.Ioc().Get(ioc.KeyCache).(cache.Cache)
	ll := c.CreateLock("plugindata:table:lock")
	ll.Lock()
	defer ll.UnLock()
	tbs = nil
	if err := db.Write().Where("plugin_name = ? and plugin_table = ?", pluginName, table).Find(&tbs).Error; err != nil {
		return "", err
	}
	if len(tbs) > 0 {
		tb := tbs[0]
		if tb.NeedExpand == 3 {
			return "", fmt.Errorf("plugin table expanding")
		}
		return tb.RealTable, nil
	}
	var activeOldTables []string
	if err := db.Write().Model(&plugin.PluginTableExpandTask{}).
		Where("status in ?", []uint{1, 2}).
		Distinct().
		Pluck("old_real_table", &activeOldTables).Error; err != nil {
		return "", err
	}
	query := db.Write().
		Where("need_expand = ?", 1).
		Order("data_length asc,id asc")
	if len(activeOldTables) > 0 {
		query = query.Where("real_table not in ?", activeOldTables)
	}
	if err := query.Find(&tbs).Error; err != nil {
		return "", err
	}
	realName := 0
	if len(tbs) > 0 {
		numStr := strings.TrimPrefix(tbs[0].RealTable, "plugin_data_")
		realName = cast.ToInt(numStr)
	} else {
		realName = NextPluginDataTableIndex(db.Write())
	}
	rt := "plugin_data_" + strconv.Itoa(realName)
	data := &plugin.PluginTable{
		PluginName:  pluginName,
		PluginTable: table,
		RealTable:   rt,
		NeedExpand:  1,
		DataLength:  0,
	}
	if err := db.Write().Create(data).Error; err != nil {
		return "", err
	}
	return rt, nil
}

func NextPluginDataTableIndex(db *gorm.DB) int {
	var rows []plugin.PluginTable
	_ = db.Model(&plugin.PluginTable{}).Select("distinct real_table").Find(&rows).Error
	maxIndex := -1
	for _, row := range rows {
		numStr := strings.TrimPrefix(row.RealTable, "plugin_data_")
		if num, err := strconv.Atoi(numStr); err == nil && num > maxIndex {
			maxIndex = num
		}
	}
	return maxIndex + 1
}

const pluginTableExpandLimit int64 = 100000
const pluginTableExpandBatchSize = 1000
const pluginTableExpandMaxRetry uint = 5

func checkPluginTableExpand() {
	c := ioc.Ioc().Get(ioc.KeyCache).(cache.Cache)
	lock := c.CreateLock("plugindata:expand:check")
	if !lock.TryLock() {
		return
	}
	defer lock.UnLock()
	ddb := ioc.Ioc().Get(ioc.KeyDatabase).(Database)
	db := ddb.Write()
	if db == nil {
		return
	}
	var rows []plugin.PluginTable
	if err := db.Where("need_expand = ?", 1).Find(&rows).Error; err != nil {
		return
	}
	for _, row := range rows {
		if row.RealTable == "" {
			continue
		}
		var count int64
		if err := db.Table(row.RealTable).
			Where("plugin_name = ? and plugin_table = ?", row.PluginName, row.PluginTable).
			Count(&count).Error; err != nil {
			continue
		}
		if err := db.Model(&plugin.PluginTable{}).
			Where("id = ? and need_expand = ?", row.ID, 1).
			Update("data_length", uint64(count)).Error; err != nil {
			continue
		}
		if count >= pluginTableExpandLimit {
			preparePluginTableExpand(row.ID)
		}
	}
}
func preparePluginTableExpand(triggerID uint) error {
	ddb := ioc.Ioc().Get(ioc.KeyDatabase).(Database)
	return ddb.Transaction(func(tx *gorm.DB) error {
		var trigger plugin.PluginTable
		if err := tx.Where("id = ?", triggerID).First(&trigger).Error; err != nil {
			return err
		}
		if trigger.NeedExpand != 1 || trigger.RealTable == "" {
			return nil
		}
		var activeTasks int64
		if err := tx.Model(&plugin.PluginTableExpandTask{}).
			Where("old_real_table = ? and status in ?", trigger.RealTable, activePluginTableExpandTaskStatuses()).
			Count(&activeTasks).Error; err != nil {
			return err
		}
		if activeTasks > 0 {
			return nil
		}
		var affected []plugin.PluginTable
		if err := tx.Where("real_table = ? and need_expand = ?", trigger.RealTable, 1).Order("id asc").Find(&affected).Error; err != nil {
			return err
		}
		if len(affected) == 0 {
			return nil
		}
		newRealTable := "plugin_data_" + strconv.Itoa(nextPluginDataTableIndex(tx))
		for _, item := range affected {
			if item.ID == trigger.ID {
				if err := tx.Model(&plugin.PluginTable{}).
					Where("id = ? and need_expand = ?", item.ID, 1).
					Update("need_expand", 2).Error; err != nil {
					return err
				}
				continue
			}
			var exists int64
			if err := tx.Model(&plugin.PluginTableExpandTask{}).
				Where("plugin_table_id = ? and status in ?", item.ID, []uint{1, 2, 3}).
				Count(&exists).Error; err != nil {
				return err
			}
			if exists > 0 {
				continue
			}
			if err := tx.Create(&plugin.PluginTableExpandTask{
				PluginTableID: item.ID,
				PluginName:    item.PluginName,
				PluginTable:   item.PluginTable,
				OldRealTable:  item.RealTable,
				NewRealTable:  newRealTable,
				DataLength:    item.DataLength,
				Status:        1,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
func nextPluginDataTableIndex(db *gorm.DB) int {
	next := NextPluginDataTableIndex(db)
	var tasks []plugin.PluginTableExpandTask
	_ = db.Model(&plugin.PluginTableExpandTask{}).Select("distinct new_real_table").Find(&tasks).Error
	maxIndex := next - 1
	for _, task := range tasks {
		numStr := strings.TrimPrefix(task.NewRealTable, "plugin_data_")
		if num, err := strconv.Atoi(numStr); err == nil && num > maxIndex {
			maxIndex = num
		}
	}
	return maxIndex + 1
}
func activePluginTableExpandTaskStatuses() []uint {
	return []uint{1, 2}
}

func migratePluginTableExpand() {
	now := time.Now()
	if !(now.Hour() > 0 || now.Minute() >= 10) {
		return
	}
	c := ioc.Ioc().Get(ioc.KeyCache).(cache.Cache)
	lock := c.CreateLock("plugindata:expand:migrate")
	if !lock.TryLock() {
		return
	}
	defer lock.UnLock()
	db := ioc.Ioc().Get(ioc.KeyDatabase).(Database).Write()
	if db == nil {
		return
	}
	var tasks []plugin.PluginTableExpandTask
	if err := db.Where("status in ? and retry_count < ?", activePluginTableExpandTaskStatuses(), pluginTableExpandMaxRetry).
		Order("old_real_table asc,id asc").
		Limit(20).
		Find(&tasks).Error; err != nil {
		return
	}
	for _, task := range tasks {
		if err := migrateOnePluginTableExpandTask(task.ID); err != nil {
			markPluginTableExpandTaskFailed(task, err)
		}
	}
	cleanupPluginTableExpand()
}
func migrateOnePluginTableExpandTask(taskID uint) error {
	dbs := ioc.Ioc().Get(ioc.KeyDatabase).(Database)
	db := dbs.Write()
	var task plugin.PluginTableExpandTask
	if err := db.Where("id = ?", taskID).First(&task).Error; err != nil {
		return err
	}
	if task.Status == 3 {
		return nil
	}
	if task.OldRealTable == "" || task.NewRealTable == "" {
		return fmt.Errorf("plugin table expand task missing real table")
	}
	var currentTable plugin.PluginTable
	if err := db.Where("id = ?", task.PluginTableID).First(&currentTable).Error; err != nil {
		return err
	}
	if currentTable.RealTable == task.NewRealTable && currentTable.NeedExpand == 1 {
		if err := db.Model(&plugin.PluginTableExpandTask{}).
			Where("id = ? and status in ?", task.ID, activePluginTableExpandTaskStatuses()).
			Updates(map[string]interface{}{
				"status":    2,
				"error_msg": "",
			}).Error; err != nil {
			return err
		}
		return nil
	}
	if err := db.Table(task.NewRealTable).AutoMigrate(&plugin.PluginData{}); err != nil {
		return err
	}
	if err := dbs.Transaction(func(tx *gorm.DB) error {
		tableUpdate := tx.Model(&plugin.PluginTable{}).
			Where("id = ? and need_expand in ?", task.PluginTableID, []uint{1, 3}).
			Update("need_expand", 3)
		if tableUpdate.Error != nil {
			return tableUpdate.Error
		}
		if tableUpdate.RowsAffected == 0 {
			return fmt.Errorf("plugin table is not expandable")
		}
		taskUpdate := tx.Model(&plugin.PluginTableExpandTask{}).
			Where("id = ? and status in ?", task.ID, activePluginTableExpandTaskStatuses()).
			Updates(map[string]interface{}{
				"status":        2,
				"migrated_rows": 0,
				"error_msg":     "",
			})
		if taskUpdate.Error != nil {
			return taskUpdate.Error
		}
		if taskUpdate.RowsAffected == 0 {
			return fmt.Errorf("plugin table expand task is not active")
		}
		return nil
	}); err != nil {
		return err
	}
	if err := db.Table(task.NewRealTable).
		Where("plugin_name = ? and plugin_table = ?", task.PluginName, task.PluginTable).
		Unscoped().
		Delete(&plugin.PluginData{}).Error; err != nil {
		return err
	}
	var migrated uint64
	var lastID uint
	for {
		var batch []plugin.PluginData
		if err := db.Table(task.OldRealTable).
			Where("plugin_name = ? and plugin_table = ? and id > ?", task.PluginName, task.PluginTable, lastID).
			Order("id asc").
			Limit(pluginTableExpandBatchSize).
			Find(&batch).Error; err != nil {
			return err
		}
		if len(batch) == 0 {
			break
		}
		lastID = batch[len(batch)-1].ID
		if err := db.Table(task.NewRealTable).CreateInBatches(batch, pluginTableExpandBatchSize).Error; err != nil {
			return err
		}
		migrated += uint64(len(batch))
		_ = db.Model(&plugin.PluginTableExpandTask{}).Where("id = ?", task.ID).Update("migrated_rows", migrated).Error
	}
	return dbs.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&plugin.PluginTable{}).Where("id = ? and need_expand = ?", task.PluginTableID, 3).Updates(map[string]interface{}{
			"real_table":  task.NewRealTable,
			"need_expand": 1,
			"data_length": migrated,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&plugin.PluginTableExpandTask{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
			"status":        2,
			"migrated_rows": migrated,
			"data_length":   migrated,
			"error_msg":     "",
		}).Error
	})
}

func markPluginTableExpandTaskFailed(task plugin.PluginTableExpandTask, err error) {
	db := ioc.Ioc().Get(ioc.KeyDatabase).(Database).Write()
	var currentTask plugin.PluginTableExpandTask
	if db.Where("id = ?", task.ID).First(&currentTask).Error == nil {
		task = currentTask
	}
	retryCount := task.RetryCount + 1
	status := task.Status
	if status == 0 {
		status = 1
	}
	if retryCount >= pluginTableExpandMaxRetry {
		status = 3
	}
	_ = db.Model(&plugin.PluginTableExpandTask{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
		"status":      status,
		"retry_count": retryCount,
		"error_msg":   truncateExpandError(err),
	}).Error
	if retryCount >= pluginTableExpandMaxRetry {
		_ = db.Model(&plugin.PluginTable{}).
			Where("id = ? and need_expand = ?", task.PluginTableID, 3).
			Update("need_expand", 1).Error
	}
}

func cleanupPluginTableExpand() {
	db := ioc.Ioc().Get(ioc.KeyDatabase).(Database).Write()
	var tasks []plugin.PluginTableExpandTask
	if err := db.Where("status = ?", 2).Order("old_real_table asc,id asc").Find(&tasks).Error; err != nil {
		return
	}
	for _, task := range tasks {
		var currentTable plugin.PluginTable
		if err := db.Where("id = ?", task.PluginTableID).First(&currentTable).Error; err != nil {
			continue
		}
		if currentTable.RealTable != task.NewRealTable || currentTable.NeedExpand != 1 {
			continue
		}
		if err := db.Table(task.OldRealTable).
			Where("plugin_name = ? and plugin_table = ?", task.PluginName, task.PluginTable).
			Unscoped().
			Delete(&plugin.PluginData{}).Error; err != nil {
			markPluginTableExpandTaskFailed(task, err)
			continue
		}
		_ = db.Model(&plugin.PluginTableExpandTask{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
			"status":    3,
			"error_msg": "",
		}).Error
	}
	ioc.Ioc().Get(ioc.KeyDatabase).(Database).ClearCache("plugin_data")
}
func truncateExpandError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	if len(msg) > 500 {
		return msg[:500]
	}
	return msg
}
func init() {
	ioc.Ioc().RegisterList(ioc.KeyTimerFunc, []interface{}{
		"plugin_table_expand_check",
		24 * 60,
		checkPluginTableExpand,
	})

	ioc.Ioc().RegisterList(ioc.KeyTimerFunc, []interface{}{
		"plugin_table_expand_migrate",
		60,
		migratePluginTableExpand,
	})
}
