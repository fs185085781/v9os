package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/fs185085781/v9os/internal/cache"
	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/internal/database"
	"github.com/fs185085781/v9os/internal/inface/distributed"
	infaceUser "github.com/fs185085781/v9os/internal/inface/user"
	"github.com/fs185085781/v9os/internal/ioc"
	"github.com/fs185085781/v9os/internal/ioc/uioc"
	"github.com/fs185085781/v9os/internal/logger"
	"github.com/fs185085781/v9os/internal/middleware"
	"github.com/fs185085781/v9os/internal/model/user"
	"github.com/fs185085781/v9os/internal/plugin"
	"github.com/fs185085781/v9os/internal/plugin/manager"
	"github.com/fs185085781/v9os/internal/queue"
	"github.com/fs185085781/v9os/internal/store"
	"github.com/fs185085781/v9os/pkg/locales"
	"github.com/fs185085781/v9os/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

type BaseController struct {
}

// ----语言相关--开始----
func (b *BaseController) GetLang(ctx *gin.Context) string {
	return locales.GetLang(ctx)
}

func (b *BaseController) GetText(c *gin.Context, key string) string {
	return locales.GetText(b.GetLang(c), key)
}

// ----语言相关--结束----
// ----人员判断--开始----
func (b *BaseController) CheckAdminAuth(c *gin.Context) bool {
	if !b.IsAdminByUser(cast.ToUint(c.GetString("userID"))) {
		b.ErrCode(c, 403, b.GetText(c, "common.user.usernoauths"))
		return false
	}
	return true
}

func (b *BaseController) IsAdminByUser(userId uint) bool {
	return userId == 1
}

func (b *BaseController) UserInfo(c *gin.Context) *user.User {
	userID, ok := c.Get("userID")
	if ok {
		var u user.User
		b.Database().GetByID(cast.ToUint(userID), &u)
		if u.ID > 0 {
			if b.IsAdminByUser(u.ID) {
				u.IsAdmin = 1
			} else {
				u.IsAdmin = 0
			}
			return &u
		}
	}
	b.FailMsg(c, b.GetText(c, "common.user.usernoauths"))
	return nil
}

// ----人员判断--结束----
// ----消息返回--开始----
const ResponseKey = "_gin-gonic/gin/responsekey"

func (b *BaseController) OkData(c *gin.Context, data interface{}) {
	b.jsonReturn(c, &gin.H{
		"code": 0,
		"msg":  b.GetText(c, "common.msg.success"),
		"data": data,
	}, http.StatusOK, false)
}
func (b *BaseController) OkMsg(c *gin.Context, msg string) {
	b.jsonReturn(c, &gin.H{
		"code": 0,
		"msg":  msg,
	}, http.StatusOK, false)
}
func (b *BaseController) ErrMsg(c *gin.Context, err error) {
	b.FailMsg(c, err.Error())
}
func (b *BaseController) FailMsg(c *gin.Context, msg string) {
	b.jsonReturn(c, &gin.H{
		"code": -1,
		"msg":  msg,
	}, http.StatusOK, true)
}
func (b *BaseController) Ok(c *gin.Context) {
	b.OkMsg(c, b.GetText(c, "common.msg.success"))
}
func (b *BaseController) ErrCode(c *gin.Context, code int, msg string) {
	b.jsonReturn(c, &gin.H{
		"code": -1,
		"msg":  msg,
	}, code, true)
}

func (b *BaseController) StreamStart(c *gin.Context) bool {
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		b.FailMsg(c, b.GetText(c, "common.stream.unsupported"))
		return false
	}
	c.Status(http.StatusOK)
	flusher.Flush()
	return true
}

func (b *BaseController) StreamWrite(c *gin.Context, data interface{}) bool {
	select {
	case <-c.Request.Context().Done():
		return false
	default:
	}
	body, err := json.Marshal(data)
	if err != nil {
		body, _ = json.Marshal(gin.H{"code": -1, "msg": err.Error()})
	}
	if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", body); err != nil {
		return false
	}
	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
	return true
}

func (b *BaseController) jsonReturn(c *gin.Context, data *gin.H, code int, abort bool) {
	if abort {
		c.AbortWithStatusJSON(code, data)
	} else {
		c.JSON(code, data)
	}
}

// ----消息返回--结束----

// ----数据权限--结束----
// 消息队列内部调用检查
func (b *BaseController) IsLegality(ctx *gin.Context) (bool, string) {
	timeCheck := ctx.Request.Header.Get("Timecheck")
	if timeCheck == "" {
		return false, "timecheck not found"
	}
	decryptTimeCheck, err := util.DecryptGCM(timeCheck, util.AdjustKey([]byte(b.Config().Server().CommunicationKey)))
	if err != nil {
		return false, "timecheck decrypt failed"
	}
	if util.UnixSeconds()-cast.ToInt64(decryptTimeCheck) > 30*60 {
		return false, "timecheck expired"
	}
	return true, ""
}

