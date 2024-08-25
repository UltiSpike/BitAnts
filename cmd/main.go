package main

import (
	"bit-ants/internal/service"
)

func main() {
	// 每秒最大任务完成数量控制的是最大开销
	maxWorksPerSeconds := 100
	// 协程池的协程数量保证的是工作能力
	maxWorkers := 10
	scheduler, err := service.NewScheduler(maxWorkers, maxWorksPerSeconds)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 10000; i++ {
		scheduler.Submit(func() {
		})
	}
	scheduler.WaitUntilFinish()
}
