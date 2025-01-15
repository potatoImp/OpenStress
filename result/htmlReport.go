package result

import (
	"fmt"
	"strings"

	"encoding/json"

	"OpenStress/internal/llmProvider"
	"time"
)

// GenerateSummaryReport 生成测试报告
func (c *Collector) GenerateSummaryReport(results []ResultData) string {
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

	// 计算成功率和平均响应时间
	successRate := float64(successCount) / float64(totalRequests) * 100
	avgResponseTime := totalResponseTime / time.Duration(totalRequests)

	// 计算 TPS (每秒事务数)
	var tps float64
	totalRunTime := time.Duration(lastTimestamp-firstTimestamp) * time.Millisecond
	if totalRunTime.Seconds() > 0 {
		tps = float64(totalRequests) / totalRunTime.Seconds()
	}

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

	// 生成报告
	report := fmt.Sprintf("测试报告:\n\n")
	report += fmt.Sprintf("总请求数: %d\n", totalRequests)
	report += fmt.Sprintf("成功请求数: %d (%.3f%%)\n", successCount, successRate)
	report += fmt.Sprintf("失败请求数: %d\n", failureCount)
	report += fmt.Sprintf("平均响应时间: %s\n", avgResponseTime)
	report += fmt.Sprintf("最大响应时间: %s\n", maxResponseTime)
	report += fmt.Sprintf("最小响应时间: %s\n", minResponseTime)
	report += fmt.Sprintf("总运行时间: %s\n", totalRunTime)
	report += fmt.Sprintf("TPS: %.2f\n", tps)
	report += fmt.Sprintf("每秒发送数据流量: %s\n", sentDataPerSecStr)
	report += fmt.Sprintf("每秒接收数据流量: %s\n", receivedDataPerSecStr)
	report += fmt.Sprintf("总发送数据量: %s\n", totalSentDataStr)
	report += fmt.Sprintf("总接收数据量: %s\n", totalReceivedDataStr)

	// 返回报告
	return report
}

// 提取 SystemPerformance 和 Risk 字段的函数
func extractSystemPerformanceAndRisk(data map[string]interface{}) (string, string, string, error) {
	// 1. 获取 choices 中的第一个元素
	choices, ok := data["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", "", "", fmt.Errorf("无法获取 choices 数据")
	}

	// 2. 获取第一个元素中的 message.content 字段
	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", "", "", fmt.Errorf("无法获取 choice 数据")
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return "", "", "", fmt.Errorf("无法获取 message 数据")
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", "", "", fmt.Errorf("无法获取 content 字段")
	}

	// 3. 去掉 content 中的 ```json 和 ```, 清理字符串
	content = strings.TrimPrefix(content, "```json\n")
	content = strings.TrimSuffix(content, "```")

	// 4. 将 content 字段中的 JSON 字符串解析为新的 map
	var analysisData map[string]interface{}
	err := json.Unmarshal([]byte(content), &analysisData)
	if err != nil {
		return "", "", "", fmt.Errorf("无法解析 content 中的 JSON 数据: %w", err)
	}

	// 5. 提取 SystemPerformance 和 Risk 字段
	systemPerformance, ok := analysisData["SystemPerformance"].(string)
	if !ok {
		systemPerformance = "未能获取系统性能分析"
	}

	risk, ok := analysisData["Risk"].(string)
	if !ok {
		risk = "未能获取风险分析"
	}

	nextPlan, ok := analysisData["NextPlan"].(string)
	if !ok {
		nextPlan = "未能获取下一步计划建议"
	}

	return systemPerformance, risk, nextPlan, nil
}

// 打印所有字段的函数
func printFields(data map[string]interface{}) {
	// 打印根字段 choices
	choices, ok := data["choices"].([]interface{})
	if !ok {
		fmt.Println("无法获取 choices 数据")
		return
	}
	fmt.Println("choices:")
	for i, choice := range choices {
		choiceMap, ok := choice.(map[string]interface{})
		if !ok {
			fmt.Println("无法解析 choice 数据")
			continue
		}

		// 打印每个 choice 的字段
		fmt.Printf("  Choice %d:\n", i+1)
		for key, value := range choiceMap {
			fmt.Printf("    %s: %v\n", key, value)
		}

		// 打印 message.content 字段内容
		message, ok := choiceMap["message"].(map[string]interface{})
		if !ok {
			fmt.Println("    message 字段未找到或类型错误")
			continue
		}
		content, ok := message["content"].(string)
		if !ok {
			fmt.Println("    message.content 字段未找到或类型错误")
			continue
		}

		// 去掉 ```json 和 ```
		content = strings.TrimPrefix(content, "```json\n")
		content = strings.TrimSuffix(content, "```")

		// 打印 content 字段内容
		fmt.Println("    message.content:")
		fmt.Println(content)

		// 解析 content 为 JSON 对象并打印
		var analysisData map[string]interface{}
		err := json.Unmarshal([]byte(content), &analysisData)
		if err != nil {
			fmt.Println("    无法解析 content 字段中的 JSON 数据:", err)
			continue
		} else {
			fmt.Println("    content 解析后的数据:")
			for key, value := range analysisData {
				// 打印解析后的每个字段
				fmt.Printf("      %s: %v\n", key, value)
			}
		}
	}
}

