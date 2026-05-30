package plugin

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cast"
)

// 插件运行规则实现
type PluginData interface {
	RunData(r *http.Request, param []byte) (any, error)
}

type PluginStreamData interface {
	Run(w http.ResponseWriter, r *http.Request)
}

type Res struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

var serverPort = ""
var runKey = ""
var webPort = ""
var code = ""

func GetServerPort() string {
	return serverPort
}

func parseArgs() (int, int) {
	if len(os.Args) < 2 {
		return 0, 2
	}

	port, err := strconv.Atoi(os.Args[1])
	if err != nil || port < 1 || port > 65535 {
		return 0, 1
	}
	serverPort = os.Args[2]
	runKey = os.Args[3]
	if len(os.Args) > 4 {
		webPort = os.Args[4]
	}
	return port, 0
}
func Server(name string, staticFiles embed.FS) {
	ServerAfterAction(name, staticFiles, nil)
}
func ServerAfterAction(name string, staticFiles embed.FS, fn func()) {
	ServerCallDataScopeAndAfterAction(name, staticFiles, nil, fn)
}
func ServerCallDataScope(name string, staticFiles embed.FS, fn func(userID, deptID, dataScope string)) {
	ServerCallDataScopeAndAfterAction(name, staticFiles, fn, nil)
}

func ServerCallDataScopeAndAfterAction(name string, staticFiles embed.FS, fn func(userID, deptID, dataScope string), fn2 func()) {
	code = name
	port, exitCode := parseArgs()
	if exitCode != 0 {
		os.Exit(exitCode)
	}
	RegisterLang("zh", map[string]string{
		"success": "操作成功",
	})
	RegisterLang("en", map[string]string{
		"success": "Operation successful",
	})
	mux := http.NewServeMux()
	mux.HandleFunc("/healthy", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok")
		close, _ := strconv.ParseBool(r.URL.Query().Get("close"))
		if close && r.Method == "GET" {
			os.Exit(0)
		}
	})
	for k, v := range streamFunc {
		strs := make([]string, 0)
		strs = append(strs, "/stream/"+k)
		strs = append(strs, strs[0]+"/")
		for _, str := range strs {
			mux.HandleFunc(str, func(w http.ResponseWriter, r *http.Request) {
				if fn != nil {
					fn(r.Header.Get("userID"), r.Header.Get("deptID"), r.Header.Get("dataScope"))
				}
				v.Run(w, r)
			})
		}
	}
	for k, v := range dataFunc {
		mux.HandleFunc("/plugin/"+k, func(w http.ResponseWriter, r *http.Request) {
			if fn != nil {
				fn(r.Header.Get("userID"), r.Header.Get("deptID"), r.Header.Get("dataScope"))
			}
			if v.needLogin && strings.TrimSpace(r.Header.Get("userID")) == "" {
				var res Res
				res.Code = -1
				res.Msg = "请登录后重试"
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(res)
				return
			}
			var res Res
			bodyBytes, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				res.Code = -1
				res.Msg = "无效的数据: " + err.Error()
			} else {
				result, err := v.pluginData.RunData(r, bodyBytes)
				if err != nil {
					res.Code = -1
					res.Msg = err.Error()
				} else {
					res.Code = 0
					res.Msg = GetText(r, "success")
					res.Data = result
				}
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(res)
		})
	}

	if webPort != "" {
		target, _ := url.Parse("http://localhost:" + webPort)
		mux.Handle("/page/"+name+"/", httputil.NewSingleHostReverseProxy(target))
	} else {
		staticFS, err := fs.Sub(staticFiles, "static")
		if err != nil {
			os.Exit(3)
		}
		fileServer := http.FileServer(http.FS(staticFS))
		mux.Handle("/page/"+name+"/", http.StripPrefix("/page/"+name+"/", fileServer))
	}
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	AutoMigrate()
	reportActions()
	if fn2 != nil {
		fn2()
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		os.Exit(4)
		return
	}
	err = server.Serve(ln)
	if err != nil {
		os.Exit(4)
		return
	}
	os.Exit(0)
}

func Convert(param []byte, v any) error {
	return json.Unmarshal(param, v)
}
func ConvertParam(param []byte) Param {
	var params map[string]interface{}
	err := json.Unmarshal(param, &params)
	if err != nil {
		return &paramImpl{
			params: make(map[string]interface{}),
		}
	}
	return &paramImpl{
		params: params,
	}
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
type paramImpl struct {
	params map[string]interface{}
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

// ActionMeta 权限元数据,用于自动注册权限
type ActionMeta struct {
	Feature string `json:"feature"`
	Label   string `json:"label"`
}
type PluginDataMore struct {
	pluginData PluginData
	needLogin  bool
}

var (
	dataFunc   = make(map[string]*PluginDataMore)
	metaFunc   = make(map[string]*ActionMeta)
	modelList  = make([]ModelInfo, 0)
	streamFunc = make(map[string]PluginStreamData)
)

// Register 注册插件方法,可选传入权限元数据(feature, label)
// 用法: Register("dept-save", &DeptSave{}, "部门管理", "新增")
// 无 meta 参数 = 不受权限控制
func Register(method string, action PluginData, meta ...string) {
	dataFunc[method] = &PluginDataMore{pluginData: action, needLogin: false}
	if len(meta) >= 2 {
		metaFunc[method] = &ActionMeta{Feature: meta[0], Label: meta[1]}
	}
}
func RegisterLogin(method string, action PluginData, meta ...string) {
	dataFunc[method] = &PluginDataMore{pluginData: action, needLogin: true}
	if len(meta) >= 2 {
		metaFunc[method] = &ActionMeta{Feature: meta[0], Label: meta[1]}
	}
}

func RegisterStream(method string, action PluginStreamData) {
	streamFunc[method] = action
}

type ModelInfo struct {
	Model     interface{}
	TableName string
}

func RegisterModel(model interface{}, tableName string) {
	modelList = append(modelList, ModelInfo{
		Model:     model,
		TableName: tableName,
	})
}

// reportActions 插件启动时上报权限元数据到内核
func reportActions() {
	if len(metaFunc) == 0 {
		return
	}
	actions := make([]map[string]string, 0, len(metaFunc))
	for method, meta := range metaFunc {
		actions = append(actions, map[string]string{
			"method":  method,
			"feature": meta.Feature,
			"label":   meta.Label,
		})
	}
	httpPost("/auth/register", map[string]interface{}{"actions": actions})
}

func httpPost(uri string, data map[string]interface{}) (map[string]interface{}, error) {
	resp, err := httpPostResp(uri, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var resultMap map[string]interface{}
	err = json.Unmarshal(body, &resultMap)
	if err != nil {
		return nil, err
	}
	return resultMap, nil
}

func httpPostResp(uri string, data map[string]interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	targetUrl := "http://127.0.0.1:" + serverPort + "/pluginExp" + uri + "?code=" + code + "&key=" + runKey
	resp, err := http.Post(targetUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		defer resp.Body.Close()
		return nil, fmt.Errorf("请求出错:%d", resp.StatusCode)
	}
	return resp, nil
}
