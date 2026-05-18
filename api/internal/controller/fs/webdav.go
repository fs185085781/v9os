package fs

import (
	"github.com/fs185085781/v9os/internal/controller"
	"github.com/gin-gonic/gin"
)

type WebDAVController struct {
	*controller.BaseController
}

func init() {
	c := &WebDAVController{
		BaseController: controller.GetBaseController(),
	}
	c.RegisterWebdav(c.WebdavFunc)
}
func (c *WebDAVController) WebdavFunc(ctx *gin.Context) {

}
