package util

import (
	"time"
)

// CycleScheduler 周期任务调度器
type CycleScheduler struct {
	interval time.Duration // 执行间隔
	task     func()        // 要执行的任务函数
	stopChan chan struct{} // 停止信号通道
}

// Run 启动周期任务执行
// intervalSeconds: 执行间隔（秒）
// task: 要周期执行的任务函数
func CycleRun(intervalSeconds int, task func()) *CycleScheduler {
	cs := &CycleScheduler{
		stopChan: make(chan struct{}),
	}
	cs.interval = time.Duration(intervalSeconds) * time.Second
	cs.task = task
	// 启动新的定时任务
	go cs.startTicker()
	return cs
}

// Stop 停止当前正在执行的任务
func (cs *CycleScheduler) Stop() {
	close(cs.stopChan)
}

// startTicker 内部方法，启动定时器执行任务
func (cs *CycleScheduler) startTicker() {
	ticker := time.NewTicker(cs.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 安全执行任务，捕获可能的panic
			cs.safeExecuteTask()
		case <-cs.stopChan:
			return // 收到停止信号，退出goroutine
		}
	}
}

// safeExecuteTask 安全地执行任务，防止panic导致调度器崩溃
func (cs *CycleScheduler) safeExecuteTask() {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	cs.task()
}
