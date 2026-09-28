// Package config loads application configuration from environment variables.
package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppPort       string
	DBDSN         string
	RedisAddr     string
	RedisPassword string
	CacheTTL      time.Duration
}

// Load reads configuration with local-development defaults. If a .env file
// exists in the working directory, its values fill any variable that is not
// already set in the environment (real env vars win).
func Load() Config {
	loadDotEnv(".env")
	return Config{
		AppPort:       getEnv("APP_PORT", "8080"),
		DBDSN:         getEnv("DB_DSN", "adam:adamramadhans@tcp(localhost:3306)/task_management?parseTime=true"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		CacheTTL:      time.Duration(getEnvInt("CACHE_TTL_SECONDS", 60)) * time.Second,
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

func loadDotEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return // no .env file: defaults and the real environment apply
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key != "" {
			if _, exists := os.LookupEnv(key); !exists {
				_ = os.Setenv(key, value)
			}
		}
	}
}