// GenerateHTMLReport 生成性能测试报告的HTML
func GenerateHTMLReport(stats map[string]interface{}, useLLMProvider bool, title ...string) string {
	var builder strings.Builder

	// 可选的参数，使用默认值
	pageTitle := "性能测试报告" // 默认标题
	logoPath := ""        // 默认无logo
	// analysisContent := generateDefaultAnalysis(stats) // 不通过llm进行数据分析时，则根据测试数据自动生成的默认分析内容

	// 如果传入了自定义的标题，则使用传入的标题
	if len(title) > 0 {
		pageTitle = title[0]
	}

	// 参考标准列表
	standards := []PerformanceStandard{
		{Field: "AvgResponseTime", Max: MaxAvgResponseTime, Compare: func(value interface{}) float64 {
			return value.(time.Duration).Seconds()
		}},
		{Field: "SuccessRate", Min: MinSuccessRate, Compare: func(value interface{}) float64 {
			return value.(float64)
		}},
		{Field: "TPS", Min: MaxTPS, Compare: func(value interface{}) float64 {
			return value.(float64)
		}},
		{Field: "AvgResponseTime", Max: MaxHighFreqResponseTime, Compare: func(value interface{}) float64 {
			return value.(time.Duration).Seconds()
		}},
	}

	// HTML基础结构
	builder.WriteString("<!DOCTYPE html>")
	builder.WriteString("<html lang='zh'>")
	builder.WriteString("<head>")
	builder.WriteString("<meta charset='UTF-8'>")
	builder.WriteString("<meta name='viewport' content='width=device-width, initial-scale=1.0'>")
	builder.WriteString("<title>" + pageTitle + "</title>")

	// 如果传入了logo路径，则添加logo
	if logoPath != "" {
		builder.WriteString("<link rel='icon' href='" + logoPath + "'>") // 设置logo图标
	}

	// 更新CSS和JS文件路径
	builder.WriteString("<link rel='stylesheet' href='static/styles.css'>")
	builder.WriteString("<style>")
	builder.WriteString(".error {color: red; font-weight: bold;}")      // 错误字段样式
	builder.WriteString(".warning {color: orange; font-weight: bold;}") // 警告字段样式
	builder.WriteString(".chart {height: auto; min-height: 400px;}")    // 添加自动高度，最小高度 400px
	builder.WriteString("</style>")
	builder.WriteString("<script src='https://cdn.jsdelivr.net/npm/chart.js'></script>") // 引入Chart.js库
	builder.WriteString("</head>")
	builder.WriteString("<body>")
	builder.WriteString("<div class='container'>")

	// 标题部分
	builder.WriteString("<header><h1>" + pageTitle + "</h1></header>")

	// 测试概览部分
	builder.WriteString("<section class='report-summary'>")
	builder.WriteString("<h2><span class='section-icon'>📋</span>测试概览</h2>")
	builder.WriteString("<table>")
	builder.WriteString("<tr><th>开始时间</th><td>" + time.Unix(stats["AvgTpsStartTime"].(int64), 0).Format("2006-01-02 15:04:05") + "</td></tr>")
	builder.WriteString("<tr><th>结束时间</th><td>" + time.Unix(stats["AvgTpsEndTime"].(int64), 0).Format("2006-01-02 15:04:05") + "</td></tr>")
	builder.WriteString("</table>")
	builder.WriteString("</section>")

	// 测试统计数据部分
	builder.WriteString("<section class='test-statistics'>")
	builder.WriteString("<h2><span class='section-icon'>📊</span>测试统计数据</h2>")
	builder.WriteString("<table>")

	// 统计数据列表，包括 SuccessRate
	keys := []string{"TotalRequests", "SuccessCount", "FailureCount", "SuccessRate", "AvgResponseTime", "MaxResponseTime", "MinResponseTime", "TotalRunTime", "TPS", "SentDataPerSec", "ReceivedDataPerSec", "TotalSentData", "TotalReceivedData"}

	for _, key := range keys {
		value := stats[key]
		class := ""

		// 针对每个字段比较参考标准
		for _, standard := range standards {
			if standard.Field == key {
				compareValue := standard.Compare(value)
				if standard.Min > 0 && compareValue < standard.Min {
					class = "error"
				} else if standard.Max > 0 && compareValue > standard.Max {
					class = "warning"
				}
			}
		}

		// 对 AvgResponseTime, MaxResponseTime, MinResponseTime, TotalRunTime 字段特殊处理，转换为毫秒并保留两位小数
		if key == "AvgResponseTime" || key == "MaxResponseTime" || key == "MinResponseTime" || key == "TotalRunTime" {
			value = fmt.Sprintf("%.2f ms", float64(value.(time.Duration))/float64(time.Millisecond))
		}

		// 对 SuccessRate 特殊处理，添加 % 符号
		if key == "SuccessRate" {
			value = fmt.Sprintf("%.3f%%", value)
		}

		// 生成数据行
		builder.WriteString("<tr>")
		builder.WriteString("<th>" + key + "</th>")
		if class != "" {
			builder.WriteString("<td class='" + class + "'>" + fmt.Sprintf("%v", value) + "</td>")
		} else {
			builder.WriteString("<td>" + fmt.Sprintf("%v", value) + "</td>")
		}
		builder.WriteString("</tr>")
	}

	builder.WriteString("</table>")
	builder.WriteString("</section>")

	// 统计图部分 - 使用 <img> 标签嵌入 SVG 图像
	builder.WriteString("<section class='charts'>")
	builder.WriteString("<h2><span class='section-icon'>📈</span>视图展示</h2>")

	// 添加TPS趋势图部分
	builder.WriteString("<div class='chart'><h3>TPS趋势图</h3>")
	// 使用iframe标签来嵌入tps_chart.html，并应用优化后的样式
	builder.WriteString("<iframe class='tps-chart' src='static/tps_chart.html' frameborder='0'></iframe>")
	builder.WriteString("</div>")

	// 添加response_time_chart趋势图部分
	builder.WriteString("<div class='chart'><h3>请求响应时间趋势图</h3>")
	// 使用iframe标签来嵌入response_time_chart.html，并应用优化后的样式
	builder.WriteString("<iframe class='tps-chart' src='static/response_time_chart.html' frameborder='0'></iframe>")
	builder.WriteString("</div>")

	// 添加response_time_chart趋势图部分
	builder.WriteString("<div class='chart'><h3>网络流量趋势图</h3>")
	// 使用iframe标签来嵌入flow_trend_chart.html，并应用优化后的样式
	builder.WriteString("<iframe class='tps-chart' src='static/flow_trend_chart.html' frameborder='0'></iframe>")
	builder.WriteString("</div>")
	builder.WriteString("</section>")

	// // LLMRequestParams 配置
	// llmParams := llmProvider.LLMRequestParams{
	// 	APIType:     "kimi",
	// 	BaseURL:     "https://api.moonshot.cn/v1/chat",
	// 	APIKey:      "sk-UyI3Y5zBNDNuyZ83ql6QIrQwLfwO2GYgh0s13hTBY8Fcn5ae", // 请替换为实际的 API Key
	// 	Model:       "kimi 8k",
	// 	Proxy:       "", // 如有需要可配置代理
	// 	Timeout:     60, // 请求超时（单位：秒）
	// 	PricingPlan: "free",
	// 	Prompt:      "", // 初始为空，后续会动态设置
	// }

	// // 初始化 LLMProvider，设置缓存 TTL 为 5 分钟，token 价格为 0.02 美元/千个 token
	// cacheTTL := 5 * time.Minute
	// tokenPrice := 0.02
	// llmProviderInstance := llmProvider.NewLLMProvider(llmParams, cacheTTL, tokenPrice)

	// // 调用 AnalyzePerformanceAndGetResponse 函数
	// AIanalysisContentJson, tokenCost, err := llmProviderInstance.AnalyzePerformanceAndGetResponse(stats, llmParams)
	// if err != nil {
	// 	fmt.Printf("调用 LLM API 时发生错误: %v", err)
	// }

	// // 打印响应数据和 token 花费
	// fmt.Printf("LLM 响应:\n%v\n", AIanalysisContentJson)
	// fmt.Println("===========================================")
	// printFields(AIanalysisContentJson)
	// fmt.Println("===========================================")
	// fmt.Printf("Token 花费: $%.4f\n", tokenCost)

	// // 提取 SystemPerformance 和 Risk 字段
	// systemPerformance, risk, nextPlan, err := extractSystemPerformanceAndRisk(AIanalysisContentJson)
	// if err != nil {
	// 	fmt.Printf("AI分析数据时发生错误: %v\n，将使用默认分析", err)
	// }

	// // 分析部分
	builder.WriteString("<section class='analysis concept-card'>")
	builder.WriteString("<h2><span class='analysis-icon'>📝</span>分析</h2>")
	// // builder.WriteString("<p>" + analysisContent + "</p>")
	// fmt.Println("systemPerformance:", systemPerformance)
	// fmt.Println("risk:", risk)
	// builder.WriteString("<p>" + systemPerformance + "</p>")
	// builder.WriteString("<p>" + risk + "</p>")
	// builder.WriteString("<p>" + nextPlan + "</p>")

	if useLLMProvider {
		// LLMRequestParams 配置
		llmParams := llmProvider.LLMRequestParams{
			APIType:     "kimi",
			BaseURL:     "https://api.moonshot.cn/v1/chat",
			APIKey:      "sk-UyI3Y5zBNDNuyZ83ql6QIrQwLfwO2GYgh0s13hTBY8Fcn5ae", // 请替换为实际的 API Key
			Model:       "kimi 8k",
			Proxy:       "", // 如有需要可配置代理
			Timeout:     60, // 请求超时（单位：秒）
			PricingPlan: "free",
			Prompt:      "", // 初始为空，后续会动态设置
		}

		// 初始化 LLMProvider，设置缓存 TTL 为 5 分钟，token 价格为 0.02 美元/千个 token
		cacheTTL := 5 * time.Minute
		tokenPrice := 0.02
		llmProviderInstance := llmProvider.NewLLMProvider(llmParams, cacheTTL, tokenPrice)

		// 调用 AnalyzePerformanceAndGetResponse 函数
		AIanalysisContentJson, tokenCost, err := llmProviderInstance.AnalyzePerformanceAndGetResponse(stats, llmParams)
		if err != nil {
			fmt.Printf("调用 LLM API 时发生错误: %v", err)
		}

		// 打印响应数据和 token 花费
		fmt.Printf("LLM 响应:\n%v\n", AIanalysisContentJson)
		fmt.Println("===========================================")
		printFields(AIanalysisContentJson)
		fmt.Println("===========================================")
		fmt.Printf("Token 花费: $%.4f\n", tokenCost)

		// 提取 SystemPerformance 和 Risk 字段
		systemPerformance, risk, nextPlan, err := extractSystemPerformanceAndRisk(AIanalysisContentJson)
		if err != nil {
			fmt.Printf("AI分析数据时发生错误: %v\n，将使用默认分析", err)
		}

		builder.WriteString("<p>&nbsp;&nbsp;&nbsp;&nbsp;" + systemPerformance + "</p>")
		builder.WriteString("<p>&nbsp;&nbsp;&nbsp;&nbsp;" + risk + "</p>")
		builder.WriteString("<p>&nbsp;&nbsp;&nbsp;&nbsp;" + nextPlan + "</p>")
	} else {
		analysisContent := generateDefaultAnalysis(stats)
		builder.WriteString("<p>&nbsp;&nbsp;&nbsp;&nbsp;" + analysisContent + "</p>")
	}

	builder.WriteString("</section>")

	builder.WriteString("<section class='reference-standards concept-card'>")
	builder.WriteString("<h2><span class='reference-icon'>📘</span>参考标准</h2>")
	builder.WriteString("<p>参考标准：高频接口平均响应时应小于 1 秒，普通接口平均响应时间应低于 2.5 秒，请求成功率应大于 99%。</p>")
	builder.WriteString("</section>")

	builder.WriteString("<section class='reference-standards'>")
	builder.WriteString("<h3>参考概念</h3>")

	// 增加概念的外观样式，使其不那么密集
	builder.WriteString("<div class='concept-card'><p><strong>TPS (Transactions Per Second)</strong>：指每秒钟能够处理的事务数。事务通常指一个完整的请求-响应周期，TPS 越高，说明系统的处理能力越强。常用于衡量系统的吞吐量。</p></div>")

	builder.WriteString("<div class='concept-card'><p><strong>QPS (Queries Per Second)</strong>：指每秒钟能够处理的查询数。QPS 更侧重于查询操作的性能，通常用于数据库或搜索引擎的性能测试。</p></div>")

	builder.WriteString("<div class='concept-card'><p><strong>平均响应时间 (Average Response Time)</strong>：指系统处理一个请求所需的平均时间。通常以毫秒为单位，响应时间越低，说明系统的性能越好。</p></div>")

	builder.WriteString("<div class='concept-card'><p><strong>最大响应时间 (Max Response Time)</strong>：指系统处理请求时所出现的最长响应时间，通常用于衡量系统在高负载下的稳定性。</p></div>")

	builder.WriteString("<div class='concept-card'><p><strong>最小响应时间 (Min Response Time)</strong>：指系统处理请求时所出现的最短响应时间。</p></div>")

	builder.WriteString("<div class='concept-card'><p><strong>上行流量 (Outbound Traffic)</strong>：指从系统发送到客户端或其他服务器的数据量。通常与客户端发送请求的数据量有关。</p></div>")

	builder.WriteString("<div class='concept-card'><p><strong>下行流量 (Inbound Traffic)</strong>：指从客户端或其他服务器接收的数据量。通常与系统返回响应的数据量有关。</p></div>")

	builder.WriteString("<div class='concept-card'><p><strong>请求成功率 (Success Rate)</strong>：指成功处理的请求占总请求数的比例，通常以百分比表示。成功率越高，说明系统的稳定性越好。</p></div>")

	builder.WriteString("<div class='concept-card'><p><strong>吞吐量 (Throughput)</strong>：指系统单位时间内处理的请求或数据量。吞吐量高意味着系统的处理能力强。</p></div>")

	builder.WriteString("<div class='concept-card'><p><strong>并发数 (Concurrency)</strong>：指系统同时处理的请求数。高并发场景下，系统需要处理大量的同时请求，测试并发数可以评估系统的承载能力。</p></div>")

	builder.WriteString("<div class='concept-card'><p><strong>响应时间分布 (Response Time Distribution)</strong>：指系统处理请求时响应时间的分布情况，通常会显示请求的响应时间在一定范围内的比例，用于衡量系统的稳定性。</p></div>")

	builder.WriteString("<div class='concept-card'><p><strong>稳定性 (Stability)</strong>：指系统在持续负载下的表现能力。稳定性测试通常用于验证系统是否能够在长时间高负载的情况下正常工作。</p></div>")

	builder.WriteString("</section>")

	// 结束HTML
	builder.WriteString("</div>")                                   // container
	builder.WriteString("<script src='static/script.js'></script>") // 引入新的 JavaScript 文件
	builder.WriteString("</body></html>")

	// 返回生成的HTML内容
	return builder.String()
}

