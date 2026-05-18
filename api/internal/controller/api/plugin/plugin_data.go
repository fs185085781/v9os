package api

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/fs185085781/v9os/internal/controller"
	"github.com/fs185085781/v9os/internal/model/plugin"
	"github.com/fs185085781/v9os/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type PluginDataController struct {
	*controller.BaseController
}

func init() {
	c := &PluginDataController{
		BaseController: controller.GetBaseController(),
	}
	c.RegisterApi("POST", "/plugin_data/page", c.Page, "基础表", "插件数据", "查看列表")
	c.RegisterApi("POST", "/plugin_data/save", c.Save, "基础表", "插件数据", "新增/编辑")
	c.RegisterApi("POST", "/plugin_data/detail", c.Detail, "基础表", "插件数据", "查看详情")
	c.RegisterApi("POST", "/plugin_data/delates", c.Deletes, "基础表", "插件数据", "删除")
	c.RegisterApi("POST", "/plugin_data/import", c.ImportXlsx, "基础表", "插件数据", "导入")
	c.RegisterApi("POST", "/plugin_data/export", c.ExportXlsx, "基础表", "插件数据", "导出")
	c.RegisterApi("GET", "/plugin_data/tables", c.Tables, "基础表", "插件数据", "表格数据")
}
func (c *PluginDataController) Tables(ctx *gin.Context) {
	var tbs []plugin.PluginTable
	c.Database().Read().Distinct("real_table").Find(&tbs)
	c.OkData(ctx, len(tbs))
}
func (c *PluginDataController) Page(ctx *gin.Context) {
	param := c.PageParam(ctx)
	db, err := c.pageParam(param)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	var total int64
	err = db.Model(&plugin.PluginData{}).Count(&total).Error
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	page := param.Page()
	pageSize := param.PageSize()
	var datas []plugin.PluginData
	err = db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&datas).Error
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	data := &controller.PageRes{}
	data.Data = datas
	data.Total = total
	c.OkData(ctx, data)
}

func (c *PluginDataController) pageParam(param controller.PageParam) (*gorm.DB, error) {
	db := c.Database().Read()
	if param.ParamString("keyword") != "" {
		db = db.Where("plugin_name like ?", "%"+param.ParamString("keyword")+"%")
	}

	field2column := map[string]string{
		"PluginName":  "plugin_name",
		"PluginTable": "plugin_table",
		"DataId":      "data_id",
		"UserId":      "user_id",
		"DeptId":      "dept_id",
		"Field1":      "field1",
		"Field2":      "field2",
		"Field3":      "field3",
		"Field4":      "field4",
		"Field5":      "field5",
		"Field6":      "field6",
		"Field7":      "field7",
		"Field8":      "field8",
		"Field9":      "field9",
		"Field10":     "field10",
		"Field11":     "field11",
		"Field12":     "field12",
		"Field13":     "field13",
		"Field14":     "field14",
		"Field15":     "field15",
		"Field16":     "field16",
		"Field17":     "field17",
		"Field18":     "field18",
		"Field19":     "field19",
		"Field20":     "field20",
		"TextField1":  "text_field1",
		"TextField2":  "text_field2",
		"TextField3":  "text_field3",
		"TextField4":  "text_field4",
		"TextField5":  "text_field5",
		"IndexField1": "index_field1",
		"IndexField2": "index_field2",
		"IndexField3": "index_field3",
		"IndexField4": "index_field4",
		"IndexField5": "index_field5",
		"ID":          "id",
		"CreatedAt":   "created_at",
		"UpdatedAt":   "updated_at",
	}
	for _, v := range param.Sorter() {
		o := v.Order()
		if o == "false" {
			continue
		}
		column := field2column[v.ColumnKey()]
		if o == "descend" {
			db = db.Order(column + " desc")
		} else {
			db = db.Order(column + " asc")
		}
	}
	return db.Table(param.ParamString("table")), nil
}

func (c *PluginDataController) Save(ctx *gin.Context) {
	table := ctx.Query("table")
	var data plugin.PluginData
	if err := ctx.ShouldBindBodyWithJSON(&data); err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	db := c.Database().Write().Table(table)
	var err error
	if data.ID > 0 {
		err = db.Select("*").Updates(&data).Error
	} else {
		err = db.Create(&data).Error
	}
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.OkData(ctx, data)
}

func (c *PluginDataController) Detail(ctx *gin.Context) {
	var data plugin.PluginData
	table := ctx.Query("table")
	if err := ctx.ShouldBindBodyWithJSON(&data); err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	err := c.Database().Read().Table(table).First(&data, data.ID).Error
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.OkData(ctx, data)
}

