package api

// "encoding/json"
// "net/http"

// Task 接口定义了任务的基本行为
type Task interface {
	Execute() error        // 执行任务的方法
	GetName() string       // 获取任务名称的方法
	GetExecutionTime() int // 获取任务执行时间
	GetWaitTime() int      // 获取任务执行完后的等待时间
	GetTimeout() int       // 获取超时等待时间
	GetInitTimeout() int   // 获取初始化超时等待时间
}

// HttpClientTask HTTP 客户端任务结构
type HttpClientTask struct {
	Name          string
	URL           string
	ExecutionTime int // 任务执行时间（毫秒）
	WaitTime      int // 执行完后的等待时间（毫秒）
	Timeout       int // 超时等待时间（毫秒）
	InitTimeout   int // 初始化超时等待时间（毫秒）
}

// Execute 执行 HTTP 请求的逻辑
func (h *HttpClientTask) Execute() error {
	// 实现 HTTP 请求逻辑
	return nil
}

// GetName 获取任务名称
func (h *HttpClientTask) GetName() string {
	return h.Name
}

// GetExecutionTime 获取任务执行时间
func (h *HttpClientTask) GetExecutionTime() int {
	return h.ExecutionTime
}

// GetWaitTime 获取任务执行完后的等待时间
func (h *HttpClientTask) GetWaitTime() int {
	return h.WaitTime
}

// GetTimeout 获取超时等待时间
func (h *HttpClientTask) GetTimeout() int {
	return h.Timeout
}

// GetInitTimeout 获取初始化超时等待时间
func (h *HttpClientTask) GetInitTimeout() int {
	return h.InitTimeout
}
