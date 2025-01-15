package result

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// LoadResultsFromFile 从本地文件异步加载结果数据
func (c *Collector) LoadResultsFromFile() ([]ResultData, error) {
	// 打开结果文件
	fmt.Println("Loading results from file:", c.jtlFilePath)
	file, err := os.Open(c.jtlFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open result file: %v", err)
	}
	defer file.Close()

	// 读取 CSV 文件
	reader := csv.NewReader(file)
	// 跳过文件的标题行
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("failed to read header: %v", err)
	}

	// 读取所有行
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV records: %v", err)
	}

	// 使用 channel 与协程并发分析数据
	dataChannel := make(chan ResultData, len(records)) // 使用缓冲channel来减少阻塞

	// 启动协程进行数据处理
	go func() {
		defer close(dataChannel) // 结束后关闭 channel
		for i, record := range records {
			// 确保记录有足够的字段
			if len(record) < 17 {
				fmt.Printf("Skipping incomplete record at line %d: %+v\n", i+1, record)
				continue // 忽略不完整的记录
			}

			// 解析每个字段
			id := record[0]
			var resultType ResultType
			if record[7] == "true" {
				resultType = Success
			} else if record[7] == "false" {
				resultType = Failure
			}

			// 响应时间
			responseTime, err := time.ParseDuration(record[1] + "ms") // 假设是毫秒单位
			if err != nil {
				fmt.Printf("failed to parse response time at line %d: %v\n", i+1, err)
				continue
			}

			// 状态码
			statusCode, err := strconv.Atoi(record[3])
			if err != nil {
				fmt.Printf("failed to parse status code at line %d: %v\n", i+1, err)
				continue
			}

			// 时间戳转换为开始时间
			timeStamp, err := strconv.ParseInt(record[0], 10, 64)
			if err != nil {
				fmt.Printf("failed to parse timestamp at line %d: %v\n", i+1, err)
				continue
			}
			startTime := time.Unix(0, timeStamp*int64(time.Millisecond))

			// 线程ID
			threadID, err := strconv.Atoi(record[9])
			if err != nil {
				fmt.Printf("failed to parse thread ID at line %d: %v\n", i+1, err)
				continue
			}

			// URL
			url := record[13]

			// 请求方法
			method := record[2] // 假设是 GET/POST 等方法

			// 发送和接收的数据大小
			dataSent, err := strconv.ParseInt(record[10], 10, 64)
			if err != nil {
				fmt.Printf("failed to parse data sent at line %d: %v\n", i+1, err)
				continue
			}

			dataReceived, err := strconv.ParseInt(record[11], 10, 64)
			if err != nil {
				fmt.Printf("failed to parse data received at line %d: %v\n", i+1, err)
				continue
			}

			// 数据类型
			dataType := record[6]

			// 响应信息
			responseMsg := record[5]

			// 线程组中的线程数
			grpThreads, err := strconv.Atoi(record[12])
			if err != nil {
				fmt.Printf("failed to parse group threads at line %d: %v\n", i+1, err)
				continue
			}

			// 所有线程数
			allThreads, err := strconv.Atoi(record[14])
			if err != nil {
				fmt.Printf("failed to parse all threads at line %d: %v\n", i+1, err)
				continue
			}

			// 连接花费时间
			connect, err := strconv.ParseInt(record[15], 10, 64)
			if err != nil {
				fmt.Printf("failed to parse connect time at line %d: %v\n", i+1, err)
				continue
			}

			// 生成 ResultData
			result := ResultData{
				ID:           id,
				Type:         resultType,
				ResponseTime: responseTime,
				StartTime:    startTime,
				EndTime:      startTime.Add(responseTime), // 假设结束时间等于开始时间加上响应时间
				StatusCode:   statusCode,
				ThreadID:     threadID,
				URL:          url,
				Method:       method,
				DataSent:     dataSent,
				DataReceived: dataReceived,
				DataType:     dataType,
				ResponseMsg:  responseMsg,
				GrpThreads:   grpThreads,
				AllThreads:   allThreads,
				Connect:      connect,
			}

			// 将解析的结果传递给主协程进行处理
			dataChannel <- result
		}
	}()

	var results []ResultData
	// 等待所有数据解析完成
	for result := range dataChannel {
		results = append(results, result)
	}

	return results, nil
}

