package logger

type Level string

const (
	DebugLevel Level = "debug"
	InfoLevel  Level = "info"
	WarnLevel  Level = "warn"
	ErrorLevel Level = "error"
)

type Field struct {
	Key   string
	Value interface{}
}

type Logger interface {
	Log(lvl Level, msg string, fields ...Field)
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Println(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Write(func(lvl Level, msg string, fields ...Field)) //输出函数,用于解决循环依赖问题
	Close() error
}

func Map2Fields(m map[string]interface{}) []Field {
	fields := make([]Field, 0, len(m))
	for k, v := range m {
		fields = append(fields, NewField(k, v))
	}
	return fields
}

func NewField(key string, value interface{}) Field {
	return Field{
		Key:   key,
		Value: value,
	}
}
