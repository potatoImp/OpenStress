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

func (c *Collector) LoadResultsFromFile() ([]ResultData, error) {
	fmt.Println("Loading results from file:", c.jtlFilePath)
	file, err := os.Open(c.jtlFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open result file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("failed to read header: %v", err)
	}

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV records: %v", err)
	}

	dataChannel := make(chan ResultData, len(records))

	go func() {
		defer close(dataChannel)
		for i, record := range records {
			if len(record) < 17 {
				fmt.Printf("Skipping incomplete record at line %d: %+v\n", i+1, record)
				continue
			}

			id := record[0]
			var resultType ResultType
			if record[7] == "true" {
				resultType = Success
			} else if record[7] == "false" {
				resultType = Failure
			}

			responseTime, err := time.ParseDuration(record[1] + "ms")
			if err != nil {
				fmt.Printf("failed to parse response time at line %d: %v\n", i+1, err)
				continue
			}

			statusCode, err := strconv.Atoi(record[3])
			if err != nil {
				fmt.Printf("failed to parse status code at line %d: %v\n", i+1, err)
				continue
			}

			timeStamp, err := strconv.ParseInt(record[0], 10, 64)
			if err != nil {
				fmt.Printf("failed to parse timestamp at line %d: %v\n", i+1, err)
				continue
			}
			startTime := time.Unix(0, timeStamp*int64(time.Millisecond))

			threadID, err := strconv.Atoi(record[9])
			if err != nil {
				fmt.Printf("failed to parse thread ID at line %d: %v\n", i+1, err)
				continue
			}

			url := record[13]

			method := record[2] // 假设是 GET/POST 等方法

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

			dataType := record[6]

			responseMsg := record[5]

			grpThreads, err := strconv.Atoi(record[12])
			if err != nil {
				fmt.Printf("failed to parse group threads at line %d: %v\n", i+1, err)
				continue
			}

			allThreads, err := strconv.Atoi(record[14])
			if err != nil {
				fmt.Printf("failed to parse all threads at line %d: %v\n", i+1, err)
				continue
			}

			connect, err := strconv.ParseInt(record[15], 10, 64)
			if err != nil {
				fmt.Printf("failed to parse connect time at line %d: %v\n", i+1, err)
				continue
			}

			result := ResultData{
				ID:           id,
				Type:         resultType,
				ResponseTime: responseTime,
				StartTime:    startTime,
				EndTime:      startTime.Add(responseTime),
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

			dataChannel <- result
		}
	}()

	var results []ResultData
	for result := range dataChannel {
		results = append(results, result)
	}

	return results, nil
}

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
	var maxResponseTime, minResponseTime time.Duration = 0, time.Hour * 24 * 365
	var totalSentData, totalReceivedData int64

	var firstTimestamp int64 = results[0].StartTime.UnixMilli()
	var lastTimestamp int64

	for _, result := range results {
		totalRequests++
		if result.Type == Success {
			successCount++
		} else {
			failureCount++
		}

		totalResponseTime += result.ResponseTime

		if result.ResponseTime > maxResponseTime {
			maxResponseTime = result.ResponseTime
		}

		if result.ResponseTime < minResponseTime {
			minResponseTime = result.ResponseTime
		}

		totalSentData += result.DataSent
		totalReceivedData += result.DataReceived

		lastTimestamp = result.EndTime.UnixMilli()
	}

	successRate := (float64(successCount) / float64(totalRequests)) * 100
	successRate = math.Round(successRate*1000) / 1000

	avgResponseTime := totalResponseTime / time.Duration(totalRequests)

	tpsValues, successValues, failureValues, tpsStartTime, tpsEndTime := c.CalculateTPS(results)

	var tps float64
	totalRunTime := time.Duration(lastTimestamp-firstTimestamp) * time.Millisecond
	if totalRunTime.Seconds() > 0 {
		tps = float64(totalRequests) / totalRunTime.Seconds()
	}
	tps = math.Round(tps*100) / 100

	var sentDataPerSec, receivedDataPerSec float64
	if totalRunTime.Seconds() > 0 {
		sentDataPerSec = float64(totalSentData) / totalRunTime.Seconds()
		receivedDataPerSec = float64(totalReceivedData) / totalRunTime.Seconds()
	}

	sentDataPerSecStr := formatBytes(int64(sentDataPerSec))
	receivedDataPerSecStr := formatBytes(int64(receivedDataPerSec))
	totalSentDataStr := formatBytes(totalSentData)
	totalReceivedDataStr := formatBytes(totalReceivedData)

	avgResponseTimeValues, avgSuccessResponseTimeValues, avgFailureResponseTimeValues, avgResponseStartTime, avgResponseEndTime := c.CalculateAvgResponseTime(results)

	avgResponseTimeValuesInt := convertToIntArray(avgResponseTimeValues)
	avgSuccessResponseTimeValuesInt := convertToIntArray(avgSuccessResponseTimeValues)
	avgFailureResponseTimeValuesInt := convertToIntArray(avgFailureResponseTimeValues)

	avgSentTrafficValues, avgReceivedTrafficValues, avgSuccessSentTrafficValues, avgTrafficStartTime, avgTrafficEndTime := c.CalculateAvgTraffic(results)

	stats := map[string]interface{}{
		"TotalRequests":                totalRequests,
		"SuccessCount":                 successCount,
		"FailureCount":                 failureCount,
		"SuccessRate":                  successRate,
		"AvgResponseTime":              avgResponseTime,
		"MaxResponseTime":              maxResponseTime,
		"MinResponseTime":              minResponseTime,
		"TotalRunTime":                 totalRunTime,
		"TPS":                          tps,
		"SentDataPerSec":               sentDataPerSecStr,
		"ReceivedDataPerSec":           receivedDataPerSecStr,
		"TotalSentData":                totalSentDataStr,
		"TotalReceivedData":            totalReceivedDataStr,
		"AvgTpsStartTime":              tpsStartTime,
		"AvgTpsEndTime":                tpsEndTime,
		"TPSValues":                    tpsValues,
		"SuccessValues":                successValues,
		"FailureValues":                failureValues,
		"AvgResponseTimeValues":        avgResponseTimeValuesInt,
		"AvgSuccessResponseTimeValues": avgSuccessResponseTimeValuesInt,
		"AvgFailureResponseTimeValues": avgFailureResponseTimeValuesInt,
		"AvgResponseStartTime":         avgResponseStartTime,
		"AvgResponseEndTime":           avgResponseEndTime,
		"AvgSentTrafficValues":         avgSentTrafficValues,
		"AvgReceivedTrafficValues":     avgReceivedTrafficValues,
		"AvgSuccessSentTrafficValues":  avgSuccessSentTrafficValues,
		"AvgTrafficStartTime":          avgTrafficStartTime,
		"AvgTrafficEndTime":            avgTrafficEndTime,
	}

	return stats, nil
}

