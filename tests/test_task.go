package tests

import (
	"OpenStress/pool"
	"fmt"
	"net/http"
	"time"
)

// TestTaskPool 测试任务池的功能
func TestTaskPool1() {
	maxWorkers := 1000
	taskPool := pool.NewPool(maxWorkers, true)

	// 定义高优先级任务
	highPriorityTask := func() {
		resp, err := http.Get("http://10.10.27.111:8089/index.html")
		if err != nil {
			fmt.Printf("请求失败: %v\n", err)
			return
		}
		defer resp.Body.Close()
		fmt.Printf("请求成功，状态码: %d\n", resp.StatusCode)
	}

	// 定义中优先级任务
	mediumPriorityTask := func() {
		fmt.Println("执行中优先级任务")
		time.Sleep(2 * time.Second) // 模拟任务执行时间
		fmt.Println("中优先级任务完成")
	}

	// 定义低优先级任务
	lowPriorityTask := func() {
		fmt.Println("执行低优先级任务")
		time.Sleep(3 * time.Second) // 模拟任务执行时间
		fmt.Println("低优先级任务完成")
	}

	// 提交高优先级任务
	for i := 1; i <= 10; i++ {
		taskID := fmt.Sprintf("High-Priority-Task-%d", i)
		taskPool.Submit(highPriorityTask, 3, 2, taskID, 5*time.Second) // 高优先级
	}

	// 提交中优先级任务
	for i := 1; i <= 5; i++ {
		taskID := fmt.Sprintf("Medium-Priority-Task-%d", i)
		taskPool.Submit(mediumPriorityTask, 2, 2, taskID, 5*time.Second) // 中优先级
	}

	// 提交低优先级任务
	for i := 1; i <= 5; i++ {
		taskID := fmt.Sprintf("Low-Priority-Task-%d", i)
		taskPool.Submit(lowPriorityTask, 1, 2, taskID, 5*time.Second) // 低优先级
	}

	// 启动任务池
	taskPool.Start()

	// 关闭任务池
	taskPool.Shutdown()
}