// 数据权限构建
func (b *BaseController) DataScopeCtx(db *gorm.DB, ctx *gin.Context) *gorm.DB {
	return b.DataScopeUser(db, ctx.GetString("userID"), ctx.GetString("deptID"))
}
func (b *BaseController) DataScopeUser(db *gorm.DB, userId, deptId string) *gorm.DB {
	userProvider := uioc.Get[infaceUser.UserProvider](ioc.KeyUserProvider)
	res := &DataScopeReq{
		UserID: userId,
		DeptID: deptId,
	}
	scope := userProvider.UserDatascope(cast.ToUint(userId), cast.ToUint(deptId))
	if scope != nil {
		res.Scope = scope.Scope
		res.DeptIds = scope.DeptIds
	}
	return b.DataScope(db, res)
}
func (b *BaseController) DataScope(db *gorm.DB, scope *DataScopeReq) *gorm.DB {
	if scope == nil || scope.Scope == 1 {
		return db
	}
	if len(scope.DeptIds) > 0 {
		if scope.UserID != "" {
			return db.Where("dept_id IN ? or user_id = ?", scope.DeptIds, scope.UserID)
		} else {
			return db.Where("dept_id IN ?", scope.DeptIds)
		}
	} else if scope.UserID != "" {
		return db.Where("user_id = ?", scope.UserID)
	}
	return db
}

// ----数据权限--结束----

// ----HTTP请求注册--开始----
// 对外暴露,走nginx
// RegisterApi 注册需要登录的 API 路由
// meta 可选: meta[0]=功能组(如"插件管理"), meta[1]=按钮组(如"新增/编辑")
func (b *BaseController) RegisterApi(method, path string, handler func(*gin.Context), meta ...string) {
	var baseName, feature, label string
	if len(meta) >= 3 {
		baseName = meta[0]
		feature = meta[1]
		label = meta[2]
	} else if len(meta) >= 2 {
		baseName = "数据管理"
		feature = meta[0]
		label = meta[1]
	}
	b.registerWithAuth("auth", "api", method, path, handler, baseName, feature, label)
}

// group = api的对外暴露,走nginx,其他按需
func (b *BaseController) RegisterPublic(group, method, path string, handler func(*gin.Context)) {
	b.register("", group, method, path, handler)
}

func (b *BaseController) RegisterAdminApi(method, path string, handler func(*gin.Context)) {
	b.register("auth", "api", method, path, func(c *gin.Context) {
		if !b.CheckAdminAuth(c) {
			return
		}
		handler(c)
	})
}

func (b *BaseController) RegisterWebdav(handler func(*gin.Context)) {
	allMethods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "TRACE", "CONNECT", "PATCH", "PROPFIND", "PROPPATCH", "MKCOL", "COPY", "MOVE", "LOCK", "UNLOCK"}
	for _, method := range allMethods {
		b.register("davauth", "", method, "/webdav:id", handler)
		b.register("davauth", "", method, "/webdav:id/*filepath", handler)
	}
}
func (b *BaseController) RegisterS3(handler func(*gin.Context)) {
	allMethods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "TRACE", "CONNECT", "PATCH", "PROPFIND", "PROPPATCH", "MKCOL", "COPY", "MOVE", "LOCK", "UNLOCK"}
	for _, method := range allMethods {
		b.register("s3auth", "", method, "/s3:id", handler)
		b.register("s3auth", "", method, "/s3:id/*filepath", handler)
	}
}
func (b *BaseController) register(warename, group, method, path string, handler func(*gin.Context)) {
	b.registerWithAuth(warename, group, method, path, handler, "", "", "")
}

func (b *BaseController) registerWithAuth(warename, group, method, path string, handler func(*gin.Context), authName, authFeature, authLabel string) {
	registerMap := ioc.Ioc().GetOrRegister(ioc.KeyControllerMap, &sync.Map{}).(*sync.Map)
	val, _ := registerMap.LoadOrStore(group, &ioc.GroupRoutes{})
	gr := val.(*ioc.GroupRoutes)
	// 加锁保护切片操作
	gr.Mu.Lock()
	defer gr.Mu.Unlock()
	gr.Routes = append(gr.Routes, &ioc.RouterStruct{
		Method: method,
		Path:   path,
		Handler: func(ctx *gin.Context) {
			defer func() {
				if r := recover(); r != nil {
					b.Log().Error("panic", logger.NewField("error", fmt.Sprintf("%v", r)))
					ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
						"code": -1,
						"msg":  "服务器错误",
					})
				}
			}()
			handler(ctx)
		},
		Ware:        warename,
		AuthName:    authName,
		AuthFeature: authFeature,
		AuthLabel:   authLabel,
	})
}

