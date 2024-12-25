// collector.go
// 结果收集器模块
// 本文件负责管理测试结果的收集和处理。

package result

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ResultType 定义结果类型
type ResultType int

const (
	// Success 成功结果
	Success ResultType = iota
	// Failure 失败结果
	Failure
)

// ResultData 测试结果数据结构
type ResultData struct {
	ID            string        // 唯一标识符
	Type          ResultType    // 结果类型（成功/失败）
	ResponseTime  time.Duration // 响应时间
	StartTime     time.Time     // 开始时间
	EndTime       time.Time     // 结束时间
	ErrorMessage  string        // 错误信息（如果失败）
	StatusCode    int          // HTTP状态码
	ThreadID      int          // 线程ID
	URL           string       // 请求URL
	Method        string       // 请求方法
	DataSent     int64        // 发送的数据大小
	DataReceived int64        // 接收的数据大小
}

// Collector 结果收集器结构体
type Collector struct {
	mu            sync.RWMutex
	results       []ResultData
	batchSize     int
	outputFormat  string
	jtlFilePath   string
	dataChan      chan ResultData
	done          chan struct{}
	logger        Logger
	numGoroutines int // 并发 goroutine 数量
	// 新增配置项：数据收集间隔（秒）
	collectInterval int
}

// CollectorConfig 收集器配置
// 本结构体用于管理数据收集的配置参数。
type CollectorConfig struct {
	BatchSize    int    // 每次批量写入的记录数
	OutputFormat string // 报告输出格式
	JTLFilePath  string // JTL文件的保存路径
	Logger       Logger // 日志记录接口
	NumGoroutines int    // 并发 goroutine 数量
	CollectInterval int  // 数据收集间隔（秒）
	TaskID      string    // 任务ID，用于生成唯一的文件名
}

// NewCollector 创建新的结果收集器
func NewCollector(config CollectorConfig) (*Collector, error) {
	if config.BatchSize <= 0 {
		config.BatchSize = 100 // 默认批量大小
	}

	// 确保JTL文件目录存在
	dir := filepath.Dir(config.JTLFilePath) // 定义并赋值 dir
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory for JTL file: %v", err)
	}

	// 使用 TaskID 生成唯一的 JTL 文件名
	jtlFileName := fmt.Sprintf("test_result_%s_%s.jtl", config.TaskID, time.Now().Format("20060102150405"))
	config.JTLFilePath = filepath.Join(dir, jtlFileName)

	c := &Collector{
		results:           make([]ResultData, 0),
		batchSize:         config.BatchSize,
		outputFormat:      config.OutputFormat,
		jtlFilePath:       config.JTLFilePath,
		dataChan:          make(chan ResultData, 1000),
		done:              make(chan struct{}),
		logger:            config.Logger,
		numGoroutines:     config.NumGoroutines,
		collectInterval:   config.CollectInterval,
	}

	// 启动异步处理goroutine
	go c.processData()

	// 启动定时任务，根据配置的数据收集间隔定期收集数据
	if c.collectInterval > 0 {
		ticker := time.NewTicker(time.Duration(c.collectInterval) * time.Second)
		go func() {
			for range ticker.C {
				// 定期收集数据
				c.CollectData()
			}
		}()
	}

	return c, nil
}

// InitializeCollector 初始化结果收集器，准备接收数据。
func (c *Collector) InitializeCollector() {
	// 初始化数据通道
	c.dataChan = make(chan ResultData, c.batchSize)
	c.done = make(chan struct{})

	// 启动异步处理数据的 goroutines
	go c.processData()

	// 日志记录收集器初始化成功
	c.logger.Log("INFO", "Collector initialized and ready to receive data.")
}

// CollectResult 收集测试结果
func (c *Collector) CollectResult(data ResultData) {
	select {
	case c.dataChan <- data:
		// 数据已发送到通道
	default:
		// 通道已满，记录错误
		c.logger.Log("ERROR", "data channel is full, result dropped")
	}
}

// CollectDataWithParams 定期收集数据
// 此函数应根据您的应用逻辑填充适当的参数。
func (c *Collector) CollectDataWithParams(id string, startTime time.Time, endTime time.Time, statusCode int, method string, url string, dataSent int64, dataReceived int64, threadID int) {
	// 计算响应时间
	responseTime := endTime.Sub(startTime)

	// 创建 ResultData 对象
	result := ResultData{
		ID:            id,
		Type:          Success,
		ResponseTime:  responseTime,
		StartTime:     startTime,
		EndTime:       endTime,
		StatusCode:    statusCode,
		Method:        method,
		URL:           url,
		DataSent:      dataSent,
		DataReceived:  dataReceived,
		ThreadID:      threadID,
	}

	// 将结果发送到数据通道
	select {
	case c.dataChan <- result:
		// 数据已发送到通道
	default:
		// 通道已满，记录错误
		c.logger.Log("ERROR", "data channel is full, result dropped")
	}
}

// CollectData 定期收集数据
func (c *Collector) CollectData() {
	// 这里可以根据需要生成测试数据
	id := "test-id"
	startTime := time.Now()
	endTime := startTime.Add(1 * time.Second) // 假设的结束时间
	statusCode := 200
	method := "GET"
	url := "http://example.com"
	dataSent := int64(100)
	dataReceived := int64(200)

	// 生成线程ID并记录
	threadID := 1 // 或者根据实际情况生成线程ID
	c.logger.Log("INFO", fmt.Sprintf("Collecting data for thread ID: %d", threadID))

	// 调用实际的 CollectData 函数
	c.CollectDataWithParams(id, startTime, endTime, statusCode, method, url, dataSent, dataReceived, threadID)
}

