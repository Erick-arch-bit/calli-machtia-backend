package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	PORT              string
	DATABASE_URL      string
	MONGODB_URI       string
	MONGODB_DATABASE  string
	REDIS_URL         string
	JWT_SECRET        string
	JWT_ACCESS_EXPIRY  time.Duration
	JWT_REFRESH_EXPIRY time.Duration
	STRIPE_SECRET_KEY      string
	STRIPE_WEBHOOK_SECRET  string
	CORS_ORIGIN       string
	ENV               string
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvRequired(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic("REQUIRED_ENV_NOT_SET: " + key)
	}
	return val
}

func parseDuration(s string, defaultVal time.Duration) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return defaultVal
	}
	return d
}

func Load() *Config {
	godotenv.Load()

	accessExpiry := parseDuration(getEnv("JWT_ACCESS_EXPIRY", "15m"), 15*time.Minute)
	refreshExpiry := parseDuration(getEnv("JWT_REFRESH_EXPIRY", "7d"), 7*24*time.Hour)

	return &Config{
		PORT:              getEnv("PORT", "8080"),
		DATABASE_URL:      getEnvRequired("DATABASE_URL"),
		MONGODB_URI:       getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MONGODB_DATABASE:  getEnv("MONGODB_DATABASE", "calli_machtia"),
		REDIS_URL:         getEnv("REDIS_URL", "redis://localhost:6379"),
		JWT_SECRET:        getEnv("JWT_SECRET", "dev-secret-change-in-production-abc123"),
		JWT_ACCESS_EXPIRY:  accessExpiry,
		JWT_REFRESH_EXPIRY: refreshExpiry,
		STRIPE_SECRET_KEY:      getEnv("STRIPE_SECRET_KEY", ""),
		STRIPE_WEBHOOK_SECRET:  getEnv("STRIPE_WEBHOOK_SECRET", ""),
		CORS_ORIGIN:       getEnv("CORS_ORIGIN", "https://calli-machtia.up.railway.app"),
		ENV:               getEnv("ENV", "development"),
	}
}
