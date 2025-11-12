package configs

import (
	"fmt"
	"os"
)

type Config struct {
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	DBSSLMode     string
	JWTPrivateKey string
	JWTPublicKey  string
}

func LoadConfig() *Config {
	return &Config{
		DBHost:        getEnv("POSTGRES_HOST", "localhost"),
		DBPort:        getEnv("POSTGRES_PORT", "5432"),
		DBUser:        getEnv("POSTGRES_USER", "postgres"),
		DBPassword:    getEnv("POSTGRES_PASSWORD", "secret"),
		DBName:        getEnv("POSTGRES_DB", "postgres"),
		DBSSLMode:     getEnv("POSTGRES_SSLMODE", "disable"),
		JWTPrivateKey: getEnv("JWT_PRIVATE_KEY", "secret"),
		JWTPublicKey:  getEnv("JWT_PUBLIC_KEY", "public_secret"),
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
