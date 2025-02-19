package testTasks

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"

	"OpenStress/internal/poolProvider"
	"OpenStress/pool"
	"OpenStress/result"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

// 你可以在这里添加你需要的方法
type task struct{}

func Initialize() {
	// 遍历 acase 包中的所有函数并自动调用以 Acase 开头的函数
	callAcaseFunctions()
}

func callAcaseFunctions() {
	v := reflect.ValueOf(&task{}) // 反射获取 acase 类型的指针
	for i := 0; i < v.NumMethod(); i++ {
		method := v.Method(i)
		if strings.HasPrefix(v.Type().Method(i).Name, "Task") {
			// 打印正在执行的函数名
			fmt.Printf("Executing %s\n", v.Type().Method(i).Name)
			method.Call(nil) // 调用方法
		}
	}
}

func GetOpenStressPool() (*pool.Pool, *pool.StressLogger, *result.Collector, error) {
	taskPool := poolProvider.InitializePool()
	stressLogger, err := pool.GetLogger()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get logger: %w", err)
	}
	// result 模块测试方法
	collectorConfig := result.CollectorConfig{
		BatchSize:       10000,
		OutputFormat:    "jtl",
		JTLFilePath:     "path/to/jtl/file.jtl",
		Logger:          stressLogger,
		NumGoroutines:   10,
		CollectInterval: 5,
		TaskID:          "testTask",
	}
	collector, createcollectoreErr := result.GetCollector(collectorConfig)
	if createcollectoreErr != nil {
		stressLogger.Log("ERROR", "Failed to create collector: "+createcollectoreErr.Error())
		return nil, nil, nil, fmt.Errorf("failed to create collector: %w", err)
	}
	collector.InitializeCollector()

	return taskPool, stressLogger, collector, nil
}

func destroyOpenStressPool(taskPool *pool.Pool, StressLogger *pool.StressLogger, collector *result.Collector) {
	// 加载结果数据
	results, err := collector.LoadResultsFromFile()
	if err != nil {
		fmt.Printf("Error loading results: %v\n", err)
		return
	}
	// 生成并打印测试报告
	report := collector.GenerateSummaryReport(results)
	fmt.Println(report)
	// collector.Close()

	stats, err := collector.GeneratePerformanceStats(results)
	if err != nil {
		fmt.Println("Error generating stats:", err)
		return
	}
	fmt.Println("Performance Stats:")
	fmt.Println(stats)

	// 保存HTML报告到文件
	reportPath, err := collector.SaveReportToFile(stats, "01X批次OpenStress产品基准测试报告")
	if err != nil {
		fmt.Println("Error saving report:", err)
		return
	}

	// 输出生成的报告路径
	fmt.Printf("测试报告已生成：%s\n", reportPath)
}

// WaitWithSystemMonitor 等待指定秒数，同时每30秒监控并打印系统资源使用情况
func WaitWithSystemMonitor(seconds int) {
	endTime := time.Now().Add(time.Duration(seconds) * time.Second)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// 首次立即打印一次系统信息
	printSystemInfo()

	for {
		select {
		case <-ticker.C:
			printSystemInfo()
		default:
			if time.Now().After(endTime) {
				return
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func printSystemInfo() {
	// 获取CPU使用率
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		fmt.Printf("获取CPU信息失败: %v\n", err)
		return
	}

	// 获取内存使用情况
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		fmt.Printf("获取内存信息失败: %v\n", err)
		return
	}

	// 根据操作系统获取磁盘信息
	var diskInfo *disk.UsageStat
	var diskErr error

	if runtime.GOOS == "windows" {
		// Windows系统获取C盘信息
		diskInfo, diskErr = disk.Usage("C:")
	} else {
		// Linux/Unix系统获取根目录信息
		diskInfo, diskErr = disk.Usage("/")
	}

	if diskErr != nil {
		fmt.Printf("获取磁盘信息失败: %v\n", diskErr)
		return
	}

	// 格式化输出系统信息
	fmt.Printf("\n=== 系统资源监控 [%s] ===\n", runtime.GOOS)
	fmt.Printf("CPU 使用率: %.2f%%\n", cpuPercent[0])
	fmt.Printf("内存使用: %.2f%% (已用: %.2f GB, 总共: %.2f GB)\n",
		memInfo.UsedPercent,
		float64(memInfo.Used)/(1024*1024*1024),
		float64(memInfo.Total)/(1024*1024*1024))
	fmt.Printf("磁盘使用: %.2f%% (已用: %.2f GB, 总共: %.2f GB)\n",
		diskInfo.UsedPercent,
		float64(diskInfo.Used)/(1024*1024*1024),
		float64(diskInfo.Total)/(1024*1024*1024))
	fmt.Printf("==================\n")
}
