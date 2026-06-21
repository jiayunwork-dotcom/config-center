package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort      int
	DBHost          string
	DBPort          int
	DBUser          string
	DBPassword      string
	DBName          string
	RedisHost       string
	RedisPort       int
	LongPollTimeout int
}

func Load() *Config {
	return &Config{
		ServerPort:      getEnvInt("SERVER_PORT", 8080),
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBPort:          getEnvInt("DB_PORT", 5432),
		DBUser:          getEnv("DB_USER", "config_center"),
		DBPassword:      getEnv("DB_PASSWORD", "config_center_pass"),
		DBName:          getEnv("DB_NAME", "config_center"),
		RedisHost:       getEnv("REDIS_HOST", "localhost"),
		RedisPort:       getEnvInt("REDIS_PORT", 6379),
		LongPollTimeout: getEnvInt("LONG_POLL_TIMEOUT", 30),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
