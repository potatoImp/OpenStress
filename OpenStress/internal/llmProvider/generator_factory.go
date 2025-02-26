package llmProvider

import (
	"fmt"
	"sync"
)

// RequestGeneratorFactory 请求生成器工厂
type RequestGeneratorFactory struct {
	generators map[APIType]RequestGenerator
	mu         sync.RWMutex
}

// NewRequestGeneratorFactory 创建新的请求生成器工厂
func NewRequestGeneratorFactory() *RequestGeneratorFactory {
	f := &RequestGeneratorFactory{
		generators: make(map[APIType]RequestGenerator),
	}
	f.registerGenerators()
	return f
}

// registerGenerators 注册所有支持的生成器
func (f *RequestGeneratorFactory) registerGenerators() {
	// 主流商业 LLM 服务
	f.generators[APITypeOpenAI] = NewOpenAIGenerator()
	f.generators[APITypeAzure] = NewAzureGenerator()
	f.generators[APITypeGroq] = NewGroqGenerator()
	f.generators[APITypeClaude] = NewClaudeGenerator()
	f.generators[APITypeGemini] = NewGeminiGenerator()
	f.generators[APITypePaLM] = NewPaLMGenerator()
	f.generators[APITypeDeepSeek] = NewDeepSeekGenerator()

	// 开源模型服务
	f.generators[APITypeOllama] = NewOllamaGenerator()
	f.generators[APITypeLlama] = NewLlamaGenerator()
	f.generators[APITypeMistral] = NewMistralGenerator()

	// 国内 LLM 服务
	f.generators[APITypeKimi] = NewKimiGenerator()
	f.generators[APITypeQianWen] = NewQianWenGenerator()
	f.generators[APITypeErnie] = NewErnieGenerator()
	f.generators[APITypeSpark] = NewSparkGenerator()
	f.generators[APITypeZhiPu] = NewZhiPuGenerator()
	f.generators[APITypeMiniMax] = NewMiniMaxGenerator()
	f.generators[APITypeBaichuan] = NewBaichuanGenerator()
}

// GetGenerator 获取指定类型的请求生成器
func (f *RequestGeneratorFactory) GetGenerator(apiType APIType) (RequestGenerator, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	generator, ok := f.generators[apiType]
	if !ok {
		return nil, fmt.Errorf("不支持的 API 类型: %s", apiType)
	}
	return generator, nil
}