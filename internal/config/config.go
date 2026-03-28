package config

import (
	"os"
)

type Config struct {
	DB         DBConfig
	JWTSecret  string
	AppName    string
	AppVersion string
	AppPort    string
	AppEnv     string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

// LoadConfig reads environment variables and returns a Config struct
func LoadConfig() *Config {
	return &Config{
		AppName:    getEnv("APP_NAME", "Workout Assistant"),
		AppVersion: getEnv("APP_VERSION", "1.0.0"),
		AppPort:    getEnv("APP_PORT", "8080"),
		AppEnv:     getEnv("APP_ENV", "debug"),
		JWTSecret:  getEnv("JWT_SECRET", ""),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "aws-0-ap-south-1.pooler.supabase.com"),
			Port:     getEnv("DB_PORT", "6543"),
			User:     getEnv("DB_USER", "postgres.user"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "postgres"),
		},
	}
}

// Helper to provide default values if an ENV is missing
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
