// monitor.go
// 监控模块
// 本文件负责监控和管理系统的运行状态。
// 主要功能包括：
// - 监控任务执行状态
// - 收集系统性能指标
// - 生成监控报告
// - 提供监控数据接口
//
// 日志功能：
// 1. 任务状态变更日志
//    - 记录任务的状态转换（pending -> running -> completed/failed）
//    - 记录任务执行时间和结果
//    - 使用 INFO 级别记录正常状态变更
//    - 使用 WARNING 级别记录重试操作
//    - 使用 ERROR 级别记录失败状态
//
// 2. 性能指标日志
//    - 记录系统资源使用情况（CPU、内存、goroutine数量）
//    - 使用 INFO 级别定期记录性能数据
//    - 当资源使用超过阈值时使用 WARNING 级别记录警告
//
// 3. 监控报告日志
//    - 定期生成系统运行状态摘要
//    - 包含成功/失败任务统计
//    - 包含系统资源使用趋势
//    - 使用 INFO 级别记录定期报告
//
// 4. 异常监控日志
//    - 记录异常的系统行为
//    - 记录资源泄漏警告
//    - 记录系统瓶颈
//    - 使用 WARNING 或 ERROR 级别记录异常情况
//
// 5. 日志管理
//    - 使用 pool/log.go 中的 StressLogger 进行日志记录
//    - 异步日志写入，避免影响监控性能
//    - 日志自动切割和压缩
//    - 支持日志查询和分析
//
// 技术实现细节：
// 1. 使用独立的 goroutine 进行监控
// 2. 使用 channel 进行数据采集
// 3. 实现可配置的监控间隔
// 4. 支持监控指标的自定义阈值
// 5. 提供监控数据的查询接口

package pool

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ResourceThresholds 资源阈值配置
type ResourceThresholds struct {
	MaxCPUUsage    float64 // CPU 使用率阈值
	MaxMemoryUsage uint64  // 内存使用阈值（字节）
	MaxGoroutines  int     // goroutine 数量阈值
}

// TaskStats 任务统计信息
type TaskStats struct {
	TotalTasks     int64         // 总任务数
	CompletedTasks int64         // 完成的任务数
	FailedTasks    int64         // 失败的任务数
	AverageTime    time.Duration // 平均执行时间
}

// statsData 内部使用的带锁的任务统计信息
type statsData struct {
	stats TaskStats
	mu    sync.RWMutex // 保护并发访问
}

// SystemMetrics 系统指标
type SystemMetrics struct {
	CPUUsage    float64 // CPU 使用率
	MemoryUsage uint64  // 内存使用量
	Goroutines  int     // goroutine 数量
	Timestamp   time.Time
}

// TaskStatusUpdate 任务状态更新信息
type TaskStatusUpdate struct {
	TaskID string
	// OldStatus     TaskStatus
	// NewStatus     TaskStatus
	ExecutionTime time.Duration
}

// Monitor 监控器结构体
type Monitor struct {
	logger           *StressLogger
	taskStats        *statsData
	thresholds       ResourceThresholds
	metricsChan      chan SystemMetrics
	statusUpdateChan chan TaskStatusUpdate
	stopChan         chan struct{}
	interval         time.Duration
	wg               sync.WaitGroup
}

// NewMonitor 创建新的监控器实例
func NewMonitor(logger *StressLogger, interval time.Duration, thresholds ResourceThresholds) *Monitor {
	return &Monitor{
		logger: logger,
		taskStats: &statsData{
			stats: TaskStats{},
		},
		thresholds:       thresholds,
		metricsChan:      make(chan SystemMetrics, 100),
		statusUpdateChan: make(chan TaskStatusUpdate, 1000),
		stopChan:         make(chan struct{}),
		interval:         interval,
	}
}

// Start 启动监控
func (m *Monitor) Start() {
	m.wg.Add(3)
	// 启动系统指标收集
	go m.collectMetrics()
	// 启动监控报告生成
	go m.generateReports()
	m.logger.Log("INFO", "Monitor started")
}

