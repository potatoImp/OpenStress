package result

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// adjustXAxisPoints 用于按平均间隔截取 20 个中间时间点，并根据这些时间点返回对应的 Y 轴数值
// values 数组表示从 startTime 到 endTime 之间每秒的数据，按顺序对应
func adjustXAxisPoints(startTime, endTime time.Time, values []int) ([]string, []int) {
	// 如果传入的 values 数组为空，返回错误
	if len(values) == 0 {
		fmt.Println("Error: values array is empty")
		return nil, nil
	}

	// 目标是从 startTime 到 endTime 之间均匀切割成 20 段，得到 20 个中间点和 21 个边界点
	numSegments := 20

	// 计算总时间间隔（秒）
	timeInterval := endTime.Sub(startTime).Seconds()

	// 计算每段的时间间隔
	segmentTime := timeInterval / float64(numSegments)

	// 创建 xAxis 和 yAxis 数组
	xAxis := make([]string, numSegments) // 存储 20 个中间点时间
	yAxis := make([]int, numSegments+1)  // 存储 21 个边界点对应的值

	// 均匀切割时间，获取边界时间点和中间时间点
	for i := 0; i < numSegments; i++ {
		// 计算每个段的边界时间点
		startSegment := startTime.Add(time.Second * time.Duration(float64(i)*segmentTime))
		endSegment := startTime.Add(time.Second * time.Duration(float64(i+1)*segmentTime))

		// 将边界点对应的值存储到 yAxis 数组中
		startSegmentIndex := int(startSegment.Sub(startTime).Seconds()) // 计算边界点在 values 数组中的索引
		endSegmentIndex := int(endSegment.Sub(startTime).Seconds())     // 同上

		// 边界时间点对应的值
		yAxis[i] = values[startSegmentIndex]
		if i == numSegments-1 {
			yAxis[i+1] = values[endSegmentIndex] // 最后一个边界点
		}

		// 计算中间点时间，存储在 xAxis 中
		middleTime := startSegment.Add(time.Second * time.Duration(segmentTime/2))
		xAxis[i] = middleTime.Format("15:04:05") // 转换为 "HH:MM:SS" 格式
	}

	// 最后一个边界点对应的值
	yAxis[numSegments] = values[int(endTime.Sub(startTime).Seconds())]

	// 返回 xAxis 和 yAxis
	return xAxis, yAxis
}

func GenerateTpsChartAsync(tpsValues []int, successValues []int, failureValues []int, startTime int64, endTime int64, dir string) (string, error) {
	// 将 time.Unix 转换为 time.Time 类型
	startTimeTime := time.Unix(startTime, 0)
	endTimeTime := time.Unix(endTime, 0)

	// 调整横坐标点数并获取调整后的数据
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

	// 创建折线图对象
	line := charts.NewLine()
	if line == nil {
		fmt.Println("Error: Failed to create line chart object")
		return "", fmt.Errorf("failed to create line chart object")
	}

	// // 打印调整后的数据
	// fmt.Println("Adjusted TPS Values:", tpsValuesAdjusted)
	// fmt.Println("Adjusted Success Values:", successValuesAdjusted)
	// fmt.Println("Adjusted Failure Values:", failureValuesAdjusted)

	line.SetXAxis(xAxis)
	// line.SetXAxis([]string{"14_21_36", "14_21_39", "14_21_43", "Thu", "Fri", "Sat", "Sun", "exoi", "8", "9"})
	// 添加数据系列
	line.AddSeries("Total TPS", generateLineData(tpsValuesAdjusted))
	if err := checkError("Failed to add Total TPS series"); err != nil {
		return "", err
	}

	// 取消注释以启用其他数据系列
	line.AddSeries("Success TPS", generateLineData(successValuesAdjusted))
	line.AddSeries("Failure TPS", generateLineData(failureValuesAdjusted))

	// 打印生成的数据
	// fmt.Println("Y轴数据:", generateLineData(tpsValuesAdjusted))
	// fmt.Println("Y轴数据长度:", len(generateLineData(tpsValuesAdjusted)))

	// 设置 X 轴
	// fmt.Println("X轴数据:", xAxis)
	// fmt.Println("X轴数据长度:", len(xAxis))

	// 设置全局选项
	line.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Transactions Per Second",
		Subtitle: fmt.Sprintf("Test Duration: %s to %s", startTimeTime.Format("15:04:05"), endTimeTime.Format("15:04:05")),
	}), charts.WithLegendOpts(opts.Legend{
		Bottom: "bottom", // 设置图例的位置，可以是 "top"、"bottom"、"left"、"right"
	}), charts.WithInitializationOpts(opts.Initialization{
		AssetsHost: "assets/", // 设置本地静态资源路径
	}),
	)

	// 获取渲染的 HTML 内容（不需要通过 io.Writer）
	htmlContent := line.RenderContent()
	if htmlContent == nil {
		fmt.Println("Error: Failed to render chart content")
		return "", fmt.Errorf("failed to render chart content")
	}

	// 打印渲染后的 HTML 内容
	// fmt.Println("Rendered HTML Content:")
	// fmt.Println(string(htmlContent)) // 打印整个 HTML 内容

	// 生成 HTML 文件路径
	htmlFilePath := filepath.Join(dir, "/tps_chart.html")

	// 创建文件并检查错误
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

	// 将渲染的 HTML 内容写入文件
	_, err = htmlFile.Write(htmlContent)
	if err != nil {
		fmt.Printf("Error writing HTML content to file: %v\n", err)
		return "", fmt.Errorf("failed to write HTML content to file: %v", err)
	}

	// 返回 HTML 文件路径
	return htmlFilePath, nil
}

