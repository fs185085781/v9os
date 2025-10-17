package plugin

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strconv"

	"github.com/spf13/cast"
)

// 插件运行规则实现
type PluginData interface {
	RunData(r *http.Request, param []byte) (any, error)
}

type Res struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

var serverPort = ""
var runKey = ""

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
	return port, 0
}

func Server(name string, staticFiles embed.FS) {
	port, exitCode := parseArgs()
	if exitCode != 0 {
		os.Exit(exitCode)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthy", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok")
		close, _ := strconv.ParseBool(r.URL.Query().Get("close"))
		if close && r.Method == "GET" {
			os.Exit(0)
		}
	})
	for k, v := range dataFunc {
		mux.HandleFunc("/"+k, func(w http.ResponseWriter, r *http.Request) {
			//需要添加代码,实现将r中提取json实体用于调用RunData方法,接着将返回的内容写入w
			var res Res
			bodyBytes, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				res.Code = -1
				res.Msg = "无效的数据: " + err.Error()
			} else {
				// 调用插件的RunData方法
				result, err := v.RunData(r, bodyBytes)
				if err != nil {
					res.Code = -1
					res.Msg = "执行出错: " + err.Error()
				} else {
					res.Code = 0
					res.Msg = "success"
					res.Data = result
				}
			}
			// 设置响应头并返回JSON结果
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(res)
		})
	}

	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		os.Exit(3)
	}
	fileServer := http.FileServer(http.FS(staticFS))
	mux.Handle("/plugin/"+name+"/", http.StripPrefix("/plugin/"+name+"/", fileServer))
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	AutoMigrate()
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		os.Exit(4) // 异常关闭
	}
	os.Exit(0) // 正常关闭
}

func Convert(param []byte, v any) {
	json.Unmarshal(param, v)
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

var (
	dataFunc  = make(map[string]PluginData)
	modelList = make([]ModelInfo, 0)
)

func Register(method string, action PluginData) {
	dataFunc[method] = action
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

func httpPost(uri string, data map[string]interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	targetUrl := "http://127.0.0.1:" + serverPort + "/pluginExp" + uri + "?key=" + runKey
	resp, err := http.Post(targetUrl, "application/json", bytes.NewBuffer(jsonData))
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
