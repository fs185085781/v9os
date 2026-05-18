package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fs185085781/v9os/internal/database"
	"github.com/fs185085781/v9os/internal/inface/distributed"
	"github.com/fs185085781/v9os/internal/inface/official_license"
	infaceUser "github.com/fs185085781/v9os/internal/inface/user"
	"github.com/fs185085781/v9os/internal/ioc"
	"github.com/fs185085781/v9os/internal/ioc/uioc"
	"github.com/fs185085781/v9os/internal/logger"
	"github.com/fs185085781/v9os/internal/model/base"
	"github.com/fs185085781/v9os/internal/model/plugin"
	"github.com/fs185085781/v9os/internal/queue"
	"github.com/fs185085781/v9os/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

type PluginController struct {
	*BaseController
}

func init() {
	c := &PluginController{
		BaseController: GetBaseController(),
	}
	//用于前端自动加载插件js
	c.RegisterApi("POST", "/plugin/webhooks", c.WebHooks)
	//主程序插件专用-代理插件前端
	c.RegisterPublic("page", "GET", "/*filepath", c.PluginStream) //代理插件前端(不校验登陆)
	//主程序插件专用-代理JSON请求
	c.RegisterPublic("plugin", "POST", "/*filepath", c.PluginStream) //POST代理JSON
	//主程序插件专用-代理后端下载(不校验登陆)
	c.RegisterPublic("stream", "GET", "/*filepath", c.PluginStream) //GET代理后端下载(不校验登陆)
	//主程序插件专用-代理大文件提交(不校验登陆)
	c.RegisterPublic("stream", "POST", "/*filepath", c.PluginStream) //POST代理大文件提交(不校验登陆)
	//三方插件专用-代理插件所有资源
	c.RegisterPublic("api", "GET", "/thirdplugin/:code/*filepath", c.ThirdPluginRun)
	//前端插件专用-代理插件运行入口
	c.RegisterPublic("api", "GET", "/webplugin/:code", c.WebPluginRun)
	//前端插件专用-代理插件所有资源
	c.RegisterPublic("api", "GET", "/webplugin/:code/*filepath", c.WebPluginRun)
	//前端插件专用-用户数据读/写/删
	c.RegisterPublic("api", "POST", "/webdata/:code/:method", c.WebPluginData)
	//内部调用专用用于转发事件数据
	c.RegisterPublic("pluginprivate", "POST", "/*filepath", c.PluginPrivatePost)
	//插件调用专用-数据库操作
	c.RegisterPublic("pluginExp", "POST", "/gorm/bridge", c.GormBridge)
	//插件调用专用-数据库模型绑定
	c.RegisterPublic("pluginExp", "POST", "/gorm/bind", c.GormBind)
	//插件调用专用-分布式加锁
	c.RegisterPublic("pluginExp", "POST", "/lock/tryLock", c.LockTryLock)
	//插件调用专用-分布式解锁
	c.RegisterPublic("pluginExp", "POST", "/lock/unLock", c.LockUnLock)
	//插件调用专用-缓存设置(临时kv)
	c.RegisterPublic("pluginExp", "POST", "/cache/set", c.CacheSet)
	//插件调用专用-缓存获取(临时kv)
	c.RegisterPublic("pluginExp", "POST", "/cache/get", c.CacheGet)
	//插件调用专用-日志打印
	c.RegisterPublic("pluginExp", "POST", "/log/set", c.LogSet)
	//插件调用专用-事件订阅(广播)
	c.RegisterPublic("pluginExp", "POST", "/event/subscribeBroadcast", c.EventSubscribe)
	//插件调用专用-事件订阅(点对点)
	c.RegisterPublic("pluginExp", "POST", "/event/subscribe", c.EventSubscribe)
	//插件调用专用-事件订阅(点对点绝对地址)
	c.RegisterPublic("pluginExp", "POST", "/event/subscribeAbsoluteUrl", c.EventSubscribe)
	//插件调用专用-事件取消订阅
	c.RegisterPublic("pluginExp", "POST", "/event/unsubscribe", c.EventUnsubscribe)
	//插件调用专用-事件推送
	c.RegisterPublic("pluginExp", "POST", "/event/push", c.EventPush)
	//插件调用专用-权限数据上报
	c.RegisterPublic("pluginExp", "POST", "/auth/register", c.AuthRegister)
	//插件调用专用-数据获取(持久化kv)
	c.RegisterPublic("pluginExp", "POST", "/data/get", c.DataGet)
	//插件调用专用-数据设置(持久化kv)
	c.RegisterPublic("pluginExp", "POST", "/data/set", c.DataSet)
	//插件调用专用-内核配置
	c.RegisterPublic("pluginExp", "POST", "/app/config", c.AppConfig)
	//插件调用专用-商业授权密文
	c.RegisterPublic("pluginExp", "POST", "/official_license/auth_cipher", c.OfficialLicenseAuthCipher)
}

func (c *PluginController) WebHooks(ctx *gin.Context) {
	var plugins []plugin.Plugin
	err := c.Database().Read().
		Where("plugin_type = ? AND status = ? AND web_hook <> ''", 1, 1).
		Find(&plugins).Error
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	result := make([]gin.H, 0, len(plugins))
	for _, item := range plugins {
		src := resolvePluginWebhookURL(item)
		if src == "" {
			continue
		}
		result = append(result, gin.H{
			"code":    item.Code,
			"name":    item.Name,
			"version": item.Version,
			"src":     src,
		})
	}
	c.OkData(ctx, result)
}

func resolvePluginWebhookURL(pluginModel plugin.Plugin) string {
	hook := strings.TrimSpace(pluginModel.WebHook)
	if hook == "" {
		return ""
	}
	if strings.HasPrefix(hook, "http://") || strings.HasPrefix(hook, "https://") || strings.HasPrefix(hook, "//") {
		return hook
	}
	if strings.HasPrefix(hook, "/") {
		return hook
	}
	return "/plugin/" + pluginModel.Code + "/" + strings.TrimPrefix(strings.TrimPrefix(hook, "./"), "/")
}

