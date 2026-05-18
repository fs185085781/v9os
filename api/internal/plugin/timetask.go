package plugin

import (
	"sync"
	"time"

	"github.com/fs185085781/v9os/internal/ioc"
	"github.com/fs185085781/v9os/internal/ioc/uioc"
	"github.com/fs185085781/v9os/pkg/util"
	"github.com/spf13/cast"
)

type TimeTaskSchedule struct {
	mu       sync.Mutex
	stopChan *chan struct{}
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
		ticker := time.NewTicker(time.Second * 20)
		defer ticker.Stop()
		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				t.RunTasks()
			}
		}
	})
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
}
