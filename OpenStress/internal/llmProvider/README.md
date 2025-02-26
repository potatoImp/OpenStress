# LLM Provider 模块说明文档

## 概述
本模块用于处理与各种 LLM (Large Language Model) 服务提供商的交互。通过统一的接口和可扩展的架构，支持多种 LLM API 的集成。

## 添加新的 LLM 提供商
要添加新的 LLM 提供商，需要完成以下步骤：

### 1. 注册 API 类型
在 `request_generator.go` 中的 APIType 常量组中添加新的类型：
```go
const (
    // 已有的类型
    APITypeOpenAI APIType = "openai"
    // 添加新类型
    APITypeNewProvider APIType = "new_provider"
)
```
### 2. 创建生成器实现
在 `request_generator.go` 中创建一个新的生成器实现，实现 `RequestGenerator` 接口：
```go
package llmProvider

type NewProviderGenerator struct {
    *baseRequestGenerator
}

func NewNewProviderGenerator() *NewProviderGenerator {
    return &NewProviderGenerator{
        baseRequestGenerator: newBaseRequestGenerator(),
    }
}

func (g *NewProviderGenerator) GenerateRequest(prompt string, messages []Message, config LLMRequestParams) (map[string]interface{}, error) {
    if err := ValidateConfig(config); err != nil {
        return nil, err
    }

    // 构建该提供商特定的请求格式
    requestData := map[string]interface{}{
        "model": config.Model,
        "temperature": g.defaultTemp,
        "messages": formatMessagesForProvider(messages, prompt),
        // 其他特定参数
    }

    return requestData, nil
}
```

### 3. 在工厂中注册
在 generator_factory.go 中注册新的生成器：
```go
func (f *RequestGeneratorFactory) registerGenerators() {
    f.generators[APITypeNewProvider] = NewNewProviderGenerator()
}
```

## 使用方法
### 1. 基础配置
```go 
config := LLMRequestParams{
    APIType:  "new_provider",
    BaseURL:  "https://api.newprovider.com/v1",
    APIKey:   "your-api-key",
    Model:    "model-name",
    Timeout:  30,
}
```

### 2. 创建对话
```go 
// 初始化工厂
factory := NewRequestGeneratorFactory()

// 获取生成器
generator, err := factory.GetGenerator(APITypeNewProvider) // APITypeNewProvider 来自 request_generator.go的常量
if err != nil {
    log.Fatal(err)
}

// 准备消息
messages := []Message{
    CreateSystemMessage("系统提示"),
    CreateUserMessage("用户问题"),
}

// 生成请求数据
requestData, err := generator.GenerateRequest(
    "当前问题",
    messages,
    config,
)
```