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
	AppEnv            string
	APIAddr           string
	DatabaseURL       string
	RedisAddr         string
	RedisPassword     string
	RedisDB           int
	WebOrigin         string
	SessionCookieName string
	SessionTTL        time.Duration
	PasswordResetTTL  time.Duration
	SMTPHost          string
	SMTPPort          int
	SMTPUsername      string
	SMTPPassword      string
	SMTPFromName      string
	SMTPFromEmail     string
	ScanQueue         string
	ScanTimeout       time.Duration
	WorkerConcurrency int
	ChromePath        string
}

func Load() (Config, error) {
	loadEnvFile(".env")

	cfg := Config{
		AppEnv:            getString("APP_ENV", "development"),
		APIAddr:           getString("API_ADDR", ":8080"),
		DatabaseURL:       strings.TrimSpace(os.Getenv("DATABASE_URL")),
		RedisAddr:         getString("REDIS_ADDR", "localhost:6379"),
		RedisPassword:     os.Getenv("REDIS_PASSWORD"),
		RedisDB:           getInt("REDIS_DB", 0),
		WebOrigin:         getString("WEB_ORIGIN", "http://localhost:3000"),
		SessionCookieName: getString("SESSION_COOKIE_NAME", "aksescheck_session"),
		SessionTTL:        getDuration("SESSION_TTL", 7*24*time.Hour),
		PasswordResetTTL:  getDuration("PASSWORD_RESET_TTL", 30*time.Minute),
		SMTPHost:          strings.TrimSpace(os.Getenv("SMTP_HOST")),
		SMTPPort:          getInt("SMTP_PORT", 587),
		SMTPUsername:      strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		SMTPPassword:      os.Getenv("SMTP_PASSWORD"),
		SMTPFromName:      getString("SMTP_FROM_NAME", "AksesCheck ID"),
		SMTPFromEmail:     strings.TrimSpace(os.Getenv("SMTP_FROM_EMAIL")),
		ScanQueue:         getString("SCAN_QUEUE", "scanner"),
		ScanTimeout:       getDuration("SCAN_TIMEOUT", 60*time.Second),
		WorkerConcurrency: getInt("WORKER_CONCURRENCY", 2),
		ChromePath:        strings.TrimSpace(os.Getenv("CHROME_PATH")),
	}

	if cfg.SMTPFromEmail == "" {
		cfg.SMTPFromEmail = cfg.SMTPUsername
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL wajib diisi")
	}

	if cfg.SessionTTL < time.Hour {
		return Config{}, errors.New("SESSION_TTL minimal satu jam")
	}

	if cfg.PasswordResetTTL < 10*time.Minute ||
		cfg.PasswordResetTTL > 24*time.Hour {
		return Config{}, errors.New(
			"PASSWORD_RESET_TTL harus antara 10 menit dan 24 jam",
		)
	}

	if cfg.SMTPPort < 1 || cfg.SMTPPort > 65535 {
		return Config{}, errors.New("SMTP_PORT tidak valid")
	}

	if cfg.WorkerConcurrency < 1 {
		return Config{}, errors.New("WORKER_CONCURRENCY minimal satu")
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
			if strings.HasPrefix(value, `"`) &&
				strings.HasSuffix(value, `"`) {
				if unquoted, unquoteErr := strconv.Unquote(value); unquoteErr == nil {
					value = unquoted
				}
			} else if strings.HasPrefix(value, "'") &&
				strings.HasSuffix(value, "'") {
				value = value[1 : len(value)-1]
			}
		}

		_ = os.Setenv(key, value)
	}
}

func getString(key, fallback string) string {
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
	if err != nil {
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
