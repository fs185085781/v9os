package api

import (
	"github.com/fs185085781/v9os/internal/controller"
	"github.com/fs185085781/v9os/pkg/locales"
	"github.com/gin-gonic/gin"
)

func init() {
	c := &LangController{
		BaseController: controller.GetBaseController(),
	}
	//获取多语言数据
	c.RegisterPublic("api", "GET", "/lang/get", c.Get)
}

type LangController struct {
	*controller.BaseController
}

func (c *LangController) Get(ctx *gin.Context) {
	lang := c.GetLang(ctx)
	c.OkData(ctx, locales.GetModelLang(lang))
}
