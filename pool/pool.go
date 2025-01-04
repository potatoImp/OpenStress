package pool

import (
	"fmt"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
)

// 引入日志模块
var stressLogger *StressLogger

// Task 任务结构体，定义每个任务的基本信息
type Task struct {
	ID string
	fn func(threadID int32) error // 任务执行的函数，接收一个 threadID 参数并返回错误
}

// Pool 任务池结构体
type Pool struct {
	maxWorkers  int
	taskList    []Task        // 任务列表，直接存储任务
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

// AddTask 添加单个任务到任务列表
func (p *Pool) AddTask(fn func(threadID int32) error) {
	// 创建一个新的任务
	task := Task{
		ID: fmt.Sprintf("task-%d", len(p.taskList)+1), // 自动生成任务ID
		fn: fn,
	}

	// 将任务添加到任务列表
	p.taskList = append(p.taskList, task)
}

// execute 执行任务
func (task *Task) execute(threadID int32) error {
	return task.fn(threadID) // 执行任务
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
						// 遍历任务列表并执行任务
						for _, task := range p.taskList {
							task.execute(threadID) // 执行任务
						}

					case <-p.stopChannel: // 收到停止信号，退出
						return
					}
				}
			})

			if err != nil {
				// 如果提交任务失败，打印错误
				fmt.Printf("Failed to start worker %d: %v\n", i, err)
			}
		}

		// 任务池控制循环
		for {
			select {
			case <-timeout: // 超时，停止任务池
				close(p.stopChannel) // 发送停止信号
				return
			case <-p.stopChannel: // 收到停止信号，停止任务池
				return
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
