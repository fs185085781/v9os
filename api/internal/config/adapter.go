package config

import (
	"time"

	"github.com/fs185085781/v9os/internal/ioc"
	"github.com/fs185085781/v9os/pkg/util"
)

type MachineConfig struct {
	Version     string `json:"version"`
	NeedUpdate  bool   `json:"need_update"`
	MachineId   string `json:"machine_id"`
	Port        int    `json:"port"`
	LoadIp      string `json:"load_ip"`
	WaitNetwork bool   `json:"wait_network"`
	SoftType    string `json:"soft_type"`
}

// ServerConfig HTTP服务配置
type ServerConfig struct {
	ReadTimeout      time.Duration `json:"read_timeout"`
	WriteTimeout     time.Duration `json:"write_timeout"`
	RequestLog       bool          `json:"request_log"`
	PasswordKey      string        `json:"password_key"` //禁止API修改
	StoreHost        string        `json:"store_host"`
	StoreType        string        `json:"store_type"`
	SystemId         string        `json:"system_id"`         //禁止API修改
	CommunicationKey string        `json:"communication_key"` //禁止API修改
	ProxyToken       string        `json:"proxy_token"`
	ProxyHost        string        `json:"proxy_host"`
	UpgradeChannel   string        `json:"upgrade_channel"` //升级渠道,默认stable,可选beta
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver       string   `json:"driver"` //sqlite mysql gaussdb sqlserver clickhouse postgres
	DSN          []string `json:"dsn"`
	MaxIdleConns int      `json:"max_idle_conns"`
	MaxOpenConns int      `json:"max_open_conns"`
	Cache        bool     `json:"cache"`
	SoftDelete   bool     `json:"soft_delete"`
	ShowSql      bool     `json:"show_sql"`
}

// CachebaseConfig Cachebase配置
type CachebaseConfig struct {
	Driver string            `json:"driver"` //file redis
	File   *FileCacheConfig  `json:"file"`
	Redis  *RedisCacheConfig `json:"redis"`
}
type FileCacheConfig struct {
	Dir string `json:"dir"`
}
type RedisCacheConfig struct {
	Mode         string        `json:"mode"`        //standalone  sentinel  cluster
	MasterName   string        `json:"master_name"` //sentinel下必须
	Addrs        []string      `json:"addrs"`
	Password     string        `json:"password"`
	DB           int           `json:"db"`
	PoolSize     int           `json:"pool_size"`
	MinIdleConns int           `json:"min_idle_conns"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	DialTimeout  time.Duration `json:"dial_timeout"`
}

// QueuebaseConfig Queuebase配置
type QueuebaseConfig struct {
	Driver string             `json:"driver"` //mem rocket redis
	Mem    *MemQueueConfig    `json:"mem"`
	Rocket *RocketQueueConfig `json:"rocket"`
	Redis  *RedisQueueConfig  `json:"redis"`
}
type MemQueueConfig struct {
	Capacity int `json:"capacity"`
}
type RocketQueueConfig struct {
	Addrs     []string `json:"addrs"`
	AccessKey string   `json:"access_key"`
	Secret    string   `json:"secret"`
	Topic     string   `json:"topic"`
}
type RedisQueueConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
	PoolSize int    `json:"pool_size"`
	Topic    string `json:"topic"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Secret                string        `json:"secret"`
	SecretTime            time.Time     `json:"secret_time"`
	ExpireDuration        time.Duration `json:"expire_duration"`
	RefreshExpireDuration time.Duration `json:"refresh_expire_duration"`
	LastSecret            string        `json:"last_secret"` //每隔10天更新一次密钥，这20天的登录会兼容旧密钥
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string   `json:"level"`
	Output     []string `json:"output"`      //file console db
	MaxSize    int      `json:"max_size"`    //单文件存储多大MB
	MaxBackups int      `json:"max_backups"` //最大备份几个文件
	MaxAge     int      `json:"max_age"`     //最大存储多久 天
	Dir        string   `json:"dir"`         //日志目录
}

type CORSConfig struct {
	Enabled bool `json:"enabled"`
}

type RateLimitConfig struct {
	Enabled bool    `json:"enabled"`
	RPS     float64 `json:"rps"`
	Burst   int     `json:"burst"`
}

type DistributedConfig struct {
	Enabled       bool   `json:"enabled"`
	DistributedId string `json:"distributed_id"`
}

type Config interface {
	Machine() *MachineConfig
	Server() *ServerConfig
	Database() *DatabaseConfig
	Cachebase() *CachebaseConfig
	Queuebase() *QueuebaseConfig
	Distributed() *DistributedConfig
	Auth() *AuthConfig
	Log() *LogConfig
	CORS() *CORSConfig
	RateLimit() *RateLimitConfig
	ConfigAll() *ConfigAll
	Save() error
}

func init() {
	// 每隔1小时更新密钥检查
	ioc.Ioc().RegisterList(ioc.KeyTimerFunc, []interface{}{
		"update_secret",
		60,
		func() {
			cfg := ioc.Ioc().Get(ioc.KeyConfig).(Config)
			if util.UnixSeconds()-cfg.Auth().SecretTime.Unix() < 15*24*60*60 {
				return
			}
			cfg.Auth().LastSecret = cfg.Auth().Secret
			cfg.Auth().Secret = util.UUID()
			cfg.Auth().SecretTime = time.Now()
			cfg.Save()
		},
	})
}
