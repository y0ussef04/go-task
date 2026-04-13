package config

import (
	"fmt"
	"os"
)

// Config holds application configuration
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	AppPort    string
}

// Load reads configuration from environment variables with sensible defaults
func Load() *Config {
	return &Config{
		DBHost:     getEnv("DB_HOST", "mysql"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", "secret"),
		AppPort:    getEnv("APP_PORT", "8081"),
	}
}

// DSN returns the MySQL Data Source Name without a specific database
func (c *Config) DSN(dbName string) string {
	base := fmt.Sprintf("%s:%s@tcp(%s:%s)/", c.DBUser, c.DBPassword, c.DBHost, c.DBPort)
	if dbName != "" {
		return base + dbName + "?parseTime=true"
	}
	return base + "?parseTime=true"
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
