package testTasks

import (
	"fmt"
	"reflect"
	"strings"

	"OpenStress/internal/poolProvider"
	"OpenStress/pool"
	"OpenStress/result"
)

// 你可以在这里添加你需要的方法
type task struct{}

func Initialize() {
	// 遍历 acase 包中的所有函数并自动调用以 Acase 开头的函数
	callAcaseFunctions()
}

func callAcaseFunctions() {
	v := reflect.ValueOf(&task{}) // 反射获取 acase 类型的指针
	for i := 0; i < v.NumMethod(); i++ {
		method := v.Method(i)
		if strings.HasPrefix(v.Type().Method(i).Name, "Task") {
			// 打印正在执行的函数名
			fmt.Printf("Executing %s\n", v.Type().Method(i).Name)
			method.Call(nil) // 调用方法
		}
	}
}

func GetOpenStressPool() (*pool.Pool, *pool.StressLogger, *result.Collector, error) {
	taskPool := poolProvider.InitializePool()
	stressLogger, err := pool.GetLogger()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get logger: %w", err)
	}
	// result 模块测试方法
	collectorConfig := result.CollectorConfig{
		BatchSize:       10000,
		OutputFormat:    "jtl",
		JTLFilePath:     "path/to/jtl/file.jtl",
		Logger:          stressLogger,
		NumGoroutines:   10,
		CollectInterval: 5,
		TaskID:          "testTask",
	}
	collector, createcollectoreErr := result.GetCollector(collectorConfig)
	if createcollectoreErr != nil {
		stressLogger.Log("ERROR", "Failed to create collector: "+createcollectoreErr.Error())
		return nil, nil, nil, fmt.Errorf("failed to create collector: %w", err)
	}
	collector.InitializeCollector()

	return taskPool, stressLogger, collector, nil
}

func destroyOpenStressPool(taskPool *pool.Pool, StressLogger *pool.StressLogger, collector *result.Collector) {
	// 加载结果数据
	results, err := collector.LoadResultsFromFile()
	if err != nil {
		fmt.Printf("Error loading results: %v\n", err)
		return
	}
	// 生成并打印测试报告
	report := collector.GenerateSummaryReport(results)
	fmt.Println(report)
	// collector.Close()

	stats, err := collector.GeneratePerformanceStats(results)
	if err != nil {
		fmt.Println("Error generating stats:", err)
		return
	}
	fmt.Println("Performance Stats:")
	fmt.Println(stats)

	// 保存HTML报告到文件
	reportPath, err := collector.SaveReportToFile(stats, "01X批次OpenStress产品基准测试报告")
	if err != nil {
		fmt.Println("Error saving report:", err)
		return
	}

	// 输出生成的报告路径
	fmt.Printf("测试报告已生成：%s\n", reportPath)
}