func convertToIntArray(floatArray []float64) []int {
	var intArray []int
	for _, value := range floatArray {
		intArray = append(intArray, int(value/1000))
	}
	return intArray
}

func (c *Collector) CalculateTPS(results []ResultData) ([]int, []int, []int, int64, int64) {
	tpsData := make(map[int64]int)
	successData := make(map[int64]int)
	failureData := make(map[int64]int)

	var startTime, endTime int64

	for _, result := range results {
		sec := result.StartTime.Unix()

		if startTime == 0 || sec < startTime {
			startTime = sec
		}
		if sec > endTime {
			endTime = sec
		}

		tpsData[sec]++
		if result.Type == Success {
			successData[sec]++
		} else if result.Type == Failure {
			failureData[sec]++
		}
	}

	var xAxis []int64
	for sec := startTime; sec <= endTime; sec++ {
		xAxis = append(xAxis, sec)
	}

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
	totalResponseTime := make(map[int64]int64)
	successResponseTime := make(map[int64]int64)
	failureResponseTime := make(map[int64]int64)
	successCount := make(map[int64]int)
	failureCount := make(map[int64]int)

	var startTime, endTime int64

	for _, result := range results {
		sec := result.StartTime.Unix()

		if startTime == 0 || sec < startTime {
			startTime = sec
		}
		if sec > endTime {
			endTime = sec
		}

		totalResponseTime[sec] += int64(result.ResponseTime)
		if result.Type == Success {
			successResponseTime[sec] += int64(result.ResponseTime)
			successCount[sec]++
		} else if result.Type == Failure {
			failureResponseTime[sec] += int64(result.ResponseTime)
			failureCount[sec]++
		}
	}

	var xAxis []int64
	for sec := startTime; sec <= endTime; sec++ {
		xAxis = append(xAxis, sec)
	}

	var avgResponseTime []float64
	var avgSuccessResponseTime []float64
	var avgFailureResponseTime []float64

	for _, sec := range xAxis {
		if successCount[sec] > 0 {
			avgSuccessResponseTime = append(avgSuccessResponseTime, float64(successResponseTime[sec])/float64(successCount[sec]))
		} else {
			avgSuccessResponseTime = append(avgSuccessResponseTime, 0)
		}

		if failureCount[sec] > 0 {
			avgFailureResponseTime = append(avgFailureResponseTime, float64(failureResponseTime[sec])/float64(failureCount[sec]))
		} else {
			avgFailureResponseTime = append(avgFailureResponseTime, 0)
		}

		if successCount[sec]+failureCount[sec] > 0 {
			avgResponseTime = append(avgResponseTime, float64(totalResponseTime[sec])/float64(successCount[sec]+failureCount[sec]))
		} else {
			avgResponseTime = append(avgResponseTime, 0)
		}
	}

	return avgResponseTime, avgSuccessResponseTime, avgFailureResponseTime, startTime, endTime
}

