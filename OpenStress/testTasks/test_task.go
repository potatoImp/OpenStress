package testTasks

import (
	"fmt"
	"net/http"

	"time"

	"OpenStress/result"
)

func (testTask *task) TaskPool() {
	// maxWorkers := 100
	// taskPool := pool.NewPool(maxWorkers)

	taskPool, stressLogger, collector, getPoolErr := GetOpenStressPool()

	if getPoolErr != nil {
		fmt.Println("failed to get logger: %w", getPoolErr)
	}

	// 定义高优先级任务
	highPriorityTask := func(threadID int32) {
		// time.Sleep(1 * time.Second) // 模拟任务执行时间
		startTime := time.Now()

		resp, err := http.Get("http://10.10.27.111:8089/index.html")
		if err != nil {
			// 连接失败时处理错误
			// fmt.Println("Request failed:", err)
			return // 可以提前返回，避免执行到 defer 语句
		}
		defer resp.Body.Close()
		if err != nil {
			collector.SaveFailureResult(result.ResultData{
				ID:           "test1",
				Type:         result.Failure,
				ResponseTime: 0,
				StartTime:    startTime,
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
		defer resp.Body.Close()
		// fmt.Printf("请求成功，状态码: %d\n", resp.StatusCode)
		collector.SaveSuccessResult(result.ResultData{
			ID:           "test1",
			Type:         result.Success,
			ResponseTime: 0,
			StartTime:    startTime,
			// EndTime:      time.Now().Add(120 * time.Millisecond),
			EndTime:      time.Now(),
			StatusCode:   200,
			Method:       "GET",
			URL:          "http://example.com",
			DataSent:     1024,
			DataReceived: 2048,
			ThreadID:     int(threadID),
		})
	}

	// 提交高优先级任务
	for i := 1; i <= 100000; i++ {
		taskID := fmt.Sprintf("请求resources-8080-%d", i)
		taskPool.Submit(highPriorityTask, 3, taskID, 1*time.Second) // 高优先级
	}

	stressLogger.Log("INFO", "测试任务开始执行")
	// 启动任务池
	// taskPool.Start()
	// taskPool.StartByDuration()
	// 每500毫秒检查一次任务状态
	// taskPool.WaitForTasksComplete(500 * time.Millisecond)
	// 关闭任务池
	taskPool.Shutdown()

	//启动任务定时轮询检查任务队列中是否还有任务，若没有则等1秒后开始执行销毁
	// 初始化报告
	// collector.InitializeReport()
	time.Sleep(1 * time.Second)
	destroyOpenStressPool(taskPool, stressLogger, collector)
}