// formatBytes 将字节数转换为适当的单位
func formatBytes(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(bytes)/1024)
	} else if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(bytes)/1024/1024)
	} else if bytes < 1024*1024*1024*1024 {
		return fmt.Sprintf("%.2f GB", float64(bytes)/1024/1024/1024)
	} else {
		return fmt.Sprintf("%.2f TB", float64(bytes)/1024/1024/1024/1024)
	}
}

func (c *Collector) GeneratePerformanceStats(results []ResultData) (map[string]interface{}, error) {
	var totalRequests, successCount, failureCount int
	var totalResponseTime time.Duration
	var maxResponseTime, minResponseTime time.Duration = 0, time.Hour * 24 * 365 // 初始为很大值
	var totalSentData, totalReceivedData int64

	var firstTimestamp int64 = results[0].StartTime.UnixMilli() // 第一条记录的时间戳
	var lastTimestamp int64                                     // 最后一条记录的时间戳

	// 统计各项数据
	for _, result := range results {
		totalRequests++
		if result.Type == Success {
			successCount++
		} else {
			failureCount++
		}

		// 累加响应时间
		totalResponseTime += result.ResponseTime

		// 最大响应时间
		if result.ResponseTime > maxResponseTime {
			maxResponseTime = result.ResponseTime
		}

		// 最小响应时间
		if result.ResponseTime < minResponseTime {
			minResponseTime = result.ResponseTime
		}

		// 累加发送和接收的数据
		totalSentData += result.DataSent
		totalReceivedData += result.DataReceived

		// 更新最后一个时间戳
		lastTimestamp = result.EndTime.UnixMilli()
	}

	// 计算成功率，保留三位小数
	successRate := (float64(successCount) / float64(totalRequests)) * 100
	successRate = math.Round(successRate*1000) / 1000 // 四舍五入到小数点后三位

	// 计算平均响应时间
	avgResponseTime := totalResponseTime / time.Duration(totalRequests)

	// 使用 CalculateTPS 计算每秒的 TPS 数据
	tpsValues, successValues, failureValues, tpsStartTime, tpsEndTime := c.CalculateTPS(results)

	// 计算每秒事务数（TPS），基于 CalculateTPS 的结果
	var tps float64
	totalRunTime := time.Duration(lastTimestamp-firstTimestamp) * time.Millisecond

	if totalRunTime.Seconds() > 0 {
		tps = float64(totalRequests) / totalRunTime.Seconds()
	}
	tps = math.Round(tps*100) / 100 // 四舍五入到小数点后二位

	// 计算每秒发送和接收的数据流量 (单位为字节)
	var sentDataPerSec, receivedDataPerSec float64
	if totalRunTime.Seconds() > 0 {
		sentDataPerSec = float64(totalSentData) / totalRunTime.Seconds()
		receivedDataPerSec = float64(totalReceivedData) / totalRunTime.Seconds()
	}

	// 将每秒发送和接收的字节数转换为适当的单位
	sentDataPerSecStr := formatBytes(int64(sentDataPerSec))
	receivedDataPerSecStr := formatBytes(int64(receivedDataPerSec))
	totalSentDataStr := formatBytes(totalSentData)
	totalReceivedDataStr := formatBytes(totalReceivedData)

	// 计算平均响应时间（每秒）
	avgResponseTimeValues, avgSuccessResponseTimeValues, avgFailureResponseTimeValues, avgResponseStartTime, avgResponseEndTime := c.CalculateAvgResponseTime(results)

	// 将响应时间数组转换为整数数组
	avgResponseTimeValuesInt := convertToIntArray(avgResponseTimeValues)
	avgSuccessResponseTimeValuesInt := convertToIntArray(avgSuccessResponseTimeValues)
	avgFailureResponseTimeValuesInt := convertToIntArray(avgFailureResponseTimeValues)

	// 计算平均流量（每秒）
	avgSentTrafficValues, avgReceivedTrafficValues, avgSuccessSentTrafficValues, avgTrafficStartTime, avgTrafficEndTime := c.CalculateAvgTraffic(results)

	// 返回所有统计数据
	stats := map[string]interface{}{
		"TotalRequests":      totalRequests,
		"SuccessCount":       successCount,
		"FailureCount":       failureCount,
		"SuccessRate":        successRate, // 保留三位小数的 float64
		"AvgResponseTime":    avgResponseTime,
		"MaxResponseTime":    maxResponseTime,
		"MinResponseTime":    minResponseTime,
		"TotalRunTime":       totalRunTime,
		"TPS":                tps, // 保留两位小数的 float64
		"SentDataPerSec":     sentDataPerSecStr,
		"ReceivedDataPerSec": receivedDataPerSecStr,
		"TotalSentData":      totalSentDataStr,
		"TotalReceivedData":  totalReceivedDataStr,
		"AvgTpsStartTime":    tpsStartTime,
		"AvgTpsEndTime":      tpsEndTime,
		"TPSValues":          tpsValues,
		"SuccessValues":      successValues,
		"FailureValues":      failureValues,
		// 包含每秒的平均响应时间相关数据
		"AvgResponseTimeValues":        avgResponseTimeValuesInt,
		"AvgSuccessResponseTimeValues": avgSuccessResponseTimeValuesInt,
		"AvgFailureResponseTimeValues": avgFailureResponseTimeValuesInt,
		"AvgResponseStartTime":         avgResponseStartTime,
		"AvgResponseEndTime":           avgResponseEndTime,
		// 包含每秒的平均流量相关数据
		"AvgSentTrafficValues":        avgSentTrafficValues,
		"AvgReceivedTrafficValues":    avgReceivedTrafficValues,
		"AvgSuccessSentTrafficValues": avgSuccessSentTrafficValues,
		"AvgTrafficStartTime":         avgTrafficStartTime,
		"AvgTrafficEndTime":           avgTrafficEndTime,
	}

	return stats, nil
}

