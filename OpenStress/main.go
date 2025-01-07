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

	// pool 模块测试方法
	tests.TestTask_AD()
	// tests.TestTaskPool1()

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