func (c *PluginController) EventSubscribe(ctx *gin.Context) {
	pluginName, _ := c.PluginManage(1).GetPluginName(ctx.Query("code"), ctx.Query("key"))
	if pluginName == "" {
		c.FailMsg(ctx, "plugin not found")
		return
	}
	uri := ctx.Request.URL.Path
	arrs := strings.Split(uri, "/")
	module := arrs[3]
	stype := -1
	switch module {
	case "subscribeBroadcast":
		stype = queue.StypePluginBroadcast
	case "subscribe":
		stype = queue.StypePluginUnicast
	case "subscribeAbsoluteUrl":
		stype = queue.StypeAbsoluteURL
	}
	if stype == -1 {
		c.FailMsg(ctx, "module not found")
		return
	}
	//plugin event method
	param := c.param(ctx)
	event := param.ParamString("event")
	method := param.ParamString("method")
	url := "/pluginprivate/" + pluginName + "/" + method
	if stype == queue.StypeAbsoluteURL {
		url = method
	}
	err := c.Queue().Subscribe(event, url, pluginName, stype)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.Ok(ctx)
}
func (c *PluginController) EventUnsubscribe(ctx *gin.Context) {
	pluginName, _ := c.PluginManage(1).GetPluginName(ctx.Query("code"), ctx.Query("key"))
	if pluginName == "" {
		c.FailMsg(ctx, "plugin not found")
		return
	}
	//plugin event method
	param := c.param(ctx)
	event := param.ParamString("event")
	method := param.ParamString("method")
	url := "/pluginprivate/" + pluginName + "/" + method
	err := c.Queue().Unsubscribe(event, url)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.Ok(ctx)
}
func (c *PluginController) EventPush(ctx *gin.Context) {
	pluginName, _ := c.PluginManage(1).GetPluginName(ctx.Query("code"), ctx.Query("key"))
	if pluginName == "" {
		c.FailMsg(ctx, "plugin not found")
		return
	}
	param := c.param(ctx)
	event := param.ParamString("event")
	err := c.Queue().Publish(&queue.Message{
		EventType: event,
		Data:      param.Param("data"),
	})
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.Ok(ctx)
}

func (c *PluginController) CacheSet(ctx *gin.Context) {
	pluginName, _ := c.PluginManage(1).GetPluginName(ctx.Query("code"), ctx.Query("key"))
	if pluginName == "" {
		c.FailMsg(ctx, "plugin not found")
		return
	}
	param := c.param(ctx)
	key := param.ParamString("key")
	val := param.ParamString("val")
	t := param.ParamInt("time")
	if t <= 0 {
		t = 60
	}
	err := c.Cache().SetValue(pluginName+":"+key, []byte(val), time.Duration(t)*time.Minute)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.Ok(ctx)
}

func (c *PluginController) CacheGet(ctx *gin.Context) {
	pluginName, _ := c.PluginManage(1).GetPluginName(ctx.Query("code"), ctx.Query("key"))
	if pluginName == "" {
		c.FailMsg(ctx, "plugin not found")
		return
	}
	param := c.param(ctx)
	key := param.ParamString("key")
	val, err := c.Cache().GetValue(pluginName + ":" + key)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.OkData(ctx, string(val))
}

func (c *PluginController) LogSet(ctx *gin.Context) {
	pluginName, _ := c.PluginManage(1).GetPluginName(ctx.Query("code"), ctx.Query("key"))
	if pluginName == "" {
		c.FailMsg(ctx, "plugin not found")
		return
	}
	param := c.param(ctx)
	msg := param.ParamString("msg")
	fields := param.Param("fields")
	level := param.ParamString("level")
	if level == "" {
		level = "info"
	}
	fields2 := logger.Map2Fields(fields.(map[string]interface{}))
	switch level {
	case "info":
		c.Log().Info(msg, fields2...)
	case "warn":
		c.Log().Warn(msg, fields2...)
	case "error":
		c.Log().Error(msg, fields2...)
	case "debug":
		c.Log().Debug(msg, fields2...)
	}
	c.Ok(ctx)
}

func (c *PluginController) LockTryLock(ctx *gin.Context) {
	pluginName, _ := c.PluginManage(1).GetPluginName(ctx.Query("code"), ctx.Query("key"))
	if pluginName == "" {
		c.FailMsg(ctx, "plugin not found")
		return
	}
	param := c.param(ctx)
	key := param.ParamString("key")
	lock := c.Cache().CreateLock("lock:plugin:" + pluginName + ":" + key)
	if lock.TryLock() {
		c.OkData(ctx, lock.GetVal())
		return
	}
	c.FailMsg(ctx, "trylock false")
}

func (c *PluginController) LockUnLock(ctx *gin.Context) {
	pluginName, _ := c.PluginManage(1).GetPluginName(ctx.Query("code"), ctx.Query("key"))
	if pluginName == "" {
		c.FailMsg(ctx, "plugin not found")
		return
	}
	param := c.param(ctx)
	key := param.ParamString("key")
	val := param.ParamString("val")
	lock := c.Cache().GetLock("lock:plugin:"+pluginName+":"+key, val)
	if lock == nil {
		c.FailMsg(ctx, "lock not found")
		return
	}
	lock.UnLock()
	c.Ok(ctx)
}
func (c *PluginController) PluginPrivatePost(ctx *gin.Context) {
	il, str := c.IsLegality(ctx)
	if !il {
		//返回200状态码是方便丢弃消息队列的消息,防止队列重试
		c.FailMsg(ctx, str)
		return
	}
	ctx.Set("InternalSystem", true)
	c.PluginStream(ctx)
}

const pluginMachineProxyHeader = "X-Plugin-Machine-Proxy"

func (c *PluginController) proxyPluginByMachineWhitelist(ctx *gin.Context, pluginCode string, pluginType int) bool {
	pluginCode = strings.TrimSpace(pluginCode)
	if pluginCode == "" {
		return false
	}
	distributedProvider := uioc.Get[distributed.DistributedProvider](ioc.KeyDistributedProvider)
	if distributedProvider == nil || !distributedProvider.Enabled() {
		return false
	}
	localMachineID := strings.TrimSpace(distributedProvider.Nodes().LocalMachineID())
	if machineAllowsPluginInMemory(localMachineID, pluginCode) {
		return false
	}
	if ctx.GetHeader(pluginMachineProxyHeader) != "" {
		c.Log().Warn("plugin machine whitelist rejected proxied request", logger.NewField("pluginCode", pluginCode), logger.NewField("pluginType", pluginType), logger.NewField("machineId", localMachineID))
		c.ErrCode(ctx, http.StatusForbidden, "current machine is not allowed to run plugin: "+pluginCode)
		return true
	}
	host, err := c.resolvePluginMachineHost(distributedProvider, pluginCode, pluginType, localMachineID)
	if err != nil {
		c.Log().Warn("plugin machine whitelist proxy target not found", logger.NewField("pluginCode", pluginCode), logger.NewField("pluginType", pluginType), logger.NewField("machineId", localMachineID), logger.NewField("err", err))
		c.ErrMsg(ctx, err)
		return true
	}
	c.Log().Info("plugin request proxied by machine whitelist", logger.NewField("pluginCode", pluginCode), logger.NewField("pluginType", pluginType), logger.NewField("fromMachineId", localMachineID), logger.NewField("targetHost", host))
	ctx.Request.Header.Set(pluginMachineProxyHeader, "1")
	util.Proxy(ctx, "http://"+host)
	return true
}

