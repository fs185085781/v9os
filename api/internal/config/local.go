package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/fs185085781/v9os/pkg/util"
)

type localConfig struct {
	all        *ConfigAll
	base       *configBase
	configFile string
}

func (l *localConfig) ConfigAll() *ConfigAll {
	return l.all
}

func newLocalConfig(base *configBase) (Config, error) {
	l := &localConfig{base: base, configFile: filepath.Join(util.RunDir(), "config", "config.json")}
	data, err := os.ReadFile(l.configFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		os.MkdirAll(filepath.Dir(l.configFile), 0755)
		l.all = createDefaultConfig()
		err = l.Save()
		if err != nil {
			return nil, err
		}
	}
	needUpdate := false
	if l.all == nil {
		var changed bool
		data, changed = checkConfigChange(data)
		var tmp ConfigAll
		err = json.Unmarshal(data, &tmp)
		if err != nil {
			return nil, err
		}
		l.all = &tmp
		if changed {
			needUpdate = true
			l.Save()
		}
	}
	if needUpdate {
		updateNeed = true
	}
	return l, nil
}

// Auth implements Config.
func (l *localConfig) Auth() *AuthConfig {
	if l.all.Auth != nil {
		return l.all.Auth
	}
	return nil
}

// CORS implements Config.
func (l *localConfig) CORS() *CORSConfig {
	if l.all.CORS != nil {
		return l.all.CORS
	}
	return nil
}

// Cachebase implements Config.
func (l *localConfig) Cachebase() *CachebaseConfig {
	if l.all.Cachebase != nil {
		return l.all.Cachebase
	}
	return nil
}

// Database implements Config.
func (l *localConfig) Database() *DatabaseConfig {
	if l.all.Database != nil {
		return l.all.Database
	}
	return nil
}

func (l *localConfig) Distributed() *DistributedConfig {
	if l.all.Distributed != nil {
		return l.all.Distributed
	}
	return nil
}

// Log implements Config.
func (l *localConfig) Log() *LogConfig {
	if l.all.Log != nil {
		return l.all.Log
	}
	return nil
}

// Machine implements Config.
func (l *localConfig) Machine() *MachineConfig {
	return &MachineConfig{
		Version:     l.base.Version,
		NeedUpdate:  updateNeed,
		MachineId:   l.base.MachineId,
		Port:        l.base.Port,
		LoadIp:      l.base.LoadIp,
		WaitNetwork: l.base.WaitNetwork,
		SoftType:    softType,
	}
}

// Queuebase implements Config.
func (l *localConfig) Queuebase() *QueuebaseConfig {
	if l.all.Queuebase != nil {
		return l.all.Queuebase
	}
	return nil
}

// RateLimit implements Config.
func (l *localConfig) RateLimit() *RateLimitConfig {
	if l.all.RateLimit != nil {
		return l.all.RateLimit
	}
	return nil
}

// Server implements Config.
func (l *localConfig) Server() *ServerConfig {
	if l.all.Server != nil {
		return l.all.Server
	}
	return nil
}

func (l *localConfig) Save() error {
	data, err := json.MarshalIndent(l.all, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(l.configFile, data, 0644)
	if err != nil {
		return err
	}
	return nil
}
