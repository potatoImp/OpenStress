package configs

import (
	"fmt"
)

// 全局配置变量
var llmInitConfig *LLMConfig

// func init() {
// 	// 程序启动初始化
// 	llmInitConfig, llmConfigErr := ReadLLMConfig()
// 	if llmConfigErr != nil {
// 		fmt.Println("Failed to load config:", llmConfigErr)
// 	}
// 	fmt.Println("Config loaded successfully in init")
// 	// 打印配置内容以验证
// 	fmt.Printf("Loaded LLM Config: %+v\n", llmInitConfig)

// }

func Initialize() {
	// 程序启动初始化
	llmInitConfig, llmConfigErr := ReadLLMConfig()
	if llmConfigErr != nil {
		fmt.Println("Failed to load config:", llmConfigErr)
	}
	fmt.Println("Config loaded successfully in init")
	// 打印配置内容以验证
	fmt.Printf("Loaded LLM Config: %+v\n", llmInitConfig)
}

// Getter 函数
func GetAPIKey() string {
	if llmInitConfig != nil {
		return llmInitConfig.LLM.APIKey
	}
	return ""
}

func GetAPIType() string {
	if llmInitConfig != nil {
		return string(llmInitConfig.LLM.APIType)
	}
	return "" // 默认返回空字符串
}

func GetBaseURL() string {
	if llmInitConfig != nil {
		return llmInitConfig.LLM.BaseURL
	}
	return ""
}

func GetMaxToken() int {
	if llmInitConfig != nil {
		return llmInitConfig.LLM.MaxToken
	}
	return 0
}

func GetTemperature() float64 {
	if llmInitConfig != nil {
		return llmInitConfig.LLM.Temperature
	}
	return 0.0
}

func GetTopP() float64 {
	if llmInitConfig != nil {
		return llmInitConfig.LLM.TopP
	}
	return 0.0
}

func GetTopK() int {
	if llmInitConfig != nil {
		return llmInitConfig.LLM.TopK
	}
	return 0
}

func GetRepetitionPenalty() float64 {
	if llmInitConfig != nil {
		return llmInitConfig.LLM.RepetitionPenalty
	}
	return 0.0
}

func GetPresencePenalty() float64 {
	if llmInitConfig != nil {
		return llmInitConfig.LLM.PresencePenalty
	}
	return 0.0
}

func GetFrequencyPenalty() float64 {
	if llmInitConfig != nil {
		return llmInitConfig.LLM.FrequencyPenalty
	}
	return 0.0
}

func GetStream() bool {
	if llmInitConfig != nil {
		return llmInitConfig.LLM.Stream
	}
	return false
}

func GetTimeout() int {
	if llmInitConfig != nil {
		return llmInitConfig.LLM.Timeout
	}
	return 0
}

func GetRegionName() string {
	if llmInitConfig != nil {
		return llmInitConfig.LLM.RegionName
	}
	return ""
}

func GetCalcUsage() bool {
	if llmInitConfig != nil {
		return llmInitConfig.LLM.CalcUsage
	}
	return false
}

func GetUseSystemPrompt() bool {
	if llmInitConfig != nil {
		return llmInitConfig.LLM.UseSystemPrompt
	}
	return false
}
