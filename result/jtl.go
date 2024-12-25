// jtl.go
// JTL文件处理模块
// 本文件负责处理JTL格式文件的读写操作。

package result

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
)

// JTLRecord JTL记录结构
type JTLRecord struct {
	Timestamp     int64  // 时间戳
	Elapsed       int64  // 耗时（毫秒）
	Label         string // 标签
	ResponseCode  int    // 响应码
	ResponseMsg   string // 响应信息
	ThreadName    string // 线程名
	DataType      string // 数据类型
	Success       bool   // 是否成功
	FailureMsg    string // 失败信息
	Bytes         int64  // 字节数
	SentBytes     int64  // 发送字节数
	GrpThreads    int    // 线程组中的线程数
	AllThreads    int    // 所有线程数
	URL           string // URL
	Latency       int64  // 延迟
	IdleTime      int64  // 空闲时间
	Connect       int64  // 连接时间
}

// writeToJTL 将一批结果写入JTL文件
func (c *Collector) writeToJTL(batch []ResultData) error {
	file, err := os.OpenFile(c.jtlFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open JTL file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 如果文件是新创建的，写入表头
	if stat, _ := file.Stat(); stat.Size() == 0 {
		headers := []string{
			"timeStamp",
			"elapsed",
			"label",
			"responseCode",
			"responseMessage",
			"threadName",
			"dataType",
			"success",
			"failureMessage",
			"bytes",
			"sentBytes",
			"grpThreads",
			"allThreads",
			"URL",
			"Latency",
			"IdleTime",
			"Connect",
		}
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("failed to write headers: %v", err)
		}
	}

	// 写入数据
	for _, data := range batch {
		record := []string{
			strconv.FormatInt(data.StartTime.UnixNano()/1e6, 10),
			strconv.FormatInt(data.ResponseTime.Milliseconds(), 10),
			data.Method,
			strconv.Itoa(data.StatusCode),
			"",
			fmt.Sprintf("Thread-%d", data.ThreadID),
			"",
			strconv.FormatBool(data.Type == Success),
			data.ErrorMessage,
			strconv.FormatInt(data.DataReceived, 10),
			strconv.FormatInt(data.DataSent, 10),
			"1",
			"1",
			data.URL,
			"0",
			"0",
			"0",
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %v", err)
		}
	}

	return nil
}

// generateJTLFileName 生成JTL文件名
func generateJTLFileName() string {
	return fmt.Sprintf("test_result_%s.jtl", time.Now().Format("20060102150405"))
}