func machineAllowsPluginInMemory(machineID string, pluginCode string) bool {
	machineID = strings.TrimSpace(machineID)
	pluginCode = strings.TrimSpace(pluginCode)
	if machineID == "" || pluginCode == "" {
		return false
	}
	machineMap := uioc.MachineAllowedPluginMap()
	if machineMap == nil {
		return false
	}
	return machineMap[machineID][pluginCode]
}

func (c *PluginController) resolvePluginMachineHost(distributedProvider distributed.DistributedProvider, pluginCode string, pluginType int, localMachineID string) (string, error) {
	_ = pluginType
	candidates := make([]string, 0)
	pluginMap := uioc.PluginAllowedMachineMap()
	for machineID := range pluginMap[pluginCode] {
		if strings.TrimSpace(machineID) == "" || machineID == localMachineID {
			continue
		}
		host, ok := distributedProvider.Nodes().Resolve(machineID)
		if ok && strings.TrimSpace(host) != "" {
			candidates = append(candidates, host)
		}
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("plugin %s has no online allowed machine", pluginCode)
	}
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})
	return candidates[0], nil
}

func (c *PluginController) PluginStream(ctx *gin.Context) {
	distributedID := ctx.Request.Header.Get("Distributed-Id")
	if distributedID != "" {
		targetHost, local, err := uioc.Get[distributed.DistributedProvider](ioc.KeyDistributedProvider).Affinity().ResolveAffinity(distributedID)
		if err != nil {
			c.ErrMsg(ctx, err)
			return
		}
		if !local {
			util.Proxy(ctx, targetHost)
			return
		}
	}
	arrs := strings.Split(ctx.Request.RequestURI, "/")
	pluginName := arrs[2]
	if c.proxyPluginByMachineWhitelist(ctx, pluginName, 1) {
		return
	}
	host, err := c.PluginManage(1).PluginHost(pluginName, ctx)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	module := arrs[1]
	action := ""
	url := ""
	switch module {
	case "page":
		url = host + ctx.Request.RequestURI
	case "stream":
		url = host + strings.Join(append(arrs[:2], arrs[3:]...), "/")
	case "plugin", "pluginprivate":
		module = "plugin"
		action = strings.Split(ctx.Request.URL.Path, "/")[3]
		url = host + "/" + module + "/" + action
	}
	if url == "" {
		c.FailMsg(ctx, "url not found")
		return
	}
	param := ctx.Request.Body
	defer param.Close()
	header := make(map[string][]string)
	for key, values := range ctx.Request.Header {
		for _, value := range values {
			header[key] = append(header[key], value)
		}
	}
	userProvider := uioc.Get[infaceUser.UserProvider](ioc.KeyUserProvider)
	if action != "" {
		// 权限校验
		userID := ctx.GetString("userID")
		if userID != "" && !ctx.GetBool("InternalSystem") {
			if !userProvider.CheckActionAuth(userID, pluginName+"/"+action) {
				ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"code": -1,
					"msg":  "无权限访问该功能",
				})
				return
			}
		}
		if lang := ctx.Request.Header.Get("lang"); lang != "" {
			header["lang"] = []string{lang}
		}
		if userID != "" {
			deptID := ctx.GetString("deptID")
			header["userID"] = []string{userID}
			header["deptID"] = []string{deptID}
			scope := userProvider.UserDatascope(cast.ToUint(userID), cast.ToUint(deptID))
			if scope != nil {
				scopeJSON, _ := json.Marshal(scope)
				header["dataScope"] = []string{string(scopeJSON)}
			}
		}
	}
	resp, err := util.HttpCommon(nil, ctx.Request.Method, url, param, header)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "请求失败: %v", err)
		return
	}
	defer resp.Body.Close()
	// 复制响应头到Gin的ResponseWriter
	for key, values := range resp.Header {
		for _, value := range values {
			ctx.Writer.Header().Add(key, value)
		}
	}
	// 设置原始状态码
	ctx.Status(resp.StatusCode)
	// 复制响应体
	buffer := make([]byte, 32*1024)
	_, err = io.CopyBuffer(ctx.Writer, resp.Body, buffer)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "复制响应体失败: %v", err)
		return
	}
}

type GormBridgeReq struct {
	TxId       string           `json:"txId"`
	Method     string           `json:"method"`
	ResultType string           `json:"resultType"`
	Steps      []GormActionStep `json:"steps"`
	Table      string           `json:"table"`
	Scope      *DataScopeReq    `json:"scope,omitempty"`
}
type DataScopeReq struct {
	Scope   int      `json:"scope"`   // 1=全部 2=本部门及下级 3=仅本部门 4=仅本人 5=自定义部门
	DeptIds []string `json:"deptIds"` // scope=2或5时的部门ID列表
	UserID  string   `json:"userID"`
	DeptID  string   `json:"deptID"`
}
type GormActionStep struct {
	Func string        //调用的方法
	Args []interface{} //参数
}
type GormBridgeResp struct {
	Data         interface{} `json:"data"`
	RowsAffected int64       `json:"rowsAffected"`
	Error        string      `json:"error"`
}

var transactionMap sync.Map

type transaction struct {
	db     *gorm.DB
	cancel context.CancelFunc
}

