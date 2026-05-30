package config

import (
	"crypto/rand"
	_ "embed"
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/fs185085781/v9os/internal/ioc"
	"github.com/fs185085781/v9os/pkg/util"
)

var version = ""
var updateNeed = false
var softType = ""

//go:embed version.json
var configContent []byte

type configBase struct {
	Remotes     []string `json:"remotes"`      //远程地址,使用该地址表示分布式,多机相同
	RemoteSave  string   `json:"remote_save"`  //远程配置保存地址,多机相同
	RemoteAuth  string   `json:"remote_auth"`  //远程配置认证串,多机相同
	Local       bool     `json:"local"`        //是否为本地配置,多机相同
	Version     string   `json:"version"`      //版本号,当前机独享
	MachineId   string   `json:"machine_id"`   //机器ID,当前机独享
	Port        int      `json:"port"`         //端口号,当前机独享
	LoadIp      string   `json:"load_ip"`      //负载主机,当前机独享
	WaitNetwork bool     `json:"wait_network"` //等待网络(启动不死的秘诀),当前机独享
}

type ConfigAll struct {
	Auth        *AuthConfig        `json:"auth"`
	CORS        *CORSConfig        `json:"cors"`
	Cachebase   *CachebaseConfig   `json:"cachebase"`
	Database    *DatabaseConfig    `json:"database"`
	Distributed *DistributedConfig `json:"distributed"`
	Log         *LogConfig         `json:"log"`
	Queuebase   *QueuebaseConfig   `json:"mqbase"`
	RateLimit   *RateLimitConfig   `json:"rate_limit"`
	Server      *ServerConfig      `json:"server"`
}

func NewConfig(st string) (Config, error) {
	var versionConfig map[string]string
	json.Unmarshal(configContent, &versionConfig)
	version = versionConfig["version"]
	configContent = nil
	softType = st
	var cnf configBase
	initFile := filepath.Join(util.RunDir(), "init.json")
	data, err := os.ReadFile(initFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			dcnf := &configBase{
				Local:       true,
				Version:     "0",
				MachineId:   util.UUID(),
				Port:        9099,
				LoadIp:      "",
				WaitNetwork: false,
			}
			data, err = json.MarshalIndent(dcnf, "", "  ")
			if err != nil {
				return nil, err
			}
			err = os.WriteFile(initFile, data, 0644)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	err = json.Unmarshal(data, &cnf)
	if err != nil {
		return nil, err
	}
	if cnf.Version != version {
		cnf.Version = version
		updateNeed = true
		data, err = json.MarshalIndent(cnf, "", "  ")
		if err != nil {
			return nil, err
		}
		//延迟执行,保证中间依赖版本更新的内容受到保护
		ioc.Ioc().RegisterList(ioc.KeyAfterFunc, func() {
			os.WriteFile(initFile, data, 0644)
		})
	} else {
		updateNeed = false
	}
	if cnf.Local {
		return newLocalConfig(&cnf)
	}
	if len(cnf.Remotes) > 0 {
		return newRemoteConfig(&cnf)
	}
	return nil, errors.New("本地配置和远程配置均不存在")
}

func deepMergeJson(base, secondary map[string]interface{}) bool {
	changed := false
	for key, secValue := range secondary {
		if baseValue, exists := base[key]; exists {
			baseMap, baseIsMap := baseValue.(map[string]interface{})
			secMap, secIsMap := secValue.(map[string]interface{})
			if baseIsMap && secIsMap {
				if deepMergeJson(baseMap, secMap) {
					changed = true
				}
			}
		} else {
			base[key] = secValue
			changed = true
		}
	}
	return changed
}

func createDefaultConfig() *ConfigAll {
	randStr := func(n int) string {
		const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		ret := make([]byte, n)
		for i := 0; i < n; i++ {
			num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
			ret[i] = letters[num.Int64()]
		}
		return string(ret)
	}
	return &ConfigAll{
		Auth: &AuthConfig{
			Secret:                randStr(32),
			SecretTime:            time.Now(),
			ExpireDuration:        24 * time.Hour,
			RefreshExpireDuration: 365 * 24 * time.Hour,
		},
		CORS: &CORSConfig{
			Enabled: true,
		},
		Cachebase: &CachebaseConfig{
			Driver: "file",
			File: &FileCacheConfig{
				Dir: "config/cache",
			},
			Redis: &RedisCacheConfig{
				Mode:         "standalone",
				Addrs:        []string{"localhost:6379"},
				Password:     "",
				DB:           0,
				PoolSize:     20,
				MinIdleConns: 5,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
				DialTimeout:  5 * time.Second,
			},
		},
		Database: &DatabaseConfig{
			Driver:       "sqlite",
			DSN:          []string{"config/" + randStr(8) + ".db?_busy_timeout=5000"},
			MaxIdleConns: 10,
			MaxOpenConns: 100,
			Cache:        true,
			SoftDelete:   false,
			ShowSql:      false,
		},
		Distributed: &DistributedConfig{
			Enabled:       false,
			DistributedId: util.UUID(),
		},
		Log: &LogConfig{
			Level:      "error",
			Output:     []string{"file", "console"},
			MaxSize:    100,
			MaxBackups: 15,
			MaxAge:     30,
			Dir:        "config/log",
		},
		Queuebase: &QueuebaseConfig{
			Driver: "mem",
			Mem: &MemQueueConfig{
				Capacity: 500000,
			},
			Rocket: &RocketQueueConfig{
				Addrs:     []string{"localhost:9002"},
				AccessKey: "accesskey",
				Secret:    "secret",
				Topic:     "v9os",
			},
			Redis: &RedisQueueConfig{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
				PoolSize: 5,
				Topic:    "v9os",
			},
		},
		RateLimit: &RateLimitConfig{
			Enabled: true,
			RPS:     100,
			Burst:   50,
		},
		Server: &ServerConfig{
			RequestLog:       true,
			ReadTimeout:      5 * time.Minute,
			WriteTimeout:     5 * time.Minute,
			PasswordKey:      randStr(20),
			StoreHost:        "https://support.v9os.com/store/store.php?path=",
			StoreType:        "v9os",
			SystemId:         util.UUID(),
			CommunicationKey: randStr(20),
			ProxyToken:       util.UUID(),
			UpgradeChannel:   "stable",
		},
	}
}

// CompareConfigChange 对比两份配置JSON，将secondary中缺失的字段合并到base中
// 返回合并后的数据和是否有变化
func CompareConfigChange(base, secondary []byte) ([]byte, bool) {
	var baseMap, secMap map[string]interface{}
	if json.Unmarshal(base, &baseMap) != nil || json.Unmarshal(secondary, &secMap) != nil {
		return base, false
	}
	if deepMergeJson(baseMap, secMap) {
		if data, err := json.MarshalIndent(baseMap, "", "  "); err == nil {
			return data, true
		}
	}
	return base, false
}

// ConfigChanged 对比两个ConfigAll是否有差异
func ConfigChanged(old, new *ConfigAll) bool {
	return !reflect.DeepEqual(old, new)
}

func checkConfigChange(now []byte) ([]byte, bool) {
	dnf, err := json.MarshalIndent(createDefaultConfig(), "", "  ")
	if err != nil {
		return now, false
	}
	return CompareConfigChange(now, dnf)
}
