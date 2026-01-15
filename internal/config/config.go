package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoDBURI        string
	MongoDBDatabase   string
	TelegramBotToken  string
	OpenAIAPIKey      string
	EncryptionKey     string
	Port              string
	Environment       string
	SyncIntervalHours int
	LogLevel          string
}

func Load() (*Config, error) {
	// Load .env file if exists (for local development)
	_ = godotenv.Load()

	cfg := &Config{
		MongoDBURI:        getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDBDatabase:   getEnv("MONGODB_DATABASE", "seller_assistant"),
		TelegramBotToken:  getEnv("TELEGRAM_BOT_TOKEN", ""),
		OpenAIAPIKey:      getEnv("OPENAI_API_KEY", ""),
		EncryptionKey:     getEnv("ENCRYPTION_KEY", ""),
		Port:              getEnv("PORT", "8080"),
		Environment:       getEnv("ENVIRONMENT", "production"),
		SyncIntervalHours: getEnvAsInt("SYNC_INTERVAL_HOURS", 6),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.MongoDBURI == "" {
		return fmt.Errorf("MONGODB_URI is required")
	}
	if c.MongoDBDatabase == "" {
		return fmt.Errorf("MONGODB_DATABASE is required")
	}
	if c.TelegramBotToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	if c.OpenAIAPIKey == "" {
		return fmt.Errorf("OPENAI_API_KEY is required")
	}
	if c.EncryptionKey == "" {
		return fmt.Errorf("ENCRYPTION_KEY is required")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