// generateCSS 生成默认的CSS样式
func generateCSS() string {
	return `
/* General Reset */
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: 'Arial', sans-serif;
    background: #f0f4f8;  /* 淡灰蓝色背景 */
    color: #333;
    line-height: 1.6;
    padding: 20px;
}

/* Container */
.container {
    width: 100%;
    max-width: 1200px;
    margin: 0 auto;
    background-color: #fff;
    border-radius: 12px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);  /* 卡片阴影 */
    padding: 20px;
}

header {
    text-align: center;
    margin-bottom: 30px;
}

h1 {
    font-size: 36px;
    color: #4b6cb7;  /* 亮蓝色 */
    text-transform: uppercase;
    font-weight: 700;
}

/* Section Title */
h2 {
    margin-top: 30px;
    color: #4b6cb7;  /* 亮蓝色 */
    font-size: 24px;
    font-weight: 600;
}
h3 {
    margin-top: 30px;
    font-size: 22px;
    font-weight: 500;
	text-align: center;  /* 让文字居中对齐 */
}

/* Table Styling */
table {
    width: 100%;
    border-collapse: collapse;
    margin-top: 20px;
    border-radius: 8px;
    overflow: hidden;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

table th, table td {
    padding: 12px;
    text-align: left;
    font-size: 16px;
}

table th {
    background: linear-gradient(145deg, #4b6cb7, #9e7dff); /* 渐变背景 */
    color: white;
}

table td {
    background-color: #f9f9f9;
    border-bottom: 1px solid #e1e1e1;
}

/* Charts Section */
.charts {
    // margin-top: 30px;
	margin-top: 50px !important; /* 强制设置与上方元素的距离 */
	width: 100%;
    height: 100%;
    border: none;
}

.tps-chart {
    width: 100%;    /* 使iframe自适应容器宽度 */
    height: 550px;  /* 设置默认高度 */
    background: #fff;
    border: 2px solid #4b6cb7; /* 亮蓝色边框 */
    border-radius: 12px;  /* 圆角边框 */
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1); /* 添加阴影效果 */
    display: block; /* 让iframe成为块级元素，便于控制 */
    margin-left: auto;
    margin-right: auto;
    overflow: hidden;  /* 禁止iframe本身滚动 */
}

/* iframe内的滚动条 */
.tps-chart::-webkit-scrollbar {
    width: 4px;
    height: 4px;  /* 水平方向上的滚动条 */
}

.tps-chart::-webkit-scrollbar-thumb {
    background: #4b6cb7;  /* 滚动条颜色 */
    border-radius: 10px;
}

.tps-chart::-webkit-scrollbar-track {
    background: #f0f4f8;  /* 滚动条轨道背景 */
    border-radius: 10px;
}

.reference-standards {
    padding: 20px;
    margin: 20px 0;
    background-color: #f9f9f9;
    border-left: 5px solid #28a745;
    border-radius: 5px;
    font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
}
.reference-standards h2 {
    color: #28a745;
    font-size: 1.5em;
    margin-bottom: 10px;
}
.reference-standards p {
    line-height: 1.8;
    font-size: 1.1em;
    color: #333;
}
.reference-icon {
    font-size: 1.5em;
    color: #28a745;
    margin-right: 10px;
}

.concept-card {
    background-color: #f5f5f5; /* 浅灰色背景 */
    border-radius: 8px;
    padding: 15px;
    margin-bottom: 15px;
    color: #6c757d; /* 浅灰色字体 */
    box-shadow: 0 2px 4px rgba(0,0,0,0.1); /* 添加阴影效果 */
    transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.concept-card:hover {
    transform: translateY(-5px); /* 悬浮时上移 */
    box-shadow: 0 4px 8px rgba(0,0,0,0.2); /* 增强阴影效果 */
}

.concept-card p {
    margin: 0;
    font-size: 14px;
}

.concept-card strong {
    color: #333; /* 加粗的文字颜色 */
}
.section-icon {
	font-size: 1.5em;
	margin-right: 10px;
}
.report-summary h2 .section-icon { color: #17a2b8; } /* 测试概览 */
.test-statistics h2 .section-icon { color: #ffc107; } /* 测试统计数据 */
.charts h2 .section-icon { color: #007bff; } /* 视图展示 */
/* Analysis Section */
.analysis {
    margin-top: 30px;
    background-color: #f9f9f9;
    padding: 20px;
    border-radius: 10px;
	border-left: 5px solid #007BFF;
    box-shadow: 0 4px 15px rgba(0, 0, 0, 0.05);
    font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
}

.analysis h2 {
    color: #007BFF;
    font-size: 1.5em;
    margin-bottom: 10px;
}

.analysis p {
    font-size: 18px;
    color: #666;
}

.analysis-icon {
	font-size: 1.5em;
	color: #007BFF;
	margin-right: 10px;
}

/* Responsive Design */
@media (max-width: 768px) {
    .container {
        padding: 10px;
    }

    h1 {
        font-size: 28px;
    }

    h2 {
        font-size: 20px;
    }

    table th, table td {
        font-size: 14px;
    }

    .tps-chart {
        height: 500px;  /* 在小屏幕上适当调整iframe的高度 */
    }
}
`
}

// generateScript 生成 static/script.js 的内容
func generateScript() string {
	return `
document.addEventListener("DOMContentLoaded", function() {
    const iframe = document.querySelector('.tps-chart');
    
    function adjustIframeHeight() {
        const iframeDocument = iframe.contentDocument || iframe.contentWindow.document;
        const body = iframeDocument.body;
        const html = iframeDocument.documentElement;

        // 获取整个文档的高度
        const docHeight = Math.max(
            body.scrollHeight, body.offsetHeight,
            html.clientHeight, html.scrollHeight, html.offsetHeight
        );
        
        // 设置iframe的高度
        iframe.style.height = docHeight + 'px';
    }

    // 初始化时调整iframe高度
    adjustIframeHeight();

    // 监听iframe内容变化，调整高度
    const observer = new MutationObserver(adjustIframeHeight);
    observer.observe(iframe.contentDocument || iframe.contentWindow.document, {
        childList: true,
        subtree: true,
        attributes: true
    });
});
`
}
