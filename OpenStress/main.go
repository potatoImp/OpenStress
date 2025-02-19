package main

import (
	"OpenStress/configs"
	"OpenStress/pool"
	"reflect"

	// "time"

	"OpenStress/testTasks"
	"fmt"
	// "OpenStress/result"
)

var logger *pool.StressLogger

func init() {

	// 调用 InitializeLogConfig 函数
	logDetails, logConfigErr := configs.InitializeLogConfig()
	if logConfigErr != nil {
		fmt.Println("Failed to initialize log config:", logConfigErr)
		return
	}

	// 打印配置值及其类型
	fmt.Println("Log Configuration:")
	printConfigValues(*logDetails)

	// 初始化日志记录器
	// logDir := "./logs/"
	// logFile := "app.log"
	moduleName := "MainModule"

	// stressLogger, err := pool.InitializeLogger(logDir, logFile, moduleName)
	stressLogger, err := pool.InitializeLogger(logDetails.Directory, logDetails.Filename, moduleName, logDetails.Level, logDetails.MaxSize, logDetails.MaxAge)

	if err != nil {
		fmt.Println("Error initializing logger:", err)
		return
	}
	defer stressLogger.Close()

	// 初始化配置
	_, baseConfigErr := configs.InitializeBaseConfig()
	if baseConfigErr != nil {
		fmt.Println("Error initializing base config:", baseConfigErr)
		return
	}

	fmt.Println("Base config initialized successfully.》》》》》》》》》》》》》》》》》》》》》》》")
	// 获取 BaseDetails
	baseDetails := configs.GetBaseDetails()
	if baseDetails != nil {
		fmt.Printf("Users: %d\n", baseDetails.Users)
		fmt.Printf("Task Default Repeats: %d\n", baseDetails.TaskDefaultRepeats)
		fmt.Printf("Task Interval: %d\n", baseDetails.TaskInterval)
		// 打印其他配置字段
	}
	// 记录日志
	// stressLogger.Log("INFO", "This is an info message.")
	// stressLogger.Log("WARN", "This is a warning message.")
	// stressLogger.Log("ERROR", "This is an error message.")
	// stressLogger.Log("DEBUG", "This is a debug message.")

	// pool.Initialize()
	configs.Initialize()
	testTasks.Initialize()
}

func mainInitialize() {

	// 调用 InitializeLogConfig 函数
	// logDetails, err := configs.InitializeLogConfig()
	// if err != nil {
	// 	fmt.Println("Failed to initialize log config:", err)
	// 	return
	// }

	// // 打印配置值及其类型
	// fmt.Println("Log Configuration:")
	// printConfigValues(*logDetails)

	// // 初始化日志记录器
	// logDir := "./logs/"
	// logFile := "app.log"
	// moduleName := "MainModule"

	// stressLogger, err := pool.InitializeLogger(logDir, logFile, moduleName)
	// // stressLogger, err := pool.InitializeLogger(logDetails.Directory, logDetails.Filename, moduleName)
	// if err != nil {
	// 	fmt.Println("Error initializing logger:", err)
	// 	return
	// }
	// defer stressLogger.Close()
}

func main() {

	// mainInitialize()
	// configs.Initialize()
	// testTasks.Initialize()

}

// 打印配置值及其类型的辅助函数
func printConfigValues(logDetails configs.LogDetails) {
	v := reflect.ValueOf(logDetails)
	typeOfS := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fmt.Printf("%s: %v (%s)\n", typeOfS.Field(i).Name, v.Field(i).Interface(), v.Field(i).Type())
	}
}

// // 初始化日志记录器
// logDir := "./logs/"
// logFile := "app.log"
// moduleName := "MainModule"

// stressLogger, err := pool.InitializeLogger(logDir, logFile, moduleName)
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

// // pool.Initialize()
// configs.Initialize()
// testTasks.Initialize()

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

// 初始化日志记录器
// logDir := "./logs/"
// logFile := "app.log"
// moduleName := "MainModule"

// stressLogger, err := pool.InitializeLogger(logDir, logFile, moduleName)
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
// tests.TestTaskPool1()

// 这里可以添加更多的逻辑来使用 config
// 例如，您可以根据配置初始化 LLM 提供者等

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

// }

// // handleError 处理错误并记录日志
// func handleError(err error) {
// 	if err != nil {
// 		// 初始化日志记录器
// 		logDir := "./logs/"
// 		logFile := "app.log"
// 		moduleName := "MainModule"

// 		stressLogger, logErr := pool.InitializeLogger(logDir, logFile, moduleName)
// 		if logErr != nil {
// 			fmt.Println("Error initializing logger:", logErr)
// 			return
// 		}
// 		defer stressLogger.Close()
// 		// 这里可以调用日志记录器记录错误信息
// 		stressLogger.Log("ERROR", err.Error())
// 		stressLogger.Log("INFO", "Test log message")
// 	}
// }
