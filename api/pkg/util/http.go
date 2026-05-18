package util

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Minute, // 总请求超时时间
	Transport: &http.Transport{
		// 连接池配置
		MaxIdleConns:        0,                // 总最大空闲连接数
		MaxIdleConnsPerHost: 10,               // 每个目标主机最大空闲连接数
		MaxConnsPerHost:     0,                // 每个目标主机最大总连接数
		IdleConnTimeout:     30 * time.Second, // 空闲连接超时时间
		// 拨号配置
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second, // 建立TCP连接超时
			KeepAlive: 30 * time.Second, // 保持连接存活的时间
		}).DialContext,
		// 其他优化
		TLSHandshakeTimeout: 10 * time.Second, // TLS握手超时
		ForceAttemptHTTP2:   true,             // 尝试使用HTTP/2
	},
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		// 返回这个错误，客户端就不会自动重定向
		return http.ErrUseLastResponse
	},
}

type httpContextKey string

const HttpAutoRedirectKey httpContextKey = "http_auto_redirect"

var httpAutoRedirectClient = &http.Client{
	Timeout:   30 * time.Minute,
	Transport: httpClient.Transport,
}

var healthCheckClient = &http.Client{
	Timeout: 2 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        5,
		MaxIdleConnsPerHost: 1,                // 对同一主机只保持一个空闲连接
		IdleConnTimeout:     30 * time.Second, // 空闲连接快速释放
	},
}

func PostStreamResp(ctx context.Context, url string, data io.Reader, headers map[string][]string) (*http.Response, error) {
	return HttpCommon(ctx, "POST", url, data, headers)
}
func PostStream(ctx context.Context, url string, data io.Reader, headers map[string][]string) ([]byte, error) {
	return returnBytes(PostStreamResp(ctx, url, data, headers))
}
func Post(ctx context.Context, url string, data []byte, headers map[string][]string) ([]byte, error) {
	return returnBytes(PostResp(ctx, url, data, headers))
}
func PostResp(ctx context.Context, url string, data []byte, headers map[string][]string) (*http.Response, error) {
	return HttpCommon(ctx, "POST", url, bytes.NewBuffer(data), headers)
}
func Get(ctx context.Context, url string, headers map[string][]string) ([]byte, error) {
	return returnBytes(GetResp(ctx, url, headers))
}
func GetResp(ctx context.Context, url string, headers map[string][]string) (*http.Response, error) {
	return HttpCommon(ctx, "GET", url, nil, headers)
}

func GetHealthy(url string) bool {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}
	resp, err := healthCheckClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func returnBytes(resp *http.Response, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// 检查HTTP状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http status code %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func HttpCommon(ctx context.Context, method, url string, data io.Reader, headers map[string][]string) (*http.Response, error) {
	var req *http.Request
	var err error
	if ctx != nil {
		req, err = http.NewRequestWithContext(ctx, method, url, data)
	} else {
		req, err = http.NewRequest(method, url, data)
	}
	if err != nil {
		return nil, err
	}
	normalizedHeaders := make(map[string][]string, len(headers))
	for key, values := range headers {
		normalizedKey := strings.ToLower(key) // 统一转为小写
		normalizedHeaders[normalizedKey] = values
	}
	// 设置自定义Header
	for key, values := range normalizedHeaders {
		// 首先清除可能存在的旧值，确保每次设置都是精确的
		req.Header.Del(key)
		for _, value := range values {
			req.Header.Add(key, value) // 对每个值使用Add
		}
	}
	// 设置默认Content-Type
	if _, exists := normalizedHeaders["content-type"]; !exists && data != nil && (method == "POST" || method == "PUT" || method == "PATCH") {
		req.Header.Set("Content-Type", "application/json")
	}
	client := httpClient
	if ctx != nil {
		if autoRedirect, ok := ctx.Value(HttpAutoRedirectKey).(bool); ok && autoRedirect {
			client = httpAutoRedirectClient
		}
	}
	return client.Do(req)
}
func ctxError(c *gin.Context, code int, msg string) {
	c.AbortWithStatusJSON(code, &gin.H{
		"code": -1,
		"msg":  msg,
	})
}
func copyHeaders(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}
func Proxy(ctx *gin.Context, newHost string) {
	targetURL := newHost + ctx.Request.URL.Path
	if ctx.Request.URL.RawQuery != "" {
		targetURL += "?" + ctx.Request.URL.RawQuery
	}
	req, err := http.NewRequest(ctx.Request.Method, targetURL, ctx.Request.Body)
	if err != nil {
		ctxError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	copyHeaders(req.Header, ctx.Request.Header)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		ctxError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	defer resp.Body.Close()
	copyHeaders(ctx.Writer.Header(), resp.Header)
	ctx.Writer.WriteHeader(resp.StatusCode)
	io.Copy(ctx.Writer, resp.Body)
}
