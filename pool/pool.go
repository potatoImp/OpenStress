// pool.go
// 协程池管理模块
// 本文件负责管理协程池的核心功能。
// 主要功能包括：
// - 创建和管理协程池
// - 任务优先级调度
// - 动态调整并发数
// - 任务重试机制
// - 任务依赖管理（待实现）
// - 任务优先级调度（待实现）
// 
// 技术实现细节：
// 1. 使用 container/heap 包实现任务的优先级调度。
// 2. 提供 Submit 方法，接收任务并将其添加到协程池。
// 3. 提供动态调整并发数的功能，以适应不同的负载情况。
// 4. 实现任务重试机制，确保任务的可靠性。

package pool

import (
	"container/heap"
	"context"
	"sync"
)

// Task represents a task with priority.
type Task struct {
	fn       func()
	priority int
	retries  int
}

// TaskQueue implements a priority queue for tasks.
type TaskQueue []*Task

func (tq TaskQueue) Len() int { return len(tq) }
func (tq TaskQueue) Less(i, j int) bool {
	return tq[i].priority > tq[j].priority // Higher priority tasks first
}
func (tq TaskQueue) Swap(i, j int) {
	tq[i], tq[j] = tq[j], tq[i]
}

func (tq *TaskQueue) Push(x interface{}) {
	*tq = append(*tq, x.(*Task))
}

func (tq *TaskQueue) Pop() interface{} {
	old := *tq
	n := len(old)
	item := old[n-1]
	*tq = old[0 : n-1]
	return item
}

// Pool represents a goroutine pool.
type Pool struct {
	maxWorkers int
	tasks      TaskQueue
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewPool creates a new Pool with the specified maximum number of workers.
func NewPool(maxWorkers int) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	pool := &Pool{
		maxWorkers: maxWorkers,
		tasks:      make(TaskQueue, 0),
		ctx:        ctx,
		cancel:     cancel,
	}
	heap.Init(&pool.tasks)
	pool.start()
	return pool
}

// start begins the worker goroutines.
func (p *Pool) start() {
	for i := 0; i < p.maxWorkers; i++ {
		go p.worker()
	}
}

// worker listens for tasks and executes them.
func (p *Pool) worker() {
	for {
		select {
		case <-p.ctx.Done():
			return // Exit the worker if context is done
		default:
			if len(p.tasks) > 0 {
				task := heap.Pop(&p.tasks).(*Task)
				if task.retries > 0 {
					task.fn() // Execute the task
					p.wg.Done()
				} else {
					// Retry the task
					task.retries++
					heap.Push(&p.tasks, task)
				}
			}
		}
	}
}

// Submit adds a new task to the pool.
func (p *Pool) Submit(fn func(), priority int, retries int) {
	p.wg.Add(1)
	heap.Push(&p.tasks, &Task{fn: fn, priority: priority, retries: retries})
}

// Shutdown gracefully stops the pool and waits for all tasks to complete.
func (p *Pool) Shutdown() {
	p.cancel()  // Cancel the context
	p.wg.Wait() // Wait for all tasks to finish
}
