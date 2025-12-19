package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Env string

	HTTPAddr string

	DatabaseURL string
	JWTSecret   string

	// mTLS
	MTLSCertFile string
	MTLSKeyFile  string
	MTLSCAFile   string

	// Toss API base URLs
	TossAPIBaseURL string // default: https://apps-in-toss-api.toss.im
	TossPayBaseURL string // default: https://pay-apps-in-toss-api.toss.im

	HTTPTimeout time.Duration
}

func Load() Config {
	return Config{
		Env:            getenv("APP_ENV", "dev"),
		HTTPAddr:       getenv("HTTP_ADDR", ":8080"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		JWTSecret:      getenv("JWT_SECRET", "dev-secret-change-me"),

		MTLSCertFile: os.Getenv("TOSS_MTLS_CERT_FILE"),
		MTLSKeyFile:  os.Getenv("TOSS_MTLS_KEY_FILE"),
		MTLSCAFile:   os.Getenv("TOSS_MTLS_CA_FILE"),

		TossAPIBaseURL: getenv("TOSS_API_BASE_URL", "https://apps-in-toss-api.toss.im"),
		TossPayBaseURL: getenv("TOSS_PAY_BASE_URL", "https://pay-apps-in-toss-api.toss.im"),

		HTTPTimeout: time.Duration(getenvInt("HTTP_TIMEOUT_SECONDS", 12)) * time.Second,
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
