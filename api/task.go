package api

// "encoding/json"
// "net/http"

// The Task interface defines the basic behavior of a task
type Task interface {
	Execute() error        // Method to execute the task.
	GetName() string       // Method to get the task name.
	GetExecutionTime() int // Get the task execution time.
	GetWaitTime() int      // Get the wait time after the task execution.
	GetTimeout() int       // Get the timeout wait time.
	GetInitTimeout() int   // Get the initialization timeout wait time.
}

// HttpClientTask structure for HTTP client tasks
type HttpClientTask struct {
	Name          string
	URL           string
	ExecutionTime int // Task execution time (milliseconds).
	WaitTime      int // Wait time after execution (milliseconds).
	Timeout       int // Timeout wait time (milliseconds).
	InitTimeout   int // Initialization timeout wait time (milliseconds).
}

// Execute Implement HTTP request logic
func (h *HttpClientTask) Execute() error {
	// Implement HTTP request logic
	return nil
}

// GetName Get the task name.
func (h *HttpClientTask) GetName() string {
	return h.Name
}

// GetExecutionTime Get the task execution time.
func (h *HttpClientTask) GetExecutionTime() int {
	return h.ExecutionTime
}

// GetWaitTime Get the wait time after the task execution.
func (h *HttpClientTask) GetWaitTime() int {
	return h.WaitTime
}

// GetTimeout Get the timeout wait time.
func (h *HttpClientTask) GetTimeout() int {
	return h.Timeout
}

// GetInitTimeout Get the initialization timeout wait time.
func (h *HttpClientTask) GetInitTimeout() int {
	return h.InitTimeout
}