// ----HTTP请求注册--结束----

// ----Websocket请求注册--开始----
func (b *BaseController) RegisterWebsocket(handler plugin.IWebSocketHandler) {
	wsManager := plugin.GetWsManager()
	//对外暴露,走nginx
	b.register("", "api", "GET", "/ws/"+handler.ChannelName(), func(c *gin.Context) {
		cid := c.Query("cid")
		if cid == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cid is required"})
			return
		}
		tmpId, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "userID not found"})
			return
		}
		userId := tmpId.(string)
		conn, err := wsManager.Upgrade(c.Writer, c.Request)
		if err != nil {
			return
		}
		client := &plugin.WSClient{
			Conn:   conn,
			UserID: userId,
			Send:   make(chan []byte, 256),
			Cid:    cid,
		}
		wsManager.AddClient(handler.ChannelName(), client)
		handler.OnOpen(client)
		util.Go(func() { wsManager.WritePump(client) })
		util.Go(func() { wsManager.ReadPump(client, handler) })
	})
	//不对外暴露,对内核暴露,不走nginx
	b.register("", "private", "POST", "/wspush/"+handler.ChannelName(), func(c *gin.Context) {
		il, str := b.IsLegality(c)
		if !il {
			//返回200状态码是方便丢弃消息队列的消息,防止队列重试
			b.FailMsg(c, str)
			return
		}
		bytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			b.ErrCode(c, http.StatusBadRequest, err.Error())
			return
		}
		defer c.Request.Body.Close()
		var msg plugin.WebsocketMessage
		if err := json.Unmarshal(bytes, &msg); err != nil {
			b.ErrCode(c, http.StatusBadRequest, err.Error())
			return
		}
		//消息转发
		userMapAny, _ := wsManager.GetClientMap(handler.ChannelName(), true)
		userMap := userMapAny.(map[string]map[string]*plugin.WSClient)
		cidMap := userMap[msg.To]
		isOk := false
		for _, client := range cidMap {
			if wsManager.SendClient(client, bytes) {
				isOk = true
			}
		}
		b.OkData(c, map[string]any{
			"delivered": isOk,
		})
	})
	//订阅一个 "/wspush"+handler.GetPath() 的事件,延迟执行保证事件订阅器一定完成了初始化
	ioc.Ioc().RegisterList(ioc.KeyAfterFunc, func() {
		b.Queue().Subscribe("wsevent:"+handler.ChannelName(), "/private/wspush/"+handler.ChannelName(), "", queue.StypeWebsocket)
	})
}

func (b *BaseController) SendWsMsg(channelName string, userId string, msgType string, msg string) error {
	error := b.Queue().Publish(&queue.Message{
		EventType: "wsevent:" + channelName,
		Data: plugin.WebsocketMessage{
			To:       userId,
			Msg:      msg,
			Type:     msgType,
			DateTime: time.Now(),
		},
	})
	return error
}

// ----Websocket请求注册--结束----

// ----Ioc依赖注入--开始----
func (b *BaseController) Config() config.Config {
	return uioc.Config()
}
func (b *BaseController) Log() logger.Logger {
	return uioc.Log()
}
func (b *BaseController) Database() database.Database {
	return uioc.Database()
}
func (b *BaseController) Cache() cache.Cache {
	return uioc.Cache()
}
func (b *BaseController) Queue() queue.Queue {
	return uioc.Queue()
}
func (b *BaseController) PluginManage(pluginType int) manager.IPluginManage {
	return uioc.Get[*manager.AllPluginManage](ioc.KeyPluginManage).Switch(pluginType)
}
func (b *BaseController) Auth() *middleware.User {
	return uioc.Get[*middleware.User](ioc.KeyMiddlewareAuth)
}
func (b *BaseController) WebdavAuth() *middleware.WebdavAuth {
	return uioc.Get[*middleware.WebdavAuth](ioc.KeyMiddlewareWebdavAuth)
}
func (b *BaseController) S3Auth() *middleware.S3Auth {
	return uioc.Get[*middleware.S3Auth](ioc.KeyMiddlewareS3Auth)
}
func (b *BaseController) Store() store.Store {
	return uioc.Get[store.Store](ioc.KeyStore)
}

func (b *BaseController) Distributed() distributed.DistributedProvider {
	return uioc.Get[distributed.DistributedProvider](ioc.KeyDistributedProvider)
}

