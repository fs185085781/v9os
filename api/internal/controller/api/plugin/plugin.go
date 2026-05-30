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

type PluginController struct {
	*controller.BaseController
}

func init() {
	c := &PluginController{
		BaseController: controller.GetBaseController(),
	}
	c.RegisterApi("POST", "/plugin/page", c.Page, "基础表", "插件", "查看列表")
	c.RegisterApi("POST", "/plugin/save", c.Save, "基础表", "插件", "新增/编辑")
	c.RegisterApi("POST", "/plugin/detail", c.Detail, "基础表", "插件", "查看详情")
	c.RegisterApi("POST", "/plugin/delates", c.Deletes, "基础表", "插件", "删除")
	c.RegisterApi("POST", "/plugin/import", c.ImportXlsx, "基础表", "插件", "导入")
	c.RegisterApi("POST", "/plugin/export", c.ExportXlsx, "基础表", "插件", "导出")
}
func (c *PluginController) Page(ctx *gin.Context) {
	param := c.PageParam(ctx)
	db, err := c.pageParam(param)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	var total int64
	err = db.Model(&plugin.Plugin{}).Count(&total).Error
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	page := param.Page()
	pageSize := param.PageSize()
	var datas []plugin.Plugin
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

func (c *PluginController) pageParam(param controller.PageParam) (*gorm.DB, error) {
	db := c.Database().Read()
	if param.ParamString("keyword") != "" {
		db = db.Where("first_machine like ?", "%"+param.ParamString("keyword")+"%")
	}

	field2column := map[string]string{
		"FirstMachine": "first_machine",
		"RuntimeError": "runtime_error",
		"Name":         "name",
		"Description":  "description",
		"CloseDelay":   "close_delay",
		"Code":         "code",
		"Status":       "status",
		"Remark":       "remark",
		"Version":      "version",
		"PluginType":   "plugin_type",
		"WebHook":      "web_hook",
		"LimitVersion": "limit_version",
		"IconUrl":      "icon_url",
		"AccessUrl":    "access_url",
		"DebugPort":    "debug_port",
		"OpenExts":     "open_exts",
		"EditExts":     "edit_exts",
		"ExpandExts":   "expand_exts",
		"ID":           "id",
		"CreatedAt":    "created_at",
		"UpdatedAt":    "updated_at",
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
	return db, nil
}

func (c *PluginController) Save(ctx *gin.Context) {
	var data plugin.Plugin
	if err := ctx.ShouldBindBodyWithJSON(&data); err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	var err error
	if data.ID > 0 {
		err = c.Database().Update(&data)
	} else {
		err = c.Database().Create(&data)
	}

	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.OkData(ctx, data)
}

func (c *PluginController) Detail(ctx *gin.Context) {
	var data plugin.Plugin
	if err := ctx.ShouldBindBodyWithJSON(&data); err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	err := c.Database().GetByID(data.ID, &data)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.OkData(ctx, data)
}

func (c *PluginController) Deletes(ctx *gin.Context) {
	var ids []int
	if err := ctx.ShouldBindBodyWithJSON(&ids); err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	db := c.Database().Write()
	err := db.Model(&plugin.Plugin{}).Delete("id in ?", ids).Error
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.Ok(ctx)
}

func (c *PluginController) ImportXlsx(ctx *gin.Context) {
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
		{FieldName: "FirstMachine", CnName: c.GetText(ctx, "model.plugin.first_machine")},
		{FieldName: "RuntimeError", CnName: c.GetText(ctx, "model.plugin.runtime_error")},
		{FieldName: "Name", CnName: c.GetText(ctx, "model.plugin.name")},
		{FieldName: "Description", CnName: c.GetText(ctx, "model.plugin.description")},
		{FieldName: "CloseDelay", CnName: c.GetText(ctx, "model.plugin.close_delay")},
		{FieldName: "Code", CnName: c.GetText(ctx, "model.plugin.code")},
		{FieldName: "Status", CnName: c.GetText(ctx, "model.plugin.status")},
		{FieldName: "Remark", CnName: c.GetText(ctx, "model.plugin.remark")},
		{FieldName: "Version", CnName: c.GetText(ctx, "model.plugin.version")},
		{FieldName: "PluginType", CnName: c.GetText(ctx, "model.plugin.plugin_type")},
		{FieldName: "WebHook", CnName: c.GetText(ctx, "model.plugin.web_hook")},
		{FieldName: "LimitVersion", CnName: c.GetText(ctx, "model.plugin.limit_version")},
		{FieldName: "IconUrl", CnName: c.GetText(ctx, "model.plugin.icon_url")},
		{FieldName: "AccessUrl", CnName: c.GetText(ctx, "model.plugin.access_url")},
		{FieldName: "DebugPort", CnName: c.GetText(ctx, "model.plugin.debug_port")},
		{FieldName: "OpenExts", CnName: c.GetText(ctx, "model.plugin.open_exts")},
		{FieldName: "EditExts", CnName: c.GetText(ctx, "model.plugin.edit_exts")},
		{FieldName: "ExpandExts", CnName: c.GetText(ctx, "model.plugin.expand_exts")},
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
				tx2 = tx.Model(&plugin.Plugin{}).Create(&v)
				if tx2.RowsAffected > 0 {
					create++
				}
			} else {
				v["UpdatedAt"] = util.UnixMilliseconds()
				tx2 = tx.Model(&plugin.Plugin{}).Where("id = ?", v["ID"]).Save(&v)
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

func (c *PluginController) ExportXlsx(ctx *gin.Context) {
	param := c.PageParam(ctx)
	c.exportXlsx(param, ctx)
}

func (c *PluginController) exportXlsx(param controller.PageParam, ctx *gin.Context) {
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	fileName := fmt.Sprintf("%s%s.xlsx", c.GetText(ctx, "model.plugin.model"), c.GetText(ctx, "common.export"))
	encodedFileName := url.QueryEscape(fileName)
	ctx.Header("Content-Disposition", `attachment; filename="`+fileName+`"; filename*=UTF-8''`+encodedFileName)
	ctx.Header("File-Name", encodedFileName)
	ctx.Header("Content-Transfer-Encoding", "binary")
	headers := []util.ExportField{
		{FieldName: "FirstMachine", CnName: c.GetText(ctx, "model.plugin.first_machine"), FieldType: "input"},
		{FieldName: "RuntimeError", CnName: c.GetText(ctx, "model.plugin.runtime_error"), FieldType: "input"},
		{FieldName: "Name", CnName: c.GetText(ctx, "model.plugin.name"), FieldType: "input"},
		{FieldName: "Description", CnName: c.GetText(ctx, "model.plugin.description"), FieldType: "textarea"},
		{FieldName: "CloseDelay", CnName: c.GetText(ctx, "model.plugin.close_delay"), FieldType: "input"},
		{FieldName: "Code", CnName: c.GetText(ctx, "model.plugin.code"), FieldType: "input"},
		{FieldName: "Status", CnName: c.GetText(ctx, "model.plugin.status"), FieldType: "select", Selects: map[string]string{
			"0": c.GetText(ctx, "model.plugin.status_select_0"),
			"1": c.GetText(ctx, "model.plugin.status_select_1"),
		}},
		{FieldName: "Remark", CnName: c.GetText(ctx, "model.plugin.remark"), FieldType: "input"},
		{FieldName: "Version", CnName: c.GetText(ctx, "model.plugin.version"), FieldType: "input"},
		{FieldName: "PluginType", CnName: c.GetText(ctx, "model.plugin.plugin_type"), FieldType: "select", Selects: map[string]string{
			"1": c.GetText(ctx, "model.plugin.plugin_type_select_1"),
			"2": c.GetText(ctx, "model.plugin.plugin_type_select_2"),
			"3": c.GetText(ctx, "model.plugin.plugin_type_select_3"),
			"4": c.GetText(ctx, "model.plugin.plugin_type_select_4"),
		}},
		{FieldName: "WebHook", CnName: c.GetText(ctx, "model.plugin.web_hook"), FieldType: "input"},
		{FieldName: "LimitVersion", CnName: c.GetText(ctx, "model.plugin.limit_version"), FieldType: "input"},
		{FieldName: "IconUrl", CnName: c.GetText(ctx, "model.plugin.icon_url"), FieldType: "input"},
		{FieldName: "AccessUrl", CnName: c.GetText(ctx, "model.plugin.access_url"), FieldType: "input"},
		{FieldName: "DebugPort", CnName: c.GetText(ctx, "model.plugin.debug_port"), FieldType: "input"},
		{FieldName: "OpenExts", CnName: c.GetText(ctx, "model.plugin.open_exts"), FieldType: "input"},
		{FieldName: "EditExts", CnName: c.GetText(ctx, "model.plugin.edit_exts"), FieldType: "input"},
		{FieldName: "ExpandExts", CnName: c.GetText(ctx, "model.plugin.expand_exts"), FieldType: "input"},
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
		err = util.ExportXlsx(headers, ctx.Writer, func() ([]plugin.Plugin, bool, error) {
			var datas []plugin.Plugin
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