func (c *Collector) CalculateAvgTraffic(results []ResultData) ([]int, []int, []int, int64, int64) {
	totalSent := make(map[int64]int64)
	totalReceived := make(map[int64]int64)
	successSent := make(map[int64]int64)
	successReceived := make(map[int64]int64)
	failureSent := make(map[int64]int64)
	failureReceived := make(map[int64]int64)
	successCount := make(map[int64]int)
	failureCount := make(map[int64]int)

	var startTime, endTime int64

	for _, result := range results {
		sec := result.StartTime.Unix()

		if startTime == 0 || sec < startTime {
			startTime = sec
		}
		if sec > endTime {
			endTime = sec
		}

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

	var xAxis []int64
	for sec := startTime; sec <= endTime; sec++ {
		xAxis = append(xAxis, sec)
	}

	var avgSentTraffic []int
	var avgReceivedTraffic []int
	var avgSuccessSentTraffic []int
	var avgSuccessReceivedTraffic []int

	for _, sec := range xAxis {
		if successCount[sec]+failureCount[sec] > 0 {
			avgSentTraffic = append(avgSentTraffic, int(totalSent[sec]/int64(successCount[sec]+failureCount[sec])))
			avgReceivedTraffic = append(avgReceivedTraffic, int(totalReceived[sec]/int64(successCount[sec]+failureCount[sec])))
		} else {
			avgSentTraffic = append(avgSentTraffic, 0)
			avgReceivedTraffic = append(avgReceivedTraffic, 0)
		}

		if successCount[sec] > 0 {
			avgSuccessSentTraffic = append(avgSuccessSentTraffic, int(successSent[sec]/int64(successCount[sec])))
			avgSuccessReceivedTraffic = append(avgSuccessReceivedTraffic, int(successReceived[sec]/int64(successCount[sec])))
		} else {
			avgSuccessSentTraffic = append(avgSuccessSentTraffic, 0)
			avgSuccessReceivedTraffic = append(avgSuccessReceivedTraffic, 0)
		}
	}

	return avgSentTraffic, avgReceivedTraffic, avgSuccessSentTraffic, startTime, endTime
}

func (c *Collector) GenerateChart(tpsValues, successValues, failureValues []int, startTime, endTime int64) {
	line := charts.NewLine()

	xAxis := make([]string, len(tpsValues))
	for i := 0; i < len(tpsValues); i++ {
		xAxis[i] = time.Unix(startTime+int64(i), 0).Format("15:04:05")
	}

	line.AddSeries("Total TPS", generateLineData(tpsValues))
	line.AddSeries("Success TPS", generateLineData(successValues))
	line.AddSeries("Failure TPS", generateLineData(failureValues))

	line.SetXAxis(xAxis)

	line.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Transactions Per Second",
		Subtitle: fmt.Sprintf("Test Duration: %s to %s", time.Unix(startTime, 0).Format("15:04:05"), time.Unix(endTime, 0).Format("15:04:05")),
	}))

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