// convertToIntArray 将浮动的时间值数组转换为整数数组
func convertToIntArray(floatArray []float64) []int {
	var intArray []int
	for _, value := range floatArray {
		intArray = append(intArray, int(value/1000))
	}
	// fmt.Println(intArray)
	return intArray
}

func (c *Collector) CalculateTPS(results []ResultData) ([]int, []int, []int, int64, int64) {
	// 按秒聚合数据
	tpsData := make(map[int64]int)     // 每秒的请求总数
	successData := make(map[int64]int) // 每秒的成功请求数
	failureData := make(map[int64]int) // 每秒的失败请求数

	var startTime, endTime int64

	for _, result := range results {
		// 计算时间戳（按秒计算）
		sec := result.StartTime.Unix()

		if startTime == 0 || sec < startTime {
			startTime = sec
		}
		if sec > endTime {
			endTime = sec
		}

		// 聚合 TPS
		tpsData[sec]++
		if result.Type == Success {
			successData[sec]++
		} else if result.Type == Failure {
			failureData[sec]++
		}
	}

	// 生成横坐标数据（秒）
	var xAxis []int64
	for sec := startTime; sec <= endTime; sec++ {
		xAxis = append(xAxis, sec)
	}

	// 汇总每秒的请求数量、成功请求数量、失败请求数量
	var tpsValues []int
	var successValues []int
	var failureValues []int

	for _, sec := range xAxis {
		tpsValues = append(tpsValues, tpsData[sec])
		successValues = append(successValues, successData[sec])
		failureValues = append(failureValues, failureData[sec])
	}

	return tpsValues, successValues, failureValues, startTime, endTime
}