func (c *PluginController) GetSetRemoveDb(txId string, db *gorm.DB, isDelete bool) (*gorm.DB, error) {
	if txId == "" {
		return nil, fmt.Errorf("txId not found")
	}
	if isDelete {
		v, ok := transactionMap.Load(txId)
		if ok {
			ts, ok := v.(*transaction)
			if ok {
				ts.cancel()
			}
		}
		transactionMap.Delete(txId)
		return nil, nil
	}
	if db != nil {
		ctx, cancel := context.WithCancel(context.Background())
		go func(ctxInner context.Context) {
			timer := time.NewTimer(120 * time.Second)
			defer timer.Stop()
			select {
			case <-timer.C:
				transactionMap.Delete(txId)
			case <-ctxInner.Done():
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				return
			}
		}(ctx)
		transactionMap.Store(txId, &transaction{db: db, cancel: cancel})
		return db, nil
	}
	v, ok := transactionMap.Load(txId)
	if !ok {
		return nil, fmt.Errorf("txId not found")
	}
	ts, ok := v.(*transaction)
	if !ok {
		return nil, fmt.Errorf("db type error")
	}
	return ts.db, nil
}
func (c *PluginController) GormBridge(ctx *gin.Context) {
	resp := &GormBridgeResp{}
	pluginName, _ := c.PluginManage(1).GetPluginName(ctx.Query("code"), ctx.Query("key"))
	if pluginName == "" {
		resp.Error = "plugin not found"
		ctx.JSON(http.StatusOK, resp)
		return
	}
	var req GormBridgeReq
	err := ctx.ShouldBindBodyWithJSON(&req)
	if err != nil {
		resp.Error = err.Error()
		ctx.JSON(http.StatusOK, resp)
		return
	}
	if req.Method == "StartTransaction" || req.Method == "RollbackTransaction" || req.Method == "CommitTransaction" {
		//事务支持
		if req.TxId == "" {
			resp.Error = "txId is required"
			ctx.JSON(http.StatusOK, resp)
			return
		}
		if req.Method == "StartTransaction" {
			_, err = c.GetSetRemoveDb(req.TxId, c.Database().Write().Begin(), false)
			if err != nil {
				resp.Error = err.Error()
				ctx.JSON(http.StatusOK, resp)
				return
			}
		} else {
			db, err2 := c.GetSetRemoveDb(req.TxId, nil, false)
			if err2 != nil {
				resp.Error = err2.Error()
				ctx.JSON(http.StatusOK, resp)
				return
			}
			if req.Method == "CommitTransaction" {
				db.Commit()
			} else {
				db.Rollback()
			}
			resp.RowsAffected = db.RowsAffected
			c.GetSetRemoveDb(req.TxId, nil, true)
		}
		ctx.JSON(http.StatusOK, resp)
		return
	}
	if req.Table == "" {
		resp.Error = "table is required"
		ctx.JSON(http.StatusOK, resp)
		return
	}
	var columns []plugin.PluginColumn
	err = c.Database().Read().Where("plugin_name = ? and plugin_table = ?", pluginName, req.Table).Find(&columns).Error
	if err != nil {
		resp.Error = err.Error()
		ctx.JSON(http.StatusOK, resp)
		return
	}
	columns = append(columns, plugin.PluginColumn{
		PluginName:     pluginName,
		PluginTable:    req.Table,
		MainFieldName:  "DataId",
		MainColumnName: "data_id",
		FieldName:      "ID",
		ColumnName:     "id",
		FieldType:      "string",
		IsIndex:        1,
	})
	fieldNameMap := make(map[string]*plugin.PluginColumn)
	columnNameMap := make(map[string]*plugin.PluginColumn)
	mainFieldNameMap := make(map[string]*plugin.PluginColumn)
	for _, column := range columns {
		fieldNameMap[column.FieldName] = &column
		columnNameMap[column.ColumnName] = &column
		mainFieldNameMap[column.MainFieldName] = &column
	}
	tableName, err := database.PluginTableName(pluginName, req.Table)
	if err != nil {
		resp.Error = err.Error()
		ctx.JSON(http.StatusOK, resp)
		return
	}
	var db *gorm.DB
	if req.TxId != "" {
		pdb, err2 := c.GetSetRemoveDb(req.TxId, nil, false)
		if err2 != nil {
			resp.Error = err2.Error()
			ctx.JSON(http.StatusOK, resp)
			return
		}
		db = pdb.Session(&gorm.Session{}).Table(tableName)
	}
	var singleRes *plugin.PluginData
	var sliceRes *[]plugin.PluginData
	if req.ResultType == "none" {
		if db == nil {
			db = c.Database().Write().Table(tableName)
		}
	} else {
		if db == nil {
			db = c.Database().Read().Table(tableName)
		}
		switch req.ResultType {
		case "single":
			singleRes = &plugin.PluginData{}
		case "slice":
			sliceRes = &[]plugin.PluginData{}
		}
	}
	if req.ResultType != "none" {
		//查询防止越权
		db = db.Where("plugin_name = ? and plugin_table = ?", pluginName, req.Table)
		// 注入数据权限过滤条件
		db = c.DataScope(db, req.Scope)
	}
	isEnd := false
	getArgs := func(args []interface{}, index int) []interface{} {
		if index == 1 && len(args) < 2 {
			return nil
		}
		if index == 0 && len(args) < 1 {
			return nil
		}
		if args[index] == nil {
			return nil
		}
		conditionSlice, _ := args[index].([]interface{})
		return conditionSlice
	}
	needClearCache := false
	for _, step := range req.Steps {
		var entity *plugin.PluginData
		if step.Func == "Create" || step.Func == "Updates" || step.Func == "Delete" {
			//转化成实例
			entity, err = pluginMapToEntity(fieldNameMap, step.Args[0])
			if err != nil {
				resp.Error = err.Error()
				ctx.JSON(http.StatusOK, resp)
				return
			}
			entity.PluginName = pluginName
			entity.PluginTable = req.Table
		}
		// 写入操作时自动填充用户ID和部门ID
		if step.Func == "Create" && req.Scope != nil {
			if req.Scope.UserID != "" {
				entity.UserId = req.Scope.UserID
			}
			if req.Scope.DeptID != "" {
				entity.DeptId = req.Scope.DeptID
			}
		}
		if step.Func == "Update" || step.Func == "Updates" || step.Func == "Delete" {
			//更新防止越权
			db = db.Where("plugin_name = ? and plugin_table = ?", pluginName, req.Table)
			//注入数据权限过滤条件
			db = c.DataScope(db, req.Scope)
		}
		if step.Func == "Updates" {
			//user_id/dept_id 由内核管控,插件无权修改
			db = db.Omit("user_id", "dept_id")
		}
		switch step.Func {
		case "Where":
			args := getArgs(step.Args, 1)
			if args == nil {
				//接Entity,如&User{Name: "jinzhu", Age: 20} 已测试
				entity, err = pluginMapToEntity(fieldNameMap, step.Args[0])
				if err != nil {
					resp.Error = err.Error()
					ctx.JSON(http.StatusOK, resp)
					return
				}
				entity.PluginName = pluginName
				entity.PluginTable = req.Table
				db = db.Where(entity)
			} else {
				//接字段和值,如"name = ?", "jinzhu" 已测试
				db = db.Where(pluginSqlToMain(columnNameMap, step.Args[0]), args...)
			}
		case "Select":
			step.Args[0] = strings.Join(cast.ToStringSlice(step.Args[0]), " , ")
			db = db.Select(pluginSqlToMain(columnNameMap, step.Args[0]))
		case "Group":
			db = db.Group(pluginSqlToMain(columnNameMap, step.Args[0]))
		case "Having":
			args := getArgs(step.Args, 1)
			if args == nil {
				db = db.Having(pluginSqlToMain(columnNameMap, step.Args[0]))
			} else {
				db = db.Having(pluginSqlToMain(columnNameMap, step.Args[0]), args...)
			}
		case "Order":
			db = db.Order(pluginSqlToMain(columnNameMap, step.Args[0]))
		case "Limit":
			db = db.Limit(cast.ToInt(step.Args[0]))
		case "Offset":
			db = db.Offset(cast.ToInt(step.Args[0]))
		case "Distinct":
			step.Args[0] = strings.Join(cast.ToStringSlice(step.Args[0]), ",")
			db = db.Distinct(pluginSqlToMain(columnNameMap, step.Args[0]))
		case "Create":
			if entity.DataId == "" {
				resp.Error = "ID is required"
				ctx.JSON(http.StatusOK, resp)
				return
			}
			isEnd = true
			db = db.Create(entity)
		case "Update":
			isEnd = true
			col := pluginSqlToMain(columnNameMap, step.Args[0])
			if col == "user_id" || col == "dept_id" {
				resp.Error = "user_id/dept_id 由内核管控，插件无权修改"
				ctx.JSON(http.StatusOK, resp)
				return
			}
			db = db.Update(col, step.Args[1])
		case "Updates":
			isEnd = true
			db = db.Updates(entity)
		case "Delete":
			isEnd = true
			args := getArgs(step.Args, 1)
			if args == nil {
				db = db.Delete(entity)
			} else {
				db = db.Delete(entity, args...)
			}
		case "First":
			isEnd = true
			args := getArgs(step.Args, 0)
			if args == nil {
				db = db.First(singleRes)
			} else {
				db = db.First(singleRes, args...)
			}
		case "Take":
			isEnd = true
			args := getArgs(step.Args, 0)
			if args == nil {
				db = db.Take(singleRes)
			} else {
				db = db.Take(singleRes, args...)
			}
		case "Last":
			isEnd = true
			args := getArgs(step.Args, 0)
			if args == nil {
				db = db.Last(singleRes)
			} else {
				db = db.Last(singleRes, args...)
			}
		case "Find":
			isEnd = true
			args := getArgs(step.Args, 0)
			if args == nil {
				db = db.Find(sliceRes)
			} else {
				db = db.Find(sliceRes, args...)
			}
		case "Count":
			isEnd = true
			var count int64
			db = db.Model(&plugin.PluginData{}).Count(&count)
			resp.Data = count
		}
	}
	if !isEnd {
		resp.Error = "未找到结束方法"
		ctx.JSON(http.StatusOK, resp)
		return
	}
	if db.Error != nil {
		resp.Error = db.Error.Error()
		ctx.JSON(http.StatusOK, resp)
		return
	}
	if needClearCache {
		c.Database().ClearCache("plugin_data")
	}
	if resp.Data == nil {
		if singleRes != nil {
			resp.Data = mainListToPluginList(mainFieldNameMap, []plugin.PluginData{*singleRes})[0]
		} else if sliceRes != nil {
			resp.Data = mainListToPluginList(mainFieldNameMap, *sliceRes)
		}
	}
	resp.RowsAffected = db.RowsAffected
	ctx.JSON(http.StatusOK, resp)
}

