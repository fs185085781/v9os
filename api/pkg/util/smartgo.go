package util

import (
	"container/list"
	"math"
	"runtime"
	"sync"
)

// 代替go的协程,可以无脑使用,内部会根据系统状态动态调整执行频率,无需担心panic
type SmartGoStruct struct {
	taskList    *list.List
	mutex       sync.Mutex
	cycle       *CycleScheduler
	maxList     int
	onceLen     int
	lastRunTime int64
	lazySeconds int64
}

func (s *SmartGoStruct) runTask(task func()) {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	task()
}

func (s *SmartGoStruct) asyncRunTask(task func()) {
	go s.runTask(task)
}

func (s *SmartGoStruct) systemIsOk() bool {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	gcCpu := m.GCCPUFraction
	if gcCpu >= 0.3 {
		seconds := int64(math.Pow(2, 10*gcCpu-3) * 5)
		onceLen := int(8 - 10*gcCpu)
		if onceLen < 1 {
			onceLen = 1
		}
		s.lazySeconds = seconds
		s.onceLen = onceLen
		return false
	}
	recentLongGC := false
	pauseThreshold := uint64(5 * 1000 * 1000)
	for i := range 3 {
		index := (m.NumGC + 255 - uint32(i)) % 256
		if m.PauseNs[index] > pauseThreshold {
			recentLongGC = true
			break
		}
	}
	if recentLongGC {
		s.lazySeconds = 10
		s.onceLen = 3
		return false
	}
	s.onceLen = 5
	return true

}

func (s *SmartGoStruct) checkAndRunTask() {
	if !s.systemIsOk() {
		now := UnixSeconds()
		if now-s.lastRunTime < s.lazySeconds {
			return
		}
		s.lastRunTime = now
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.taskList.Len() == 0 {
		if s.cycle != nil {
			s.cycle.Stop()
			s.cycle = nil
		}
		return
	}
	tasksToProcess := min(s.taskList.Len(), s.onceLen)
	for range tasksToProcess {
		e := s.taskList.Front()
		if e != nil {
			task := s.taskList.Remove(e).(func())
			s.asyncRunTask(task)
		}
	}
}

var smartGoInstance *SmartGoStruct

func Go(task func()) {
	if task == nil {
		return
	}
	//初始化实例
	if smartGoInstance == nil {
		globalMu.Lock()
		defer globalMu.Unlock()
		if smartGoInstance == nil {
			smartGoInstance = &SmartGoStruct{
				taskList: list.New(),
				maxList:  200,
			}
		}
	}
	//如果系统空闲则直接执行
	if smartGoInstance.systemIsOk() {
		smartGoInstance.asyncRunTask(task)
	} else {
		//将任务加入队列
		smartGoInstance.mutex.Lock()
		defer smartGoInstance.mutex.Unlock()
		if smartGoInstance.taskList.Len() > smartGoInstance.maxList {
			smartGoInstance.asyncRunTask(task)
			return
		}
		smartGoInstance.taskList.PushBack(task)
		//初始化监视器
		if smartGoInstance.cycle == nil {
			smartGoInstance.cycle = CycleRun(3, smartGoInstance.checkAndRunTask)
		}
	}
}
