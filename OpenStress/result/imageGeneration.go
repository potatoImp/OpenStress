package result

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func adjustXAxisPoints(startTime, endTime time.Time, values []int) ([]string, []int) {
	if len(values) == 0 {
		fmt.Println("Error: values array is empty")
		return nil, nil
	}

	numSegments := 10
	timeInterval := endTime.Sub(startTime).Seconds()
	segmentTime := timeInterval / float64(numSegments)
	xAxis := make([]string, numSegments)
	yAxis := make([]int, numSegments+1)

	for i := 0; i < numSegments; i++ {
		startSegment := startTime.Add(time.Second * time.Duration(float64(i)*segmentTime))
		endSegment := startTime.Add(time.Second * time.Duration(float64(i+1)*segmentTime))

		startSegmentIndex := int(startSegment.Sub(startTime).Seconds()) // 计算边界点在 values 数组中的索引
		endSegmentIndex := int(endSegment.Sub(startTime).Seconds())     // 同上

		yAxis[i] = values[startSegmentIndex]
		if i == numSegments-1 {
			yAxis[i+1] = values[endSegmentIndex] // 最后一个边界点
		}

		middleTime := startSegment.Add(time.Second * time.Duration(segmentTime/2))
		xAxis[i] = middleTime.Format("15:04:05") // 转换为 "HH:MM:SS" 格式
	}

	yAxis[numSegments] = values[int(endTime.Sub(startTime).Seconds())]

	return xAxis, yAxis
}

func GenerateTpsChartAsync(tpsValues []int, successValues []int, failureValues []int, startTime int64, endTime int64, dir string) (string, error) {
	startTimeTime := time.Unix(startTime, 0)
	endTimeTime := time.Unix(endTime, 0)

	xAxis, tpsValuesAdjusted := adjustXAxisPoints(startTimeTime, endTimeTime, tpsValues)
	if xAxis == nil || len(tpsValuesAdjusted) == 0 {
		fmt.Println("Error: Failed to adjust xAxis or tpsValues")
		return "", fmt.Errorf("failed to adjust xAxis or tpsValues")
	}

	_, successValuesAdjusted := adjustXAxisPoints(startTimeTime, endTimeTime, successValues)
	if len(successValuesAdjusted) == 0 {
		fmt.Println("Error: Failed to adjust successValues")
		return "", fmt.Errorf("failed to adjust successValues")
	}

	_, failureValuesAdjusted := adjustXAxisPoints(startTimeTime, endTimeTime, failureValues)
	if len(failureValuesAdjusted) == 0 {
		fmt.Println("Error: Failed to adjust failureValues")
		return "", fmt.Errorf("failed to adjust failureValues")
	}

	line := charts.NewLine()
	if line == nil {
		fmt.Println("Error: Failed to create line chart object")
		return "", fmt.Errorf("failed to create line chart object")
	}

	line.SetXAxis(xAxis)
	line.AddSeries("Total TPS", generateLineData(tpsValuesAdjusted))
	if err := checkError("Failed to add Total TPS series"); err != nil {
		return "", err
	}

	line.AddSeries("Success TPS", generateLineData(successValuesAdjusted))
	line.AddSeries("Failure TPS", generateLineData(failureValuesAdjusted))

	line.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Transactions Per Second",
		Subtitle: fmt.Sprintf("Test Duration: %s to %s", startTimeTime.Format("15:04:05"), endTimeTime.Format("15:04:05")),
	}), charts.WithLegendOpts(opts.Legend{
		Bottom: "bottom",
	}))

	htmlContent := line.RenderContent()
	if htmlContent == nil {
		fmt.Println("Error: Failed to render chart content")
		return "", fmt.Errorf("failed to render chart content")
	}

	htmlFilePath := filepath.Join(dir, "/tps_chart.html")

	htmlFile, err := os.Create(htmlFilePath)
	if err != nil {
		fmt.Printf("Error creating HTML file: %v\n", err)
		return "", fmt.Errorf("failed to create HTML file: %v", err)
	}
	defer func() {
		if cerr := htmlFile.Close(); cerr != nil {
			fmt.Printf("Error closing HTML file: %v\n", cerr)
		}
	}()

	_, err = htmlFile.Write(htmlContent)
	if err != nil {
		fmt.Printf("Error writing HTML content to file: %v\n", err)
		return "", fmt.Errorf("failed to write HTML content to file: %v", err)
	}

	return htmlFilePath, nil
}

func checkError(msg string) error {
	if r := recover(); r != nil {
		fmt.Printf("Error: %s: %v\n", msg, r)
		return fmt.Errorf("%s: %v", msg, r)
	}
	return nil
}

