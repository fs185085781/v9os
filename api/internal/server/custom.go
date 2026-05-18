package server

import (
	"net/http"
	"path"
)

// customFS 包装器，正确实现 static.ServeFileSystem 接口
type customFS struct {
	fs http.FileSystem
}

// Exists 方法必须符合 static.ServeFileSystem 接口的定义
// prefix: 静态文件服务的URL前缀（例如 "/static"）
// filepath: 请求的文件路径（相对于静态文件服务的根目录）
func (c *customFS) Exists(prefix string, filepath string) bool {
	// 构建完整的文件系统路径
	fullPath := path.Join(prefix, filepath)
	// 尝试打开文件来判断是否存在
	f, err := c.fs.Open(fullPath)
	if err != nil {
		return false
	}
	defer f.Close()
	return true
}

// Open 方法实现 http.FileSystem 接口
func (c *customFS) Open(name string) (http.File, error) {
	return c.fs.Open(name)
}
