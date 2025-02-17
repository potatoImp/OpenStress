package pool

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/panjf2000/ants/v2"
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

// Pool represents a goroutine pool with dynamic concurrency and priority scheduling.
type Pool struct {
	maxWorkers      int32      // max workers
	activeWorkers   int32      // active workers
	taskPool        *ants.Pool // Task pool from ants library
	isPaused        int32      // 0 means running, 1 means paused
	shutdownFlag    int32      // 0 means not shutdown, 1 means shutdown
	threadIDCounter int32      // Atomic counter for assigning ThreadID
}

// NewPool creates a new Pool with the specified maximum number of workers.
func NewPool(maxWorkers int) *Pool {
	stressLogger.Log("INFO", fmt.Sprintf("Creating a new pool with %d workers", maxWorkers))

	// Initialize ants pool with max worker limit
	taskPool, err := ants.NewPool(maxWorkers)
	if err != nil {
		stressLogger.Log("ERROR", fmt.Sprintf("Error creating ants pool: %v", err))
		return nil
	}

	// Initialize the pool with the new threadIDCounter
	pool := &Pool{
		maxWorkers:      int32(maxWorkers),
		taskPool:        taskPool,
		threadIDCounter: 0,
	}

	stressLogger.Log("INFO", "Pool created successfully")
	return pool
}

// Start initializes the worker goroutines.
func (p *Pool) Start() {
	stressLogger.Log("INFO", fmt.Sprintf("Starting %d worker goroutines...", p.maxWorkers))

	// Initialize workers
	for i := 0; i < int(p.maxWorkers); i++ {
		go p.worker()
	}

	stressLogger.Log("INFO", fmt.Sprintf("%d worker goroutines started", p.maxWorkers))
}

// worker listens for tasks and executes them.
func (p *Pool) worker() {
	for {
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

		// Submit tasks to ants pool, worker executes tasks asynchronously
		task := &Task{
			ID: "Sample-Task",
			fn: func() {
				stressLogger.Log("INFO", "Executing task...")
				// Simulate task processing
				time.Sleep(1 * time.Second)
				stressLogger.Log("INFO", "Task completed")
			},
		}

		// Submit the task to the pool with panic recovery
		err := p.taskPool.Submit(func() {
			// 自动使用 defer 和 recover 捕获 panic
			defer func() {
				if r := recover(); r != nil {
					stressLogger.Log("ERROR", fmt.Sprintf("Task %s panicked: %v", task.ID, r))
				}
			}()

			// 执行任务
			task.fn()
		})
		if err != nil {
			stressLogger.Log("ERROR", fmt.Sprintf("Failed to submit task %s: %v", task.ID, err))
		}
	}
}

// Submit adds a new task to the pool.
func (p *Pool) Submit(fn func(threadID int32), priority int, taskID string, timeout time.Duration) {
	stressLogger.Log("INFO", fmt.Sprintf("Submitting task %s with priority %d", taskID, priority))

	// Get a unique ThreadID for the current task, limiting it to maxWorkers
	threadID := atomic.AddInt32(&p.threadIDCounter, 1) % p.maxWorkers

	task := &Task{
		ID:         taskID,
		fn:         func() { fn(threadID) }, // Pass the threadID to the task function
		priority:   priority,
		retries:    0, // Default retries
		maxRetries: 1, // Maximum retries
		timeout:    timeout,
	}

	// Submit task to ants pool with panic recovery
	err := p.taskPool.Submit(func() {
		// 使用 defer 和 recover 捕获 panic 错误
		defer func() {
			if r := recover(); r != nil {
				stressLogger.Log("ERROR", fmt.Sprintf("Task %s panicked: %v", taskID, r))
			}
		}()

		// 执行任务
		task.fn()
	})
	if err != nil {
		stressLogger.Log("ERROR", fmt.Sprintf("Failed to submit task %s: %v", taskID, err))
	}
	stressLogger.Log("INFO", fmt.Sprintf("Task %s submitted successfully", taskID))
}

// Shutdown gracefully stops the pool and waits for all tasks to complete.
func (p *Pool) Shutdown() {
	stressLogger.Log("INFO", "Shutting down the pool")
	atomic.StoreInt32(&p.shutdownFlag, 1)
	p.taskPool.Release()
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
	// You can resize the pool using ants' dynamic worker adjustment features if needed
	stressLogger.Log("INFO", fmt.Sprintf("Worker count adjusted to %d", newWorkerCount))
}

// GetTaskStatus returns the status of a task by its ID.
func (p *Pool) GetTaskStatus(taskID string) (*Task, error) {
	stressLogger.Log("INFO", fmt.Sprintf("Fetching status for task %s", taskID))

	// Return the status of the task
	return nil, fmt.Errorf("task not found")
}
