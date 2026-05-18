package api

import (
	"net/http"
	"strings"

	"github.com/fs185085781/v9os/internal/controller"
	"github.com/fs185085781/v9os/internal/model/plugin"
	"github.com/fs185085781/v9os/pkg/util"
	"github.com/gin-gonic/gin"
)

type ComponentController struct {
	*controller.BaseController
}

func init() {
	c := &ComponentController{
		BaseController: controller.GetBaseController(),
	}
	//获取组件拦截信息
	c.RegisterPublic("api", "GET", "/component/*filepath", c.ComponentGet)
}
func (c *ComponentController) ComponentGet(ctx *gin.Context) {
	name := strings.Replace(ctx.Request.RequestURI, "/api/component/", "", 1)
	pluginName := strings.Split(name, "/")[0]
	var pluginModel plugin.Plugin
	err := c.Database().Read().Where("code = ?", pluginName).First(&pluginModel).Error
	hasPlugin := false
	if err == nil && pluginModel.Status == 1 && pluginModel.PluginType == 1 {
		hasPlugin = true
	}
	if !hasPlugin {
		data := map[string]interface{}{
			"ComType": "vue",
		}
		c.OkData(ctx, data)
		return
	}
	action := name[strings.Index(name, "/")+1:]
	host, err := c.PluginManage(1).PluginHost(pluginName, ctx)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	url := host + "/plugin/" + action
	param := ctx.Request.Body
	defer param.Close()
	body, err := util.PostStream(nil, url, param, nil)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	ctx.Data(http.StatusOK, "application/json", body)
}
