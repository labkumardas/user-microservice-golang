package config

import (
	"os"
	"strconv"
)

// AppConfig holds all application configuration
type AppConfig struct {
	AppEnv         string
	ServerPort     string
	MongoURI       string
	MongoDBName    string
	MongoTimeout   int
	JWTSecret      string
	JWTExpiryHours int
	BcryptCost     int
}

// Load reads configuration from environment variables
func Load() *AppConfig {
	return &AppConfig{
		AppEnv:         getEnv("APP_ENV", "development"),
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		MongoURI:       getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName:    getEnv("MONGO_DB_NAME", "userservice"),
		MongoTimeout:   getEnvAsInt("MONGO_TIMEOUT", 10),
		JWTSecret:      getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpiryHours: getEnvAsInt("JWT_EXPIRY_HOURS", 24),
		BcryptCost:     getEnvAsInt("BCRYPT_COST", 12),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