func mainListToPluginList(mainFieldNameMap map[string]*plugin.PluginColumn, list []plugin.PluginData) []interface{} {
	res := make([]interface{}, 0)
	for _, item := range list {
		itemValue := reflect.ValueOf(item)
		if itemValue.Kind() == reflect.Ptr {
			if itemValue.IsNil() {
				continue
			}
			itemValue = itemValue.Elem()
		}
		if itemValue.Kind() != reflect.Struct {
			continue
		}

		itemType := itemValue.Type()
		resultMap := make(map[string]interface{})
		for i := 0; i < itemValue.NumField(); i++ {
			field := itemValue.Field(i)
			fieldType := itemType.Field(i)
			if !fieldType.IsExported() {
				continue
			}
			fieldName := fieldType.Name
			column, exists := mainFieldNameMap[fieldName]
			if exists {
				switch column.FieldType {
				case "string":
					resultMap[column.FieldName] = cast.ToString(field.Interface())
				case "int":
					resultMap[column.FieldName] = cast.ToInt(field.Interface())
				case "float64":
					resultMap[column.FieldName] = cast.ToFloat64(field.Interface())
				case "int64":
					resultMap[column.FieldName] = cast.ToInt64(field.Interface())
				default:
					resultMap[column.FieldName] = field.Interface()
				}
			} else if fieldType.Name == "BaseModel" {
				base := field.Interface().(base.BaseModel)
				resultMap["ID"] = base.ID
				resultMap["CreatedAt"] = base.CreatedAt
				resultMap["UpdatedAt"] = base.UpdatedAt
				resultMap["DeletedAt"] = base.DeletedAt
			}
		}
		res = append(res, resultMap)
	}
	return res
}

func pluginSqlToMain(columnNameMap map[string]*plugin.PluginColumn, sql interface{}) string {
	sqlStr := strings.TrimSpace(cast.ToString(sql))
	sqlStr = strings.ReplaceAll(sqlStr, "=", " = ")
	sqlStr = strings.ReplaceAll(sqlStr, ">", " > ")
	sqlStr = strings.ReplaceAll(sqlStr, "<", " < ")
	sqlStr = strings.ReplaceAll(sqlStr, ">=", " >= ")
	sqlStr = strings.ReplaceAll(sqlStr, "<=", " <= ")
	sqlStr = strings.ReplaceAll(sqlStr, "<>", " <> ")
	sqlStr = strings.ReplaceAll(sqlStr, "!=", " <> ")
	sqlStr = strings.ReplaceAll(sqlStr, "(", " ) ")
	sqlStr = strings.ReplaceAll(sqlStr, ")", " ) ")
	for {
		if strings.Contains(sqlStr, "  ") {
			sqlStr = strings.ReplaceAll(sqlStr, "  ", " ")
		} else {
			break
		}
	}
	sqlStr = " " + sqlStr + " "
	for k, v := range columnNameMap {
		sqlStr = strings.ReplaceAll(sqlStr, " "+k+" ", " "+v.MainColumnName+" ")
	}
	sqlStr = strings.TrimSpace(sqlStr)
	return sqlStr
}

