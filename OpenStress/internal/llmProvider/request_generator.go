package llmProvider

import (
	"fmt"
	"time"
)

// Message 表示一条对话消息
type Message struct {
	Role    string    `json:"role"`    // system, user, assistant
	Content string    `json:"content"` // 消息内容
	Time    time.Time `json:"time"`    // 消息时间戳
}

// APIType 定义支持的 API 类型
type APIType string

const (
	// 主流商业 LLM 服务
	APITypeOpenAI   APIType = "openai"   // OpenAI API
	APITypeAzure    APIType = "azure"    // Azure OpenAI
	APITypeGroq     APIType = "groq"     // Groq API
	APITypeClaude   APIType = "claude"   // Anthropic Claude
	APITypeGemini   APIType = "gemini"   // Google Gemini
	APITypePaLM     APIType = "palm"     // Google PaLM
	APITypeDeepSeek APIType = "deepseek" // DeepSeek AI

	// 开源模型服务
	APITypeOllama  APIType = "ollama"  // Ollama
	APITypeLlama   APIType = "llama"   // Llama
	APITypeMistral APIType = "mistral" // Mistral AI

	// 国内 LLM 服务
	APITypeKimi     APIType = "kimi"     // Moonshot AI
	APITypeQianWen  APIType = "qianwen"  // 阿里通义千问
	APITypeErnie    APIType = "ernie"    // 百度文心一言
	APITypeSpark    APIType = "spark"    // 讯飞星火
	APITypeZhiPu    APIType = "zhipu"    // 智谱 AI
	APITypeMiniMax  APIType = "minimax"  // MiniMax
	APITypeBaichuan APIType = "baichuan" // 百川智能
)

// RequestGenerator 定义请求数据生成器接口
type RequestGenerator interface {
	GenerateRequest(prompt string, messages []Message, config LLMRequestParams) (map[string]interface{}, error)
}

// 基础请求生成器
type baseRequestGenerator struct {
	defaultTemp float64
	maxTokens   int
}

// RequestOptions 定义请求选项
type RequestOptions struct {
	Temperature      float64
	MaxTokens        int
	TopP             float64
	FrequencyPenalty float64
	PresencePenalty  float64
	Stop             []string
}

func newBaseRequestGenerator() *baseRequestGenerator {
	fmt.Println("newBaseRequestGenerator来了")
	fmt.Println(&baseRequestGenerator{
		defaultTemp: 0.3,
		maxTokens:   10000,
	})
	return &baseRequestGenerator{
		defaultTemp: 0.3,
		maxTokens:   10000,
	}
}

// ValidateConfig 验证配置的有效性
func ValidateConfig(config LLMRequestParams) error {
	if config.APIType == "" {
		return fmt.Errorf("API类型不能为空")
	}
	if config.BaseURL == "" {
		return fmt.Errorf("BaseURL不能为空")
	}
	if config.Model == "" {
		return fmt.Errorf("模型名称不能为空")
	}
	if config.Timeout <= 0 {
		return fmt.Errorf("超时时间必须大于0")
	}
	return nil
}

// CreateSystemMessage 创建系统消息
func CreateSystemMessage(content string) Message {
	return Message{
		Role:    "system",
		Content: content,
		Time:    time.Now(),
	}
}

// CreateUserMessage 创建用户消息
func CreateUserMessage(content string) Message {
	return Message{
		Role:    "user",
		Content: content,
		Time:    time.Now(),
	}
}

// CreateAssistantMessage 创建助手消息
func CreateAssistantMessage(content string) Message {
	return Message{
		Role:    "assistant",
		Content: content,
		Time:    time.Now(),
	}
}
