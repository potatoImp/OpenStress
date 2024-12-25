package pool

// error.go
// 错误处理模块
// 本文件负责处理协程池中的错误，确保任务执行的稳定性。
// 主要功能包括：
// - 定义错误类型
// - 捕获和记录错误信息
// - 提供统一的错误处理接口
// - 支持自定义错误类型（待实现）
//
// 技术实现细节：
// 1. 定义自定义错误类型，用于描述不同的错误情况。
// 2. 提供统一的错误处理接口，方便错误的捕获和处理。
// 3. 支持自定义错误类型，允许用户扩展错误处理能力。
// 4. 提供错误日志记录功能，将错误信息记录到日志中。
// 5. 实现错误分类和统计功能，方便错误分析和处理。
//
// 常见的接口错误返回内容：
// - 400 Bad Request: 请求参数错误
// - 401 Unauthorized: 用户未授权
// - 403 Forbidden: 访问被禁止
// - 404 Not Found: 请求的资源未找到
// - 500 Internal Server Error: 服务器内部错误
// - 503 Service Unavailable: 服务不可用
//
// 其他常见接口返回错误：
// - 408 Request Timeout: 请求超时
// - 429 Too Many Requests: 请求过多
// - 501 Not Implemented: 服务器不支持的功能
// - 502 Bad Gateway: 网关错误
// - 504 Gateway Timeout: 网关超时

import (
	"fmt"
	"log"
)

// CustomError 自定义错误类型
// 用于描述不同的错误情况
type CustomError struct {
	Message string
	Code    int
}

// Error 实现 error 接口
// 返回错误信息
func (e *CustomError) Error() string {
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

// HandleError 处理错误并记录日志
// 记录日志级别为 ERROR
func HandleError(err error) {
	if err != nil {
		logger, logErr := NewStressLogger("logs/", "error.log", "ErrorModule")
		if logErr != nil {
			log.Println("Failed to create logger:", logErr)
			return
		}
		logger.Log("ERROR", "Error occurred: " + err.Error()) // 记录错误信息
	}
}

// ExampleFunction 示例函数
// 演示如何使用自定义错误类型
func ExampleFunction() {
	// 模拟一个错误
	err := &CustomError{Message: "Something went wrong", Code: 500}
	HandleError(err)
}
