package main

import (
	"OpenStress/pool"
	// "time"

	"OpenStress/tests"
	"fmt"
	// "OpenStress/result"
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

	// pool 模块测试方法
	// tests.TestTask_AD()
	tests.TestTaskPool1()

	// // result 模块测试方法
	// collectorConfig := result.CollectorConfig{
	// 	BatchSize:       10,
	// 	OutputFormat:    "jtl",
	// 	JTLFilePath:     "path/to/jtl/file.jtl",
	// 	Logger:          logger,
	// 	NumGoroutines:   2,
	// 	CollectInterval: 5,
	// 	TaskID:          "testTask",
	// }
	// collector, err := result.NewCollector(collectorConfig)
	// if err != nil {
	// 	logger.Log("ERROR", "Failed to create collector: "+err.Error())
	// }
	// collector.InitializeCollector()

	// // 模拟收集数据
	// collector.CollectDataWithParams("test1", time.Now(), time.Now(), 200, "GET", "http://example.com", 1024, 2048, 1)

	// collector.SaveFailureResult(result.ResultData{
	// 	ID:           "test1",
	// 	Type:         result.Failure,
	// 	ResponseTime: 120 * time.Millisecond,
	// 	StartTime:    time.Now(),
	// 	EndTime:      time.Now().Add(120 * time.Millisecond),
	// 	StatusCode:   200,
	// 	Method:       "GET",
	// 	URL:          "http://example.com",
	// 	DataSent:     1024,
	// 	DataReceived: 2048,
	// 	ThreadID:     1,
	// })

	// collector.SaveSuccessResult(result.ResultData{
	// 	ID:           "test1",
	// 	Type:         result.Success,
	// 	ResponseTime: 120 * time.Millisecond,
	// 	StartTime:    time.Now(),
	// 	EndTime:      time.Now().Add(120 * time.Millisecond),
	// 	StatusCode:   200,
	// 	Method:       "GET",
	// 	URL:          "http://example.com",
	// 	DataSent:     1024,
	// 	DataReceived: 2048,
	// 	ThreadID:     1,
	// })

	// collector.Close()

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
