package main

import (
	"OpenStress/pool"
	"fmt"
)

func main() {
	// 创建一个新的任务池
	taskPool := pool.NewPool(5) // 假设最大工作线程数为 5
	defer taskPool.Shutdown()   // 确保在退出时优雅地关闭任务池

	// 加载任务到任务池
	// pool.LoadTasks(taskPool)

	pool.LoadTasks2(taskPool)
	// // 启动任务池
	// taskPool.Start()

	// 这里可以添加其他逻辑，例如等待用户输入或其他操作
	fmt.Println("Tasks have been loaded and the pool is running.")
}