func (c *PluginDataController) Deletes(ctx *gin.Context) {
	table := ctx.Query("table")
	var ids []int
	if err := ctx.ShouldBindBodyWithJSON(&ids); err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	db := c.Database().Write()
	err := db.Table(table).Model(&plugin.PluginData{}).Delete("id in ?", ids).Error
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.Ok(ctx)
}

func (c *PluginDataController) ImportXlsx(ctx *gin.Context) {
	table := ctx.Query("table")
	file, err := ctx.FormFile("file")
	if err != nil {
		param := c.PageParam(ctx)
		if param.ParamBool("isTemplate") {
			c.exportXlsx(param, ctx)
			return
		}
		c.ErrMsg(ctx, err)
		return
	}
	srcFile, err := file.Open()
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	defer srcFile.Close()
	f, err := excelize.OpenReader(srcFile)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	defer f.Close()
	sheets := f.GetSheetMap()
	if len(sheets) == 0 {
		c.FailMsg(ctx, c.GetText(ctx, "common.execlnosheet"))
		return
	}
	firstSheetName := sheets[1]
	rows, err := f.GetRows(firstSheetName)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}

	if len(rows) == 0 {
		c.FailMsg(ctx, c.GetText(ctx, "common.nosheet"))
		return
	}
	headers := []util.ExportField{
		{FieldName: "PluginName", CnName: c.GetText(ctx, "model.plugin_data.plugin_name")},
		{FieldName: "PluginTable", CnName: c.GetText(ctx, "model.plugin_data.plugin_table")},
		{FieldName: "DataId", CnName: c.GetText(ctx, "model.plugin_data.data_id")},
		{FieldName: "UserId", CnName: c.GetText(ctx, "model.plugin_data.user_id")},
		{FieldName: "DeptId", CnName: c.GetText(ctx, "model.plugin_data.dept_id")},
		{FieldName: "Field1", CnName: c.GetText(ctx, "model.plugin_data.field1")},
		{FieldName: "Field2", CnName: c.GetText(ctx, "model.plugin_data.field2")},
		{FieldName: "Field3", CnName: c.GetText(ctx, "model.plugin_data.field3")},
		{FieldName: "Field4", CnName: c.GetText(ctx, "model.plugin_data.field4")},
		{FieldName: "Field5", CnName: c.GetText(ctx, "model.plugin_data.field5")},
		{FieldName: "Field6", CnName: c.GetText(ctx, "model.plugin_data.field6")},
		{FieldName: "Field7", CnName: c.GetText(ctx, "model.plugin_data.field7")},
		{FieldName: "Field8", CnName: c.GetText(ctx, "model.plugin_data.field8")},
		{FieldName: "Field9", CnName: c.GetText(ctx, "model.plugin_data.field9")},
		{FieldName: "Field10", CnName: c.GetText(ctx, "model.plugin_data.field10")},
		{FieldName: "Field11", CnName: c.GetText(ctx, "model.plugin_data.field11")},
		{FieldName: "Field12", CnName: c.GetText(ctx, "model.plugin_data.field12")},
		{FieldName: "Field13", CnName: c.GetText(ctx, "model.plugin_data.field13")},
		{FieldName: "Field14", CnName: c.GetText(ctx, "model.plugin_data.field14")},
		{FieldName: "Field15", CnName: c.GetText(ctx, "model.plugin_data.field15")},
		{FieldName: "Field16", CnName: c.GetText(ctx, "model.plugin_data.field16")},
		{FieldName: "Field17", CnName: c.GetText(ctx, "model.plugin_data.field17")},
		{FieldName: "Field18", CnName: c.GetText(ctx, "model.plugin_data.field18")},
		{FieldName: "Field19", CnName: c.GetText(ctx, "model.plugin_data.field19")},
		{FieldName: "Field20", CnName: c.GetText(ctx, "model.plugin_data.field20")},
		{FieldName: "TextField1", CnName: c.GetText(ctx, "model.plugin_data.text_field1")},
		{FieldName: "TextField2", CnName: c.GetText(ctx, "model.plugin_data.text_field2")},
		{FieldName: "TextField3", CnName: c.GetText(ctx, "model.plugin_data.text_field3")},
		{FieldName: "TextField4", CnName: c.GetText(ctx, "model.plugin_data.text_field4")},
		{FieldName: "TextField5", CnName: c.GetText(ctx, "model.plugin_data.text_field5")},
		{FieldName: "IndexField1", CnName: c.GetText(ctx, "model.plugin_data.index_field1")},
		{FieldName: "IndexField2", CnName: c.GetText(ctx, "model.plugin_data.index_field2")},
		{FieldName: "IndexField3", CnName: c.GetText(ctx, "model.plugin_data.index_field3")},
		{FieldName: "IndexField4", CnName: c.GetText(ctx, "model.plugin_data.index_field4")},
		{FieldName: "IndexField5", CnName: c.GetText(ctx, "model.plugin_data.index_field5")},
		{FieldName: "ID", CnName: c.GetText(ctx, "model.common.id")},
		{FieldName: "CreatedAt", CnName: c.GetText(ctx, "model.common.createdat")},
		{FieldName: "UpdatedAt", CnName: c.GetText(ctx, "model.common.updatedat")},
	}
	cnNameToFieldMap := make(map[string]string)
	for _, header := range headers {
		cnNameToFieldMap[header.CnName] = header.FieldName
	}
	columnMapping := make(map[int]string)
	headerRow := rows[0]
	for colIndex, cnName := range headerRow {
		if fieldName, exists := cnNameToFieldMap[cnName]; exists {
			columnMapping[colIndex] = fieldName
		}
	}
	var result []map[string]interface{}
	for rowIndex := 1; rowIndex < len(rows); rowIndex++ {
		row := rows[rowIndex]
		rowData := make(map[string]interface{})
		for colIndex, cellValue := range row {
			if fieldName, exists := columnMapping[colIndex]; exists {
				rowData[fieldName] = cellValue
			}
		}
		if len(rowData) > 0 {
			result = append(result, rowData)
		}
	}
	create := 0
	update := 0
	err = c.Database().Transaction(func(tx *gorm.DB) error {
		for _, v := range result {
			var tx2 *gorm.DB
			if v["ID"] == nil {
				v["CreatedAt"] = util.UnixMilliseconds()
				v["UpdatedAt"] = util.UnixMilliseconds()
				tx2 = tx.Table(table).Model(&plugin.PluginData{}).Create(&v)
				if tx2.RowsAffected > 0 {
					create++
				}
			} else {
				v["UpdatedAt"] = util.UnixMilliseconds()
				tx2 = tx.Table(table).Model(&plugin.PluginData{}).Where("id = ?", v["ID"]).Save(&v)
				if tx2.RowsAffected > 0 {
					update++
				}
			}
			if tx2.Error != nil {
				return tx2.Error
			}
		}
		return nil
	})
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.OkMsg(ctx, fmt.Sprintf(c.GetText(ctx, "common.importsuccess"), create, update))
}