func pluginMapToEntity(columnMap map[string]*plugin.PluginColumn, mapData interface{}) (*plugin.PluginData, error) {
	nMapData := make(map[string]interface{})
	if mapData != nil {
		for k, v := range mapData.(map[string]interface{}) {
			nk := k
			if colName, ok := columnMap[k]; ok {
				nk = colName.MainFieldName
			}
			nMapData[nk] = v
		}
	}
	var result plugin.PluginData
	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &result,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return nil, err
	}
	err = decoder.Decode(nMapData)
	if err != nil {
		return nil, err
	}
	var baseModel base.BaseModel
	config = &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &baseModel,
	}
	decoder, err = mapstructure.NewDecoder(config)
	if err == nil {
		err = decoder.Decode(nMapData)
		if err == nil {
			result.BaseModel = baseModel
		}
	}
	return &result, nil
}

type GormBindReq struct {
	Table  string
	Fields []map[string]string
}

func (c *PluginController) GormBind(ctx *gin.Context) {
	pluginName, _ := c.PluginManage(1).GetPluginName(ctx.Query("code"), ctx.Query("key"))
	if pluginName == "" {
		c.FailMsg(ctx, "plugin not found")
		return
	}
	var req GormBindReq
	err := ctx.ShouldBindBodyWithJSON(&req)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	// 1. 从数据库获取现有字段配置
	var columns []plugin.PluginColumn
	err = c.Database().Write().Where("plugin_name = ? and plugin_table = ?", pluginName, req.Table).Find(&columns).Error
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	// 2. 字段数量验证
	normalFieldNum := 20
	textFieldNum := 5
	indexFieldNum := 5
	// 统计已使用的普通字段和索引字段数量
	usedNormalFieldCount := 0
	usedIndexFieldCount := 0
	usedTextFieldCount := 0
	for _, field := range req.Fields {
		if field["text"] == "true" {
			usedTextFieldCount++
		} else if field["index"] == "true" {
			usedIndexFieldCount++
		} else {
			usedNormalFieldCount++
		}
	}
	// 验证字段数量限制
	if usedNormalFieldCount > normalFieldNum {
		c.ErrMsg(ctx, fmt.Errorf("普通字段数量超过限制，最大支持%d个字段", normalFieldNum))
		return
	}
	if usedTextFieldCount > textFieldNum {
		c.ErrMsg(ctx, fmt.Errorf("文本字段数量超过限制，最大支持%d个字段", textFieldNum))
		return
	}
	if usedIndexFieldCount > indexFieldNum {
		c.ErrMsg(ctx, fmt.Errorf("索引字段数量超过限制，最大支持%d个索引字段", indexFieldNum))
		return
	}

	// 3. 准备字段映射
	fieldPrefix := "Field"
	columnPrefix := "field"
	textFieldPrefix := "TextField"
	textColumnPrefix := "text_field"
	indexFieldPrefix := "IndexField"
	indexColumnPrefix := "index_field"
	// 初始化可用字段映射
	availableNormalFields := make(map[string]bool)
	availableTextFields := make(map[string]bool)
	availableIndexFields := make(map[string]bool)
	fieldToColumn := make(map[string]string)
	initAvailable := func(available map[string]bool, prefix, colPrefix string, n int) {
		for i := 1; i <= n; i++ {
			k := prefix + strconv.Itoa(i)
			available[k] = true
			fieldToColumn[k] = colPrefix + strconv.Itoa(i)
		}
	}
	initAvailable(availableNormalFields, fieldPrefix, columnPrefix, normalFieldNum)
	initAvailable(availableTextFields, textFieldPrefix, textColumnPrefix, textFieldNum)
	initAvailable(availableIndexFields, indexFieldPrefix, indexColumnPrefix, indexFieldNum)
	// 4. 构建查找映射（用于快速查找）
	existingMap := make(map[string]plugin.PluginColumn)
	reqMap := make(map[string]map[string]string)
	// 构建数据库现有记录的映射（以 MainFieldName 或 FieldName 作为键）
	for _, col := range columns {
		existingMap[col.FieldName] = col
		if col.IsText == 1 {
			availableTextFields[col.MainFieldName] = false
		} else if col.IsIndex == 1 {
			availableIndexFields[col.MainFieldName] = false
		} else {
			availableNormalFields[col.MainFieldName] = false
		}
	}
	// 构建请求字段的映射
	for _, field := range req.Fields {
		reqMap[field["field"]] = field
	}
	// 内置字段映射：column 和 field 与内核一致，直接映射，不占槽位
	builtinFieldMap := map[string]string{
		"user_id": "UserId",
		"dept_id": "DeptId",
	}
	for fieldKey, reqCol := range reqMap {
		mainCol, isBuiltin := builtinFieldMap[reqCol["column"]]
		if !isBuiltin || reqCol["field"] != mainCol {
			continue
		}
		// 从 reqMap 移除，避免进入普通差集逻辑
		delete(reqMap, fieldKey)
		// 若 existingMap 中已存在则跳过，否则创建
		if _, exists := existingMap[fieldKey]; exists {
			continue
		}
		col := plugin.PluginColumn{
			PluginName:     pluginName,
			PluginTable:    req.Table,
			FieldName:      reqCol["field"],
			FieldType:      reqCol["type"],
			ColumnName:     reqCol["column"],
			IsIndex:        2,
			IsText:         2,
			MainFieldName:  mainCol,
			MainColumnName: reqCol["column"],
		}
		if err := c.Database().Write().Create(&col).Error; err != nil {
			c.ErrMsg(ctx, err)
			return
		}
	}
	// 5. 计算差集和交集
	var toAdd []map[string]string      // 需要新增的字段
	var toDelete []plugin.PluginColumn // 需要删除的字段
	var toUpdate []plugin.PluginColumn // 需要更新的字段
	type migrateItem struct {
		oldCol    plugin.PluginColumn
		newReqCol map[string]string
	}
	var toMigrate []migrateItem // 槽位类型变化但需保留数据

	// 可保留数据的单向升级路径：index->text, index->normal, normal->text
	canMigrate := func(from, to string) bool {
		return (from == "index" && to == "text") ||
			(from == "index" && to == "normal") ||
			(from == "normal" && to == "text")
	}

	allocField := func(available map[string]bool, prefix string, n int, reqCol map[string]string) {
		for i := 1; i <= n; i++ {
			k := prefix + strconv.Itoa(i)
			if available[k] {
				reqCol["main_field_name"] = k
				available[k] = false
				break
			}
		}
	}
	doAlloc := func(reqCol map[string]string) {
		if reqCol["text"] == "true" {
			allocField(availableTextFields, textFieldPrefix, textFieldNum, reqCol)
		} else if reqCol["index"] == "true" {
			allocField(availableIndexFields, indexFieldPrefix, indexFieldNum, reqCol)
		} else {
			allocField(availableNormalFields, fieldPrefix, normalFieldNum, reqCol)
		}
	}

	slotType := func(isIndex, isText uint) string {
		if isText == 1 {
			return "text"
		} else if isIndex == 1 {
			return "index"
		}
		return "normal"
	}
	reqSlotType := func(reqCol map[string]string) string {
		if reqCol["text"] == "true" {
			return "text"
		} else if reqCol["index"] == "true" {
			return "index"
		}
		return "normal"
	}
	releaseSlot := func(col plugin.PluginColumn) {
		if col.IsText == 1 {
			availableTextFields[col.MainFieldName] = true
		} else if col.IsIndex == 1 {
			availableIndexFields[col.MainFieldName] = true
		} else {
			availableNormalFields[col.MainFieldName] = true
		}
	}

	// 5.1 找出需要删除和更新的字段
	for key, col := range existingMap {
		// 内置字段不参与差集删除逻辑
		if _, isBuiltin := builtinFieldMap[col.ColumnName]; isBuiltin {
			continue
		}
		reqCol, exists := reqMap[key]
		if !exists {
			toDelete = append(toDelete, col)
			releaseSlot(col)
		} else {
			from := slotType(col.IsIndex, col.IsText)
			to := reqSlotType(reqCol)
			if from != to {
				releaseSlot(col)
				doAlloc(reqCol)
				if canMigrate(from, to) {
					toMigrate = append(toMigrate, migrateItem{col, reqCol})
				} else {
					toDelete = append(toDelete, col)
					toAdd = append(toAdd, reqCol)
				}
			} else if col.FieldType != reqCol["type"] || col.ColumnName != reqCol["column"] {
				col.FieldType = reqCol["type"]
				col.ColumnName = reqCol["column"]
				toUpdate = append(toUpdate, col)
			}
		}
	}
	// 5.2 找出需要新增的字段（existingMap 中不存在的全新字段）
	for key, reqCol := range reqMap {
		_, exists := existingMap[key]
		if !exists {
			doAlloc(reqCol)
			toAdd = append(toAdd, reqCol)
		}
	}
	// 6. 执行数据库操作
	tx := c.Database().Write().Begin()
	// 6.1 执行删除
	deleteData := make(map[string]interface{})
	for _, col := range toDelete {
		if err = tx.Where("id = ?", col.ID).Delete(&plugin.PluginColumn{}).Error; err != nil {
			tx.Rollback()
			c.ErrMsg(ctx, err)
			return
		}
		deleteData[col.MainColumnName] = ""
	}
	// 删除通用数据
	tableName, err := database.PluginTableName(pluginName, req.Table)
	if err != nil {
		tx.Rollback()
		c.ErrMsg(ctx, err)
		return
	}
	err = tx.Table(tableName).Where("plugin_name = ? and plugin_table = ?", pluginName, req.Table).Updates(deleteData).Error
	if err != nil {
		tx.Rollback()
		c.ErrMsg(ctx, err)
		return
	}
	// 6.2 执行新增
	for _, reqCol := range toAdd {
		isIndex := uint(2)
		isText := uint(2)
		if reqCol["text"] == "true" {
			isText = 1
		} else if reqCol["index"] == "true" {
			isIndex = 1
		}
		col := plugin.PluginColumn{
			PluginName:     pluginName,
			PluginTable:    req.Table,
			FieldName:      reqCol["field"],
			FieldType:      reqCol["type"],
			ColumnName:     reqCol["column"],
			IsIndex:        isIndex,
			IsText:         isText,
			MainFieldName:  reqCol["main_field_name"],
			MainColumnName: fieldToColumn[reqCol["main_field_name"]],
		}
		if err := tx.Create(&col).Error; err != nil {
			tx.Rollback()
			c.ErrMsg(ctx, err)
			return
		}
	}

	// 6.3 执行更新
	for _, col := range toUpdate {
		err := tx.Save(&col).Error
		if err != nil {
			tx.Rollback()
			c.ErrMsg(ctx, err)
			return
		}
	}
	// 6.4 执行迁移（保留数据的槽位类型变化）
	for _, m := range toMigrate {
		oldMainCol := m.oldCol.MainColumnName
		newMainFieldName := m.newReqCol["main_field_name"]
		newMainCol := fieldToColumn[newMainFieldName]
		isIndex := uint(2)
		isText := uint(2)
		if m.newReqCol["index"] == "true" {
			isIndex = 1
		} else if m.newReqCol["text"] == "true" {
			isText = 1
		}
		// 删除旧 plugin_column 记录
		if err := tx.Where("id = ?", m.oldCol.ID).Delete(&plugin.PluginColumn{}).Error; err != nil {
			tx.Rollback()
			c.ErrMsg(ctx, err)
			return
		}
		// 创建新 plugin_column 记录
		newCol := plugin.PluginColumn{
			PluginName:     pluginName,
			PluginTable:    req.Table,
			FieldName:      m.newReqCol["field"],
			FieldType:      m.newReqCol["type"],
			ColumnName:     m.newReqCol["column"],
			IsIndex:        isIndex,
			IsText:         isText,
			MainFieldName:  newMainFieldName,
			MainColumnName: newMainCol,
		}
		if err := tx.Create(&newCol).Error; err != nil {
			tx.Rollback()
			c.ErrMsg(ctx, err)
			return
		}
		// 迁移数据：new_col = old_col, old_col = ''
		err = tx.Table(tableName).
			Where("plugin_name = ? and plugin_table = ?", pluginName, req.Table).
			Updates(map[string]interface{}{newMainCol: gorm.Expr(oldMainCol), oldMainCol: ""}).Error
		if err != nil {
			tx.Rollback()
			c.ErrMsg(ctx, err)
			return
		}
	}
	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.ErrMsg(ctx, err)
		return
	}
	c.Ok(ctx)
}

