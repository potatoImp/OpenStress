package result

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func (c *Collector) SaveReportToFile(stats map[string]interface{}, customName ...string) (string, error) {
	currentTime := time.Now().Format("2006-01-02_15-04-05")

	var name string
	if len(customName) > 0 && customName[0] != "" {
		name = customName[0]
	} else {
		name = "performance_report"
	}

	dir := fmt.Sprintf("path/to/htmlReport/%s_%s", name, currentTime)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}

	htmlFilePath := filepath.Join(dir, fmt.Sprintf("%s_%s.html", name, currentTime))

	staticDirPath := filepath.Join(dir, "/static/")
	fmt.Println(staticDirPath)
	err = os.MkdirAll(staticDirPath, 0777)
	if err != nil {
		return "", fmt.Errorf("failed to create static directory: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>..")

		var tpsValues, successValues, failureValues []int

		if tpsValuesRaw, ok := stats["TPSValues"].([]int); ok {
			for _, v := range tpsValuesRaw {
				tpsValues = append(tpsValues, v)
			}
		} else {
			fmt.Println("Error: TPSValues is not of type []int")
		}

		if successValuesRaw, ok := stats["SuccessValues"].([]int); ok {
			for _, v := range successValuesRaw {
				successValues = append(successValues, v)
			}
		} else {
			fmt.Println("Error: SuccessValues is not of type []int")
		}

		if failureValuesRaw, ok := stats["FailureValues"].([]int); ok {
			for _, v := range failureValuesRaw {
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

		var avgResponseTimeValues, avgSuccessResponseTimeValues, avgFailureResponseTimeValues []int

		if avgResponseTimeValuesRaw, ok := stats["AvgResponseTimeValues"].([]int); ok {
			for _, v := range avgResponseTimeValuesRaw {
				avgResponseTimeValues = append(avgResponseTimeValues, v)
			}
		} else {
			fmt.Println("Error: AvgResponseTimeValues is not of type []int")
		}

		if avgSuccessResponseTimeValuesRaw, ok := stats["AvgSuccessResponseTimeValues"].([]int); ok {
			for _, v := range avgSuccessResponseTimeValuesRaw {
				avgSuccessResponseTimeValues = append(avgSuccessResponseTimeValues, v)
			}
		} else {
			fmt.Println("Error: AvgSuccessResponseTimeValues is not of type []int")
		}

		if avgFailureResponseTimeValuesRaw, ok := stats["AvgFailureResponseTimeValues"].([]int); ok {
			for _, v := range avgFailureResponseTimeValuesRaw {
				avgFailureResponseTimeValues = append(avgFailureResponseTimeValues, v)
			}
		} else {
			fmt.Println("Error: AvgFailureResponseTimeValues is not of type []int")
		}

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

		var avgSentTrafficValues, avgReceivedTrafficValues []int

		if avgSentTrafficValuesRaw, ok := stats["AvgSentTrafficValues"].([]int); ok {
			for _, v := range avgSentTrafficValuesRaw {
				avgSentTrafficValues = append(avgSentTrafficValues, v)
			}
		} else {
			fmt.Println("Error: AvgSentTrafficValues is not of type []int")
		}

		if avgReceivedTrafficValuesRaw, ok := stats["AvgReceivedTrafficValues"].([]int); ok {
			for _, v := range avgReceivedTrafficValuesRaw {
				avgReceivedTrafficValues = append(avgReceivedTrafficValues, v)
			}
		} else {
			fmt.Println("Error: AvgReceivedTrafficValues is not of type []int")
		}

		_, GenerateFlowTrendCharterr := GenerateFlowTrendChartAsync(
			avgSentTrafficValues,
			avgReceivedTrafficValues,
			stats["AvgTrafficStartTime"].(int64),
			stats["AvgTrafficEndTime"].(int64),
			staticDirPath,
		)

		if GenerateFlowTrendCharterr != nil {
			fmt.Printf("Error generating flow trend chart: %v", GenerateFlowTrendCharterr)
		}
	}()

	reportContent := GenerateHTMLReport(stats, name)

	file, err := os.Create(htmlFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create HTML report: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(reportContent)
	if err != nil {
		return "", fmt.Errorf("failed to write HTML content: %v", err)
	}

	cssFilePath := filepath.Join(staticDirPath, "styles.css")
	cssContent := generateCSS()
	err = os.WriteFile(cssFilePath, []byte(cssContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write CSS file: %v", err)
	}

	jsFilePath := filepath.Join(staticDirPath, "script.js")
	jsContent := generateScript()
	err = os.WriteFile(jsFilePath, []byte(jsContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write JavaScript file: %v", err)
	}

	wg.Wait()
	return htmlFilePath, nil
}
