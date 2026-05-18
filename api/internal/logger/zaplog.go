package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fs185085781/v9os/internal/config"
	"github.com/fs185085781/v9os/pkg/util"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// 初始化配置适配
func newZapLog(cnf *config.LogConfig) (Logger, error) {
	// 构建输出目标
	outputs := make([]zapcore.WriteSyncer, 0)
	// 控制台输出
	if contains(cnf.Output, "console") {
		outputs = append(outputs, zapcore.AddSync(os.Stdout))
	}
	// 文件输出
	if contains(cnf.Output, "file") {
		path := util.RunDir()
		// 创建日志目录
		if err := os.MkdirAll(filepath.Join(path, cnf.Dir), 0755); err != nil {
			return nil, fmt.Errorf("创建日志目录失败: %w", err)
		}
		// 添加文件滚动策略
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   filepath.Join(path, cnf.Dir, "v9os.log"),
			MaxSize:    cnf.MaxSize, // MB
			MaxBackups: cnf.MaxBackups,
			MaxAge:     cnf.MaxAge,
			LocalTime:  true,
		})
		outputs = append(outputs, fileWriter)
	}
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	// 创建核心写入器
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.NewMultiWriteSyncer(outputs...),
		zap.NewAtomicLevelAt(parseLevel(Level(cnf.Level))),
	)

	// 构建日志实例
	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zap.ErrorLevel),
	)
	outs := make([]io.Writer, 0)
	for _, out := range outputs {
		outs = append(outs, out)
	}
	// 初始化回调状态
	zlog := &zapLog{
		log:   logger,
		outs:  outs,
		hasDb: contains(cnf.Output, "db"),
	}
	output := strings.Join(cnf.Output, ",")
	zlog.Println("[" + cnf.Level + "日志]已初始化,输出目标:" + output)
	return zlog, nil
}

// 辅助函数检查输出配置
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// Close实现带互斥锁的安全关闭
func (z *zapLog) Close() error {
	if z.log != nil {
		if err := z.log.Sync(); err != nil {
			return fmt.Errorf("日志关闭失败: %w", err)
		}
		z.log = nil // 释放资源
	}
	return nil
}

// 添加互斥锁保护dbFn
type zapLog struct {
	hasDb bool
	once  sync.Once
	log   *zap.Logger
	outs  []io.Writer
	dbFn  func(lvl Level, msg string, fields ...Field)
}

// 所有日志方法添加回调执行
func (z *zapLog) Debug(msg string, fields ...Field) {
	z.log.Debug(msg, convertFields(fields)...)
	z.executeCallback(DebugLevel, msg, fields)
}

// 回调执行逻辑
func (z *zapLog) executeCallback(lvl Level, msg string, fields []Field) {
	if !z.hasDb {
		return
	}
	if z.dbFn == nil {
		return
	}
	currentLevel := z.log.Level()
	if currentLevel == zapcore.DebugLevel {
		//因Debug模式会打印sql,sql又打印日志,形成死循环,因此Debug模式不支持写到数据库
		return
	}
	if parseLevel(lvl) < currentLevel {
		return
	}
	util.Go(func() {
		z.dbFn(lvl, msg, fields...)
	})
}

// Write方法添加锁保护
func (z *zapLog) Write(fn func(lvl Level, msg string, fields ...Field)) {
	z.once.Do(func() {
		z.dbFn = fn
	})
}
func (z *zapLog) Error(msg string, fields ...Field) {
	z.log.Error(msg, convertFields(fields)...)
	z.executeCallback(ErrorLevel, msg, fields)
}
func (z *zapLog) Log(lvl Level, msg string, fields ...Field) {
	z.log.Log(parseLevel(lvl), msg, convertFields(fields)...)
	z.executeCallback(lvl, msg, fields)
}
func (z *zapLog) Warn(msg string, fields ...Field) {
	z.log.Warn(msg, convertFields(fields)...)
	z.executeCallback(WarnLevel, msg, fields)
}
func (z *zapLog) Info(msg string, fields ...Field) {
	z.log.Info(msg, convertFields(fields)...)
	z.executeCallback(InfoLevel, msg, fields)
}

func (z *zapLog) Println(msg string, fields ...Field) {
	str := ""
	for _, val := range fields {
		str += fmt.Sprintf(" %s=%v", val.Key, val.Value)
	}
	for _, out := range z.outs {
		fmt.Fprintf(out, "%s     info   %s     %s\n", time.Now().Format("2006-01-02 15:04:05"), msg, str)
	}
}

// 字段类型转换
func convertFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		zapFields[i] = zap.Any(f.Key, f.Value)
	}
	return zapFields
}

// 日志级别解析
func parseLevel(lvl Level) zapcore.Level {
	switch lvl {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	default:
		return zap.ErrorLevel
	}
}
