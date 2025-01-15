package llmProvider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type PerformanceStats struct {
	TotalRequests               int64   `json:"TotalRequests"`
	SuccessCount                int64   `json:"SuccessCount"`
	FailureCount                int64   `json:"FailureCount"`
	SuccessRate                 float64 `json:"SuccessRate"`
	AvgResponseTime             int64   `json:"AvgResponseTime"`
	MaxResponseTime             int64   `json:"MaxResponseTime"`
	MinResponseTime             int64   `json:"MinResponseTime"`
	TotalRunTime                int64   `json:"TotalRunTime"`
	TPS                         float64 `json:"TPS"`
	SentDataPerSec              string  `json:"SentDataPerSec"`
	ReceivedDataPerSec          string  `json:"ReceivedDataPerSec"`
	TotalSentData               string  `json:"TotalSentData"`
	TotalReceivedData           string  `json:"TotalReceivedData"`
	AvgSentTrafficValues        []int64 `json:"AvgSentTrafficValues"`
	AvgReceivedTrafficValues    []int64 `json:"AvgReceivedTrafficValues"`
	AvgSuccessSentTrafficValues []int64 `json:"AvgSuccessSentTrafficValues"`
}

type LLMRequestParams struct {
	APIType     string `json:"api_type"`
	BaseURL     string `json:"base_url"`
	APIKey      string `json:"api_key"`
	Model       string `json:"model"`
	Proxy       string `json:"proxy"`
	Timeout     int    `json:"timeout"`
	PricingPlan string `json:"pricing_plan"`
	Prompt      string `json:"prompt"`
}

type LLMProvider struct {
	Config     LLMRequestParams
	cache      sync.Map // 使用sync.Map作为缓存
	cacheTTL   time.Duration
	tokenPrice float64 // 每千个token的费用
}

func NewLLMProvider(config LLMRequestParams, cacheTTL time.Duration, tokenPrice float64) *LLMProvider {
	return &LLMProvider{Config: config, cacheTTL: cacheTTL, tokenPrice: tokenPrice}
}

// 格式化性能数据为 LLM 可解析的文本
func (p *LLMProvider) formatPrompt(stats PerformanceStats) string {
	// 将时间相关字段除以1000000，转换为秒
	avgResponseTimeSec := stats.AvgResponseTime / 1000000
	maxResponseTimeSec := stats.MaxResponseTime / 1000000
	minResponseTimeSec := stats.MinResponseTime / 1000000
	totalRunTimeSec := stats.TotalRunTime / 1000000

	return fmt.Sprintf(`
	   请使用如下 JSON 格式输出你的回复：
 
       {
        "SystemPerformance": "系统性能表现相关分析",
        "Risk": "可能存在的风险",
		"NextPlan": "下一步的计划"
       }
		以下是性能测试的汇总结果数据，请根据这些数据分析当前系统的表现，并指出可能存在的风险，以及下一步的参考测试方向：
		
		总请求数: %d
		成功请求数: %d
		失败请求数: %d
		成功率: %.3f
		平均响应时间: %d 毫秒
		最大响应时间: %d 毫秒
		最小响应时间: %d 毫秒
		总运行时间: %d 毫秒
		每秒事务数(TPS): %.2f
		每秒发送的数据量: %s
		每秒接收的数据量: %s
		总发送的数据量: %s
		总接收的数据量: %s
		
		每秒的平均发送流量值: %v
		每秒的平均接收流量值: %v
		每秒成功请求发送流量的平均值: %v
		`,
		stats.TotalRequests,
		stats.SuccessCount,
		stats.FailureCount,
		stats.SuccessRate,
		avgResponseTimeSec, // 修改后的时间单位为毫秒
		maxResponseTimeSec, // 修改后的时间单位为毫秒
		minResponseTimeSec, // 修改后的时间单位为毫秒
		totalRunTimeSec,    // 修改后的时间单位为毫秒
		stats.TPS,
		stats.SentDataPerSec,
		stats.ReceivedDataPerSec,
		stats.TotalSentData,
		stats.TotalReceivedData,
		stats.AvgSentTrafficValues,
		stats.AvgReceivedTrafficValues,
		stats.AvgSuccessSentTrafficValues,
	)
}

// 生成缓存的键
func (p *LLMProvider) generateCacheKey(prompt string) string {
	return fmt.Sprintf("%x", prompt) // 根据prompt内容生成唯一的缓存键
}

// 检查缓存中是否存在有效的结果
func (p *LLMProvider) getCachedResponse(cacheKey string) (map[string]interface{}, bool) {
	val, ok := p.cache.Load(cacheKey)
	if !ok {
		return nil, false
	}

	cachedData, valid := val.(cachedItem)
	if !valid || time.Now().After(cachedData.expiryTime) {
		// 缓存过期
		p.cache.Delete(cacheKey)
		return nil, false
	}

	return cachedData.response, true
}

