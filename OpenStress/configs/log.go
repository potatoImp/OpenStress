package configs

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type LogConfig struct {
	Log LogDetails `yaml:"log"`
}

type LogDetails struct {
	// Directory specifies the directory where log files are stored.
	// Example: "./logs/"
	Directory string `yaml:"directory"`

	// Filename specifies the name of the log file.
	// Example: "OpenStress.log"
	Filename string `yaml:"filename"`

	// Level specifies the logging level.
	// Possible values: "DEBUG", "INFO", "WARN", "ERROR"
	// Example: "INFO"
	Level string `yaml:"level"`

	// MaxSize specifies the maximum size in megabytes of the log file before it gets rotated.
	// Unit: Megabytes (MB)
	// Example: 100
	MaxSize int `yaml:"max_size"`

	// MaxAge specifies the maximum number of days to retain old log files.
	// Example: 28
	MaxAge int `yaml:"max_age"`
}

// var defaultLogConfig = LogConfig{
// 	Log: LogDetails{
// 		Directory: "./logs/",
// 		Filename:  "OpenStress.log",
// 		Level:     "INFO",
// 		MaxSize:   100,
// 		MaxAge:    28,
// 	},
// }

func createDefaultLogConfig(configPath string) error {
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create log config directory: %w", err)
	}

	// Manually construct the YAML content with comments
	defaultConfigContent := `# Log configuration file
log:
  # Directory specifies the directory where log files are stored.
  # Example: "./logs/"
  directory: "./logs/"
  
  # Filename specifies the name of the log file.
  # Example: "OpenStress.log"
  filename: "OpenStress.log"
  
  # Level specifies the logging level.
  # Example: "INFO"
  level: "INFO"
  
  # MaxSize specifies the maximum size in megabytes of the log file before it gets rotated.
  # Example: 100
  max_size: 100
  
  # MaxAge specifies the maximum number of days to retain old log files.
  # Example: 28
  max_age: 28
`

	if err := os.WriteFile(configPath, []byte(defaultConfigContent), 0644); err != nil {
		return fmt.Errorf("failed to write default log config file: %w", err)
	}

	return nil
}

func InitializeLogConfig(customDir ...string) (*LogDetails, error) {
	dir := getConfigDirectory(customDir)
	configPath := filepath.Join(dir, "config", "log.yaml")

	fileContent, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Log config file not found, creating default config...")
			if err := createDefaultLogConfig(configPath); err != nil {
				return nil, fmt.Errorf("failed to create default log config: %w", err)
			}
			fileContent, err = os.ReadFile(configPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read newly created log config file: %w", err)
			}
		} else if os.IsPermission(err) {
			return nil, fmt.Errorf("permission denied when reading log config file: %w", err)
		} else {
			return nil, fmt.Errorf("failed to read log config file at %s: %w", configPath, err)
		}
	}

	var config LogConfig
	if err := yaml.Unmarshal(fileContent, &config); err != nil {
		fmt.Println("Failed to parse YAML log config, recreating default config...")
		if err := os.Remove(configPath); err != nil {
			return nil, fmt.Errorf("failed to remove invalid log config file: %w", err)
		}
		if err := createDefaultLogConfig(configPath); err != nil {
			return nil, fmt.Errorf("failed to recreate default log config: %w", err)
		}
		fileContent, err = os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read newly created log config file: %w", err)
		}
		if err := yaml.Unmarshal(fileContent, &config); err != nil {
			return nil, fmt.Errorf("failed to parse newly created YAML log config: %w", err)
		}
	}

	return &config.Log, nil
}
