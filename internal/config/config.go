package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port string
	Env  string

	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	DBSSLMode            string
	DBMaxOpenConns       int
	DBMaxIdleConns       int
	DBConnMaxLifetimeMin int
}

func Load() *Config {
	return &Config{
		Port: getEnv("PORT", "8080"),
		Env:  getEnv("APP_ENV", "development"),

		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "5432"),
		DBUser:               getEnv("DB_USER", "postgres"),
		DBPassword:           getEnv("DB_PASSWORD", "postgres"),
		DBName:               getEnv("DB_NAME", "xyz_football"),
		DBSSLMode:            getEnv("DB_SSLMODE", "disable"),
		DBMaxOpenConns:       getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:       getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
		DBConnMaxLifetimeMin: getEnvAsInt("DB_CONN_MAX_LIFETIME_MIN", 5),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return fallback
}