func (c *Collector) CalculateAvgResponseTime(results []ResultData) ([]float64, []float64, []float64, int64, int64) {
	// 按秒聚合数据
	totalResponseTime := make(map[int64]int64)   // 每秒的总响应时间（单位：毫秒）
	successResponseTime := make(map[int64]int64) // 每秒的成功请求的响应时间（单位：毫秒）
	failureResponseTime := make(map[int64]int64) // 每秒的失败请求的响应时间（单位：毫秒）
	successCount := make(map[int64]int)          // 每秒成功请求的数量
	failureCount := make(map[int64]int)          // 每秒失败请求的数量

	var startTime, endTime int64

	for _, result := range results {
		// 计算时间戳（按秒计算）
		sec := result.StartTime.Unix()

		if startTime == 0 || sec < startTime {
			startTime = sec
		}
		if sec > endTime {
			endTime = sec
		}

		// 聚合响应时间（直接以毫秒为单位进行加总）
		totalResponseTime[sec] += int64(result.ResponseTime)

		if result.Type == Success {
			successResponseTime[sec] += int64(result.ResponseTime)
			successCount[sec]++
		} else if result.Type == Failure {
			failureResponseTime[sec] += int64(result.ResponseTime)
			failureCount[sec]++
		}

	}

	// 生成横坐标数据（秒）
	var xAxis []int64
	for sec := startTime; sec <= endTime; sec++ {
		xAxis = append(xAxis, sec)
	}

	// 汇总每秒的平均响应时间、成功请求的平均响应时间、失败请求的平均响应时间
	var avgResponseTime []float64
	var avgSuccessResponseTime []float64
	var avgFailureResponseTime []float64

	for _, sec := range xAxis {
		// 计算每秒的平均响应时间
		if successCount[sec]+failureCount[sec] > 0 {
			avgResponseTime = append(avgResponseTime, float64(totalResponseTime[sec])/float64(successCount[sec]+failureCount[sec])/1000)
		} else {
			avgResponseTime = append(avgResponseTime, 0)
		}

		// 计算每秒的成功请求的平均响应时间
		if successCount[sec] > 0 {
			avgSuccessResponseTime = append(avgSuccessResponseTime, float64(successResponseTime[sec])/float64(successCount[sec])/1000)
		} else {
			avgSuccessResponseTime = append(avgSuccessResponseTime, 0)
		}

		// 计算每秒的失败请求的平均响应时间
		if failureCount[sec] > 0 {
			avgFailureResponseTime = append(avgFailureResponseTime, float64(failureResponseTime[sec])/float64(failureCount[sec])/1000)
		} else {
			avgFailureResponseTime = append(avgFailureResponseTime, 0)
		}
	}

	return avgResponseTime, avgSuccessResponseTime, avgFailureResponseTime, startTime, endTime
}

