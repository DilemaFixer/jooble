package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type (
	Config struct {
		Log     LogConfig     `yaml:"log"`
		DB      DBConfig      `yaml:"db"`
		Chrome  ChromeConfig  `yaml:"chrome"`
		Parsing ParsingConfig `yaml:"parsing"`
		Signal  SignalConfig  `yaml:"signal"`
	}

	LogConfig struct {
		LogToFile  bool   `yaml:"log_to_file"`
		FilePath   string `yaml:"file_path"`
		LogLevel   int    `yaml:"log_level"`
		MaxSize    int    `yaml:"max_size"`
		MaxBackups int    `yaml:"max_backups"`
		MaxAge     int    `yaml:"max_age"`
		Compress   bool   `yaml:"compress"`
	}

	DBConfig struct {
		Path         string `yaml:"path"`
		Limit        uint   `yaml:"limit"`
		ClearingStep uint   `yaml:"clear_step"`
	}

	ChromeConfig struct {
		ExePath        string `yaml:"exe_path"`
		UserDataFolder string `yaml:"user_data_folder"`
	}

	ParsingConfig struct {
		Url   string `yaml:"url"`
		Delay int    `yaml:"delay"` // in m
	}

	SignalConfig struct {
		Token      string `yaml:"token"`
		CustomerId int64  `yaml:"customer_id"`
	}
)

func Load(configPath string) (*Config, error) {
	if IsDocker() {
		return LoadFromEnv()
	}

	return LoadFromFile(configPath)
}

func (c *Config) SetDefaults() {
	if c.Log.FilePath == "" && c.Log.LogToFile {
		c.Log.FilePath = "./logs/app.log"
	}
	if c.Log.MaxSize == 0 {
		c.Log.MaxSize = 10
	}
	if c.Log.MaxBackups == 0 {
		c.Log.MaxBackups = 5
	}
	if c.Log.MaxAge == 0 {
		c.Log.MaxAge = 7
	}

	if c.DB.Path == "" {
		c.DB.Path = "./db/data.db"
	}
	if c.DB.Limit == 0 {
		c.DB.Limit = 50
	}
	if c.DB.ClearingStep == 0 {
		c.DB.ClearingStep = 10
	}

	if c.Parsing.Delay == 0 {
		c.Parsing.Delay = 1
	}
}

func (c *Config) Validate() error {
	if c.Log.LogToFile && c.Log.FilePath == "" {
		return fmt.Errorf("log.file_path is required when log.log_to_file is true")
	}
	if c.Log.LogLevel < 0 || c.Log.LogLevel > 5 {
		return fmt.Errorf("log.log_level must be between 0 and 5, got %d", c.Log.LogLevel)
	}

	if c.DB.Path == "" {
		return fmt.Errorf("db.path is required")
	}
	if c.DB.Limit == 0 {
		return fmt.Errorf("db.limit is required")
	}
	if c.DB.ClearingStep == 0 {
		return fmt.Errorf("db.clearing_step is required")
	}

	if c.Chrome.ExePath == "" {
		return fmt.Errorf("chrome.exe_path is required")
	}
	if _, err := os.Stat(c.Chrome.ExePath); os.IsNotExist(err) {
		return fmt.Errorf("chrome.exe_path does not exist: %s", c.Chrome.ExePath)
	}
	if c.Chrome.UserDataFolder == "" {
		return fmt.Errorf("chrome.user_data_folder is required")
	}

	if c.Parsing.Url == "" {
		return fmt.Errorf("parsing.url is required")
	}
	if c.Parsing.Delay < 1 {
		return fmt.Errorf("parsing.delay must be at least 1 minute, got %d", c.Parsing.Delay)
	}

	return nil
}
func (p *ParsingConfig) GetDelayDuration() time.Duration {
	return time.Duration(p.Delay) * time.Minute
}

func (c *Config) Save(configPath string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
