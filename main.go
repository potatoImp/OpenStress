package main

import (
	"OpenStress/pool"
	"fmt"
	"time"
)

var logger *pool.StressLogger

func main() {

	// 初始化日志记录器
	logDir := "./logs/"
	logFile := "app.log"
	var err error
	logger, err = pool.InitializeLogger(logDir, logFile, "MainModule")
	if err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		return
	}
	defer logger.Close() // 确保在程序结束时关闭日志记录器
	// // 创建一个新的任务池
	// taskPool := pool.NewPool(5) // 假设最大工作线程数为 5
	// defer taskPool.Shutdown()   // 确保在退出时优雅地关闭任务池

	// // 加载任务到任务池
	// // pool.LoadTasks(taskPool)

	// pool.LoadTasks2(taskPool)
	// // // 启动任务池
	// // taskPool.Start()

	// // 这里可以添加其他逻辑，例如等待用户输入或其他操作
	// fmt.Println("Tasks have been loaded and the pool is running.")

	// // 初始化日志记录器
	// logDir := "./logs/"
	// logFile := "app.log"
	// moduleName := "MainModule"

	// stressLogger, err := pool.NewStressLogger(logDir, logFile, moduleName)
	// if err != nil {
	// 	fmt.Println("Error initializing logger:", err)
	// 	return
	// }
	// defer stressLogger.Close()

	// // 记录日志
	// stressLogger.Log("INFO", "This is an info message.")
	// stressLogger.Log("WARN", "This is a warning message.")
	// stressLogger.Log("ERROR", "This is an error message.")
	// stressLogger.Log("DEBUG", "This is a debug message.")

	// // 模拟一些处理
	// time.Sleep(2 * time.Second)

	// fmt.Println("Logging completed.")

	// // 创建一个 CustomError 实例
	// err := &pool.CustomError{
	// 	Message:   "Something went wrong222",
	// 	Code:      500,
	// 	Timestamp: time.Now(),
	// 	Stack:     "main.go:42", // 示例调用栈信息
	// }

	// // 输出错误信息
	// fmt.Println(err.Error())

	// // 模拟错误处理并记录日志
	// handleError(err)

	// 创建一个新的协程池
	// 创建一个新的协程池
	maxWorkers := 5
	pool := pool.NewPool(maxWorkers)

	// 定义任务
	taskFunc := func(taskID string, duration time.Duration) func() {
		return func() {
			fmt.Printf("Task %s is running...\n", taskID)
			time.Sleep(duration) // 模拟任务执行
			fmt.Printf("Task %s completed.\n", taskID)
		}
	}

	// 提交不同优先级的任务到池中
	// 高优先级任务
	for i := 1; i <= 3; i++ {
		taskID := fmt.Sprintf("High-Priority-Task-%d", i)
		pool.Submit(taskFunc(taskID, 1*time.Second), 3, 2, taskID, 5*time.Second) // 高优先级
	}

	// 中优先级任务
	for i := 1; i <= 3; i++ {
		taskID := fmt.Sprintf("Medium-Priority-Task-%d", i)
		pool.Submit(taskFunc(taskID, 2*time.Second), 2, 2, taskID, 5*time.Second) // 中优先级
	}

	// 低优先级任务
	for i := 1; i <= 4; i++ {
		taskID := fmt.Sprintf("Low-Priority-Task-%d", i)
		pool.Submit(taskFunc(taskID, 3*time.Second), 1, 2, taskID, 5*time.Second) // 低优先级
	}

	// 启动协程池
	pool.Start()
	fmt.Println("Pool started.")

	// 等待一段时间后关闭池
	time.Sleep(15 * time.Second)
	pool.Shutdown()
	fmt.Println("Pool shutdown.")
}

// handleError 处理错误并记录日志
func handleError(err error) {
	if err != nil {
		// 初始化日志记录器
		logDir := "./logs/"
		logFile := "app.log"
		moduleName := "MainModule"

		stressLogger, logErr := pool.InitializeLogger(logDir, logFile, moduleName)
		if logErr != nil {
			fmt.Println("Error initializing logger:", logErr)
			return
		}
		defer stressLogger.Close()
		// 这里可以调用日志记录器记录错误信息
		stressLogger.Log("ERROR", err.Error())
		stressLogger.Log("INFO", "Test log message")
	}
}