// Stop 停止监控
func (m *Monitor) Stop() {
	close(m.stopChan)
	m.wg.Wait()
	m.logger.Log("INFO", "Monitor stopped")
}

// // RecordTaskStatus 记录任务状态变更（异步）
// func (m *Monitor) RecordTaskStatus(taskID string, oldStatus, newStatus TaskStatus, executionTime time.Duration) {
// 	// 创建状态更新对象
// 	update := TaskStatusUpdate{
// 		TaskID:        taskID,
// 		OldStatus:     oldStatus,
// 		NewStatus:     newStatus,
// 		ExecutionTime: executionTime,
// 	}

// 	// 异步发送状态更新
// 	select {
// 	case m.statusUpdateChan <- update:
// 		// 成功发送到通道
// 	default:
// 		// 通道已满，记录警告
// 		m.logger.Log("WARNING", fmt.Sprintf("Status update channel full, dropping update for task %s", taskID))
// 	}
// }

// collectMetrics 收集系统指标
func (m *Monitor) collectMetrics() {
	defer m.wg.Done()
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			metrics := m.getSystemMetrics()
			m.metricsChan <- metrics
			m.checkThresholds(metrics)
		}
	}
}

// getSystemMetrics 获取系统指标
func (m *Monitor) getSystemMetrics() SystemMetrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	metrics := SystemMetrics{
		MemoryUsage: memStats.Alloc,
		Goroutines:  runtime.NumGoroutine(),
		Timestamp:   time.Now(),
		// Note: CPU 使用率的计算需要更复杂的实现
		// 这里简化处理
		CPUUsage: 0.0,
	}

	return metrics
}

// checkThresholds 检查系统指标是否超过阈值
func (m *Monitor) checkThresholds(metrics SystemMetrics) {
	if metrics.CPUUsage > m.thresholds.MaxCPUUsage {
		m.logger.Log("WARNING", fmt.Sprintf("CPU usage (%.2f%%) exceeded threshold (%.2f%%)", metrics.CPUUsage, m.thresholds.MaxCPUUsage))
	}
	if metrics.MemoryUsage > m.thresholds.MaxMemoryUsage {
		m.logger.Log("WARNING", fmt.Sprintf("Memory usage (%d bytes) exceeded threshold (%d bytes)", metrics.MemoryUsage, m.thresholds.MaxMemoryUsage))
	}
	if metrics.Goroutines > m.thresholds.MaxGoroutines {
		m.logger.Log("WARNING", fmt.Sprintf("Number of goroutines (%d) exceeded threshold (%d)", metrics.Goroutines, m.thresholds.MaxGoroutines))
	}
}

// generateReports 生成监控报告
func (m *Monitor) generateReports() {
	defer m.wg.Done()
	ticker := time.NewTicker(time.Minute * 5) // 每5分钟生成一次报告
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.generateReport()
		}
	}
}

// generateReport 生成单次报告
func (m *Monitor) generateReport() {
	m.taskStats.mu.RLock()
	stats := m.taskStats.stats
	m.taskStats.mu.RUnlock()

	successRate := float64(0)
	if stats.TotalTasks > 0 {
		successRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
	}

	report := fmt.Sprintf(`System Status Report:
	Total Tasks: %d
	Completed Tasks: %d
	Failed Tasks: %d
	Success Rate: %.2f%%
	Average Execution Time: %v`,
		stats.TotalTasks,
		stats.CompletedTasks,
		stats.FailedTasks,
		successRate,
		stats.AverageTime)

	m.logger.Log("INFO", report)
}

// GetTaskStats 获取任务统计信息的副本
func (m *Monitor) GetTaskStats() TaskStats {
	m.taskStats.mu.RLock()
	defer m.taskStats.mu.RUnlock()
	// 返回统计信息的副本，而不是包含锁的结构体
	return m.taskStats.stats
}

// GetLatestMetrics 获取最新的系统指标
func (m *Monitor) GetLatestMetrics() (SystemMetrics, bool) {
	select {
	case metrics := <-m.metricsChan:
		return metrics, true
	default:
		return SystemMetrics{}, false
	}
}
