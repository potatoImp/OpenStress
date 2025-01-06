package tests

import (
	"OpenStress/pool"
	"fmt"

	// "net/http"
	"time"

	"OpenStress/result"

	"github.com/jcmturner/gokrb5/client"
	"gopkg.in/jcmturner/gokrb5.v7/config"
)

// TestTaskPool 测试任务池的功能
func TestTask_AD() {
	maxWorkers := 1000
	taskPool := pool.NewPool(maxWorkers)

	stressLogger, _ := pool.GetLogger()
	// result 模块测试方法
	collectorConfig := result.CollectorConfig{
		BatchSize:       20000,
		OutputFormat:    "jtl",
		JTLFilePath:     "path/to/jtl/file.jtl",
		Logger:          stressLogger,
		NumGoroutines:   10,
		CollectInterval: 50,
		TaskID:          "testTask",
	}
	collector, err := result.NewCollector(collectorConfig)
	if err != nil {
		stressLogger.Log("ERROR", "Failed to create collector: "+err.Error())
	}
	collector.InitializeCollector()

	// 外部配置加载
	var krb5conf = `
[libdefaults]
    default_realm = WTEST.COM
    udp_preference_limit = 1
[realms]
    TEST.COM = {
        kdc = 10.10.27.65
        admin_server = 10.10.27.65
    }
[domain_realm]
    .example.com = WTEST.COM
    example.com = WTEST.COM
`
	var conf, _ = config.NewConfigFromString(krb5conf)
	fmt.Println("krb5配置信息初始化完成：", krb5conf)
	// 定义高优先级任务
	highPriorityTask := func(threadID int32) {

		// 		krb5conf := `
		// [libdefaults]
		//     default_realm = TEST.COM
		//     udp_preference_limit = 1
		// [realms]
		//     TEST.COM = {
		//         kdc = 10.10.27.145
		//         admin_server = 10.10.27.145
		//     }
		// [domain_realm]
		//     .example.com = EXAMPLE.COM
		//     example.com = EXAMPLE.COM
		//     `
		// conf, err := config.NewConfigFromString(krb5conf)
		if err != nil {
			fmt.Println("Error loading krb5 configuration: ", err)

		}
		startTime := time.Now()
		// 创建 Kerberos 客户端
		Kerberos_client := client.NewClientWithPassword("Administrator", "WTEST.COM", "Emm@2022", conf)

		// 登录
		err = Kerberos_client.Login()
		if err != nil {
			fmt.Printf("Error logging in with user: %v", err)
		} else {
			//succ_cnt++
		}

		spn := "HTTP/www.example.com"
		Kerberos_client.GetServiceTicket(spn)
		if err != nil {
			collector.SaveFailureResult(result.ResultData{
				ID:           "test1",
				Type:         result.Failure,
				ResponseTime: 1 * time.Millisecond,
				StartTime:    startTime,
				EndTime:      time.Now(),
				StatusCode:   404,
				Method:       "GET",
				URL:          "Kerberos://10.10.27.145:8089/auth",
				DataSent:     1024,
				DataReceived: 2048,
				ThreadID:     int(threadID),
			})
			fmt.Printf("请求失败: %v\n", err)
			return
		}
		// defer resp.Body.Close()
		// fmt.Printf("请求成功，状态码: %d\n", resp.StatusCode)
		collector.SaveSuccessResult(result.ResultData{
			ID:           "test1",
			Type:         result.Success,
			ResponseTime: 2 * time.Millisecond,
			StartTime:    startTime,
			EndTime:      time.Now(),
			StatusCode:   200,
			Method:       "GET",
			URL:          "Kerberos://10.10.27.145:8089/auth",
			DataSent:     1024,
			DataReceived: 2048,
			ThreadID:     int(threadID),
		})
	}

	// 提交高优先级任务
	for i := 1; i <= 100000; i++ {
		taskID := fmt.Sprintf("请求resources-8080-%d", i)
		taskPool.Submit(highPriorityTask, 3, taskID, 5*time.Second) // 高优先级
	}

	// 启动任务池
	taskPool.Start()

	// 关闭任务池
	taskPool.Shutdown()

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

	// 保存HTML报告到文件
	reportPath, err := collector.SaveReportToFile(stats, "SGP-XCAD产品认证性能测试报告")
	if err != nil {
		fmt.Println("Error saving report:", err)
		return
	}
	// 输出生成的报告路径
	fmt.Printf("测试报告已生成：%s\n", reportPath)

	collector.CloseCollector()
}
