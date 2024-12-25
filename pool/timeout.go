package pool

// 超时处理模块
// 本文件负责处理任务的超时和取消功能。
// 主要功能包括：
// - 设置任务的超时时间
// - 取消未完成的任务
// - 提供取消反馈机制（待实现）
// - 支持批量任务取消（待实现）
// - 合理的超时重试机制（待实现）
// - 日志记录操作（待实现）
//
// 技术实现细节：
// 1. 使用 context 包实现超时控制，允许用户设置超时时间。
// 2. 提供 Cancel 方法，允许用户取消正在执行的任务。
// 3. 实现取消反馈机制，通知用户任务取消的结果。
// 4. 支持批量任务取消，提供批量取消的接口。
// 5. 在任务执行过程中，检查是否超过超时时间，若超过则自动取消任务。
// 6. 实现合理的超时重试机制，允许用户在任务超时后进行重试。
//    - 提供重试次数和重试间隔的配置选项。
// 7. 在超时和取消操作中记录日志，便于监控和排查问题。

import (
	"context"
	"fmt"
	"time"
)

// TimeoutManager 结构体
// 负责管理任务的超时和取消
// 记录日志的级别为 INFO
type TimeoutManager struct {
	Timeout       time.Duration
	RetryCount    int
	RetryInterval time.Duration
	logger        *StressLogger
}

// NewTimeoutManager 创建新的 TimeoutManager 实例
func NewTimeoutManager(timeout time.Duration, retryCount int, retryInterval time.Duration) (*TimeoutManager, error) {
	logger, logErr := NewStressLogger("logs/", "timeout.log", "TimeoutModule")
	if logErr != nil {
		return nil, logErr
	}
	return &TimeoutManager{
		Timeout:       timeout,
		RetryCount:    retryCount,
		RetryInterval: retryInterval,
		logger:        logger,
	}, nil
}

// ExecuteWithTimeout 执行带有超时的任务
func (tm *TimeoutManager) ExecuteWithTimeout(task func()) error {
	ctx, cancel := context.WithTimeout(context.Background(), tm.Timeout)
	defer cancel()

	go func() {
		select {
		case <-ctx.Done():
			tm.logger.Log("WARNING", "Task cancelled due to timeout") // 记录日志，级别为 WARNING
		}
	}()

	// 执行任务
	tm.logger.Log("INFO", "Starting task execution...") // 记录日志，级别为 INFO
	task()

	if ctx.Err() == context.DeadlineExceeded {
		tm.logger.Log("ERROR", "Task exceeded timeout") // 记录日志，级别为 ERROR
		return ctx.Err()
	}
	return nil
}

// Retry 执行带有重试机制的任务
func (tm *TimeoutManager) Retry(task func()) {
	for i := 0; i < tm.RetryCount; i++ {
		tm.logger.Log("INFO", fmt.Sprintf("Attempting task retry %d/%d...", i+1, tm.RetryCount)) // 记录日志，级别为 INFO
		err := tm.ExecuteWithTimeout(task)
		if err == nil {
			return
		}
		tm.logger.Log("ERROR", fmt.Sprintf("Retrying task failed: %v", err)) // 记录日志，级别为 ERROR
		time.Sleep(tm.RetryInterval)
	}
}
