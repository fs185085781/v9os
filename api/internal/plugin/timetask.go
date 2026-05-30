package plugin

import (
	"sync"
	"time"

	"github.com/fs185085781/v9os/internal/cache"
	"github.com/fs185085781/v9os/internal/ioc"
	"github.com/fs185085781/v9os/internal/ioc/uioc"
	"github.com/fs185085781/v9os/pkg/util"
	"github.com/spf13/cast"
)

const (
	timeTaskLeaderName  = "timetask"
	timeTaskTick        = 20 * time.Second
	timeTaskRetryDelay  = 3 * time.Second
	timeTaskHealthCheck = 5 * time.Second
)

type TimeTaskSchedule struct {
	mu       sync.Mutex
	stopChan *chan struct{}
	leader   cache.Lock
}

func NewTimeTaskSchedule() *TimeTaskSchedule {
	return &TimeTaskSchedule{}
}

func (t *TimeTaskSchedule) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.stopChan != nil {
		return
	}
	stopChan := make(chan struct{})
	t.stopChan = &stopChan
	util.Go(func() {
		leaderKey := "leader:" + timeTaskLeaderName
		for {
			if t.isStopped(stopChan) {
				return
			}
			leaderLock := uioc.Cache().CreateLock(leaderKey)
			if leaderLock == nil {
				if t.waitOrStop(stopChan, timeTaskRetryDelay) {
					return
				}
				continue
			}
			if !leaderLock.TryLock() {
				if t.waitOrStop(stopChan, timeTaskRetryDelay) {
					return
				}
				continue
			}
			if t.runAsLeader(stopChan, leaderKey, leaderLock) {
				return
			}
			if t.waitOrStop(stopChan, timeTaskRetryDelay) {
				return
			}
		}
	})
}

func (t *TimeTaskSchedule) runAsLeader(stopChan <-chan struct{}, leaderKey string, leaderLock cache.Lock) (stopped bool) {
	t.setLeaderLock(leaderLock)
	defer func() {
		leaderLock.UnLock()
		t.clearLeaderLock(leaderLock)
		if recover() != nil {
			stopped = false
		}
	}()
	taskTicker := time.NewTicker(timeTaskTick)
	healthTicker := time.NewTicker(timeTaskHealthCheck)
	defer taskTicker.Stop()
	defer healthTicker.Stop()
	for {
		select {
		case <-stopChan:
			return true
		case <-healthTicker.C:
			if !leaderLock.IsAlive() {
				return false
			}
		case <-taskTicker.C:
			if !leaderLock.IsAlive() {
				return false
			}
			t.RunTasks()
		}
	}
}

func (t *TimeTaskSchedule) waitOrStop(stopChan <-chan struct{}, d time.Duration) bool {
	select {
	case <-stopChan:
		return true
	case <-time.After(d):
		return false
	}
}

func (t *TimeTaskSchedule) isStopped(stopChan <-chan struct{}) bool {
	select {
	case <-stopChan:
		return true
	default:
		return false
	}
}

func (t *TimeTaskSchedule) setLeaderLock(lock cache.Lock) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.leader = lock
}

func (t *TimeTaskSchedule) clearLeaderLock(lock cache.Lock) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.leader == lock {
		t.leader = nil
	}
}

func (t *TimeTaskSchedule) RunTasks() {
	a := ioc.Ioc().Get(ioc.KeyTimerFunc)
	if a == nil {
		return
	}
	list := a.([]interface{})
	if len(list) == 0 {
		return
	}
	for _, task := range list {
		items := task.([]interface{})
		name := cast.ToString(items[0])
		perMinute := cast.ToInt64(items[1])
		fn := items[2].(func())
		t.RunTask(name, perMinute, fn)
	}
}

func (t *TimeTaskSchedule) RunTask(name string, perMinute int64, fn func()) {
	c := uioc.Cache()
	lk := c.CreateLock("timetask:lock:" + name)
	if !lk.TryLock() {
		return
	}
	defer lk.UnLock()
	var lastTime int64
	c.GetObject("timetask:runtime:"+name, &lastTime)
	if util.UnixSeconds()-lastTime >= perMinute*60 {
		c.SetObject("timetask:runtime:"+name, util.UnixSeconds(), time.Minute*time.Duration(2*perMinute))
		util.Go(fn)
	}
}

type TimeTask struct {
	Name      string
	PerMinute int64
	Fn        func()
}

func (t *TimeTaskSchedule) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.stopChan == nil {
		return
	}
	close(*t.stopChan)
	t.stopChan = nil
	if t.leader != nil {
		t.leader.UnLock()
		t.leader = nil
	}
}
