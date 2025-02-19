package configs

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v2"
)

type BaseConfig struct {
	Base BaseDetails `yaml:"base"`
}

type BaseDetails struct {
	// 用户数，协程池数
	// 默认值: 100
	Users int `yaml:"users"`

	// 任务默认循环次数
	// 默认值: 1
	// 设置多个可重复执行，设置为非正整数无效
	TaskDefaultRepeats int `yaml:"task_default_repeats"`

	// 任务间隔，单位为毫秒
	// 默认值: 1000
	TaskInterval int `yaml:"task_interval"`

	// 定时执行时长，单位为秒
	// 默认值: 0
	// 非正数代表不走定时执行逻辑；设置为正整数则走定时执行逻辑
	ScheduledDuration int `yaml:"scheduled_duration"`

	// 测试数据格式
	// 默认值: none
	// 可选值: none, jtl
	TestDataFormat string `yaml:"test_data_format"`

	// HTML测试报告
	// 默认值: false
	// 可选值: true, false
	HtmlTestReport bool `yaml:"html_test_report"`

	// AI分析
	// 默认值: false
	// 可选值: true, false
	AiAnalysis bool `yaml:"ai_analysis"`

	// 知识库链接
	// 默认值: none
	KnowledgeBaseLink string `yaml:"knowledge_base_link"`

	// Web界面
	// 默认值: false
	// 可选值: true, false
	WebInterface bool `yaml:"web_interface"`
}

func createDefaultBaseConfig(configPath string) error {
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create base config directory: %w", err)
	}

	// Manually construct the YAML content with comments
	defaultConfigContent := `# Base configuration file
base:
  # 用户数，协程池数
  # 默认值: 100
  users: 100
  
  # 任务默认循环次数
  # 默认值: 1
  task_default_repeats: 1
  
  # 任务间隔，单位为毫秒
  # 默认值: 1000
  task_interval: 1000
  
  # 定时执行时长，单位为秒
  # 默认值: 0
  scheduled_duration: 0
  
  # 测试数据格式
  # 默认值: none
  test_data_format: none
  
  # HTML测试报告
  # 默认值: false
  html_test_report: false
  
  # AI分析
  # 默认值: false
  ai_analysis: false
  
  # 知识库链接
  # 默认值: none
  knowledge_base_link: none
  
  # Web界面
  # 默认值: false
  web_interface: false
`

	if err := os.WriteFile(configPath, []byte(defaultConfigContent), 0644); err != nil {
		return fmt.Errorf("failed to write default base config file: %w", err)
	}

	return nil
}

var (
	baseDetailsInstance *BaseDetails
	once                sync.Once
)

func InitializeBaseConfig(customDir ...string) (*BaseDetails, error) {
	var initErr error // 使用局部变量捕获错误
	once.Do(func() {
		dir := getConfigDirectory(customDir)
		configPath := filepath.Join(dir, "config", "base.yaml")

		fileContent, err := os.ReadFile(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("Base config file not found, creating default config...")
				if err := createDefaultBaseConfig(configPath); err != nil {
					initErr = fmt.Errorf("failed to create default base config: %w", err)
					return
				}
				fileContent, err = os.ReadFile(configPath)
				if err != nil {
					initErr = fmt.Errorf("failed to read newly created base config file: %w", err)
					return
				}
			} else if os.IsPermission(err) {
				initErr = fmt.Errorf("permission denied when reading base config file: %w", err)
				return
			} else {
				initErr = fmt.Errorf("failed to read base config file at %s: %w", configPath, err)
				return
			}
		}

		var config BaseConfig
		if err := yaml.Unmarshal(fileContent, &config); err != nil {
			fmt.Println("Failed to parse YAML base config, recreating default config...")
			if err := os.Remove(configPath); err != nil {
				initErr = fmt.Errorf("failed to remove invalid base config file: %w", err)
				return
			}
			if err := createDefaultBaseConfig(configPath); err != nil {
				initErr = fmt.Errorf("failed to recreate default base config: %w", err)
				return
			}
			fileContent, err = os.ReadFile(configPath)
			if err != nil {
				initErr = fmt.Errorf("failed to read newly created base config file: %w", err)
				return
			}
			if err := yaml.Unmarshal(fileContent, &config); err != nil {
				initErr = fmt.Errorf("failed to parse newly created YAML base config: %w", err)
				return
			}
		}

		baseDetailsInstance = &config.Base
	})

	return baseDetailsInstance, initErr // 返回捕获的错误
}

func GetBaseDetails() *BaseDetails {
	return baseDetailsInstance
}