// AuthRegister 插件启动时上报权限元数据
func (c *PluginController) AuthRegister(ctx *gin.Context) {
	pluginCode, _ := c.PluginManage(1).GetPluginName(ctx.Query("code"), ctx.Query("key"))
	if pluginCode == "" {
		c.FailMsg(ctx, "plugin not found")
		return
	}
	param := c.param(ctx)
	actionsRaw := param.Param("actions")
	if actionsRaw == nil {
		c.Ok(ctx)
		return
	}
	actionsSlice, ok := actionsRaw.([]interface{})
	if !ok || len(actionsSlice) == 0 {
		c.Ok(ctx)
		return
	}
	userProvider := uioc.Get[infaceUser.UserProvider](ioc.KeyUserProvider)
	userProvider.SyncPluginAuths(pluginCode, "", actionsSlice)
	c.Ok(ctx)
}

func (c *PluginController) DataGet(ctx *gin.Context) {
	pluginName, _ := c.PluginManage(1).GetPluginName(ctx.Query("code"), ctx.Query("key"))
	if pluginName == "" {
		c.FailMsg(ctx, "plugin not found")
		return
	}
	param := c.param(ctx)
	key := param.ParamString("key")
	if key == "" {
		c.FailMsg(ctx, "key is required")
		return
	}
	val, err := c.Database().GetValue(pluginName + "_" + key)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	if val == "" {
		c.FailMsg(ctx, "not found")
		return
	}
	c.OkData(ctx, val)
}