// ----Ioc依赖注入--结束----

// ----分页基础相关--开始----
type PageRes struct {
	Total int64       `json:"total"`
	Data  interface{} `json:"data"`
}
type Param interface {
	ParamString(key string) string
	ParamInt(key string) int
	ParamBool(key string) bool
	ParamInt64(key string) int64
	ParamFloat64(key string) float64
	Param(key string) interface{}
	Map() map[string]interface{}
}
type PageParam interface {
	Param
	Page() int
	PageSize() int
	Sorter() []PageParamSort
}
type PageParamSort interface {
	ColumnKey() string
	Order() string
}
type paramImpl struct {
	params map[string]interface{}
}
type pageParamImpl struct {
	page     int
	pageSize int
	sorter   []*pageParamSortImpl
	param    *paramImpl
}

func (p *pageParamImpl) ParamBool(key string) bool {
	return p.param.ParamBool(key)
}
func (p *pageParamImpl) ParamFloat64(key string) float64 {
	return p.param.ParamFloat64(key)
}
func (p *pageParamImpl) ParamInt(key string) int {
	return p.param.ParamInt(key)
}
func (p *pageParamImpl) ParamInt64(key string) int64 {
	return p.param.ParamInt64(key)
}
func (p *pageParamImpl) ParamString(key string) string {
	return p.param.ParamString(key)
}
func (p *pageParamImpl) Param(key string) interface{} {
	return p.param.Param(key)
}
func (p *pageParamImpl) Map() map[string]interface{} {
	return p.param.Map()
}
func (p *pageParamImpl) Page() int {
	return p.page
}
func (p *pageParamImpl) PageSize() int {
	return p.pageSize
}
func (p *pageParamImpl) Sorter() []PageParamSort {
	list := make([]PageParamSort, 0)
	for _, sort := range p.sorter {
		list = append(list, sort)
	}
	return list
}
func (p *paramImpl) ParamString(key string) string {
	return cast.ToString(p.params[key])
}

func (p *paramImpl) ParamInt(key string) int {
	return cast.ToInt(p.params[key])
}

func (p *paramImpl) ParamBool(key string) bool {
	return cast.ToBool(p.params[key])
}

func (p *paramImpl) ParamInt64(key string) int64 {
	return cast.ToInt64(p.params[key])
}

func (p *paramImpl) ParamFloat64(key string) float64 {
	return cast.ToFloat64(p.params[key])
}

func (p *paramImpl) Param(key string) interface{} {
	return p.params[key]
}

func (p *paramImpl) Map() map[string]interface{} {
	return p.params
}

type pageParamSortImpl struct {
	columnKey string
	order     string
}

func (s *pageParamSortImpl) ColumnKey() string {
	return s.columnKey
}

func (s *pageParamSortImpl) Order() string {
	return s.order
}

func (b *BaseController) Param(ctx *gin.Context) Param {
	return b.param(ctx)
}

func (b *BaseController) ParamBytes(bytes []byte) Param {
	var params map[string]interface{}
	err := json.Unmarshal(bytes, &params)
	if err != nil {
		return &paramImpl{
			params: make(map[string]interface{}),
		}
	}
	return &paramImpl{
		params: params,
	}
}

func (b *BaseController) param(ctx *gin.Context) *paramImpl {
	var params map[string]interface{}
	err := ctx.ShouldBindBodyWithJSON(&params)
	if err != nil {
		return &paramImpl{
			params: make(map[string]interface{}),
		}
	}
	return &paramImpl{
		params: params,
	}
}

func (b *BaseController) PageParam(ctx *gin.Context) PageParam {
	param := b.param(ctx)
	impl := &pageParamImpl{
		param: param,
	}
	impl.page = param.ParamInt("page")
	impl.pageSize = param.ParamInt("pageSize")
	if impl.page < 1 {
		impl.page = 1
	}
	if impl.pageSize < 1 {
		impl.pageSize = 10
	}
	sorter := param.params["sorter"]
	if sorter != nil {
		sorterList, ok := sorter.([]interface{})
		if !ok {
			return impl
		}
		for _, sort := range sorterList {
			sortMap, ok := sort.(map[string]interface{})
			if !ok {
				continue
			}
			key := cast.ToString(sortMap["columnKey"])
			if key == "" {
				continue
			}
			impl.sorter = append(impl.sorter, &pageParamSortImpl{
				columnKey: key,
				order:     cast.ToString(sortMap["order"]),
			})
		}
	}
	return impl
}

// ----分页基础相关--结束----

// ----单例BaseController构建--开始----
var baseController = &BaseController{}

func GetBaseController() *BaseController {
	return baseController
}

// ----单例BaseController构建--结束----
