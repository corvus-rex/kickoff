package config

import (
	"os"
	"strconv"
)

// DefaultJWTSecret is the fallback used only for local development.
// main.go refuses to start with this value outside APP_ENV=development.
const DefaultJWTSecret = "dev-secret-change-me"

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

	JWTSecret        string
	JWTExpiryMinutes int
	SeedUsers        bool
}

func Load() *Config {
	env := getEnv("APP_ENV", "development")

	seedDefault := env == "development"

	return &Config{
		Port: getEnv("PORT", "8080"),
		Env:  env,

		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "5432"),
		DBUser:               getEnv("DB_USER", "postgres"),
		DBPassword:           getEnv("DB_PASSWORD", "postgres"),
		DBName:               getEnv("DB_NAME", "xyz_football"),
		DBSSLMode:            getEnv("DB_SSLMODE", "disable"),
		DBMaxOpenConns:       getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:       getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
		DBConnMaxLifetimeMin: getEnvAsInt("DB_CONN_MAX_LIFETIME_MIN", 5),

		JWTSecret:        getEnv("JWT_SECRET", DefaultJWTSecret),
		JWTExpiryMinutes: getEnvAsInt("JWT_EXPIRY_MINUTES", 60),
		SeedUsers:        getEnvAsBool("SEED_USERS", seedDefault),
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

func getEnvAsBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return fallback
}