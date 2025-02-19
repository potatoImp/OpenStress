package result

import (
	"encoding/json"
	"fmt"
	"strings"
)

func extractPerformanceAnalysis(AIType string, data map[string]interface{}) (string, string, string, string, error) {
	if AIType != "ollama" && AIType != "openai" {
		systemPerformance, risk, nextPlan, err := extractSystemPerformanceAndRisk(data)
		return "", systemPerformance, risk, nextPlan, err
	}
	thinkContent, systemPerformance, risk, nextPlan, err := extractThinkSystemPerformanceAndRisk(data)
	return thinkContent, systemPerformance, risk, nextPlan, err
}

// 提取 SystemPerformance 和 Risk 字段的函数
func extractThinkSystemPerformanceAndRisk(data map[string]interface{}) (string, string, string, string, error) {
	// 1. 获取 choices 中的第一个元素
	choices, ok := data["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", "", "", "", fmt.Errorf("无法获取 choices 数据")
	}

	// 2. 获取第一个元素中的 message.content 字段
	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", "", "", "", fmt.Errorf("无法获取 choice 数据")
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return "", "", "", "", fmt.Errorf("无法获取 message 数据")
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", "", "", "", fmt.Errorf("无法获取 content 字段")
	}

	// 3. 提取 <think> 标签中的内容
	thinkContent := ""
	thinkStart := strings.Index(content, "<think>")
	thinkEnd := strings.Index(content, "</think>")
	if thinkStart != -1 && thinkEnd != -1 && thinkEnd > thinkStart {
		thinkContent = content[thinkStart+len("<think>") : thinkEnd]
	}
	fmt.Println(";;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;", thinkContent)

	// 4. 去掉 content 中的 ```json 和 ```, 清理字符串
	jsonStart := strings.Index(content, "```json")
	jsonEnd := strings.LastIndex(content, "```")
	if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
		content = content[jsonStart+len("```json") : jsonEnd]
	}

	// 5. 将 content 字段中的 JSON 字符串解析为新的 map
	var analysisData map[string]interface{}
	err := json.Unmarshal([]byte(content), &analysisData)
	if err != nil {
		return thinkContent, "", "", "", fmt.Errorf("无法解析 content 中的 JSON 数据: %w", err)
	}

	// 6. 提取 SystemPerformance 和 Risk 字段
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

	return thinkContent, systemPerformance, risk, nextPlan, nil
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

func parseGeneralLLMJson(data map[string]interface{}) (string, string, string, error) {
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
