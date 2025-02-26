package llmProvider

import "fmt"

// OpenAIGenerator OpenAI API请求生成器
type OpenAIGenerator struct {
	*baseRequestGenerator
}

func NewOpenAIGenerator() *OpenAIGenerator {
	return &OpenAIGenerator{
		baseRequestGenerator: newBaseRequestGenerator(),
	}
}

func (g *OpenAIGenerator) GenerateRequest(prompt string, messages []Message, config LLMRequestParams) (map[string]interface{}, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	requestData := map[string]interface{}{
		"model":       config.Model,
		"messages":    formatMessages(messages, prompt),
		"temperature": g.defaultTemp,
		"max_tokens":  g.maxTokens,
	}

	return requestData, nil
}

// AzureGenerator Azure OpenAI请求生成器
type AzureGenerator struct {
	*baseRequestGenerator
}

func NewAzureGenerator() *AzureGenerator {
	return &AzureGenerator{
		baseRequestGenerator: newBaseRequestGenerator(),
	}
}

func (g *AzureGenerator) GenerateRequest(prompt string, messages []Message, config LLMRequestParams) (map[string]interface{}, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	requestData := map[string]interface{}{
		"messages":    formatMessages(messages, prompt),
		"temperature": g.defaultTemp,
		"max_tokens":  g.maxTokens,
		"stream":      false,
	}

	return requestData, nil
}

// GroqGenerator Groq API请求生成器
type GroqGenerator struct {
	*baseRequestGenerator
}

func NewGroqGenerator() *GroqGenerator {
	return &GroqGenerator{
		baseRequestGenerator: newBaseRequestGenerator(),
	}
}

func (g *GroqGenerator) GenerateRequest(prompt string, messages []Message, config LLMRequestParams) (map[string]interface{}, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	requestData := map[string]interface{}{
		"model":       config.Model,
		"messages":    formatMessages(messages, prompt),
		"temperature": g.defaultTemp,
	}

	return requestData, nil
}

// ClaudeGenerator Anthropic Claude请求生成器
type ClaudeGenerator struct {
	*baseRequestGenerator
}

func NewClaudeGenerator() *ClaudeGenerator {
	return &ClaudeGenerator{
		baseRequestGenerator: newBaseRequestGenerator(),
	}
}

func (g *ClaudeGenerator) GenerateRequest(prompt string, messages []Message, config LLMRequestParams) (map[string]interface{}, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	requestData := map[string]interface{}{
		"model":       config.Model,
		"messages":    formatMessages(messages, prompt),
		"temperature": g.defaultTemp,
		"max_tokens":  g.maxTokens,
	}

	return requestData, nil
}

// GeminiGenerator Google Gemini请求生成器
type GeminiGenerator struct {
	*baseRequestGenerator
}

func NewGeminiGenerator() *GeminiGenerator {
	return &GeminiGenerator{
		baseRequestGenerator: newBaseRequestGenerator(),
	}
}

func (g *GeminiGenerator) GenerateRequest(prompt string, messages []Message, config LLMRequestParams) (map[string]interface{}, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	requestData := map[string]interface{}{
		"model":       config.Model,
		"contents":    formatMessagesForGemini(messages, prompt),
		"temperature": g.defaultTemp,
	}

	return requestData, nil
}

// PaLMGenerator Google PaLM请求生成器
type PaLMGenerator struct {
	*baseRequestGenerator
}

func NewPaLMGenerator() *PaLMGenerator {
	return &PaLMGenerator{
		baseRequestGenerator: newBaseRequestGenerator(),
	}
}

func (g *PaLMGenerator) GenerateRequest(prompt string, messages []Message, config LLMRequestParams) (map[string]interface{}, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	requestData := map[string]interface{}{
		"model":       config.Model,
		"prompt":      formatMessagesForPaLM(messages, prompt),
		"temperature": g.defaultTemp,
	}

	return requestData, nil
}

// DeepSeekGenerator DeepSeek API请求生成器
type DeepSeekGenerator struct {
	*baseRequestGenerator
}

func NewDeepSeekGenerator() *DeepSeekGenerator {
	fmt.Println("DeepSeekGenerator来了")
	return &DeepSeekGenerator{
		baseRequestGenerator: newBaseRequestGenerator(),
	}
}

func (g *DeepSeekGenerator) GenerateRequest(prompt string, messages []Message, config LLMRequestParams) (map[string]interface{}, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	requestData := map[string]interface{}{
		"model":       config.Model,
		"messages":    formatMessages(messages, prompt),
		"temperature": g.defaultTemp,
	}

	return requestData, nil
}

// 辅助函数：格式化消息
func formatMessages(messages []Message, prompt string) []map[string]interface{} {
	formattedMessages := make([]map[string]interface{}, len(messages))
	for i, msg := range messages {
		formattedMessages[i] = map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}

	// 添加当前提示作为最后一条用户消息
	if prompt != "" {
		formattedMessages = append(formattedMessages, map[string]interface{}{
			"role":    "user",
			"content": prompt,
		})
	}

	return formattedMessages
}

// 为Gemini格式化消息
func formatMessagesForGemini(messages []Message, prompt string) []map[string]interface{} {
	formattedMessages := make([]map[string]interface{}, len(messages))
	for i, msg := range messages {
		formattedMessages[i] = map[string]interface{}{
			"role": msg.Role,
			"parts": []map[string]interface{}{
				{"text": msg.Content},
			},
		}
	}

	// 添加当前提示
	if prompt != "" {
		formattedMessages = append(formattedMessages, map[string]interface{}{
			"role": "user",
			"parts": []map[string]interface{}{
				{"text": prompt},
			},
		})
	}

	return formattedMessages
}

// 为PaLM格式化消息
func formatMessagesForPaLM(messages []Message, prompt string) string {
	var result string
	for _, msg := range messages {
		prefix := ""
		switch msg.Role {
		case "system":
			prefix = "System: "
		case "user":
			prefix = "User: "
		case "assistant":
			prefix = "Assistant: "
		}
		result += prefix + msg.Content + "\n"
	}

	// 添加当前提示
	if prompt != "" {
		result += "User: " + prompt
	}

	return result
}
