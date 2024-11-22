package cron

import "time"

// Interface 执行顺序 Before->Process->Run->Defer
// 只建议重写 Init 和 Process
// Init 添加固定执行的脚本，使用 AddBaseTaskList 添加，尽量通过该方法添加
// Process 添加运行中才能确定执行的脚本，使用 AddProcessAddTaskList 添加
type Interface interface {
	GetIsRun() bool
	AddBaseTaskList(task func())
	AddProcessAddTaskList(task func())

	// Init 初始化只会执行一次
	Init(base Interface)
	Process()

	Before(name, lockName string, deadLockExpireTime time.Duration, maxConcurrencyCnt int)
	Run()
	Defer()
}