const (
	MaxAvgResponseTime      = 2.5
	MaxTPS                  = 2000
	MinSuccessRate          = 99.0
	MaxHighFreqResponseTime = 1.0
)

type PerformanceStandard struct {
	Field   string
	Min     float64
	Max     float64
	Compare func(value interface{}) float64
}

func generateDefaultAnalysis(stats map[string]interface{}) string {
	successRate := stats["SuccessRate"].(float64)
	avgResponseTime := stats["AvgResponseTime"].(time.Duration)
	tps := stats["TPS"].(float64)
	sentDataPerSec := stats["SentDataPerSec"].(string)
	receivedDataPerSec := stats["ReceivedDataPerSec"].(string)

	avgResponseTimeMillis := float64(avgResponseTime) / float64(time.Millisecond)
	avgResponseTimeFormatted := fmt.Sprintf("%.2f", avgResponseTimeMillis)

	var successAnalysis string
	successRateFormatted := fmt.Sprintf("%.3f", successRate)
	if successRate >= 99 {
		successAnalysis = "本次测试的请求成功率非常高，达到了 " + successRateFormatted + "%，表明系统能够高效处理请求。"
	} else if successRate >= 90 {
		successAnalysis = "本次测试的请求成功率达到了 " + successRateFormatted + "%，系统表现良好，但仍有一定的优化空间。"
	} else {
		successAnalysis = "本次测试的请求成功率为 " + successRateFormatted + "%，说明系统可能存在一定的瓶颈或故障，需要进一步排查。"
	}

	var responseTimeAnalysis string
	if avgResponseTimeMillis <= 1000 {
		responseTimeAnalysis = "系统的平均响应时间非常低，达到了 " + avgResponseTimeFormatted + " 毫秒，符合高频接口的性能标准。"
	} else if avgResponseTimeMillis <= 2000 {
		responseTimeAnalysis = "系统的平均响应时间为 " + avgResponseTimeFormatted + " 毫秒，符合普通接口的性能标准。"
	} else {
		responseTimeAnalysis = "系统的平均响应时间为 " + avgResponseTimeFormatted + " 毫秒，可能会影响用户体验，需要进一步优化。"
	}

	var tpsAnalysis string
	tpsFormatted := fmt.Sprintf("%.2f", tps)
	if tps >= 5000 {
		tpsAnalysis = fmt.Sprintf("TPS（事务每秒）达到了 %s，说明系统能够承载较高的负载。", tpsFormatted)
	} else if tps >= 2000 {
		tpsAnalysis = fmt.Sprintf("TPS（事务每秒）为 %s，系统能够处理中等负载的请求。", tpsFormatted)
	} else {
		tpsAnalysis = fmt.Sprintf("TPS（事务每秒）为 %s，系统在高负载下的表现较为平缓，可能会有性能瓶颈。", tpsFormatted)
	}

	var dataFlowAnalysis string
	dataFlowAnalysis = "每秒发送的数据流量为 " + sentDataPerSec + "，每秒接收的数据流量为 " + receivedDataPerSec + "，系统的数据吞吐量良好。"

	analysis := successAnalysis + " " + responseTimeAnalysis + " " + tpsAnalysis + " " + dataFlowAnalysis
	return analysis
}
