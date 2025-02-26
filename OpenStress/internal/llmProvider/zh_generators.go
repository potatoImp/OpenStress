package llmProvider

// KimiGenerator Moonshot AI请求生成器
type KimiGenerator struct {
	*baseRequestGenerator
}

func NewKimiGenerator() *KimiGenerator {
	return &KimiGenerator{
		baseRequestGenerator: newBaseRequestGenerator(),
	}
}

func (g *KimiGenerator) GenerateRequest(prompt string, messages []Message, config LLMRequestParams) (map[string]interface{}, error) {
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

// QianWenGenerator 阿里通义千问请求生成器
type QianWenGenerator struct {
	*baseRequestGenerator
}

func NewQianWenGenerator() *QianWenGenerator {
	return &QianWenGenerator{
		baseRequestGenerator: newBaseRequestGenerator(),
	}
}

func (g *QianWenGenerator) GenerateRequest(prompt string, messages []Message, config LLMRequestParams) (map[string]interface{}, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	requestData := map[string]interface{}{
		"model": config.Model,
		"input": map[string]interface{}{
			"messages": formatMessages(messages, prompt),
		},
		"parameters": map[string]interface{}{
			"temperature": g.defaultTemp,
			"max_tokens":  g.maxTokens,
		},
	}

	return requestData, nil
}

// ErnieGenerator 百度文心一言请求生成器
type ErnieGenerator struct {
	*baseRequestGenerator
}

func NewErnieGenerator() *ErnieGenerator {
	return &ErnieGenerator{
		baseRequestGenerator: newBaseRequestGenerator(),
	}
}

func (g *ErnieGenerator) GenerateRequest(prompt string, messages []Message, config LLMRequestParams) (map[string]interface{}, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	requestData := map[string]interface{}{
		"messages":    formatMessages(messages, prompt),
		"temperature": g.defaultTemp,
		"top_p":       0.8,
	}

	return requestData, nil
}

// SparkGenerator 讯飞星火请求生成器
type SparkGenerator struct {
	*baseRequestGenerator
}

func NewSparkGenerator() *SparkGenerator {
	return &SparkGenerator{
		baseRequestGenerator: newBaseRequestGenerator(),
	}
}

func (g *SparkGenerator) GenerateRequest(prompt string, messages []Message, config LLMRequestParams) (map[string]interface{}, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	requestData := map[string]interface{}{
		"header": map[string]interface{}{
			"app_id": config.AccessKey,
		},
		"parameter": map[string]interface{}{
			"chat": map[string]interface{}{
				"temperature": g.defaultTemp,
				"max_tokens":  g.maxTokens,
			},
		},
		"payload": map[string]interface{}{
			"message": map[string]interface{}{
				"text": formatMessagesForSpark(messages, prompt),
			},
		},
	}

	return requestData, nil
}

// ZhiPuGenerator 智谱AI请求生成器
type ZhiPuGenerator struct {
	*baseRequestGenerator
}

func NewZhiPuGenerator() *ZhiPuGenerator {
	return &ZhiPuGenerator{
		baseRequestGenerator: newBaseRequestGenerator(),
	}
}

func (g *ZhiPuGenerator) GenerateRequest(prompt string, messages []Message, config LLMRequestParams) (map[string]interface{}, error) {
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

// MiniMaxGenerator MiniMax请求生成器
type MiniMaxGenerator struct {
	*baseRequestGenerator
}

func NewMiniMaxGenerator() *MiniMaxGenerator {
	return &MiniMaxGenerator{
		baseRequestGenerator: newBaseRequestGenerator(),
	}
}

func (g *MiniMaxGenerator) GenerateRequest(prompt string, messages []Message, config LLMRequestParams) (map[string]interface{}, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	requestData := map[string]interface{}{
		"model":       config.Model,
		"messages":    formatMessages(messages, prompt),
		"temperature": g.defaultTemp,
		"max_tokens":  g.maxTokens,
		"stream":      false,
	}

	return requestData, nil
}

// BaichuanGenerator 百川智能请求生成器
type BaichuanGenerator struct {
	*baseRequestGenerator
}

func NewBaichuanGenerator() *BaichuanGenerator {
	return &BaichuanGenerator{
		baseRequestGenerator: newBaseRequestGenerator(),
	}
}

func (g *BaichuanGenerator) GenerateRequest(prompt string, messages []Message, config LLMRequestParams) (map[string]interface{}, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	requestData := map[string]interface{}{
		"model":            config.Model,
		"messages":         formatMessages(messages, prompt),
		"temperature":      g.defaultTemp,
		"top_p":            0.7,
		"with_search_info": false,
	}

	return requestData, nil
}

// 为Spark格式化消息
func formatMessagesForSpark(messages []Message, prompt string) string {
	var result string
	for _, msg := range messages {
		result += msg.Content + "\n"
	}

	if prompt != "" {
		result += prompt
	}

	return result
}
