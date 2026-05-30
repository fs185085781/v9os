package fs

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/fs185085781/v9os/internal/controller"
	"github.com/fs185085781/v9os/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type FileController struct {
	*controller.BaseController
}

func init() {
	c := &FileController{
		BaseController: controller.GetBaseController(),
	}
	//获取文件流
	c.RegisterPublic("api", "GET", "/file/get/:key", c.GetFile)
	//上传文件
	c.RegisterApi("POST", "/file/upload", c.UploadFile)
}
func (c *FileController) GetFile(ctx *gin.Context) {
	if c.Config().Distributed().Enabled {
		c.ErrCode(ctx, http.StatusForbidden, "Distributed mode is not supported yet")
		return
	}
	decryptTimeCheck, err := util.DecryptGCM(ctx.Param("key"), util.AdjustKey([]byte(c.Config().Server().CommunicationKey)))
	if err != nil {
		c.ErrCode(ctx, http.StatusForbidden, err.Error())
		return
	}
	var mapData map[string]string
	err = json.Unmarshal([]byte(decryptTimeCheck), &mapData)
	if err != nil {
		c.ErrCode(ctx, http.StatusForbidden, err.Error())
		return
	}
	path := filepath.Join(util.RunDir(), "uploads", mapData["uid"], mapData["scene"]+"."+mapData["ext"])
	if _, err := os.Stat(path); err != nil {
		c.ErrCode(ctx, http.StatusNotFound, err.Error())
		return
	}
	ctx.Header("Cache-Control", "public, max-age=31536000")
	ctx.File(path)
}

func (c *FileController) UploadFile(ctx *gin.Context) {
	if c.Config().Distributed().Enabled {
		c.FailMsg(ctx, "Distributed mode is not supported yet")
		return
	}
	fileData, err := ctx.FormFile("file")
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	ext := filepath.Ext(fileData.Filename)
	scene := ctx.PostForm("scene")
	userId := cast.ToString(c.UserInfo(ctx).ID)
	mapData := make(map[string]string)
	mapData["uid"] = userId
	mapData["scene"] = scene
	mapData["ext"] = ext
	str, err := json.Marshal(mapData)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	timeCheck, err := util.EncryptGCM(string(str), util.AdjustKey([]byte(c.Config().Server().CommunicationKey)))
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	path := filepath.Join(util.RunDir(), "uploads", mapData["uid"], mapData["scene"]+"."+mapData["ext"])
	if err := ctx.SaveUploadedFile(fileData, path); err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.OkData(ctx, "/api/file/get/"+timeCheck)
}
