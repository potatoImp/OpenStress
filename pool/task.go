// task.go
// 任务管理模块
// 本文件负责定义任务结构和状态，以及任务的生命周期管理。
//
// 主要功能包括：
// - 定义任务结构体和接口
// - 管理任务状态和生命周期
// - 任务执行日志记录
// - 任务重试机制
// - 任务依赖管理
// - 任务优先级管理
// - 任务超时控制
//
// 技术实现细节：
// 1. 定义任务结构体，包括：
//   - 任务ID：唯一标识任务
//   - 任务状态：pending、running、completed、failed
//   - 任务参数：支持泛型参数传递
//   - 执行函数：实际的任务执行逻辑
//   - 重试配置：最大重试次数和重试间隔
//   - 超时设置：任务执行超时时间
//   - 优先级：任务执行优先级
//   - 依赖关系：前置和后置任务
//
// 2. 任务状态管理：
//
//   - Pending: 任务已创建，等待执行
//
//   - Running: 任务正在执行中
//
//   - Completed: 任务已成功完成
//
//   - Failed: 任务执行失败
//
//   - Cancelled: 任务被取消
//
//   - Timeout: 任务执行超时
//
//     3. 日志记录实现：
//     使用 StressLogger 进行日志管理，主要特点：
//     a. 日志初始化：
//
//   - 在任务创建时通过 NewStressLogger 初始化日志记录器
//
//   - 自动创建日志目录和文件
//
//   - 支持日志文件自动切割和压缩
//
//   - 支持配置最大文件大小和保留时间
//
//     b. 日志记录格式：
//
//   - 时间戳：精确到毫秒级
//
//   - 日志级别：[INFO]/[WARNING]/[ERROR]
//
//   - 模块标识：[TaskModule]
//
//   - 详细信息：具体的日志消息
//
//     c. 日志记录时机：
//
//   - 任务创建：记录任务ID和初始状态
//
//   - 任务启动：记录开始执行时间
//
//   - 任务完成：记录执行结果和耗时
//
//   - 任务失败：记录错误信息
//
//   - 状态变更：记录状态转换
//
//   - 重试操作：记录重试次数和原因
//
//     d. 异步日志处理：
//
//   - 使用通道进行异步日志写入
//
//   - 缓冲区大小为100条日志
//
//   - 支持优雅关闭和等待日志写入完成
//
//     e. 日志文件管理：
//
//   - 最大文件大小：10MB
//
//   - 最大备份数量：3个
//
//   - 最大保留天数：28天
//
//   - 自动压缩备份文件
//
// 4. 错误处理：
//   - 使用 StressLogger 记录错误信息
//   - 支持任务重试机制
//   - 记录详细的错误上下文
//
// 5. 性能优化：
//   - 利用 StressLogger 的异步日志机制
//   - 使用缓冲通道避免日志阻塞
//   - 支持日志文件自动切割，避免单文件过大
package pool

import (
	"OpenStress/tasks"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"
)

// TaskStatus 定义任务状态
type TaskStatus int

const (
	TaskPending TaskStatus = iota
	TaskRunning
	TaskCompleted
	TaskFailed
	TaskCancelled
	TaskTimeout
)

// String 返回TaskStatus的字符串表示
func (s TaskStatus) String() string {
	switch s {
	case TaskPending:
		return "PENDING"
	case TaskRunning:
		return "RUNNING"
	case TaskCompleted:
		return "COMPLETED"
	case TaskFailed:
		return "FAILED"
	case TaskCancelled:
		return "CANCELLED"
	case TaskTimeout:
		return "TIMEOUT"
	default:
		return "UNKNOWN"
	}
}

