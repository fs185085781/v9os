package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

// 实现虚拟的Gorm
type GormDb struct {
	ActionSteps  []GormActionStep
	RowsAffected int64
	TxId         string
	Table        string
}
type GormActionStep struct {
	Func string        //调用的方法
	Args []interface{} //参数
}

func Db(table string) *GormDb {
	return &GormDb{TxId: "", Table: table, ActionSteps: make([]GormActionStep, 0)}
}

func cloneAndAddStep(d *GormDb, method string, args []interface{}) *GormDb {
	steps := make([]GormActionStep, 0)
	if d.ActionSteps != nil {
		steps = append(steps, d.ActionSteps...)
	}
	steps = append(steps, GormActionStep{
		Func: method,
		Args: args,
	})
	return &GormDb{TxId: d.TxId, Table: d.Table, ActionSteps: steps}
}
func cloneAndRun(d *GormDb, method string, dest interface{}, args []interface{}) error {
	tx := cloneAndAddStep(d, method, args)
	resultType := "none"
	if dest != nil {
		destValue := reflect.ValueOf(dest)
		if destValue.Kind() != reflect.Ptr {
			return fmt.Errorf("dest must be a pointer")
		}
		destElem := destValue.Elem()
		switch destElem.Kind() {
		case reflect.Slice:
			resultType = "slice"
		case reflect.Struct:
			resultType = "single"
		}
	}
	v, rs, err := httpGormRun(tx.TxId, "SqlRun", resultType, tx.Table, tx.ActionSteps)
	tx.RowsAffected = rs
	if err != nil {
		return err
	}
	if v != nil && dest != nil {
		b, err2 := json.Marshal(v)
		if err2 != nil {
			return err2
		}
		return json.Unmarshal(b, dest)
	}
	return nil
}

// 链式调用方法
func (d *GormDb) Where(query interface{}, args ...interface{}) *GormDb {
	return cloneAndAddStep(d, "Where", []interface{}{query, args})
}
func (d *GormDb) Select(query interface{}, args ...interface{}) *GormDb {
	return cloneAndAddStep(d, "Select", []interface{}{query, args})
}
func (d *GormDb) Group(name string) *GormDb {
	return cloneAndAddStep(d, "Group", []interface{}{name})
}
func (d *GormDb) Having(query interface{}, args ...interface{}) *GormDb {
	return cloneAndAddStep(d, "Having", []interface{}{query, args})
}
func (d *GormDb) Order(value interface{}) *GormDb {
	return cloneAndAddStep(d, "Order", []interface{}{value})
}
func (d *GormDb) Limit(limit int) *GormDb {
	return cloneAndAddStep(d, "Limit", []interface{}{limit})
}
func (d *GormDb) Offset(offset int) *GormDb {
	return cloneAndAddStep(d, "Offset", []interface{}{offset})
}
func (d *GormDb) Distinct(args ...interface{}) *GormDb {
	return cloneAndAddStep(d, "Distinct", []interface{}{args})
}
func (d *GormDb) Session(table string) *GormDb {
	return &GormDb{TxId: d.TxId, Table: table, ActionSteps: make([]GormActionStep, 0)}
}

// 结束调用,不返回结果
func (d *GormDb) Create(value interface{}) error {
	return cloneAndRun(d, "Create", nil, []interface{}{value})
}
func (d *GormDb) Save(value interface{}) error {
	return cloneAndRun(d, "Save", nil, []interface{}{value})
}
func (d *GormDb) Update(column string, value interface{}) error {
	return cloneAndRun(d, "Update", nil, []interface{}{column, value})
}
func (d *GormDb) Updates(values interface{}) error {
	return cloneAndRun(d, "Updates", nil, []interface{}{values})
}
func (d *GormDb) Delete(value interface{}, conds ...interface{}) error {
	return cloneAndRun(d, "Delete", nil, []interface{}{value, conds})
}