// 设置缓存
func (p *LLMProvider) setCache(cacheKey string, response map[string]interface{}) {
	expiryTime := time.Now().Add(p.cacheTTL)
	p.cache.Store(cacheKey, cachedItem{response: response, expiryTime: expiryTime})
}

type cachedItem struct {
	response   map[string]interface{}
	expiryTime time.Time
}

// 设置请求参数
func (p *LLMProvider) generateRequestData(prompt string) (map[string]interface{}, error) {
	// 基础的请求数据结构
	requestData := map[string]interface{}{
		"temperature": 0.3, // 设置默认的温度
	}

	// 根据 APIType 构建不同的请求数据结构
	switch p.Config.APIType {
	case "kimi":
		// 针对 kimi 的特殊请求数据结构
		requestData["model"] = "moonshot-v1-8k"
		requestData["messages"] = []map[string]interface{}{
			{
				"role":    "system",
				"content": "你是一名专业的性能测试专家，由OponStress提供的智能助手，你更擅长中文和英文的对话。你会为用户提供安全，有帮助，准确的回答。同时，你会拒绝一切涉及恐怖主义，种族歧视，黄色暴力等问题的回答。将根据用户的提问给出专业确定的性能分析结论，不回复模糊的结论",
			},
			{
				"role":    "user",
				"content": prompt, // 将传入的 prompt 填充到此处
			},
		}
	default:
		// 其他 API 类型使用默认的请求结构
		requestData["model"] = p.Config.Model
		requestData["prompt"] = prompt
	}

	return requestData, nil
}

// 调用 LLM API 进行性能分析
func (p *LLMProvider) CallLLMAPI(prompt string) (map[string]interface{}, float64, error) {
	// 生成请求数据
	requestData, err := p.generateRequestData(prompt)
	if err != nil {
		return nil, 0, fmt.Errorf("请求数据生成失败: %w", err)
	}

	// 将请求数据编码为 JSON
	requestDataBytes, err := json.Marshal(requestData)
	if err != nil {
		return nil, 0, fmt.Errorf("请求参数编码失败: %w", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", p.Config.BaseURL+"/completions", bytes.NewBuffer(requestDataBytes))
	if err != nil {
		return nil, 0, fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.Config.APIKey)

	// 执行 HTTP 请求
	client := &http.Client{
		Timeout: time.Duration(p.Config.Timeout) * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("发送 HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查返回状态
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("请求失败，HTTP状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, 0, fmt.Errorf("解析 LLM API 响应失败: %w", err)
	}

	// 获取 token 使用量
	usage, ok := response["usage"].(map[string]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("无法解析 token 使用信息")
	}

	totalTokens, ok := usage["total_tokens"].(float64)
	if !ok {
		return nil, 0, fmt.Errorf("无法获取 total_tokens")
	}

	// 计算 token 花费
	tokenCost := totalTokens / 1000 * p.tokenPrice

	return response, tokenCost, nil
}

// 外部调用接口：传入性能数据与 LLMRequestParams，格式化后请求并返回响应
func (p *LLMProvider) AnalyzePerformanceAndGetResponse(stats map[string]interface{}, llmParams LLMRequestParams) (map[string]interface{}, float64, error) {
	// 格式化性能数据为 LLM 可解析的文本
	// 将 stats 转为 PerformanceStats 类型
	performanceStats, err := mapToPerformanceStats(stats)
	if err != nil {
		return nil, 0, fmt.Errorf("无法将 stats 转换为 PerformanceStats: %w", err)
	}

	formattedPrompt := p.formatPrompt(performanceStats)

	// 设置 LLMRequestParams
	p.Config = llmParams

	fmt.Println("格式化后的 prompt:")
	fmt.Println(formattedPrompt)
	// 调用 LLM API 获取响应
	response, tokenCost, err := p.CallLLMAPI(formattedPrompt)
	if err != nil {
		return nil, 0, fmt.Errorf("调用 LLM API 失败: %w", err)
	}

	return response, tokenCost, nil
}

// 辅助函数：将 map[string]interface{} 转换为 PerformanceStats 类型
func mapToPerformanceStats(stats map[string]interface{}) (PerformanceStats, error) {
	var ps PerformanceStats
	data, err := json.Marshal(stats)
	if err != nil {
		return ps, fmt.Errorf("序列化 stats 失败: %w", err)
	}

	err = json.Unmarshal(data, &ps)
	if err != nil {
		return ps, fmt.Errorf("反序列化 stats 失败: %w", err)
	}

	return ps, nil
}