// TaskDetail 任务结构体
// 用于管理和执行单个任务的所有相关信息和状态
type TaskDetail struct {
	// ID 任务的唯一标识符
	// 用于在日志和错误信息中标识特定任务
	ID string

	// Status 任务当前状态
	// 可能的值：Pending（待执行）、Running（执行中）、
	// Completed（已完成）、Failed（失败）、
	// Cancelled（已取消）、Timeout（超时）
	Status TaskStatus

	// Execute 任务的执行函数
	// 包含实际需要执行的业务逻辑
	// 返回 error 表示执行是否成功
	Execute func() error

	// RetryCount 当前重试次数
	// 记录任务已经重试了多少次
	RetryCount int

	// MaxRetries 最大重试次数
	// 任务失败后最多允许重试的次数
	MaxRetries int

	// RetryDelay 重试间隔时间
	// 两次重试之间的等待时间
	RetryDelay time.Duration

	// Timeout 任务超时时间
	// 如果任务执行时间超过此值，将被标记为超时
	Timeout time.Duration

	// Priority 任务优先级
	// 数值越大优先级越高，用于任务调度
	Priority int

	// Dependencies 任务依赖列表
	// 存储当前任务依赖的其他任务
	// 只有依赖的任务全部完成，当前任务才能执行
	Dependencies []*TaskDetail

	// StartTime 任务开始执行的时间
	// 用于计算任务执行时长和超时判断
	StartTime time.Time

	// EndTime 任务结束时间
	// 包括正常完成、失败、取消等所有结束状态
	EndTime time.Time

	// Error 任务执行过程中的错误信息
	// 如果任务执行失败，这里存储具体的错误原因
	Error error

	// mu 互斥锁
	// 用于保护任务状态的并发访问
	// 确保任务状态的修改是线程安全的
	mu sync.Mutex
}

var logger *StressLogger

// InitLogger 初始化全局日志记录器
func InitLogger(logDir, logFile string) error {
	var err error
	logger, err = NewStressLogger(logDir, logFile, "TaskModule")
	return err
}

// NewTaskDetail 创建新任务
func NewTaskDetail(id string, execute func() error) (*TaskDetail, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger not initialized")
	}

	task := &TaskDetail{
		ID:         id,
		Status:     TaskPending,
		Execute:    execute,
		MaxRetries: 3,
		RetryDelay: time.Second,
		Priority:   0,
	}

	logger.Log("INFO", fmt.Sprintf("Task created with ID: %s", id))
	return task, nil
}

// Start 开始执行任务
func (t *TaskDetail) Start() error {
	t.mu.Lock()
	if t.Status != TaskPending {
		t.mu.Unlock()
		return fmt.Errorf("task %s is not in pending status", t.ID)
	}
	t.Status = TaskRunning
	t.StartTime = time.Now()
	t.mu.Unlock()

	logger.Log("INFO", fmt.Sprintf("Task %s started at %v", t.ID, t.StartTime))

	// 检查依赖任务
	if err := t.checkDependencies(); err != nil {
		t.setStatus(TaskFailed)
		return err
	}

	// 执行任务
	if t.Timeout > 0 {
		return t.executeWithTimeout()
	}
	return t.executeTask()
}

// executeTask 执行任务的核心逻辑
func (t *TaskDetail) executeTask() error {
	err := t.Execute()
	t.mu.Lock()
	t.EndTime = time.Now()
	if err != nil {
		t.Error = err
		t.Status = TaskFailed
		logger.Log("ERROR", fmt.Sprintf("Task %s failed: %v", t.ID, err))
		t.mu.Unlock()
		return t.Retry()
	}
	t.Status = TaskCompleted
	t.mu.Unlock()

	duration := t.EndTime.Sub(t.StartTime)
	logger.Log("INFO", fmt.Sprintf("Task %s completed successfully in %v", t.ID, duration))
	return nil
}

// executeWithTimeout 带超时的任务执行
func (t *TaskDetail) executeWithTimeout() error {
	done := make(chan error)
	go func() {
		done <- t.executeTask()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(t.Timeout):
		t.setStatus(TaskTimeout)
		logger.Log("ERROR", fmt.Sprintf("Task %s timed out after %v", t.ID, t.Timeout))
		return fmt.Errorf("task %s timed out", t.ID)
	}
}

// Retry 重试任务
func (t *TaskDetail) Retry() error {
	t.mu.Lock()
	if t.RetryCount >= t.MaxRetries {
		t.mu.Unlock()
		logger.Log("ERROR", fmt.Sprintf("Task %s exceeded maximum retry attempts (%d)", t.ID, t.MaxRetries))
		return fmt.Errorf("exceeded maximum retry attempts")
	}
	t.RetryCount++
	currentRetry := t.RetryCount
	t.mu.Unlock()

	logger.Log("WARNING", fmt.Sprintf("Retrying task %s (attempt %d/%d)", t.ID, currentRetry, t.MaxRetries))
	time.Sleep(t.RetryDelay)

	return t.executeTask()
}

