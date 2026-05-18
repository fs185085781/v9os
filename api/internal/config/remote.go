package config

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/fs185085781/v9os/pkg/util"
)

type remoteConfig struct {
	all  *ConfigAll
	base *configBase
}

func (l *remoteConfig) ConfigAll() *ConfigAll {
	return l.all
}

// Auth implements Config.
func (l *remoteConfig) Auth() *AuthConfig {
	if l.all.Auth != nil {
		return l.all.Auth
	}
	return nil
}

// CORS implements Config.
func (l *remoteConfig) CORS() *CORSConfig {
	if l.all.CORS != nil {
		return l.all.CORS
	}
	return nil
}

// Cachebase implements Config.
func (l *remoteConfig) Cachebase() *CachebaseConfig {
	if l.all.Cachebase != nil {
		return l.all.Cachebase
	}
	return nil
}

// Database implements Config.
func (l *remoteConfig) Database() *DatabaseConfig {
	if l.all.Database != nil {
		return l.all.Database
	}
	return nil
}

func (l *remoteConfig) Distributed() *DistributedConfig {
	if l.all.Distributed != nil {
		return l.all.Distributed
	}
	return nil
}

// Log implements Config.
func (l *remoteConfig) Log() *LogConfig {
	if l.all.Log != nil {
		return l.all.Log
	}
	return nil
}

// Machine implements Config.
func (l *remoteConfig) Machine() *MachineConfig {
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
func (l *remoteConfig) Queuebase() *QueuebaseConfig {
	if l.all.Queuebase != nil {
		return l.all.Queuebase
	}
	return nil
}

// RateLimit implements Config.
func (l *remoteConfig) RateLimit() *RateLimitConfig {
	if l.all.RateLimit != nil {
		return l.all.RateLimit
	}
	return nil
}

// Server implements Config.
func (l *remoteConfig) Server() *ServerConfig {
	if l.all.Server != nil {
		return l.all.Server
	}
	return nil
}
func newRemoteConfig(base *configBase) (Config, error) {
	getConfig := func() (ConfigAll, []byte, error) {
		var jsonData ConfigAll
		var lastErr error
		var data []byte
		for _, url := range base.Remotes {
			body, err := util.Post(nil, url, []byte("{}"), map[string][]string{
				"Authorization": {base.RemoteAuth},
			})
			if err != nil {
				lastErr = err
				continue
			}
			err = json.Unmarshal(body, &jsonData)
			if err != nil {
				lastErr = err
				continue
			}
			data = body
			lastErr = nil
			break
		}
		return jsonData, data, lastErr
	}
	var jsonData ConfigAll
	var data []byte
	var lastErr error
	for {
		jsonData, data, lastErr = getConfig()
		if !base.WaitNetwork || lastErr == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}
	l := &remoteConfig{
		all:  &jsonData,
		base: base,
	}
	needUpdate := false
	if base.RemoteSave != "" {
		var changed bool
		data, changed = checkConfigChange(data)
		if changed {
			err := json.Unmarshal(data, &jsonData)
			if err != nil {
				return nil, err
			}
			l.all = &jsonData
			//保存到配置地址
			needUpdate = true
			l.Save()
		}
	}
	if needUpdate {
		updateNeed = true
	}
	return l, nil
}
func (l *remoteConfig) Save() error {
	if l.base.RemoteSave == "" {
		return errors.New("remote save url is empty")
	}
	data, err := json.MarshalIndent(l.all, "", "  ")
	if err != nil {
		return err
	}
	_, err = util.Post(nil, l.base.RemoteSave, data, map[string][]string{
		"Authorization": {l.base.RemoteAuth},
	})
	return err
}
