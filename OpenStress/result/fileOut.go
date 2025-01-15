package result

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SaveReportToFile 保存报告到HTML文件
func (c *Collector) SaveReportToFile(stats map[string]interface{}, customName ...string) (string, error) {
	// 获取当前日期时间，格式化为 yyyy-MM-dd_HH-mm-ss
	currentTime := time.Now().Format("2006-01-02_15-04-05")

	// 判断是否传递了自定义名称，如果没有，使用默认名称
	var name string
	if len(customName) > 0 && customName[0] != "" {
		name = customName[0]
	} else {
		name = "performance_report"
	}

	// 创建与文件同名的目录
	dir := fmt.Sprintf("path/to/htmlReport/%s_%s", name, currentTime)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}

	// 定义保存的HTML文件路径
	htmlFilePath := filepath.Join(dir, fmt.Sprintf("%s_%s.html", name, currentTime))

	// 创建 static 目录
	staticDirPath := filepath.Join(dir, "/static/")
	// fmt.Println(staticDirPath)
	err = os.MkdirAll(staticDirPath, 0777)
	if err != nil {
		return "", fmt.Errorf("failed to create static directory: %v", err)
	}

	// 创建 static 目录
	staticAssetsDirPath := filepath.Join(dir, "/static/assets")
	// fmt.Println(staticAssetsDirPath)
	err = os.MkdirAll(staticAssetsDirPath, 0777)
	if err != nil {
		return "", fmt.Errorf("failed to create static directory: %v", err)
	}

	// 使用 WaitGroup 来等待主线程完成初始化
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done() // 在 goroutine 完成时通知主线程

		// 初始化切片
		var tpsValues, successValues, failureValues []int

		// 遍历并提取 tpsValues（如果需要处理，可以在这里做额外的转换或操作）
		if tpsValuesRaw, ok := stats["TPSValues"].([]int); ok {
			for _, v := range tpsValuesRaw {
				// 在这里可以对 tps 值做进一步处理，例如加倍、过滤等
				// 这里只是简单的添加到新切片中
				tpsValues = append(tpsValues, v)
			}
		} else {
			fmt.Println("Error: TPSValues is not of type []int")
		}

		// 遍历并提取 successValues（如果需要处理，可以在这里做额外的转换或操作）
		if successValuesRaw, ok := stats["SuccessValues"].([]int); ok {
			for _, v := range successValuesRaw {
				// 这里可以对 success 值进行操作，例如加倍、过滤等
				// 这里只是简单的添加到新切片中
				successValues = append(successValues, v)
			}
		} else {
			fmt.Println("Error: SuccessValues is not of type []int")
		}

		// 遍历并提取 failureValues（如果需要处理，可以在这里做额外的转换或操作）
		if failureValuesRaw, ok := stats["FailureValues"].([]int); ok {
			for _, v := range failureValuesRaw {
				// 这里可以对 failure 值进行操作，例如加倍、过滤等
				// 这里只是简单的添加到新切片中
				failureValues = append(failureValues, v)
			}
		} else {
			fmt.Println("Error: FailureValues is not of type []int")
		}
		_, GenerateTpsCharterr := GenerateTpsChartAsync(tpsValues,
			successValues,
			failureValues,
			stats["AvgTpsStartTime"].(int64),
			stats["AvgTpsEndTime"].(int64),
			staticDirPath)
		if GenerateTpsCharterr != nil {
			fmt.Printf("Error generating chart: %v", err)
		}

		// 初始化切片
		var avgResponseTimeValues, avgSuccessResponseTimeValues, avgFailureResponseTimeValues []int

		// 遍历并提取 avgResponseTimeValues（如果需要处理，可以在这里做额外的转换或操作）
		if avgResponseTimeValuesRaw, ok := stats["AvgResponseTimeValues"].([]int); ok {
			for _, v := range avgResponseTimeValuesRaw {
				// 在这里可以对 avgResponseTime 值做进一步处理，例如加倍、过滤等
				// 这里只是简单的添加到新切片中
				avgResponseTimeValues = append(avgResponseTimeValues, v)
			}
		} else {
			fmt.Println("Error: AvgResponseTimeValues is not of type []int")
		}

		// 遍历并提取 avgSuccessResponseTimeValues（如果需要处理，可以在这里做额外的转换或操作）
		if avgSuccessResponseTimeValuesRaw, ok := stats["AvgSuccessResponseTimeValues"].([]int); ok {
			for _, v := range avgSuccessResponseTimeValuesRaw {
				// 这里可以对 avgSuccessResponseTime 值进行操作，例如加倍、过滤等
				// 这里只是简单的添加到新切片中
				avgSuccessResponseTimeValues = append(avgSuccessResponseTimeValues, v)
			}
		} else {
			fmt.Println("Error: AvgSuccessResponseTimeValues is not of type []int")
		}

		// 遍历并提取 avgFailureResponseTimeValues（如果需要处理，可以在这里做额外的转换或操作）
		if avgFailureResponseTimeValuesRaw, ok := stats["AvgFailureResponseTimeValues"].([]int); ok {
			for _, v := range avgFailureResponseTimeValuesRaw {
				// 这里可以对 avgFailureResponseTime 值进行操作，例如加倍、过滤等
				// 这里只是简单的添加到新切片中
				avgFailureResponseTimeValues = append(avgFailureResponseTimeValues, v)
			}
		} else {
			fmt.Println("Error: AvgFailureResponseTimeValues is not of type []int")
		}

		// 调用 GenerateResponseTimeChartAsync 函数并传递参数
		_, GenerateResponseTimeCharterr := GenerateResponseTimeChartAsync(
			avgResponseTimeValues,
			avgSuccessResponseTimeValues,
			avgFailureResponseTimeValues,
			stats["AvgResponseStartTime"].(int64),
			stats["AvgResponseEndTime"].(int64),
			staticDirPath,
		)

		if GenerateResponseTimeCharterr != nil {
			fmt.Printf("Error generating chart: %v", GenerateResponseTimeCharterr)
		}

		// 初始化切片
		var avgSentTrafficValues, avgReceivedTrafficValues []int

		// 遍历并提取 avgSentTrafficValues（如果需要处理，可以在这里做额外的转换或操作）
		if avgSentTrafficValuesRaw, ok := stats["AvgSentTrafficValues"].([]int); ok {
			for _, v := range avgSentTrafficValuesRaw {
				// 在这里可以对 avgSentTraffic 值做进一步处理，例如加倍、过滤等
				// 这里只是简单的添加到新切片中
				avgSentTrafficValues = append(avgSentTrafficValues, v)
			}
		} else {
			fmt.Println("Error: AvgSentTrafficValues is not of type []int")
		}

		// 遍历并提取 avgReceivedTrafficValues（如果需要处理，可以在这里做额外的转换或操作）
		if avgReceivedTrafficValuesRaw, ok := stats["AvgReceivedTrafficValues"].([]int); ok {
			for _, v := range avgReceivedTrafficValuesRaw {
				// 这里可以对 avgReceivedTraffic 值进行操作，例如加倍、过滤等
				// 这里只是简单的添加到新切片中
				avgReceivedTrafficValues = append(avgReceivedTrafficValues, v)
			}
		} else {
			fmt.Println("Error: AvgReceivedTrafficValues is not of type []int")
		}

		// 调用 GenerateFlowTrendChartAsync 函数并传递参数
		_, GenerateFlowTrendCharterr := GenerateFlowTrendChartAsync(
			avgSentTrafficValues,                 // 已处理的 avgSentTrafficValues
			avgReceivedTrafficValues,             // 已处理的 avgReceivedTrafficValues
			stats["AvgTrafficStartTime"].(int64), // 从 stats 提取的时间参数
			stats["AvgTrafficEndTime"].(int64),   // 从 stats 提取的时间参数
			staticDirPath,
		)

		if GenerateFlowTrendCharterr != nil {
			fmt.Printf("Error generating flow trend chart: %v", GenerateFlowTrendCharterr)
		}
	}()

	// 生成HTML报告
	reportContent := GenerateHTMLReport(stats, false, name)

	// 创建HTML文件
	file, err := os.Create(htmlFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create HTML report: %v", err)
	}
	defer file.Close()

	// 写入报告内容
	_, err = file.WriteString(reportContent)
	if err != nil {
		return "", fmt.Errorf("failed to write HTML content: %v", err)
	}

	// 生成并保存 styles.css
	cssFilePath := filepath.Join(staticDirPath, "styles.css")
	cssContent := generateCSS() // 调用生成CSS的函数
	err = os.WriteFile(cssFilePath, []byte(cssContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write CSS file: %v", err)
	}

	// 生成并保存 script.js
	jsFilePath := filepath.Join(staticDirPath, "script.js")
	jsContent := generateScript() // 调用生成JS的函数
	err = os.WriteFile(jsFilePath, []byte(jsContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write JavaScript file: %v", err)
	}

	// 生成并保存 echarts.min.js
	echartsMinJsFilePath := filepath.Join(staticAssetsDirPath, "echarts.min.js")
	echartsMinJsContent := generateEchartsMinJs() // 调用生成JS的函数
	err = os.WriteFile(echartsMinJsFilePath, []byte(echartsMinJsContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write JavaScript file: %v", err)
	}

	// 等待 goroutine 完成
	wg.Wait()
	// 返回文件路径
	return htmlFilePath, nil
}