// Cancel 取消任务
func (t *TaskDetail) Cancel() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Status == TaskRunning || t.Status == TaskPending {
		t.Status = TaskCancelled
		logger.Log("WARNING", fmt.Sprintf("Task %s cancelled", t.ID))
		return nil
	}
	return fmt.Errorf("task %s cannot be cancelled in status %v", t.ID, t.Status.String())
}

// AddDependency 添加任务依赖
func (t *TaskDetail) AddDependency(dep *TaskDetail) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Dependencies = append(t.Dependencies, dep)
	logger.Log("INFO", fmt.Sprintf("Added dependency: Task %s now depends on Task %s", t.ID, dep.ID))
}

// SetPriority 设置任务优先级
func (t *TaskDetail) SetPriority(priority int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Priority = priority
	logger.Log("INFO", fmt.Sprintf("Task %s priority set to %d", t.ID, priority))
}

// SetTimeout 设置任务超时时间
func (t *TaskDetail) SetTimeout(timeout time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Timeout = timeout
	logger.Log("INFO", fmt.Sprintf("Task %s timeout set to %v", t.ID, timeout))
}

// checkDependencies 检查依赖任务是否完成
func (t *TaskDetail) checkDependencies() error {
	for _, dep := range t.Dependencies {
		if dep.Status != TaskCompleted {
			logger.Log("ERROR", fmt.Sprintf("Dependency task %s is not completed (status: %v)", dep.ID, dep.Status))
			return fmt.Errorf("dependency task %s is not completed", dep.ID)
		}
	}
	return nil
}

// setStatus 设置任务状态
func (t *TaskDetail) setStatus(status TaskStatus) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Status = status
	logger.Log("INFO", fmt.Sprintf("Task %s status changed to %v", t.ID, status))
}

// LoadTasks 自动加载任务到任务池
func LoadTasks(pool *Pool) {
	// 使用反射获取 tasks 包中的类型
	fmt.Println("Loading tasks...")
	taskType := reflect.TypeOf(tasks.Task{}) // 获取 Task 结构体的类型
	// fmt.Println(TaskDetail{})
	fmt.Println(taskType)
	// fmt.Println(taskType.NumMethod())
	for i := 0; i < taskType.NumMethod(); i++ {
		method := taskType.Method(i)
		if method.Type.NumIn() == 0 && strings.HasPrefix(method.Name, "Task_") {
			// 生成任务数据结构
			taskID := method.Name
			fn := method.Func.Interface().(func())
			priority := 1   // 可以根据需要设置优先级
			maxRetries := 3 // 可以根据需要设置重试次数
			// 提交任务到任务池
			pool.Submit(fn, priority, maxRetries, taskID)

			// 打印加载的任务信息
			fmt.Printf("Loaded task: %s\n", taskID)
		}
	}
}

// LoadTasks 自动加载任务到任务池
func LoadTasks2(pool *Pool) {
	fmt.Println("Loading tasks...11111111111111")

	wd, pwdErr := os.Getwd()
	if pwdErr != nil {
		fmt.Printf("Error getting current directory: %v\n", pwdErr)
		return
	}

	asksDir := filepath.Join(wd, "tasks") // 构建绝对路径

	// 扫描 tasks 目录下的所有 .go 文件
	err := filepath.Walk(asksDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), ".go") {
			// 解析 Go 源代码
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
			if err != nil {
				return err
			}

			// 遍历文件中的所有声明
			for _, decl := range node.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok {
					if strings.HasPrefix(fn.Name.Name, "Task_") {
						fmt.Printf("Found function: %s\n", fn.Name.Name) // 打印函数名称

						// 生成任务数据结构
						taskID := fn.Name.Name
						fnValue := reflect.ValueOf(tasks.Task{}).MethodByName(taskID)
						if fnValue.IsValid() && fnValue.Type().NumIn() == 0 {
							// 提交任务到任务池
							pool.Submit(fnValue.Interface().(func()), 1, 3, taskID)
							fmt.Printf("Loaded task: %s\n", taskID)
						}
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error loading tasks: %v\n", err)
	}
}
