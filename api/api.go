// api.go
// API interface module
// This file is responsible for providing the API interface for users to submit tasks.
// Main functions include:
// - Submit tasks to the coroutine pool
// - Set maximum concurrency and rate limiting
// - Query task execution status
// - Query the list of executable tasks
// - Start, pause, and stop concurrent tasks
// - Query currently executing tasks
// - Authentication and authorization (to be implemented)
// - API documentation generation (to be implemented)
//
// Technical implementation details:
// 1. Use the net/http package to implement the HTTP API interface.
// 2. Provide the SubmitTask method to receive task parameters and submit tasks to the coroutine pool.
// 3. Provide SetMaxConcurrency and SetRateLimit methods to allow users to set parameters.
// 4. Provide the GetTaskStatus method to query the execution status of a specific task.
// 5. Provide the GetAvailableTasks method to return the current list of executable tasks.
// 6. Provide StartPool, PausePool, and StopPool methods to control the lifecycle of the coroutine pool.
// 7. Provide the GetRunningTasks method to return the list of currently executing tasks.
// 8. Implement authentication and authorization mechanisms to ensure API security.
// 9. Generate detailed API documentation with usage examples and interface descriptions.
//
// Features to be implemented:
// 1. Authentication and authorization
//    - Implement user authentication mechanisms to ensure that only authorized users can submit and manage tasks.
//    - Provide different levels of permission control to ensure API security.
// 2. API documentation generation
//    - Automatically generate detailed API documentation, including request and response examples for each interface.
//    - Provide usage examples to help users get started quickly.
// 3. Error handling mechanism
//    - Unified error handling mechanism to ensure all APIs return clear error messages.
//    - Provide error codes and descriptions to help users understand issues.
// 4. Task management feature extensions
//    - Provide task cancellation functionality to allow users to cancel tasks during execution.
//    - Provide task retry functionality to allow users to retry failed tasks.
// 5. Monitoring and statistics features
//    - Provide statistics on task execution, such as success rate, average execution time, etc.
//    - Provide real-time monitoring interfaces to allow users to query current system load and task status.
//
// Use the net/http package to implement the HTTP API interface.
// Provide clear API routing design for easy user invocation.
// Implement middleware mechanisms to support logging and request processing.

package api

// TODO: Implement task submission and management API

import (
	"encoding/json"
	"net/http"
	"sync"

	"OpenStress/pool" // Assume this is the coroutine pool implementation
	// "OpenStress/pool/error" // Import error handling module
)

var (
	maxConcurrency int
	rateLimit      int
	taskPool       *pool.Pool // Assume this is the coroutine pool instance
	mu             sync.Mutex
)

// TaskRequest represents the request structure for submitting tasks
type TaskRequest struct {
	TaskName string                 `json:"task_name"`
	Params   map[string]interface{} `json:"params"`
	TaskID   string                 `json:"task_id"` // New field
}

// SubmitTask submits tasks to the coroutine pool
func SubmitTask(w http.ResponseWriter, r *http.Request) {
	var req TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Define task priority and retry times
	priority := 1 // Example priority
	retries := 3  // Example retry times

	// Submit task to coroutine pool
	taskPool.Submit(func() {
		// Execute specific task logic here
	}, priority, retries, req.TaskID)

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "task submitted"})
}

// SetMaxConcurrency sets the maximum concurrency
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

// SetRateLimit sets the rate limiting
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

// GetTaskStatus queries the execution status of a task
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

// GetAvailableTasks queries the list of executable tasks
func GetAvailableTasks(w http.ResponseWriter, r *http.Request) {
	tasks := taskPool.GetAvailableTasks()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}

// StartPool starts the coroutine pool
func StartPool(w http.ResponseWriter, r *http.Request) {
	taskPool.Start()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "task pool started"})
}

// PausePool pauses the coroutine pool
func PausePool(w http.ResponseWriter, r *http.Request) {
	taskPool.Pause()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "task pool paused"})
}

// StopPool stops the coroutine pool
func StopPool(w http.ResponseWriter, r *http.Request) {
	taskPool.Stop()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "task pool stopped"})
}

// GetRunningTasks queries the list of currently executing tasks
func GetRunningTasks(w http.ResponseWriter, r *http.Request) {
	tasks := taskPool.GetRunningTasks()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}

// errorResponse unified error response format
func errorResponse(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
