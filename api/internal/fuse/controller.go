package fuse

import (
	_ "github.com/fs185085781/v9os/internal/controller"
	_ "github.com/fs185085781/v9os/internal/controller/api"
	_ "github.com/fs185085781/v9os/internal/controller/api/plugin"
	_ "github.com/fs185085781/v9os/internal/controller/api/system"
	_ "github.com/fs185085781/v9os/internal/controller/api/user"
	_ "github.com/fs185085781/v9os/internal/controller/fs"
	_ "github.com/fs185085781/v9os/internal/controller/websocket"
)

// 只是引火线,啥也不干
func ControllerFuse() {
}