func (c *PluginDataController) ExportXlsx(ctx *gin.Context) {
	param := c.PageParam(ctx)
	c.exportXlsx(param, ctx)
}

func (c *PluginDataController) exportXlsx(param controller.PageParam, ctx *gin.Context) {
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	fileName := fmt.Sprintf("%s%s.xlsx", c.GetText(ctx, "model.plugin_data.model"), c.GetText(ctx, "common.export"))
	encodedFileName := url.QueryEscape(fileName)
	ctx.Header("Content-Disposition", `attachment; filename="`+fileName+`"; filename*=UTF-8''`+encodedFileName)
	ctx.Header("File-Name", encodedFileName)
	ctx.Header("Content-Transfer-Encoding", "binary")
	headers := []util.ExportField{
		{FieldName: "PluginName", CnName: c.GetText(ctx, "model.plugin_data.plugin_name"), FieldType: "input"},
		{FieldName: "PluginTable", CnName: c.GetText(ctx, "model.plugin_data.plugin_table"), FieldType: "input"},
		{FieldName: "DataId", CnName: c.GetText(ctx, "model.plugin_data.data_id"), FieldType: "input"},
		{FieldName: "UserId", CnName: c.GetText(ctx, "model.plugin_data.user_id"), FieldType: "input"},
		{FieldName: "DeptId", CnName: c.GetText(ctx, "model.plugin_data.dept_id"), FieldType: "input"},
		{FieldName: "Field1", CnName: c.GetText(ctx, "model.plugin_data.field1"), FieldType: "input"},
		{FieldName: "Field2", CnName: c.GetText(ctx, "model.plugin_data.field2"), FieldType: "input"},
		{FieldName: "Field3", CnName: c.GetText(ctx, "model.plugin_data.field3"), FieldType: "input"},
		{FieldName: "Field4", CnName: c.GetText(ctx, "model.plugin_data.field4"), FieldType: "input"},
		{FieldName: "Field5", CnName: c.GetText(ctx, "model.plugin_data.field5"), FieldType: "input"},
		{FieldName: "Field6", CnName: c.GetText(ctx, "model.plugin_data.field6"), FieldType: "input"},
		{FieldName: "Field7", CnName: c.GetText(ctx, "model.plugin_data.field7"), FieldType: "input"},
		{FieldName: "Field8", CnName: c.GetText(ctx, "model.plugin_data.field8"), FieldType: "input"},
		{FieldName: "Field9", CnName: c.GetText(ctx, "model.plugin_data.field9"), FieldType: "input"},
		{FieldName: "Field10", CnName: c.GetText(ctx, "model.plugin_data.field10"), FieldType: "input"},
		{FieldName: "Field11", CnName: c.GetText(ctx, "model.plugin_data.field11"), FieldType: "input"},
		{FieldName: "Field12", CnName: c.GetText(ctx, "model.plugin_data.field12"), FieldType: "input"},
		{FieldName: "Field13", CnName: c.GetText(ctx, "model.plugin_data.field13"), FieldType: "input"},
		{FieldName: "Field14", CnName: c.GetText(ctx, "model.plugin_data.field14"), FieldType: "input"},
		{FieldName: "Field15", CnName: c.GetText(ctx, "model.plugin_data.field15"), FieldType: "input"},
		{FieldName: "Field16", CnName: c.GetText(ctx, "model.plugin_data.field16"), FieldType: "input"},
		{FieldName: "Field17", CnName: c.GetText(ctx, "model.plugin_data.field17"), FieldType: "input"},
		{FieldName: "Field18", CnName: c.GetText(ctx, "model.plugin_data.field18"), FieldType: "input"},
		{FieldName: "Field19", CnName: c.GetText(ctx, "model.plugin_data.field19"), FieldType: "input"},
		{FieldName: "Field20", CnName: c.GetText(ctx, "model.plugin_data.field20"), FieldType: "input"},
		{FieldName: "TextField1", CnName: c.GetText(ctx, "model.plugin_data.text_field1"), FieldType: "input"},
		{FieldName: "TextField2", CnName: c.GetText(ctx, "model.plugin_data.text_field2"), FieldType: "input"},
		{FieldName: "TextField3", CnName: c.GetText(ctx, "model.plugin_data.text_field3"), FieldType: "input"},
		{FieldName: "TextField4", CnName: c.GetText(ctx, "model.plugin_data.text_field4"), FieldType: "input"},
		{FieldName: "TextField5", CnName: c.GetText(ctx, "model.plugin_data.text_field5"), FieldType: "input"},
		{FieldName: "IndexField1", CnName: c.GetText(ctx, "model.plugin_data.index_field1"), FieldType: "input"},
		{FieldName: "IndexField2", CnName: c.GetText(ctx, "model.plugin_data.index_field2"), FieldType: "input"},
		{FieldName: "IndexField3", CnName: c.GetText(ctx, "model.plugin_data.index_field3"), FieldType: "input"},
		{FieldName: "IndexField4", CnName: c.GetText(ctx, "model.plugin_data.index_field4"), FieldType: "input"},
		{FieldName: "IndexField5", CnName: c.GetText(ctx, "model.plugin_data.index_field5"), FieldType: "input"},
		{FieldName: "ID", CnName: c.GetText(ctx, "model.common.id"), FieldType: "input"},
		{FieldName: "CreatedAt", CnName: c.GetText(ctx, "model.common.createdat"), FieldType: "datetime"},
		{FieldName: "UpdatedAt", CnName: c.GetText(ctx, "model.common.updatedat"), FieldType: "datetime"},
	}
	var err error
	if param.ParamBool("isTemplate") {
		err = util.ExportXlsx(headers, ctx.Writer, func() ([]interface{}, bool, error) {
			return nil, false, nil
		})
	} else {
		db, err2 := c.pageParam(param)
		if err2 != nil {
			c.ErrCode(ctx, 500, err2.Error())
			return
		}
		pageSize := 1000
		page := 1
		err = util.ExportXlsx(headers, ctx.Writer, func() ([]plugin.PluginData, bool, error) {
			var datas []plugin.PluginData
			offset := (page - 1) * pageSize
			err = db.Offset(offset).Limit(pageSize).Find(&datas).Error
			if err2 != nil {
				return nil, false, err2
			}
			page++
			if len(datas) == 0 || len(datas) < pageSize {
				return datas, false, nil
			}
			return datas, true, nil
		})
	}
	if err != nil {
		c.ErrCode(ctx, 500, err.Error())
		return
	}
	ctx.Status(http.StatusOK)
}