func (c *Collector) CalculateAvgTraffic(results []ResultData) ([]int, []int, []int, int64, int64) {
	// 按秒聚合数据
	totalSent := make(map[int64]int64)       // 每秒的发送数据总量
	totalReceived := make(map[int64]int64)   // 每秒的接收数据总量
	successSent := make(map[int64]int64)     // 每秒成功请求的发送数据总量
	successReceived := make(map[int64]int64) // 每秒成功请求的接收数据总量
	failureSent := make(map[int64]int64)     // 每秒失败请求的发送数据总量
	failureReceived := make(map[int64]int64) // 每秒失败请求的接收数据总量
	successCount := make(map[int64]int)      // 每秒成功请求的数量
	failureCount := make(map[int64]int)      // 每秒失败请求的数量

	var startTime, endTime int64

	// 遍历结果数据，计算每秒的聚合数据
	for _, result := range results {
		// 计算时间戳（按秒计算）
		sec := result.StartTime.Unix()

		// 更新最早和最晚的时间
		if startTime == 0 || sec < startTime {
			startTime = sec
		}
		if sec > endTime {
			endTime = sec
		}

		// 聚合流量数据
		totalSent[sec] += result.DataSent
		totalReceived[sec] += result.DataReceived

		if result.Type == Success {
			successSent[sec] += result.DataSent
			successReceived[sec] += result.DataReceived
			successCount[sec]++
		} else if result.Type == Failure {
			failureSent[sec] += result.DataSent
			failureReceived[sec] += result.DataReceived
			failureCount[sec]++
		}
	}

	// 生成横坐标数据（秒）
	var xAxis []int64
	for sec := startTime; sec <= endTime; sec++ {
		xAxis = append(xAxis, sec)
	}

	// 汇总每秒的平均流量、成功请求的平均发送流量、成功请求的平均接收流量
	var avgSentTraffic []int
	var avgReceivedTraffic []int
	var avgSuccessSentTraffic []int
	var avgSuccessReceivedTraffic []int

	for _, sec := range xAxis {
		// 计算每秒的平均发送流量
		if successCount[sec]+failureCount[sec] > 0 {
			avgSentTraffic = append(avgSentTraffic, int(totalSent[sec]/int64(successCount[sec]+failureCount[sec])))
			avgReceivedTraffic = append(avgReceivedTraffic, int(totalReceived[sec]/int64(successCount[sec]+failureCount[sec])))
		} else {
			avgSentTraffic = append(avgSentTraffic, 0)
			avgReceivedTraffic = append(avgReceivedTraffic, 0)
		}

		// 计算每秒成功请求的平均发送流量
		if successCount[sec] > 0 {
			avgSuccessSentTraffic = append(avgSuccessSentTraffic, int(successSent[sec]/int64(successCount[sec])))
			avgSuccessReceivedTraffic = append(avgSuccessReceivedTraffic, int(successReceived[sec]/int64(successCount[sec])))
		} else {
			avgSuccessSentTraffic = append(avgSuccessSentTraffic, 0)
			avgSuccessReceivedTraffic = append(avgSuccessReceivedTraffic, 0)
		}
	}

	// 返回修正后的值
	return avgSentTraffic, avgReceivedTraffic, avgSuccessSentTraffic, startTime, endTime
}

func (c *Collector) GenerateChart(tpsValues, successValues, failureValues []int, startTime, endTime int64) {
	// 创建折线图对象
	line := charts.NewLine()

	// 设置 X 轴数据（时间）
	xAxis := make([]string, len(tpsValues))
	for i := 0; i < len(tpsValues); i++ {
		xAxis[i] = time.Unix(startTime+int64(i), 0).Format("15:04:05")
	}

	// 添加数据系列
	line.AddSeries("Total TPS", generateLineData(tpsValues))
	line.AddSeries("Success TPS", generateLineData(successValues))
	line.AddSeries("Failure TPS", generateLineData(failureValues))

	// 设置 X 轴
	line.SetXAxis(xAxis)

	// 设置全局选项
	line.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Transactions Per Second",
		Subtitle: fmt.Sprintf("Test Duration: %s to %s", time.Unix(startTime, 0).Format("15:04:05"), time.Unix(endTime, 0).Format("15:04:05")),
	}))

	// 渲染图表并保存
	f, err := os.Create("tps_chart.html")
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	err = line.Render(f)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("TPS chart generated successfully!")
}

func generateLineData(values []int) []opts.LineData {
	var lineData []opts.LineData
	for _, v := range values {
		lineData = append(lineData, opts.LineData{Value: v})
	}
	return lineData
}