// processData 负责异步处理收集到的测试结果数据。
func (c *Collector) processData() {
	// 使用 WaitGroup 来等待所有 goroutine 完成
	var wg sync.WaitGroup

	// 启动多个 goroutine 来处理数据
	for i := 0; i < c.numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for data := range c.dataChan {
				// 处理每个数据项
				if data.Type == Success {
					// 保存成功结果
					if err := c.SaveSuccessResult(data); err != nil {
						c.logger.Log("ERROR", fmt.Sprintf("failed to save success result: %v", err))
					}
				} else {
					// 保存失败结果
					if err := c.SaveFailureResult(data); err != nil {
						c.logger.Log("ERROR", fmt.Sprintf("failed to save failure result: %v", err))
					}
				}
			}
		}()
	}

	// 等待所有 goroutine 完成
	wg.Wait()
}

// saveBatch 保存一批数据
func (c *Collector) saveBatch(batch []ResultData) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 将数据添加到结果集
	c.results = append(c.results, batch...)

	// 如果配置了JTL文件，则写入文件
	if c.jtlFilePath != "" {
		if err := c.writeToJTL(batch); err != nil {
			return fmt.Errorf("failed to write to JTL file: %v", err)
		}
	}
	return nil
}

// Close 关闭收集器
func (c *Collector) Close() error {
	close(c.done)
	return nil
}

// CloseCollector 关闭结果收集器，释放相关资源。
func (c *Collector) CloseCollector() error {
	// 关闭数据通道
	close(c.dataChan)

	// 关闭完成信号通道
	close(c.done)

	// 等待所有 goroutine 完成
	// 使用 WaitGroup 来等待所有 goroutine 完成
	var wg sync.WaitGroup

	// 启动多个 goroutine 来处理数据
	for i := 0; i < c.numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range c.dataChan {
			}
		}()
	}

	// 等待所有 goroutine 完成
	wg.Wait()

	// 日志记录收集器关闭信息
	c.logger.Log("INFO", "Collector has been closed and resources released.")
	return nil
}

// GetResults 获取所有结果
func (c *Collector) GetResults() []ResultData {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	results := make([]ResultData, len(c.results))
	copy(results, c.results)
	return results
}

// SaveSuccessResult 保存成功结果到结果集中，并写入JTL文件（如果配置了路径）。
func (c *Collector) SaveSuccessResult(data ResultData) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// 将成功结果添加到结果集中
	c.results = append(c.results, data)

	// 如果配置了JTL文件，则写入文件
	if c.jtlFilePath != "" {
		if err := c.writeToJTL([]ResultData{data}); err != nil {
			c.logger.Log("ERROR", fmt.Sprintf("failed to write success result to JTL file: %v", err))
			return err
		}
	}
	return nil
}

// SaveFailureResult 保存失败结果到结果集中，并写入JTL文件（如果配置了路径）。
func (c *Collector) SaveFailureResult(data ResultData) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// 将失败结果添加到结果集中
	c.results = append(c.results, data)

	// 如果配置了JTL文件，则写入文件
	if c.jtlFilePath != "" {
		if err := c.writeToJTL([]ResultData{data}); err != nil {
			c.logger.Log("ERROR", fmt.Sprintf("failed to write failure result to JTL file: %v", err))
			return err
		}
	}
	return nil
}

// GenerateReport 生成简要结果报告，支持不同格式的输出。
func (c *Collector) GenerateReport(format string) error {
	// 获取所有结果
	results := c.GetResults()

	// 根据格式生成报告
	switch format {
	case "text":
		return c.generateTextReport(results)
	case "jtl":
		return c.generateJTLReport(results)
	default:
		return fmt.Errorf("unsupported report format: %s", format)
	}
}

// generateTextReport 生成文本格式的报告。
func (c *Collector) generateTextReport(results []ResultData) error {
	// 创建报告内容
	var report strings.Builder
	report.WriteString("简要结果报告:\n")

	for _, result := range results {
		report.WriteString(fmt.Sprintf("请求ID: %s, 响应时间: %v, 状态码: %d\n", result.ID, result.ResponseTime, result.StatusCode))
	}

	// 输出报告到控制台
	fmt.Println(report.String())
	return nil
}

// generateJTLReport 生成JTL格式的报告。
func (c *Collector) generateJTLReport(results []ResultData) error {
	// TODO: 实现 JTL 格式报告生成逻辑
	var report strings.Builder
	report.WriteString("JTL Report:\n")
	for _, result := range results {
		report.WriteString(fmt.Sprintf("ID: %s, ResponseTime: %v, StatusCode: %d\n", result.ID, result.ResponseTime, result.StatusCode))
	}
	fmt.Println(report.String())
	return nil
}

// LoadConfig 从文件中加载配置
func LoadConfig(filePath string) (*CollectorConfig, error) {
	// 打开配置文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	// 解析 JSON 配置
	var config CollectorConfig
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}

// Logger 日志接口
type Logger interface {
	Log(level string, message string)
}
