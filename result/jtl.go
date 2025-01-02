// jtl.go
// JTL文件处理模块
// 本文件负责处理JTL格式文件的读写操作。

package result

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// JTLRecord JTL记录结构
type JTLRecord struct {
	Timestamp    int64  // 时间戳
	Elapsed      int64  // 耗时（毫秒）
	Label        string // 标签
	ResponseCode int    // 响应码
	ResponseMsg  string // 响应信息
	ThreadName   string // 线程名
	DataType     string // 数据类型
	Success      bool   // 是否成功
	FailureMsg   string // 失败信息
	Bytes        int64  // 字节数
	SentBytes    int64  // 发送字节数
	GrpThreads   int    // 线程组中的线程数
	AllThreads   int    // 所有线程数
	URL          string // URL
	Latency      int64  // 延迟
	IdleTime     int64  // 空闲时间
	Connect      int64  // 连接时间
}

// 替换掉数据中的逗号
func sanitizeField(field string) string {
	// 替换逗号和其他特殊字符
	return strings.ReplaceAll(field, ",", "_")
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
			sanitizeField(strconv.FormatInt(data.StartTime.UnixNano()/1e6, 10)),
			sanitizeField(strconv.FormatInt(data.ResponseTime.Milliseconds(), 10)),
			sanitizeField(data.Method),
			sanitizeField(strconv.Itoa(data.StatusCode)),
			"", // responseMessage 空
			sanitizeField(fmt.Sprintf("Thread-%d", data.ThreadID)),
			"", // dataType 空
			sanitizeField(strconv.FormatBool(data.Type == Success)),
			sanitizeField(data.ErrorMessage),
			sanitizeField(strconv.FormatInt(data.DataReceived, 10)),
			sanitizeField(strconv.FormatInt(data.DataSent, 10)),
			"1", // grpThreads 固定值
			"1", // allThreads 固定值
			sanitizeField(data.URL),
			"0", // Latency 固定值
			"0", // IdleTime 固定值
			"0", // Connect 固定值
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