// 定义常量参数（参考标准列表）
const (
	MaxAvgResponseTime      = 2.5  // 普通接口最大平均响应时间
	MaxTPS                  = 2000 // TPS最小值
	MinSuccessRate          = 99.0 // 最低请求成功率
	MaxHighFreqResponseTime = 1.0  // 高频接口最大平均响应时间
)

// 参考标准结构体
type PerformanceStandard struct {
	Field   string
	Min     float64
	Max     float64
	Compare func(value interface{}) float64
}

// generateDefaultAnalysis 根据传入的测试数据生成默认的分析内容
// 通过测试数据来动态生成一段分析报告
func generateDefaultAnalysis(stats map[string]interface{}) string {
	// 获取测试数据
	successRate := stats["SuccessRate"].(float64)
	avgResponseTime := stats["AvgResponseTime"].(time.Duration) // 修改为 time.Duration 类型
	tps := stats["TPS"].(float64)
	sentDataPerSec := stats["SentDataPerSec"].(string)
	receivedDataPerSec := stats["ReceivedDataPerSec"].(string)

	// 将 time.Duration 转换为毫秒并格式化为两位小数
	avgResponseTimeMillis := float64(avgResponseTime) / float64(time.Millisecond)
	avgResponseTimeFormatted := fmt.Sprintf("%.2f", avgResponseTimeMillis) // 保留两位小数

	// 根据成功率生成分析内容，精确到小数点后三位
	var successAnalysis string
	successRateFormatted := fmt.Sprintf("%.3f", successRate) // 格式化成功率为小数点后三位
	if successRate >= 99 {
		successAnalysis = "本次测试的请求成功率非常高，达到了 " + successRateFormatted + "%，表明系统能够高效处理请求。"
	} else if successRate >= 90 {
		successAnalysis = "本次测试的请求成功率达到了 " + successRateFormatted + "%，系统表现良好，但仍有一定的优化空间。"
	} else {
		successAnalysis = "本次测试的请求成功率为 " + successRateFormatted + "%，说明系统可能存在一定的瓶颈或故障，需要进一步排查。"
	}

	// 根据平均响应时间生成分析内容
	var responseTimeAnalysis string
	if avgResponseTimeMillis <= 1000 {
		responseTimeAnalysis = "系统的平均响应时间非常低，达到了 " + avgResponseTimeFormatted + " 毫秒，符合高频接口的性能标准。"
	} else if avgResponseTimeMillis <= 2000 {
		responseTimeAnalysis = "系统的平均响应时间为 " + avgResponseTimeFormatted + " 毫秒，符合普通接口的性能标准。"
	} else {
		responseTimeAnalysis = "系统的平均响应时间为 " + avgResponseTimeFormatted + " 毫秒，可能会影响用户体验，需要进一步优化。"
	}

	// 根据TPS生成分析内容，精确到小数点后二位
	var tpsAnalysis string
	tpsFormatted := fmt.Sprintf("%.2f", tps) // 格式化TPS为小数点后二位
	if tps >= 5000 {
		tpsAnalysis = fmt.Sprintf("TPS（事务每秒）达到了 %s，说明系统能够承载较高的负载。", tpsFormatted)
	} else if tps >= 2000 {
		tpsAnalysis = fmt.Sprintf("TPS（事务每秒）为 %s，系统能够处理中等负载的请求。", tpsFormatted)
	} else {
		tpsAnalysis = fmt.Sprintf("TPS（事务每秒）为 %s，系统在高负载下的表现较为平缓，可能会有性能瓶颈。", tpsFormatted)
	}

	// 根据数据流量生成分析内容
	var dataFlowAnalysis string
	dataFlowAnalysis = "每秒发送的数据流量为 " + sentDataPerSec + "，每秒接收的数据流量为 " + receivedDataPerSec + "，系统的数据吞吐量良好。"

	// 组合分析内容
	analysis := successAnalysis + " " + responseTimeAnalysis + " " + tpsAnalysis + " " + dataFlowAnalysis
	return analysis
}
