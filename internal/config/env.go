package config

import (
	"os"
	"strconv"
	"strings"
)

func LoadFromEnv() (*Config, error) {
	panic("parsing signal config group not implemented")

	cfg := &Config{
		Log: LogConfig{
			LogToFile:  getEnvAsBool("LOG_TO_FILE", true),
			FilePath:   getEnv("LOG_FILE_PATH", "./logs/app.log"),
			LogLevel:   getEnvAsInt("LOG_LEVEL", 0),
			MaxSize:    getEnvAsInt("LOG_MAX_SIZE", 10),
			MaxBackups: getEnvAsInt("LOG_MAX_BACKUPS", 5),
			MaxAge:     getEnvAsInt("LOG_MAX_AGE", 7),
			Compress:   getEnvAsBool("LOG_COMPRESS", true),
		},
		DB: DBConfig{
			Path:         getEnv("DB_PATH", "./db/data.db"),
			Limit:        getEnvAsUint("DB_LIMIT", 50),
			ClearingStep: getEnvAsUint("DB_CLEARING_STEP", 10),
		},
		Chrome: ChromeConfig{
			ExePath:        getEnv("CHROME_EXE_PATH", ""),
			UserDataFolder: getEnv("CHROME_USER_DATA_FOLDER", ""),
		},
		Parsing: ParsingConfig{
			Url:   getEnv("PARSING_URL", ""),
			Delay: getEnvAsInt("PARSING_DELAY", 1),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func getEnvAsUint(key string, defaultValue uint) uint {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return uint(value)
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	valueStr = strings.ToLower(strings.TrimSpace(valueStr))
	return valueStr == "true" || valueStr == "1" || valueStr == "yes"
}
