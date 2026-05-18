package fs

import (
	"github.com/fs185085781/v9os/internal/controller"
	"github.com/gin-gonic/gin"
)

type S3Controller struct {
	*controller.BaseController
}

func init() {
	c := &S3Controller{
		BaseController: controller.GetBaseController(),
	}
	c.RegisterS3(c.S3Func)
}
func (c *S3Controller) S3Func(ctx *gin.Context) {

}
