package main

import (
	"fmt"
)

// 使用 channel 实现的简单屏障示例
func main() {
	numWorkers := 5
	barrier := make(chan struct{})
	ready := make(chan struct{})
	done := make(chan struct{})

	for i := 0; i < numWorkers; i++ {
		go func(id int) {
			// 第一阶段工作
			fmt.Println("Worker", id, "phase 1 complete")

			// 等待所有协程到达屏障点
			ready <- struct{}{}
			<-barrier

			// 所有协程同时开始第二阶段
			fmt.Println("Worker", id, "phase 2 starting")
			done <- struct{}{}
		}(i)
	}

	// 等待所有工作协程到达屏障点
	for i := 0; i < numWorkers; i++ {
		<-ready
	}

	// 释放屏障，所有协程继续执行
	close(barrier)

	// 等待所有工作协程完成第二阶段
	for i := 0; i < numWorkers; i++ {
		<-done
	}
	fmt.Println("All workers completed phase 2")
}