// 辅助函数：用于检查错误并打印相应的错误信息
func checkError(msg string) error {
	if r := recover(); r != nil {
		fmt.Printf("Error: %s: %v\n", msg, r)
		return fmt.Errorf("%s: %v", msg, r)
	}
	return nil
}

func GenerateResponseTimeChartAsync(avgResponseTimeValues []int, avgSuccessResponseTimeValues []int, avgFailureResponseTimeValues []int, avgResponseStartTime int64, avgResponseEndTime int64, dir string) (string, error) {
	// 将 time.Unix 转换为 time.Time 类型
	startTimeTime := time.Unix(avgResponseStartTime, 0)
	endTimeTime := time.Unix(avgResponseEndTime, 0)

	// 调整横坐标点数并获取调整后的数据
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

	// 创建折线图对象
	line := charts.NewLine()

	// 设置 X 轴
	line.SetXAxis(xAxis)
	// line.SetXAxis([]string{"14_21_36", "14_21_39", "14_21_43", "Thu", "Fri", "Sat", "Sun", "exoi", "8", "9"})

	// 设置 X 轴
	// fmt.Println("X轴数据:", xAxis)
	// fmt.Println("X轴数据长度:", len(xAxis))

	// 打印生成的数据
	// fmt.Println("avgResponseTimeValuesAdjustedY轴数据:", generateLineData(avgResponseTimeValuesAdjusted))
	// fmt.Println("Y轴数据长度:", len(generateLineData(avgResponseTimeValuesAdjusted)))
	// fmt.Println("avgSuccessResponseTimeValuesAdjustedY轴数据:", generateLineData(avgSuccessResponseTimeValuesAdjusted))
	// fmt.Println("Y轴数据长度:", len(generateLineData(avgSuccessResponseTimeValuesAdjusted)))
	// fmt.Println("avgFailureResponseTimeValuesAdjustedY轴数据:", generateLineData(avgFailureResponseTimeValuesAdjusted))
	// fmt.Println("Y轴数据长度:", len(generateLineData(avgFailureResponseTimeValuesAdjusted)))

	// 添加数据系列
	line.AddSeries("Average Response Time", generateLineData(avgResponseTimeValuesAdjusted))
	line.AddSeries("Average Success Response Time", generateLineData(avgSuccessResponseTimeValuesAdjusted))
	line.AddSeries("Average Failure Response Time", generateLineData(avgFailureResponseTimeValuesAdjusted))

	// 设置全局选项
	line.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Response Time Over Time(ms)",
		Subtitle: fmt.Sprintf("Test Duration: %s to %s", startTimeTime.Format("15:04:05"), endTimeTime.Format("15:04:05")),
	}), charts.WithLegendOpts(opts.Legend{
		Bottom: "bottom", // 设置图例的位置，可以是 "top"、"bottom"、"left"、"right"
	}), charts.WithInitializationOpts(opts.Initialization{
		AssetsHost: "assets/", // 设置本地静态资源路径
	}),
	)

	// 获取渲染的 HTML 内容（不需要通过 io.Writer）
	htmlContent := line.RenderContent()
	if htmlContent == nil {
		fmt.Println("Error: Failed to render chart content")
		return "", fmt.Errorf("failed to render chart content")
	}

	// 打印渲染后的 HTML 内容
	// fmt.Println("Rendered HTML Content:")
	// fmt.Println(string(htmlContent)) // 打印整个 HTML 内容

	// 生成 HTML 文件路径
	htmlFilePath := filepath.Join(dir, "response_time_chart.html")
	// fmt.Println("HTML 文件路径:", htmlFilePath)

	// 创建文件并检查错误
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

	// 将渲染的 HTML 内容写入文件
	_, err = htmlFile.Write(htmlContent)
	if err != nil {
		fmt.Printf("Error writing HTML content to file: %v\n", err)
		return "", fmt.Errorf("failed to write HTML content to file: %v", err)
	}

	// 返回 HTML 文件路径
	return htmlFilePath, nil
}

