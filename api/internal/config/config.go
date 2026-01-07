package config

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds all configuration for the application
type Config struct {
	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// JWT
	JWTSecret        string
	JWTAccessExpiry  string
	JWTRefreshExpiry string

	// Server
	GRPCPort string
	HTTPPort string

	// Observability
	OTELEndpoint string
	LogLevel     string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           getEnv("DB_PORT", "5432"),
		DBUser:           getEnv("DB_USER", "postgres"),
		DBPassword:       getEnv("DB_PASSWORD", "postgres"),
		DBName:           getEnv("DB_NAME", "shopping_list_dev"),
		DBSSLMode:        getEnv("DB_SSLMODE", "disable"),
		JWTSecret:        getEnv("JWT_SECRET", "change-me-in-production"),
		JWTAccessExpiry:  getEnv("JWT_ACCESS_EXPIRY", "15m"),
		JWTRefreshExpiry: getEnv("JWT_REFRESH_EXPIRY", "7d"),
		GRPCPort:         getEnv("GRPC_PORT", "50051"),
		HTTPPort:         getEnv("HTTP_PORT", "8080"),
		OTELEndpoint:     getEnv("OTEL_ENDPOINT", "http://localhost:4317"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// NewLogger creates a new zap logger based on config
func NewLogger(level string) (*zap.Logger, error) {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapLevel)
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	return config.Build()
}

// ParseDuration parses duration strings like "15m", "7d"
func ParseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}
