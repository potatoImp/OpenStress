package llmProvider

// OllamaGenerator Ollama API请求生成器
type OllamaGenerator struct {
	*baseRequestGenerator
}

func NewOllamaGenerator() *OllamaGenerator {
	return &OllamaGenerator{
		baseRequestGenerator: newBaseRequestGenerator(),
	}
}

func (g *OllamaGenerator) GenerateRequest(prompt string, messages []Message, config LLMRequestParams) (map[string]interface{}, error) {
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

// LlamaGenerator Llama API请求生成器
type LlamaGenerator struct {
	*baseRequestGenerator
}

func NewLlamaGenerator() *LlamaGenerator {
	return &LlamaGenerator{
		baseRequestGenerator: newBaseRequestGenerator(),
	}
}

func (g *LlamaGenerator) GenerateRequest(prompt string, messages []Message, config LLMRequestParams) (map[string]interface{}, error) {
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

// MistralGenerator Mistral AI请求生成器
type MistralGenerator struct {
	*baseRequestGenerator
}

func NewMistralGenerator() *MistralGenerator {
	return &MistralGenerator{
		baseRequestGenerator: newBaseRequestGenerator(),
	}
}

func (g *MistralGenerator) GenerateRequest(prompt string, messages []Message, config LLMRequestParams) (map[string]interface{}, error) {
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
