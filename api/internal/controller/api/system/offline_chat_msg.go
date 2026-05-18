package api

import (
	"fmt"
	"net/http"
	"net/url"
	"github.com/fs185085781/v9os/internal/controller"
	"github.com/fs185085781/v9os/internal/model/system"
	"github.com/fs185085781/v9os/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type OfflineChatMsgController struct {
	*controller.BaseController
}

func init() {
	c := &OfflineChatMsgController{
		BaseController: controller.GetBaseController(),
	}
	c.RegisterApi("POST", "/offline_chat_msg/page", c.Page, "基础表", "离线聊天消息", "查看列表")
	c.RegisterApi("POST", "/offline_chat_msg/save", c.Save, "基础表", "离线聊天消息", "新增/编辑")
	c.RegisterApi("POST", "/offline_chat_msg/detail", c.Detail, "基础表", "离线聊天消息", "查看详情")
	c.RegisterApi("POST", "/offline_chat_msg/delates", c.Deletes, "基础表", "离线聊天消息", "删除")
	c.RegisterApi("POST", "/offline_chat_msg/import", c.ImportXlsx, "基础表", "离线聊天消息", "导入")
	c.RegisterApi("POST", "/offline_chat_msg/export", c.ExportXlsx, "基础表", "离线聊天消息", "导出")
}
func (c *OfflineChatMsgController) Page(ctx *gin.Context) {
	param := c.PageParam(ctx)
	db, err := c.pageParam(param)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	var total int64
	err = db.Model(&system.OfflineChatMsg{}).Count(&total).Error
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	page := param.Page()
	pageSize := param.PageSize()
	var datas []system.OfflineChatMsg
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

func (c *OfflineChatMsgController) pageParam(param controller.PageParam) (*gorm.DB, error) {
	db := c.Database().Read()
	if param.ParamString("keyword") != "" {
		db = db.Where("from like ?", "%"+param.ParamString("keyword")+"%")
	}
	
	field2column := map[string]string{
		"From": "from",
		"To": "to",
		"Msg": "msg",
		"Type": "type",
		"DateTime": "date_time",
		"ID": "id",
		"CreatedAt": "created_at",
		"UpdatedAt": "updated_at",
		
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

func (c *OfflineChatMsgController) Save(ctx *gin.Context) {
	var data system.OfflineChatMsg
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

func (c *OfflineChatMsgController) Detail(ctx *gin.Context) {
	var data system.OfflineChatMsg
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

func (c *OfflineChatMsgController) Deletes(ctx *gin.Context) {
	var ids []int
	if err := ctx.ShouldBindBodyWithJSON(&ids); err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	db := c.Database().Write()
	err := db.Model(&system.OfflineChatMsg{}).Delete("id in ?", ids).Error
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.Ok(ctx)
}

func (c *OfflineChatMsgController) ImportXlsx(ctx *gin.Context) {
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
		{FieldName: "From", CnName: c.GetText(ctx, "model.offline_chat_msg.from")},
		{FieldName: "To", CnName: c.GetText(ctx, "model.offline_chat_msg.to")},
		{FieldName: "Msg", CnName: c.GetText(ctx, "model.offline_chat_msg.msg")},
		{FieldName: "Type", CnName: c.GetText(ctx, "model.offline_chat_msg.type")},
		{FieldName: "DateTime", CnName: c.GetText(ctx, "model.offline_chat_msg.date_time")},
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
				tx2 = tx.Model(&system.OfflineChatMsg{}).Create(&v)
				if tx2.RowsAffected > 0 {
					create++
				}
			} else {
				v["UpdatedAt"] = util.UnixMilliseconds()
				tx2 = tx.Model(&system.OfflineChatMsg{}).Where("id = ?", v["ID"]).Save(&v)
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

func (c *OfflineChatMsgController) ExportXlsx(ctx *gin.Context) {
	param := c.PageParam(ctx)
	c.exportXlsx(param, ctx)
}

func (c *OfflineChatMsgController) exportXlsx(param controller.PageParam,ctx *gin.Context) {
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	fileName := fmt.Sprintf("%s%s.xlsx", c.GetText(ctx, "model.offline_chat_msg.model"), c.GetText(ctx, "common.export"))
	encodedFileName := url.QueryEscape(fileName)
	ctx.Header("Content-Disposition", `attachment; filename="`+fileName+`"; filename*=UTF-8''`+encodedFileName)
	ctx.Header("File-Name", encodedFileName)
	ctx.Header("Content-Transfer-Encoding", "binary")
	headers := []util.ExportField{
		{FieldName: "From", CnName: c.GetText(ctx, "model.offline_chat_msg.from"), FieldType: "input"},
		{FieldName: "To", CnName: c.GetText(ctx, "model.offline_chat_msg.to"), FieldType: "input"},
		{FieldName: "Msg", CnName: c.GetText(ctx, "model.offline_chat_msg.msg"), FieldType: "input"},
		{FieldName: "Type", CnName: c.GetText(ctx, "model.offline_chat_msg.type"), FieldType: "input"},
		{FieldName: "DateTime", CnName: c.GetText(ctx, "model.offline_chat_msg.date_time"), FieldType: "datetime"},
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
		err = util.ExportXlsx(headers, ctx.Writer, func() ([]system.OfflineChatMsg, bool, error) {
			var datas []system.OfflineChatMsg
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
