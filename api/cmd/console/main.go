package main

import (
	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/internal/logger"
	"github.com/fs185085781/v9os/internal/server"
)

func main() {
	cfg, err := config.NewConfig("console")
	if err != nil {
		panic(err)
	}
	log, err := logger.NewLogger(cfg.Log())
	if err != nil {
		panic(err)
	}
	serv, err := server.NewServer(cfg, log)
	if err != nil {
		panic(err)
	}
	serv.StartSync()
}
