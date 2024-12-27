package pool

import (
	"container/heap"
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

// 引入日志模块
var stressLogger *StressLogger

// Task represents a task with priority and retry settings.
type Task struct {
	ID         string
	fn         func() // Task execution function (no error returned)
	priority   int
	retries    int
	maxRetries int           // Maximum retry attempts
	timeout    time.Duration // Task execution timeout
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

// Pool represents a goroutine pool with dynamic concurrency and priority scheduling.
type Pool struct {
	maxWorkers    int32 // max workers
	activeWorkers int32 // active workers
	tasks         TaskQueue
	ctx           context.Context
	cancel        context.CancelFunc
	isPaused      int32         // 0 means running, 1 means paused
	shutdownFlag  int32         // 0 means not shutdown, 1 means shutdown
	workerExit    chan struct{} // Channel to signal workers to exit gracefully
	isRepeating   bool          // Flag to control repeating task execution
}

// NewPool creates a new Pool with the specified maximum number of workers.
func NewPool(maxWorkers int) *Pool {
	stressLogger.Log("INFO", fmt.Sprintf("Creating a new pool with %d workers", maxWorkers))

	ctx, cancel := context.WithCancel(context.Background())
	pool := &Pool{
		maxWorkers: int32(maxWorkers),
		tasks:      make(TaskQueue, 0),
		ctx:        ctx,
		cancel:     cancel,
		workerExit: make(chan struct{}),
	}
	heap.Init(&pool.tasks)
	pool.Start()

	stressLogger.Log("INFO", "Pool created successfully")
	return pool
}

// Start initializes the worker goroutines.
func (p *Pool) Start() {
	stressLogger.Log("INFO", "Starting worker goroutines...")

	// Use atomic operations to control max concurrency
	for i := 0; i < int(p.maxWorkers); i++ {
		go p.worker()
	}

	stressLogger.Log("INFO", fmt.Sprintf("%d worker goroutines started", p.maxWorkers))
}

// worker listens for tasks and executes them.
func (p *Pool) worker() {
	for {
		select {
		case <-p.workerExit:
			stressLogger.Log("DEBUG", "Worker exiting")
			return // Exit the worker if notified to stop
		default:
			if atomic.LoadInt32(&p.shutdownFlag) == 1 {
				stressLogger.Log("INFO", "Pool shutdown detected, worker exiting")
				return
			}

			// Check for pause status
			if atomic.LoadInt32(&p.isPaused) == 1 {
				stressLogger.Log("DEBUG", "Pool is paused, worker waiting")
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// If tasks are empty and we are repeating, keep executing initial tasks
			if len(p.tasks) == 0 && p.isRepeating {
				stressLogger.Log("DEBUG", "Queue is empty, but repeating tasks are active")
				// Re-push the initial tasks to the queue
				p.rePushTasks()
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Try to pop a task from the queue and execute it
			if len(p.tasks) > 0 {
				task := heap.Pop(&p.tasks).(*Task)
				stressLogger.Log("INFO", fmt.Sprintf("Executing task %s", task.ID))
				if task.retries < task.maxRetries {
					p.executeTask(task) // Execute the task (no error handling)
					// If task retries exceed max, do not retry
					atomic.AddInt32(&p.activeWorkers, -1)
				} else {
					stressLogger.Log("WARN", fmt.Sprintf("Task %s exceeded max retries", task.ID))
					atomic.AddInt32(&p.activeWorkers, -1)
				}
			} else {
				stressLogger.Log("DEBUG", "Queue is empty, worker sleeping")
				// Queue is empty, sleep for a bit
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

// executeTask executes a task and manages its timeout.
func (p *Pool) executeTask(task *Task) {
	stressLogger.Log("INFO", fmt.Sprintf("Executing task %s with timeout %v", task.ID, task.timeout))

	// Create a context with timeout for the task
	ctx, cancel := context.WithTimeout(p.ctx, task.timeout)
	defer cancel()

	// Execute task in a goroutine to handle timeout
	done := make(chan struct{}, 1)

	go func() {
		stressLogger.Log("DEBUG", fmt.Sprintf("Task %s started execution", task.ID))
		task.fn() // Execute the task function (no error handling)
		done <- struct{}{}
	}()

	select {
	case <-done:
		stressLogger.Log("INFO", fmt.Sprintf("Task %s completed successfully", task.ID))
	case <-ctx.Done():
		stressLogger.Log("ERROR", fmt.Sprintf("Task %s timeout exceeded", task.ID))
	}
}

// Submit adds a new task to the pool.
func (p *Pool) Submit(fn func(), priority int, retries int, taskID string, timeout time.Duration, isRepeating bool) {
	stressLogger.Log("INFO", fmt.Sprintf("Submitting task %s with priority %d", taskID, priority))

	task := &Task{
		ID:         taskID,
		fn:         fn,
		priority:   priority,
		retries:    retries,
		maxRetries: 3, // Maximum retries
		timeout:    timeout,
	}

	// If repeating tasks are enabled, mark the flag
	if isRepeating {
		p.isRepeating = true
	}

	heap.Push(&p.tasks, task)
	atomic.AddInt32(&p.activeWorkers, 1)

	stressLogger.Log("INFO", fmt.Sprintf("Task %s submitted successfully", taskID))
}

// rePushTasks re-pushes initial tasks to the queue to ensure they keep running.
func (p *Pool) rePushTasks() {
	stressLogger.Log("INFO", "Re-pushing initial tasks to the queue")

	// Here we just re-push the tasks to the queue (you can modify this to track specific tasks)
	for _, task := range p.tasks {
		heap.Push(&p.tasks, task)
	}
}

// Shutdown gracefully stops the pool and waits for all tasks to complete.
func (p *Pool) Shutdown() {
	stressLogger.Log("INFO", "Shutting down the pool")
	atomic.StoreInt32(&p.shutdownFlag, 1)
	p.cancel()
	close(p.workerExit)
	stressLogger.Log("INFO", "Pool shutdown completed")
}

// Pause pauses the pool, preventing any new tasks from starting.
func (p *Pool) Pause() {
	stressLogger.Log("INFO", "Pausing the pool")
	atomic.StoreInt32(&p.isPaused, 1)
	stressLogger.Log("INFO", "Pool paused")
}

// Resume resumes the pool, allowing tasks to start again.
func (p *Pool) Resume() {
	stressLogger.Log("INFO", "Resuming the pool")
	atomic.StoreInt32(&p.isPaused, 0)
	stressLogger.Log("INFO", "Pool resumed")
}

// AdjustWorkers dynamically adjusts the number of worker goroutines.
func (p *Pool) AdjustWorkers(newWorkerCount int) {
	stressLogger.Log("INFO", fmt.Sprintf("Adjusting workers to %d", newWorkerCount))

	currentWorkers := atomic.LoadInt32(&p.activeWorkers)
	if int(currentWorkers) < newWorkerCount {
		// Start more workers
		for i := 0; i < newWorkerCount-int(currentWorkers); i++ {
			go p.worker()
		}
	} else {
		// Reduce workers gracefully
		atomic.StoreInt32(&p.maxWorkers, int32(newWorkerCount))
	}

	stressLogger.Log("INFO", fmt.Sprintf("Worker count adjusted to %d", newWorkerCount))
}

// GetTaskStatus returns the status of a task by its ID.
func (p *Pool) GetTaskStatus(taskID string) (*Task, error) {
	stressLogger.Log("INFO", fmt.Sprintf("Fetching status for task %s", taskID))

	for _, task := range p.tasks {
		if task.ID == taskID {
			stressLogger.Log("INFO", fmt.Sprintf("Task %s found", taskID))
			return task, nil
		}
	}

	stressLogger.Log("ERROR", fmt.Sprintf("Task %s not found", taskID))
	return nil, fmt.Errorf("task not found")
}
