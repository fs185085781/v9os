package main

import (
	"github.com/fs185085781/v9os/internal/app"
	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/internal/logger"
)

func main() {
	cfg, err := config.NewConfig("gui")
	if err != nil {
		panic(err)
	}
	log, err := logger.NewLogger(cfg.Log())
	if err != nil {
		panic(err)
	}
	ap, err := app.NewApp(cfg, log)
	if err != nil {
		panic(err)
	}
	ap.StartSync()
}