// 结束调用,返回结果
func (d *GormDb) First(dest interface{}, conds ...interface{}) error {
	return cloneAndRun(d, "First", dest, []interface{}{conds})
}
func (d *GormDb) Take(dest interface{}, conds ...interface{}) error {
	return cloneAndRun(d, "Take", dest, []interface{}{conds})
}
func (d *GormDb) Last(dest interface{}, conds ...interface{}) error {
	return cloneAndRun(d, "Last", dest, []interface{}{conds})
}
func (d *GormDb) Find(dest interface{}, conds ...interface{}) error {
	return cloneAndRun(d, "Find", dest, []interface{}{conds})
}
func (d *GormDb) Count(count *int64) error {
	return cloneAndRun(d, "Count", count, []interface{}{})
}

// 事务调用
func Transaction(fc func(tx *GormDb) error) error {
	txId := uuid.New().String()
	//http开启事务
	_, _, err := httpGormRun(txId, "StartTransaction", "none", "", nil)
	if err != nil {
		return err
	}
	err = fc(&GormDb{TxId: txId, ActionSteps: make([]GormActionStep, 0)})
	if err != nil {
		//http回滚事务
		_, _, err2 := httpGormRun(txId, "RollbackTransaction", "none", "", nil)
		if err2 != nil {
			return err2
		}
	} else {
		//http提交事务
		_, _, err2 := httpGormRun(txId, "CommitTransaction", "none", "", nil)
		if err2 != nil {
			return err2
		}
	}
	return err
}
func httpGormRun(txId, method, resultType, table string, steps []GormActionStep) (interface{}, int64, error) {
	data := make(map[string]interface{})
	data["txId"] = txId
	data["method"] = method
	data["resultType"] = resultType
	data["steps"] = steps
	data["table"] = table
	resultMap, err := httpPost("/gorm/bridge", data)
	if err != nil {
		return nil, 0, err
	}
	var resErr error
	rerr := resultMap["error"]
	if rerr != nil {
		rerrStr := cast.ToString(rerr)
		if rerrStr != "" {
			resErr = errors.New(rerrStr)
		}
	}
	return resultMap["data"], cast.ToInt64(resultMap["rowsAffected"]), resErr
}

type BaseModel struct {
	ID        string
	CreatedAt uint64
	UpdatedAt uint64
	DeletedAt gorm.DeletedAt
}

func getStructFieldsInfo(instance interface{}) []map[string]string {
	t := reflect.TypeOf(instance).Elem()
	var fieldsInfo []map[string]string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name == "BaseModel" || field.Type.Name() == "BaseModel" {
			continue
		}
		gormTag := field.Tag.Get("gorm")
		columnName := ""
		index := "false"
		if gormTag != "" {
			tags := strings.Split(gormTag, ";")
			for _, tag := range tags {
				if strings.HasPrefix(tag, "column:") {
					columnName = strings.TrimPrefix(tag, "column:")
					break
				}
				if strings.HasPrefix(tag, "index") {
					index = "true"
					break
				}
			}
		}
		if columnName == "" {
			columnName = strings.ToLower(field.Name)
		}
		fieldInfo := map[string]string{
			"column": columnName,
			"field":  field.Name,
			"type":   field.Type.String(),
			"index":  index,
		}
		fieldsInfo = append(fieldsInfo, fieldInfo)
	}
	return fieldsInfo
}

func BindStruct(entity interface{}, table string) error {
	data := make(map[string]interface{})
	data["Fields"] = getStructFieldsInfo(entity)
	data["Table"] = table
	res, err := httpPost("/gorm/bind", data)
	if err != nil {
		return err
	}
	if res["code"] != 0 {
		return errors.New(cast.ToString(res["msg"]))
	}
	return nil
}

func AutoMigrate() {
	for _, info := range modelList {
		BindStruct(info.Model, info.TableName)
	}
}
