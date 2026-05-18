package logger

import "github.com/fs185085781/v9os/internal/config"

func NewLogger(cnf *config.LogConfig) (Logger, error) {
	return newZapLog(cnf)
}