func GenerateFlowTrendChartAsync(avgSentTrafficValues []int, avgReceivedTrafficValues []int, avgTrafficStartTime int64, avgTrafficEndTime int64, dir string) (string, error) {
	// 将 time.Unix 转换为 time.Time 类型
	startTimeTime := time.Unix(avgTrafficStartTime, 0)
	endTimeTime := time.Unix(avgTrafficEndTime, 0)

	// 调整横坐标点数
	_, avgSentTrafficValuesAdjusted := adjustXAxisPoints(startTimeTime, endTimeTime, avgSentTrafficValues)
	xAxis, avgReceivedTrafficValuesAdjusted := adjustXAxisPoints(startTimeTime, endTimeTime, avgReceivedTrafficValues)

	// 检查数据是否为空
	if len(avgSentTrafficValuesAdjusted) == 0 || len(avgReceivedTrafficValuesAdjusted) == 0 {
		return "", fmt.Errorf("failed to adjust traffic values")
	}

	// 创建折线图对象
	line := charts.NewLine()

	// 设置 X 轴
	line.SetXAxis(xAxis)
	// fmt.Println("X轴数据:", xAxis)
	// fmt.Println("X轴数据长度:", len(xAxis))

	// 添加数据系列
	line.AddSeries("Sent Traffic", generateLineData(avgSentTrafficValuesAdjusted))
	line.AddSeries("Received Traffic", generateLineData(avgReceivedTrafficValuesAdjusted))

	// fmt.Println("Y轴数据:", generateLineData(avgSentTrafficValuesAdjusted))
	// fmt.Println("Y轴数据长度:", len(generateLineData(avgSentTrafficValuesAdjusted)))

	// fmt.Println("Y轴数据:", generateLineData(avgReceivedTrafficValuesAdjusted))
	// fmt.Println("Y轴数据长度:", len(generateLineData(avgReceivedTrafficValuesAdjusted)))

	// 设置全局选项
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    "Flow Trend Over Time (byte)",
			Subtitle: fmt.Sprintf("Test Duration: %s to %s", startTimeTime.Format("15:04:05"), endTimeTime.Format("15:04:05")),
		}),
		charts.WithLegendOpts(opts.Legend{
			Bottom: "bottom", // 设置图例位置
		}),
		charts.WithInitializationOpts(opts.Initialization{
			AssetsHost: "assets/", // 设置本地静态资源路径
		}),
	)

	// 获取渲染的 HTML 内容
	htmlContent := line.RenderContent()
	if htmlContent == nil {
		return "", fmt.Errorf("failed to render chart content")
	}

	// 打印渲染后的 HTML 内容
	// fmt.Println("Rendered HTML Content:")
	// fmt.Println(string(htmlContent)) // 打印整个 HTML 内容

	// 生成 HTML 文件路径
	htmlFilePath := filepath.Join(dir, "flow_trend_chart.html")
	// fmt.Println("HTML 文件路径:", htmlFilePath)

	// 创建文件并检查错误
	htmlFile, err := os.Create(htmlFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create HTML file: %v", err)
	}
	defer func() {
		if cerr := htmlFile.Close(); cerr != nil {
			fmt.Printf("Error closing HTML file: %v\n", cerr)
		}
	}()

	// 将渲染的 HTML 内容写入文件
	_, err = htmlFile.Write(htmlContent)
	if err != nil {
		return "", fmt.Errorf("failed to write HTML content to file: %v", err)
	}

	// 返回生成的 HTML 文件路径
	return htmlFilePath, nil
}
