package result

import (
	"fmt"
	"strings"

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

// GenerateHTMLReport 生成性能测试报告的HTML
func GenerateHTMLReport(stats map[string]interface{}, title ...string) string {
	var builder strings.Builder

	// 可选的参数，使用默认值
	pageTitle := "性能测试报告"                             // 默认标题
	logoPath := ""                                    // 默认无logo
	analysisContent := generateDefaultAnalysis(stats) // 根据测试数据自动生成的默认分析内容

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
	builder.WriteString("<h2>测试概览</h2>")
	builder.WriteString("<table>")
	builder.WriteString("<tr><th>开始时间</th><td>" + time.Unix(stats["AvgTpsStartTime"].(int64), 0).Format("2006-01-02 15:04:05") + "</td></tr>")
	builder.WriteString("<tr><th>结束时间</th><td>" + time.Unix(stats["AvgTpsEndTime"].(int64), 0).Format("2006-01-02 15:04:05") + "</td></tr>")
	builder.WriteString("</table>")
	builder.WriteString("</section>")

	// 测试统计数据部分
	builder.WriteString("<section class='test-statistics'>")
	builder.WriteString("<h2>测试统计数据</h2>")
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

		// 对 AvgResponseTime, MaxResponseTime, MinResponseTime 字段特殊处理，转换为毫秒并保留两位小数
		if key == "AvgResponseTime" || key == "MaxResponseTime" || key == "MinResponseTime" {
			value = fmt.Sprintf("%.2f", float64(value.(time.Duration))/float64(time.Millisecond))
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
	builder.WriteString("<h2>视图展示</h2>")

	// 添加TPS趋势图部分
	builder.WriteString("<div class='chart'><h3>TPS趋势图</h3>")
	// 使用iframe标签来嵌入tps_chart.html，并应用优化后的样式
	builder.WriteString("<iframe class='tps-chart' src='static/tps_chart.html' frameborder='0'></iframe>")
	builder.WriteString("</div>")
	builder.WriteString("</section>")

	// 分析部分
	builder.WriteString("<section class='analysis'>")
	builder.WriteString("<h2>分析</h2>")
	builder.WriteString("<p>" + analysisContent + "</p>")
	builder.WriteString("</section>")

	// 参考标准部分
	builder.WriteString("<section class='reference-standards'>")
	builder.WriteString("<h2>参考标准</h2>")
	builder.WriteString("<p>参考标准：高频接口平均响应时应小于 1 秒，普通接口平均响应时间应低于 2.5 秒，请求成功率应大于 99%。</p>")
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

/* Analysis Section */
.analysis {
    margin-top: 30px;
    background-color: #f9f9f9;
    padding: 20px;
    border-radius: 10px;
    box-shadow: 0 4px 15px rgba(0, 0, 0, 0.05);
}

.analysis p {
    font-size: 18px;
    color: #666;
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
