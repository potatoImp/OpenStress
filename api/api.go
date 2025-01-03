// api.go
// API 接口模块
// 本文件负责提供用户提交任务的 API 接口。
// 主要功能包括：
// - 提交任务到协程池
// - 设置最大并发数和限流控制
// - 查询任务执行状态
// - 查询可执行任务列表
// - 启动、暂停与停止并发任务
// - 查询正在执行的任务
// - 身份验证与授权（待实现）
// - API 文档生成（待实现）
//
// 技术实现细节：
// 1. 使用 net/http 包实现 HTTP API 接口。
// 2. 提供 SubmitTask 方法，接收任务参数并将任务提交到协程池。
// 3. 提供 SetMaxConcurrency 和 SetRateLimit 方法，允许用户设置参数。
// 4. 提供 GetTaskStatus 方法，查询特定任务的执行状态。
// 5. 提供 GetAvailableTasks 方法，返回当前可执行的任务列表。
// 6. 提供 StartPool、PausePool 和 StopPool 方法，控制协程池的生命周期。
// 7. 提供 GetRunningTasks 方法，返回当前正在执行的任务列表。
// 8. 实现身份验证和授权机制，确保 API 的安全性。
// 9. 生成详细的 API 文档，提供使用示例和接口说明。
//
// 待实现功能：
// 1. 身份验证与授权
//    - 实现用户身份验证机制，确保只有授权用户可以提交和管理任务。
//    - 提供不同级别的权限控制，确保 API 的安全性。
// 2. API 文档生成
//    - 自动生成详细的 API 文档，包含每个接口的请求和响应示例。
//    - 提供使用示例，帮助用户快速上手。
// 3. 错误处理机制
//    - 统一的错误处理机制，确保所有 API 返回清晰的错误信息。
//    - 提供错误码和错误描述，帮助用户理解问题。
// 4. 任务管理功能扩展
//    - 提供任务取消功能，允许用户在任务执行过程中取消任务。
//    - 提供任务重试功能，允许用户对失败的任务进行重试。
// 5. 监控与统计功能
//    - 提供任务执行的统计信息，例如成功率、平均执行时间等。
//    - 提供实时监控接口，允许用户查询当前系统负载和任务状态。
//
// 技术实现细节：
// - 使用 net/http 包实现 HTTP API 接口。
// - 提供清晰的 API 路由设计，方便用户调用。
// - 实现中间件机制，支持日志记录和请求处理。

package api

// TODO: 实现任务提交和管理的 API

import (
	"encoding/json"
	"net/http"
	"sync"

	"OpenStress/pool" // 假设这是协程池的实现
	// "OpenStress/pool/error" // 引入错误处理模块
)

var (
	maxConcurrency int
	rateLimit      int
	taskPool       *pool.Pool // 假设这是协程池的实例
	mu             sync.Mutex
)

// TaskRequest 表示提交任务的请求结构
type TaskRequest struct {
	TaskName string                 `json:"task_name"`
	Params   map[string]interface{} `json:"params"`
	TaskID   string                 `json:"task_id"` // 新增字段
}

// SubmitTask 提交任务到协程池
func SubmitTask(w http.ResponseWriter, r *http.Request) {
	var req TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// 定义任务的优先级和重试次数
	priority := 1 // 示例优先级
	retries := 3  // 示例重试次数

	// 提交任务到协程池
	taskPool.Submit(func() {
		// 这里可以执行具体的任务逻辑
	}, priority, retries, req.TaskID)

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "task submitted"})
}

// SetMaxConcurrency 设置最大并发数
func SetMaxConcurrency(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MaxConcurrency int `json:"max_concurrency"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	mu.Lock()
	maxConcurrency = req.MaxConcurrency
	mu.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "max concurrency set"})
}

// SetRateLimit 设置限流控制
func SetRateLimit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RateLimit int `json:"rate_limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	mu.Lock()
	rateLimit = req.RateLimit
	mu.Unlock()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "rate limit set"})
}

// GetTaskStatus 查询任务执行状态
func GetTaskStatus(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")
	status, err := taskPool.GetTaskStatus(taskID)
	if err != nil {
		errorResponse(w, http.StatusNotFound, "Task not found")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

// GetAvailableTasks 查询可执行任务列表
func GetAvailableTasks(w http.ResponseWriter, r *http.Request) {
	tasks := taskPool.GetAvailableTasks()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}

// StartPool 启动协程池
func StartPool(w http.ResponseWriter, r *http.Request) {
	taskPool.Start()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "task pool started"})
}

// PausePool 暂停协程池
func PausePool(w http.ResponseWriter, r *http.Request) {
	taskPool.Pause()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "task pool paused"})
}

// StopPool 停止协程池
func StopPool(w http.ResponseWriter, r *http.Request) {
	taskPool.Stop()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "task pool stopped"})
}

// GetRunningTasks 查询正在执行的任务
func GetRunningTasks(w http.ResponseWriter, r *http.Request) {
	tasks := taskPool.GetRunningTasks()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}

// errorResponse 统一的错误响应格式
func errorResponse(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
