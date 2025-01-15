package tests

import (
	"OpenStress/pool"
	"fmt"

	// "net/http"
	"time"

	"OpenStress/result"
)

// TestTaskPool 测试任务池的功能
func TestTaskPool1() {
	maxWorkers := 100
	taskPool := pool.NewPool(maxWorkers)

	stressLogger, _ := pool.GetLogger()
	// result 模块测试方法
	collectorConfig := result.CollectorConfig{
		BatchSize:       10,
		OutputFormat:    "jtl",
		JTLFilePath:     "path/to/jtl/file.jtl",
		Logger:          stressLogger,
		NumGoroutines:   2,
		CollectInterval: 5,
		TaskID:          "testTask",
	}
	collector, err := result.NewCollector(collectorConfig)
	if err != nil {
		stressLogger.Log("ERROR", "Failed to create collector: "+err.Error())
	}
	collector.InitializeCollector()

	// 定义高优先级任务
	highPriorityTask := func(threadID int32) {
		time.Sleep(1 * time.Second) // 模拟任务执行时间

		// resp, err := http.Get("http://10.10.27.111:8089/index.html")
		// if err != nil {
		// 	// 连接失败时处理错误
		// 	fmt.Println("Request failed:", err)
		// 	return // 可以提前返回，避免执行到 defer 语句
		// }
		// defer resp.Body.Close()
		if err != nil {
			collector.SaveFailureResult(result.ResultData{
				ID:           "test1",
				Type:         result.Failure,
				ResponseTime: 0,
				StartTime:    time.Now(),
				EndTime:      time.Now().Add(120 * time.Millisecond),
				StatusCode:   404,
				Method:       "GET",
				URL:          "http://10.10.27.111:8089/index.html",
				DataSent:     1024,
				DataReceived: 2048,
				ThreadID:     int(threadID),
			})
			fmt.Printf("请求失败: %v\n", err)
			return
		}
		// defer resp.Body.Close()
		// fmt.Printf("请求成功，状态码: %d\n", resp.StatusCode)
		collector.SaveSuccessResult(result.ResultData{
			ID:           "test1",
			Type:         result.Success,
			ResponseTime: 0,
			StartTime:    time.Now(),
			EndTime:      time.Now().Add(120 * time.Millisecond),
			StatusCode:   200,
			Method:       "GET",
			URL:          "http://example.com",
			DataSent:     1024,
			DataReceived: 2048,
			ThreadID:     int(threadID),
		})

		collector.SaveFailureResult(result.ResultData{
			ID:           "test1",
			Type:         result.Failure,
			ResponseTime: 2 * time.Millisecond,
			StartTime:    time.Now(),
			EndTime:      time.Now().Add(120 * time.Millisecond),
			StatusCode:   404,
			Method:       "GET",
			URL:          "http://10.10.27.111:8089/index.html",
			DataSent:     1024,
			DataReceived: 2048,
			ThreadID:     int(threadID),
		})
	}

	// 定义中优先级任务
	mediumPriorityTask := func(threadID int32) {
		fmt.Println("执行中优先级任务")
		time.Sleep(2 * time.Second) // 模拟任务执行时间
		fmt.Println("中优先级任务完成")
	}

	// 定义低优先级任务
	lowPriorityTask := func(threadID int32) {
		fmt.Println("执行低优先级任务")
		time.Sleep(3 * time.Second) // 模拟任务执行时间
		fmt.Println("低优先级任务完成")
	}

	// 提交高优先级任务
	for i := 1; i <= 100; i++ {
		taskID := fmt.Sprintf("请求resources-8080-%d", i)
		taskPool.Submit(highPriorityTask, 3, taskID, 1*time.Second) // 高优先级
	}

	// 提交中优先级任务
	for i := 1; i <= 2; i++ {
		taskID := fmt.Sprintf("Medium-Priority-Task-%d", i)
		taskPool.Submit(mediumPriorityTask, 2, taskID, 5*time.Second) // 中优先级
	}

	// 提交低优先级任务
	for i := 1; i <= 2; i++ {
		taskID := fmt.Sprintf("Low-Priority-Task-%d", i)
		taskPool.Submit(lowPriorityTask, 1, taskID, 1*time.Second) // 低优先级
	}

	// 启动任务池
	taskPool.Start()

	// 关闭任务池
	taskPool.Shutdown()

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

	// collector.CloseCollector()
}
