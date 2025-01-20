package configs

import (
	"OpenStress/pool"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// LLMType 定义枚举类型
type LLMType string

const (
	OPENAI           LLMType = "openai"
	ANTHROPIC        LLMType = "anthropic"
	CLAUDE           LLMType = "claude"
	SPARK            LLMType = "spark"
	ZHIPUAI          LLMType = "zhipuai"
	FIREWORKS        LLMType = "fireworks"
	OPENLLM          LLMType = "open_llm"
	GEMINI           LLMType = "gemini"
	METAGPT          LLMType = "metagpt"
	AZURE            LLMType = "azure"
	OLLAMA           LLMType = "ollama"
	OLLAMAGENERATE   LLMType = "ollama.generate"
	OLLAMAEMBEDDINGS LLMType = "ollama.embeddings"
	OLLAMAEMBED      LLMType = "ollama.embed"
	QIANFAN          LLMType = "qianfan"
	DASHSCOPE        LLMType = "dashscope"
	MOONSHOT         LLMType = "moonshot"
	MISTRAL          LLMType = "mistral"
	Yi               LLMType = "yi"
	OPENROUTER       LLMType = "openrouter"
	BEDROCK          LLMType = "bedrock"
	ARK              LLMType = "ark"
)

type LLMConfig struct {
	LLM LLMDetails `yaml:"llm"`
}

type LLMDetails struct {
	APIKey            string  `yaml:"api_key"`
	APIType           LLMType `yaml:"api_type"`
	BaseURL           string  `yaml:"base_url"`
	MaxToken          int     `yaml:"max_token"`
	Temperature       float64 `yaml:"temperature"`
	TopP              float64 `yaml:"top_p"`
	TopK              int     `yaml:"top_k"`
	RepetitionPenalty float64 `yaml:"repetition_penalty"`
	PresencePenalty   float64 `yaml:"presence_penalty"`
	FrequencyPenalty  float64 `yaml:"frequency_penalty"`
	Stream            bool    `yaml:"stream"`
	Timeout           int     `yaml:"timeout"`
	RegionName        string  `yaml:"region_name"`
	CalcUsage         bool    `yaml:"calc_usage"`
	UseSystemPrompt   bool    `yaml:"use_system_prompt"`
}

// 默认配置值
var defaultConfig = LLMDetails{
	APIKey:            "YOUR_API_KEY",
	APIType:           OPENAI,
	BaseURL:           "https://api.openai.com/v1",
	MaxToken:          2048,
	Temperature:       0.7,
	TopP:              0.9,
	TopK:              50,
	RepetitionPenalty: 1.2,
	PresencePenalty:   0.5,
	FrequencyPenalty:  0.5,
	Stream:            false,
	Timeout:           600,
	RegionName:        "us-west-1",
	CalcUsage:         true,
	UseSystemPrompt:   true,
}

// ReadConfig 读取并解析配置文件
func ReadLLMConfig(customDir ...string) (*LLMConfig, error) {
	logger, err := pool.GetLogger()
	if err != nil {
		fmt.Println("Error getting logger:", err.Error())
		return nil, err
	}

	dir := getConfigDirectory(customDir)
	configPath := filepath.Join(dir, "config", "llm.yaml")

	fmt.Println("Reading config file from:", configPath)
	fileContent, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Config file not found, creating default config...")
			logger.Log("ERROR", "Config file not found, creating default config: "+err.Error())
			if err := createDefaultConfig(configPath); err != nil {
				return nil, fmt.Errorf("failed to create default config: %w", err)
			}
			fmt.Println("Default config created, reading the config file again...")
			fileContent, err = os.ReadFile(configPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read newly created config file: %w", err)
			}
		} else {
			fmt.Println("Error reading config file:", err.Error())
			return nil, fmt.Errorf("failed to read config file at %s: %w", configPath, err)
		}
	} else {
		fmt.Println("Config file read successfully")
	}

	// 打印读取的 YAML 文件内容
	// fmt.Println("Raw YAML content:")
	// fmt.Println(string(fileContent))

	// fmt.Println("Parsing YAML content...")
	var config LLMConfig
	if err := yaml.Unmarshal(fileContent, &config); err != nil {
		fmt.Println("Failed to parse YAML config:", err.Error())
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}
	fmt.Println("YAML content parsed successfully")

	// // 打印解析后的每个字段
	// fmt.Println("Parsed Config Fields:")
	// fmt.Printf("APIKey: %s\n", config.LLM.APIKey)
	// fmt.Printf("APIType: %s\n", config.LLM.APIType)
	// fmt.Printf("BaseURL: %s\n", config.LLM.BaseURL)
	// fmt.Printf("MaxToken: %d\n", config.LLM.MaxToken)
	// fmt.Printf("Temperature: %.2f\n", config.LLM.Temperature)
	// fmt.Printf("TopP: %.2f\n", config.LLM.TopP)
	// fmt.Printf("TopK: %d\n", config.LLM.TopK)
	// fmt.Printf("RepetitionPenalty: %.2f\n", config.LLM.RepetitionPenalty)
	// fmt.Printf("PresencePenalty: %.2f\n", config.LLM.PresencePenalty)
	// fmt.Printf("FrequencyPenalty: %.2f\n", config.LLM.FrequencyPenalty)
	// fmt.Printf("Stream: %t\n", config.LLM.Stream)
	// fmt.Printf("Timeout: %d\n", config.LLM.Timeout)
	// fmt.Printf("RegionName: %s\n", config.LLM.RegionName)
	// fmt.Printf("CalcUsage: %t\n", config.LLM.CalcUsage)
	// fmt.Printf("UseSystemPrompt: %t\n", config.LLM.UseSystemPrompt)
	// 打印其他字段

	if err := validateConfig(&config); err != nil {
		fmt.Println("Config validation failed:", err.Error())
		return nil, err
	}

	logger.Log("INFO", "Config loaded successfully")
	return &config, nil
}

// getConfigDirectory 获取配置目录
func getConfigDirectory(customDir []string) string {
	if len(customDir) == 0 || customDir[0] == "" {
		dir, getConfigDirectoryErr := os.Getwd()
		if getConfigDirectoryErr != nil {
			logger, err := pool.GetLogger()
			if err != nil {
				logger.Log("ERROR", "Error get logger:"+err.Error())
				return ""
			}
			logger.Log("ERROR", "failed to get current directory:"+getConfigDirectoryErr.Error())
		}
		return dir
	}
	return customDir[0]
}

// createDefaultConfig 创建默认配置文件
func createDefaultConfig(configPath string) error {
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	defaultConfigData, err := yaml.Marshal(defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %w", err)
	}

	if err := os.WriteFile(configPath, defaultConfigData, 0644); err != nil {
		return fmt.Errorf("failed to write default config file: %w", err)
	}

	logger, logErr := pool.GetLogger()
	if logErr != nil {
		logger.Log("ERROR", "Error getting logger: "+logErr.Error())
		return logErr
	}
	logger.Log("INFO", "Default config created at: "+configPath)
	return nil
}

// validateConfig 验证配置的有效性
func validateConfig(config *LLMConfig) error {
	// // 打印完整的配置内容，方便调试
	// fmt.Printf("Validating config: %+v\n", config.LLM)

	// 检查 API Key 是否为空
	if config.LLM.APIKey == "" {
		return errors.New("api key is missing")
	}

	// 检查 Base URL 是否为空
	if config.LLM.BaseURL == "" {
		return errors.New("base URL is missing")
	}

	return nil
}
