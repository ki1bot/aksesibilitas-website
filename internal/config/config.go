package config

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	APIAddr           string
	DatabaseURL       string
	RedisAddr         string
	RedisPassword     string
	RedisDB           int
	WebOrigin         string
	ScanQueue         string
	ScanTimeout       time.Duration
	WorkerConcurrency int
}

func Load() (Config, error) {
	loadEnvFile(".env")

	cfg := Config{
		APIAddr:           getString("API_ADDR", ":8080"),
		DatabaseURL:       strings.TrimSpace(os.Getenv("DATABASE_URL")),
		RedisAddr:         getString("REDIS_ADDR", "localhost:6379"),
		RedisPassword:     os.Getenv("REDIS_PASSWORD"),
		RedisDB:           getInt("REDIS_DB", 0),
		WebOrigin:         getString("WEB_ORIGIN", "http://localhost:3000"),
		ScanQueue:         getString("SCAN_QUEUE", "scanner"),
		ScanTimeout:       getDuration("SCAN_TIMEOUT", 60*time.Second),
		WorkerConcurrency: getInt("WORKER_CONCURRENCY", 2),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL wajib diisi")
	}

	return cfg, nil
}

func loadEnvFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		if len(value) >= 2 {
			if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
				if unquoted, unquoteErr := strconv.Unquote(value); unquoteErr == nil {
					value = unquoted
				}
			} else if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
				value = value[1 : len(value)-1]
			}
		}

		_ = os.Setenv(key, value)
	}
}

func getString(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func getInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return fallback
	}

	return parsed
}

func getDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}