func (c *PluginController) DataSet(ctx *gin.Context) {
	pluginName, _ := c.PluginManage(1).GetPluginName(ctx.Query("code"), ctx.Query("key"))
	if pluginName == "" {
		c.FailMsg(ctx, "plugin not found")
		return
	}
	param := c.param(ctx)
	key := param.ParamString("key")
	val := param.ParamString("val")
	if key == "" {
		c.FailMsg(ctx, "key is required")
		return
	}
	err := c.Database().SetValue(pluginName+"_"+key, val)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.Ok(ctx)
}

type AppConfig struct {
}

func (c *PluginController) AppConfig(ctx *gin.Context) {
	pluginName, _ := c.PluginManage(1).GetPluginName(ctx.Query("code"), ctx.Query("key"))
	if pluginName == "" {
		c.FailMsg(ctx, "plugin not found")
		return
	}
	c.OkData(ctx, &AppConfig{})
}

func (c *PluginController) OfficialLicenseAuthCipher(ctx *gin.Context) {
	pluginName, _ := c.PluginManage(1).GetPluginName(ctx.Query("code"), ctx.Query("key"))
	if pluginName == "" {
		c.FailMsg(ctx, "plugin not found")
		return
	}
	var req struct {
		Code string `json:"code"`
	}
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	provider := uioc.Get[official_license.Provider](ioc.KeyOfficialLicenseProvider)
	if provider == nil {
		c.ErrMsg(ctx, official_license.ErrUnauthorized)
		return
	}
	cipher, err := provider.AuthCipher(ctx, req.Code)
	if err != nil {
		c.ErrMsg(ctx, err)
		return
	}
	c.OkData(ctx, cipher)
}

func (c *PluginController) ThirdPluginRun(ctx *gin.Context) {
	code := strings.TrimSpace(ctx.Param("code"))
	if code == "" {
		c.FailMsg(ctx, "plugin not found")
		return
	}
	if c.proxyPluginByMachineWhitelist(ctx, code, 3) {
		return
	}
	webHost := ctx.Query("host")
	if webHost != "" {
		root := filepath.Clean(c.PluginManage(3).PluginDir(code))
		indexPath := filepath.Join(root, "index.html")
		if _, err := os.Stat(indexPath); err != nil {
			c.ErrMsg(ctx, fmt.Errorf("plugin index.html not found: %v", err))
			return
		}
		ctx.File(indexPath)
	} else {
		host, err := c.PluginManage(3).PluginHost(code, nil)
		if err != nil {
			c.ErrMsg(ctx, err)
			return
		}
		redirectURL := "/api/thirdplugin/" + code + "/?host=" + url.QueryEscape(host)
		ctx.Redirect(http.StatusFound, redirectURL)
	}
}
func (c *PluginController) WebPluginData(ctx *gin.Context) {
	u := c.UserInfo(ctx)
	if u == nil {
		return
	}
	code := strings.TrimSpace(ctx.Param("code"))
	method := strings.TrimSpace(ctx.Param("method"))
	if code == "" || method == "" {
		c.ErrCode(ctx, 404, "page not found")
		return
	}
	if method != "get" && method != "set" && method != "del" {
		c.ErrCode(ctx, 404, "page not found")
		return
	}
	userProvider := uioc.Get[infaceUser.UserProvider](ioc.KeyUserProvider)
	has := userProvider.CheckActionAuth(cast.ToString(u.ID), code+"/"+method)
	if !has {
		c.ErrCode(ctx, 403, "")
		return
	}
	param := c.Param(ctx)
	key := param.ParamString("key")
	if key == "" {
		c.FailMsg(ctx, "key not found")
		return
	}

	var data plugin.PluginWebData
	err := c.Database().Read().Where("user_id = ? and code = ? and data_key = ?", u.ID, code, key).First(&data).Error
	if !c.Database().IsOk(err) {
		c.ErrMsg(ctx, err)
		return
	}
	switch method {
	case "get":
		c.OkData(ctx, data.DataValue)
	case "set":
		data.UserID = u.ID
		data.Code = code
		data.DataKey = key
		data.DataValue = param.ParamString("val")
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
		c.Ok(ctx)
	case "del":
		if data.ID > 0 {
			err := c.Database().Delete(data.ID, &plugin.PluginWebData{})
			if err != nil {
				c.ErrMsg(ctx, err)
				return
			}
		}
		c.Ok(ctx)
	}
}
func (c *PluginController) WebPluginRun(ctx *gin.Context) {
	code := strings.TrimSpace(ctx.Param("code"))
	if code == "" {
		c.ErrCode(ctx, 404, "plugin not found")
		return
	}
	if c.proxyPluginByMachineWhitelist(ctx, code, 2) {
		return
	}
	if _, err := c.PluginManage(2).PluginHost(code, ctx); err != nil {
		c.ErrCode(ctx, 403, err.Error())
		return
	}
	root := filepath.Clean(c.PluginManage(2).PluginDir(code))
	relativePath := strings.TrimPrefix(ctx.Param("filepath"), "/")
	if relativePath == "" {
		relativePath = "index.html"
	}
	target := filepath.Clean(filepath.Join(root, filepath.FromSlash(relativePath)))
	if target != root && !strings.HasPrefix(target, root+string(os.PathSeparator)) {
		c.ErrCode(ctx, 403, "invalid web plugin path")
		return
	}
	info, err := os.Stat(target)
	if err == nil && info.IsDir() {
		target = filepath.Join(target, "index.html")
		info, err = os.Stat(target)
	}
	if err != nil && !strings.Contains(filepath.Base(relativePath), ".") {
		target = filepath.Join(root, "index.html")
		info, err = os.Stat(target)
	}
	if err != nil {
		c.ErrCode(ctx, 404, err.Error())
		return
	}
	if info.IsDir() {
		c.ErrCode(ctx, 404, "web plugin file not found")
		return
	}
	ctx.File(target)
}
