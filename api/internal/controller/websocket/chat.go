package websocket

import (
	"encoding/json"
	"time"

	"github.com/fs185085781/v9os/internal/controller"
	"github.com/fs185085781/v9os/internal/model/system"
	"github.com/fs185085781/v9os/internal/plugin"
	"github.com/fs185085781/v9os/internal/queue"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

func init() {
	c := &ChatController{
		BaseController: controller.GetBaseController(),
	}
	c.RegisterWebsocket(c)
	//获取系统通知
	c.RegisterApi("POST", "/chat/notices", c.Notices)
	//删除系统通知
	c.RegisterApi("POST", "/chat/delete_notices", c.DeleteNotices)
}

type ChatController struct {
	*controller.BaseController
}

func (h *ChatController) ChannelName() string {
	return "chat"
}

func (h *ChatController) OnOpen(client *plugin.WSClient) {

}

func (h *ChatController) OnMessage(client *plugin.WSClient, message []byte) {
	var msg plugin.WebsocketMessage
	err := json.Unmarshal(message, &msg)
	if err != nil {
		return
	}
	msg.DateTime = time.Now()
	if msg.From == "" {
		msg.From = client.UserID
	}
	_ = h.BaseController.Queue().Publish(&queue.Message{
		EventType: "wsevent:" + h.ChannelName(),
		Data:      msg,
	})
}

func (h *ChatController) OnClose(client *plugin.WSClient) {

}

func (h *ChatController) Notices(ctx *gin.Context) {
	var param struct {
		Page     int `json:"page"`
		PageSize int `json:"pageSize"`
	}
	_ = ctx.ShouldBindBodyWithJSON(&param)
	if param.Page < 1 {
		param.Page = 1
	}
	if param.PageSize < 1 {
		param.PageSize = 100
	}
	var rows []system.OfflineChatMsg
	err := h.Database().Read().
		Where("to = ? AND type != ?", ctx.GetString("userID"), "chat_message").
		Order("id desc").
		Offset((param.Page - 1) * param.PageSize).
		Limit(param.PageSize).
		Find(&rows).Error
	if err != nil {
		h.ErrMsg(ctx, err)
		return
	}
	h.OkData(ctx, rows)
}

func (h *ChatController) DeleteNotices(ctx *gin.Context) {
	var param struct {
		IDs []uint `json:"ids"`
	}
	if err := ctx.ShouldBindBodyWithJSON(&param); err != nil {
		h.ErrMsg(ctx, err)
		return
	}
	if len(param.IDs) == 0 {
		h.Ok(ctx)
		return
	}
	err := h.Database().Write().
		Where("id IN ? AND to = ?", param.IDs, cast.ToString(ctx.GetString("userID"))).
		Delete(&system.OfflineChatMsg{}).Error
	if err != nil {
		h.ErrMsg(ctx, err)
		return
	}
	h.Ok(ctx)
}