func GenerateResponseTimeChartAsync(avgResponseTimeValues []int, avgSuccessResponseTimeValues []int, avgFailureResponseTimeValues []int, avgResponseStartTime int64, avgResponseEndTime int64, dir string) (string, error) {
	startTimeTime := time.Unix(avgResponseStartTime, 0)
	endTimeTime := time.Unix(avgResponseEndTime, 0)

	xAxis, avgResponseTimeValuesAdjusted := adjustXAxisPoints(startTimeTime, endTimeTime, avgResponseTimeValues)
	if len(avgResponseTimeValuesAdjusted) == 0 {
		return "", fmt.Errorf("failed to adjust avgResponseTimeValues")
	}

	_, avgSuccessResponseTimeValuesAdjusted := adjustXAxisPoints(startTimeTime, endTimeTime, avgSuccessResponseTimeValues)
	if len(avgSuccessResponseTimeValuesAdjusted) == 0 {
		return "", fmt.Errorf("failed to adjust avgSuccessResponseTimeValues")
	}

	_, avgFailureResponseTimeValuesAdjusted := adjustXAxisPoints(startTimeTime, endTimeTime, avgFailureResponseTimeValues)
	if len(avgFailureResponseTimeValuesAdjusted) == 0 {
		return "", fmt.Errorf("failed to adjust avgFailureResponseTimeValues")
	}

	line := charts.NewLine()

	line.SetXAxis(xAxis)

	line.AddSeries("Average Response Time", generateLineData(avgResponseTimeValuesAdjusted))
	line.AddSeries("Average Success Response Time", generateLineData(avgSuccessResponseTimeValuesAdjusted))
	line.AddSeries("Average Failure Response Time", generateLineData(avgFailureResponseTimeValuesAdjusted))

	line.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Response Time Over Time(ms)",
		Subtitle: fmt.Sprintf("Test Duration: %s to %s", startTimeTime.Format("15:04:05"), endTimeTime.Format("15:04:05")),
	}), charts.WithLegendOpts(opts.Legend{
		Bottom: "bottom",
	}))

	htmlContent := line.RenderContent()
	if htmlContent == nil {
		fmt.Println("Error: Failed to render chart content")
		return "", fmt.Errorf("failed to render chart content")
	}

	htmlFilePath := filepath.Join(dir, "response_time_chart.html")

	htmlFile, err := os.Create(htmlFilePath)
	if err != nil {
		fmt.Printf("Error creating HTML file: %v\n", err)
		return "", fmt.Errorf("failed to create HTML file: %v", err)
	}
	defer func() {
		if cerr := htmlFile.Close(); cerr != nil {
			fmt.Printf("Error closing HTML file: %v\n", cerr)
		}
	}()

	_, err = htmlFile.Write(htmlContent)
	if err != nil {
		fmt.Printf("Error writing HTML content to file: %v\n", err)
		return "", fmt.Errorf("failed to write HTML content to file: %v", err)
	}

	return htmlFilePath, nil
}

func GenerateFlowTrendChartAsync(avgSentTrafficValues []int, avgReceivedTrafficValues []int, avgTrafficStartTime int64, avgTrafficEndTime int64, dir string) (string, error) {
	startTimeTime := time.Unix(avgTrafficStartTime, 0)
	endTimeTime := time.Unix(avgTrafficEndTime, 0)

	_, avgSentTrafficValuesAdjusted := adjustXAxisPoints(startTimeTime, endTimeTime, avgSentTrafficValues)
	xAxis, avgReceivedTrafficValuesAdjusted := adjustXAxisPoints(startTimeTime, endTimeTime, avgReceivedTrafficValues)

	if len(avgSentTrafficValuesAdjusted) == 0 || len(avgReceivedTrafficValuesAdjusted) == 0 {
		return "", fmt.Errorf("failed to adjust traffic values")
	}

	line := charts.NewLine()
	line.SetXAxis(xAxis)
	line.AddSeries("Sent Traffic", generateLineData(avgSentTrafficValuesAdjusted))
	line.AddSeries("Received Traffic", generateLineData(avgReceivedTrafficValuesAdjusted))

	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    "Flow Trend Over Time (byte)",
			Subtitle: fmt.Sprintf("Test Duration: %s to %s", startTimeTime.Format("15:04:05"), endTimeTime.Format("15:04:05")),
		}),
		charts.WithLegendOpts(opts.Legend{
			Bottom: "bottom",
		}),
	)

	htmlContent := line.RenderContent()
	if htmlContent == nil {
		return "", fmt.Errorf("failed to render chart content")
	}

	htmlFilePath := filepath.Join(dir, "flow_trend_chart.html")

	htmlFile, err := os.Create(htmlFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create HTML file: %v", err)
	}
	defer func() {
		if cerr := htmlFile.Close(); cerr != nil {
			fmt.Printf("Error closing HTML file: %v\n", cerr)
		}
	}()

	_, err = htmlFile.Write(htmlContent)
	if err != nil {
		return "", fmt.Errorf("failed to write HTML content to file: %v", err)
	}

	return htmlFilePath, nil
}
