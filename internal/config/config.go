package config

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

const localScannerToken = "aksescheck-local-scanner-token"

type Config struct {
	AppEnv            string
	APIAddr           string
	ScannerAddr       string
	ScannerURL        string
	ScannerToken      string
	DatabaseURL       string
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
	ScanTimeout       time.Duration
	ChromePath        string
}

func Load() (Config, error) {
	loadEnvFile(".env")

	cfg := Config{
		AppEnv:            getString("APP_ENV", "development"),
		APIAddr:           getListenAddress("API_ADDR", "127.0.0.1:8080"),
		ScannerAddr:       getListenAddress("SCANNER_ADDR", "127.0.0.1:8081"),
		ScannerURL:        getString("SCANNER_URL", "http://127.0.0.1:8081"),
		ScannerToken:      getString("SCANNER_TOKEN", localScannerToken),
		DatabaseURL:       strings.TrimSpace(os.Getenv("DATABASE_URL")),
		WebOrigin:         getString("WEB_ORIGIN", "http://127.0.0.1:3000"),
		SessionCookieName: getString("SESSION_COOKIE_NAME", "aksescheck_session"),
		SessionTTL:        getDuration("SESSION_TTL", 7*24*time.Hour),
		PasswordResetTTL:  getDuration("PASSWORD_RESET_TTL", 30*time.Minute),
		SMTPHost:          strings.TrimSpace(os.Getenv("SMTP_HOST")),
		SMTPPort:          getInt("SMTP_PORT", 587),
		SMTPUsername:      strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		SMTPPassword:      os.Getenv("SMTP_PASSWORD"),
		SMTPFromName:      getString("SMTP_FROM_NAME", "AksesCheck ID"),
		SMTPFromEmail:     strings.TrimSpace(os.Getenv("SMTP_FROM_EMAIL")),
		ScanTimeout:       getDuration("SCAN_TIMEOUT", 60*time.Second),
		ChromePath:        strings.TrimSpace(os.Getenv("CHROME_PATH")),
	}

	if cfg.SMTPFromEmail == "" {
		cfg.SMTPFromEmail = cfg.SMTPUsername
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New(
			"DATABASE_URL wajib diisi",
		)
	}

	if cfg.SessionTTL < time.Hour {
		return Config{}, errors.New(
			"SESSION_TTL minimal satu jam",
		)
	}

	if cfg.PasswordResetTTL < 10*time.Minute ||
		cfg.PasswordResetTTL > 24*time.Hour {
		return Config{}, errors.New(
			"PASSWORD_RESET_TTL harus antara 10 menit dan 24 jam",
		)
	}

	if cfg.SMTPPort < 1 ||
		cfg.SMTPPort > 65535 {
		return Config{}, errors.New(
			"SMTP_PORT tidak valid",
		)
	}

	if cfg.ScanTimeout < 10*time.Second ||
		cfg.ScanTimeout > 5*time.Minute {
		return Config{}, errors.New(
			"SCAN_TIMEOUT harus antara 10 detik dan 5 menit",
		)
	}

	if cfg.AppEnv == "production" &&
		cfg.ScannerToken == localScannerToken {
		return Config{}, errors.New(
			"SCANNER_TOKEN production wajib diganti",
		)
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
		line := strings.TrimSpace(
			scanner.Text(),
		)

		if line == "" ||
			strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(
			line,
			"=",
		)
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
			if strings.HasPrefix(
				value,
				`"`,
			) &&
				strings.HasSuffix(
					value,
					`"`,
				) {
				if unquoted, unquoteErr :=
					strconv.Unquote(value); unquoteErr == nil {
					value = unquoted
				}
			} else if strings.HasPrefix(
				value,
				"'",
			) &&
				strings.HasSuffix(
					value,
					"'",
				) {
				value = value[1 : len(value)-1]
			}
		}

		_ = os.Setenv(
			key,
			value,
		)
	}
}

func getListenAddress(
	key string,
	fallback string,
) string {
	appEnv := strings.ToLower(
		strings.TrimSpace(
			os.Getenv("APP_ENV"),
		),
	)

	if appEnv == "production" {
		port := strings.TrimSpace(
			os.Getenv("PORT"),
		)

		if port != "" {
			port = strings.TrimPrefix(
				port,
				":",
			)

			if parsedPort, err := strconv.Atoi(
				port,
			); err == nil &&
				parsedPort >= 1 &&
				parsedPort <= 65535 {
				return ":" + port
			}
		}
	}

	return getString(
		key,
		fallback,
	)
}

func getString(
	key string,
	fallback string,
) string {
	value := strings.TrimSpace(
		os.Getenv(key),
	)

	if value == "" {
		return fallback
	}

	return value
}

func getInt(
	key string,
	fallback int,
) int {
	value := strings.TrimSpace(
		os.Getenv(key),
	)

	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getDuration(
	key string,
	fallback time.Duration,
) time.Duration {
	value := strings.TrimSpace(
		os.Getenv(key),
	)

	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil ||
		parsed <= 0 {
		return fallback
	}

	return parsed
}
