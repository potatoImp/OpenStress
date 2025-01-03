package result

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type ResultType int

const (
	Success ResultType = iota
	Failure
)

type ResultData struct {
	ID           string        // 唯一标识符
	Type         ResultType    // 结果类型（成功/失败）
	ResponseTime time.Duration // 响应时间
	StartTime    time.Time     // 开始时间
	EndTime      time.Time     // 结束时间
	ErrorMessage string        // 错误信息（如果失败）
	StatusCode   int           // HTTP状态码
	ThreadID     int           // 线程ID
	URL          string        // 请求URL
	Method       string        // 请求方法
	DataSent     int64         // 发送的数据大小
	DataReceived int64         // 接收的数据大小
	DataType     string        // 数据类型
	ResponseMsg  string        // 响应信息
	GrpThreads   int           // 线程组中的线程数
	AllThreads   int           // 所有线程数
	Connect      int64         // 连接花费时间
}

type Collector struct {
	mu              sync.RWMutex
	results         []ResultData
	batchSize       int
	outputFormat    string
	jtlFilePath     string
	dataChan        chan ResultData
	done            chan struct{}
	logger          Logger
	numGoroutines   int // 并发 goroutine 数量
	collectInterval int
}

type CollectorConfig struct {
	BatchSize       int
	OutputFormat    string
	JTLFilePath     string
	Logger          Logger
	NumGoroutines   int
	CollectInterval int
	TaskID          string
}

func NewCollector(config CollectorConfig) (*Collector, error) {
	if config.BatchSize <= 0 {
		config.BatchSize = 100
	}

	dir := filepath.Dir(config.JTLFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory for JTL file: %v", err)
	}

	jtlFileName := fmt.Sprintf("test_result_%s_%s.jtl", config.TaskID, time.Now().Format("20060102150405"))
	config.JTLFilePath = filepath.Join(dir, jtlFileName)

	c := &Collector{
		results:         make([]ResultData, 0),
		batchSize:       config.BatchSize,
		outputFormat:    config.OutputFormat,
		jtlFilePath:     config.JTLFilePath,
		dataChan:        make(chan ResultData, 1000),
		done:            make(chan struct{}),
		logger:          config.Logger,
		numGoroutines:   config.NumGoroutines,
		collectInterval: config.CollectInterval,
	}

	go c.processData()

	if c.collectInterval > 0 {
		ticker := time.NewTicker(time.Duration(c.collectInterval) * time.Second)
		go func() {
			for range ticker.C {
				c.CollectData()
			}
		}()
	}

	return c, nil
}

func (c *Collector) InitializeCollector() {
	c.dataChan = make(chan ResultData, c.batchSize)
	c.done = make(chan struct{})

	go c.processData()

	c.logger.Log("INFO", "Collector initialized and ready to receive data.")
}

func (c *Collector) CollectResult(data ResultData) {
	select {
	case c.dataChan <- data:
	default:
		c.logger.Log("ERROR", "data channel is full, result dropped")
	}
}

func (c *Collector) CollectDataWithParams(id string, startTime time.Time, endTime time.Time, statusCode int, method string, url string, dataSent int64, dataReceived int64, threadID int, dataType string, responseMsg string, grpThreads int, allThreads int, connect int64) {
	responseTime := endTime.Sub(startTime)

	result := ResultData{
		ID:           id,
		Type:         Success,
		ResponseTime: responseTime,
		StartTime:    startTime,
		EndTime:      endTime,
		StatusCode:   statusCode,
		Method:       method,
		URL:          url,
		DataSent:     dataSent,
		DataReceived: dataReceived,
		ThreadID:     threadID,
		DataType:     dataType,
		ResponseMsg:  responseMsg,
		GrpThreads:   grpThreads,
		AllThreads:   allThreads,
		Connect:      connect,
	}

	select {
	case c.dataChan <- result:
	default:
		c.logger.Log("ERROR", "data channel is full, result dropped")
	}
}

func (c *Collector) CollectData() {
	// 假设的结束时间和数据
	id := "test-id"
	startTime := time.Now()
	endTime := startTime.Add(1 * time.Second)
	statusCode := 200
	method := "GET"
	url := "http://example.com"
	dataSent := int64(100)
	dataReceived := int64(200)
	threadID := 1 // 假设的线程ID
	dataType := "1"
	responseMsg := "OK"
	grpThreads := 1
	allThreads := 0
	connect := int64(10)

	c.logger.Log("INFO", fmt.Sprintf("Collecting data for thread ID: %d", threadID))
	c.CollectDataWithParams(id, startTime, endTime, statusCode, method, url, dataSent, dataReceived, threadID, dataType, responseMsg, grpThreads, allThreads, connect)
}

func (c *Collector) processData() {
	var wg sync.WaitGroup

	for i := 0; i < c.numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for data := range c.dataChan {
				if data.Type == Success {
					if err := c.SaveSuccessResult(data); err != nil {
						c.logger.Log("ERROR", fmt.Sprintf("failed to save success result: %v", err))
					}
				} else {
					if err := c.SaveFailureResult(data); err != nil {
						c.logger.Log("ERROR", fmt.Sprintf("failed to save failure result: %v", err))
					}
				}
			}
		}()
	}

	wg.Wait()
}

func (c *Collector) SaveSuccessResult(data ResultData) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.results = append(c.results, data)

	if c.jtlFilePath != "" {
		if err := c.writeToJTL([]ResultData{data}); err != nil {
			c.logger.Log("ERROR", fmt.Sprintf("failed to write success result to JTL file: %v", err))
			return err
		}
	}
	return nil
}

func (c *Collector) SaveFailureResult(data ResultData) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.results = append(c.results, data)

	if c.jtlFilePath != "" {
		if err := c.writeToJTL([]ResultData{data}); err != nil {
			c.logger.Log("ERROR", fmt.Sprintf("failed to write failure result to JTL file: %v", err))
			return err
		}
	}
	return nil
}

func (c *Collector) generateTextReport(results []ResultData) error {
	var report strings.Builder
	report.WriteString("简要结果报告:\n")

	for _, result := range results {
		report.WriteString(fmt.Sprintf("请求ID: %s, 响应时间: %v, 状态码: %d, 数据类型: %s, 响应信息: %s, 线程组数: %d, 所有线程数: %d, 连接花费时间: %d\n", result.ID, result.ResponseTime, result.StatusCode, result.DataType, result.ResponseMsg, result.GrpThreads, result.AllThreads, result.Connect))
	}

	fmt.Println(report.String())
	return nil
}

func (c *Collector) generateJTLReport(results []ResultData) error {
	var report strings.Builder
	report.WriteString("JTL Report:\n")
	for _, result := range results {
		report.WriteString(fmt.Sprintf("ID: %s, ResponseTime: %v, StatusCode: %d, DataType: %s, ResponseMsg: %s, GrpThreads: %d, AllThreads: %d, Connect: %d\n", result.ID, result.ResponseTime, result.StatusCode, result.DataType, result.ResponseMsg, result.GrpThreads, result.AllThreads, result.Connect))
	}
	fmt.Println(report.String())
	return nil
}

type Logger interface {
	Log(level string, message string)
}

func (c *Collector) Close() error {
	close(c.done)
	return nil
}

func (c *Collector) CloseCollector() error {
	close(c.dataChan)

	close(c.done)

	var wg sync.WaitGroup

	for i := 0; i < c.numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range c.dataChan {
			}
		}()
	}

	wg.Wait()

	c.logger.Log("INFO", "Collector has been closed and resources released.")
	return nil
}
