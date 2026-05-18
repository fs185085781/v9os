package server

import (
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/fs185085781/v9os/internal/ioc/uioc"
	"github.com/fs185085781/v9os/pkg/util"
	"github.com/gin-gonic/gin"
)

func proxyFilter() gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.Request
		if isProxyRequest(req) {
			handleProxy(c)
			c.Abort()
			return
		}
		c.Next()
	}
}
func isProxyRequest(r *http.Request) bool {
	token := r.Header.Get("X-Proxy-Token")
	if token == uioc.Config().Server().ProxyToken {
		return true
	}
	return false
}
func handleProxy(c *gin.Context) {
	req := c.Request
	if req.Method == http.MethodConnect {
		handleConnectProxy(c)
		return
	}
	handleHTTPProxy(c)
}
func handleConnectProxy(c *gin.Context) {
	req := c.Request
	w := c.Writer
	target := req.Host
	if !strings.Contains(target, ":") {
		target = target + ":443"
	}
	destConn, err := net.DialTimeout("tcp", target, 10*time.Second)
	if err != nil {
		http.Error(w, "无法连接到目标", http.StatusBadGateway)
		return
	}
	defer destConn.Close()
	w.WriteHeader(http.StatusOK)
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "不支持劫持", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		return
	}
	defer clientConn.Close()
	util.Go(func() {
		io.Copy(destConn, clientConn)
	})
	io.Copy(clientConn, destConn)
}
func handleHTTPProxy(c *gin.Context) {
	req := c.Request
	originalURL := req.Header.Get("X-Original-URL")
	if originalURL == "" {
		if req.URL.Host != "" && req.URL.Scheme != "" {
			originalURL = req.URL.String()
		} else {
			c.JSON(400, gin.H{"error": "无法确定目标URL"})
			return
		}
	}
	target, err := url.Parse(originalURL)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的URL", "details": err.Error()})
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = func(req *http.Request) {
		req.URL = target
		req.Host = target.Host
		req.Header.Del("X-Proxy-Token")
		req.Header.Del("X-Original-URL")
		req.Header.Set("X-Forwarded-For", c.ClientIP())
		req.Header.Set("X-Forwarded-Proto", req.URL.Scheme)
		req.Header.Set("Via", "1.1 v9os-proxy")
	}
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		c.JSON(502, gin.H{
			"error":   "代理失败",
			"message": err.Error(),
			"target":  target.String(),
		})
	}
	proxy.ServeHTTP(c.Writer, req)
}
