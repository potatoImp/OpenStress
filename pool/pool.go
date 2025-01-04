package pool

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
)

// 引入日志模块
var stressLogger *StressLogger

// Task 任务结构体，定义每个任务的基本信息
type Task struct {
	ID         string
	fn         func(threadID int32) error // 任务执行的函数，接收一个 threadID 参数并返回错误
	priority   int                        // 任务优先级（数字越大优先级越高）
	retries    int                        // 重试次数
	maxRetries int                        // 最大重试次数
	timeout    time.Duration              // 任务超时时间
}

// Pool 任务池结构体
type Pool struct {
	maxWorkers  int
	taskList    sync.Map      // 使用 sync.Map 来管理任务，避免加锁
	taskPool    *ants.Pool    // ants 协程池
	stopChannel chan struct{} // 停止信号通道
	wg          sync.WaitGroup
}

// NewPool 创建并初始化任务池
func NewPool(maxWorkers int) (*Pool, error) {
	// 使用 ants.NewPool 来创建池
	pool, err := ants.NewPool(maxWorkers)
	if err != nil {
		return nil, err
	}
	return &Pool{
		maxWorkers:  maxWorkers,
		taskPool:    pool,
		stopChannel: make(chan struct{}),
	}, nil
}

// AddTask 添加单个任务到任务列表并排序
func (p *Pool) AddTask(fn func(threadID int32) error, priority int) {
	// 创建一个新的任务
	task := Task{
		ID:         fmt.Sprintf("task-%d", time.Now().UnixNano()), // 使用时间戳作为任务ID
		fn:         fn,
		priority:   priority, // 设置优先级
		retries:    0,        // 默认重试为0
		maxRetries: 3,        // 默认最大重试次数为3
		timeout:    0,        // 默认不设置超时时间
	}

	// 将任务添加到任务列表
	taskList := make([]Task, 0)
	p.taskList.Range(func(key, value interface{}) bool {
		taskList = append(taskList, value.(Task))
		return true
	})

	// 将任务添加到本地任务列表并按优先级排序
	taskList = append(taskList, task)
	sort.SliceStable(taskList, func(i, j int) bool {
		return taskList[i].priority > taskList[j].priority
	})

	// 将任务列表存回 sync.Map
	for i, t := range taskList {
		p.taskList.Store(i, t)
	}

	stressLogger.Log("INFO", fmt.Sprintf("Task %s added to the task list.", task.ID))
}

// executeWithRetry 执行任务的重试逻辑
func (task *Task) executeWithRetry(threadID int32) error {
	var retries int
	for {
		err := task.fn(threadID) // 执行任务
		if err == nil {
			return nil // 任务成功，退出
		}

		// 达到最大重试次数时退出
		if retries >= task.maxRetries {
			stressLogger.Log("ERROR", fmt.Sprintf("Task %s failed after %d retries.", task.ID, retries))
			return err
		}

		retries++
		// 使用指数退避策略来延迟重试
		time.Sleep(time.Duration(1<<retries) * time.Second) // 延迟 2^retries 秒
	}
}

// Start 启动任务池并循环执行任务
func (p *Pool) Start(runDuration time.Duration) {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()

		timeout := time.After(runDuration)               // 超时通道
		ticker := time.NewTicker(100 * time.Millisecond) // 公共定时器
		defer ticker.Stop()

		// 启动所有工作协程
		for i := 0; i < p.maxWorkers; i++ {
			err := p.taskPool.Submit(func() {
				threadID := int32(i)
				for {
					select {
					case <-ticker.C:
						// 在这里复制任务列表到本地缓存
						localTaskList := make([]Task, 0)
						p.taskList.Range(func(key, value interface{}) bool {
							localTaskList = append(localTaskList, value.(Task))
							return true
						})

						// 按优先级排序本地任务列表
						sort.SliceStable(localTaskList, func(i, j int) bool {
							return localTaskList[i].priority > localTaskList[j].priority
						})

						// 遍历本地缓存的任务列表并执行任务
						for _, task := range localTaskList {
							task.executeWithRetry(threadID) // 执行带重试的任务
						}

					case <-p.stopChannel: // 收到停止信号，退出
						stressLogger.Log("INFO", fmt.Sprintf("Worker %d received stop signal, stopping.", i))
						return
					}
				}
			})

			if err != nil {
				stressLogger.Log("ERROR", fmt.Sprintf("Failed to start worker %d: %v", i, err))
			} else {
				stressLogger.Log("INFO", fmt.Sprintf("Worker %d started successfully", i))
			}
		}

		// 任务池控制循环
		for {
			select {
			case <-timeout: // 超时，停止任务池
				stressLogger.Log("INFO", "Task pool reached specified runtime, stopping.")
				close(p.stopChannel) // 发送停止信号
				return
			case <-p.stopChannel: // 收到停止信号，停止任务池
				stressLogger.Log("INFO", "Received stop signal, stopping task pool.")
				return
			case <-ticker.C: // 每 100 毫秒检查一次
				// 可以在这里处理其他定时任务
			}
		}
	}()

	// 等待任务池中的所有工作协程完成
	p.wg.Wait()
}

// Stop 停止任务池
func (p *Pool) Stop() {
	// 主动发送停止信号
	close(p.stopChannel)
	p.wg.Wait()
}
