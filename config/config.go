package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	Auth     AuthConfig
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
	Host string
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	// JWTSecret signs access tokens. It must come from the environment: a
	// secret committed to source control is not a secret.
	JWTSecret string
	JWTIssuer string

	// AccessTokenTTL is short by design. An access token cannot be revoked, so
	// its lifetime is the window an attacker gets with a stolen one.
	AccessTokenTTL time.Duration

	// RefreshTokenTTL bounds how long a session can be kept alive silently.
	RefreshTokenTTL time.Duration

	// MaxFailedLogins is how many consecutive wrong passwords lock an account.
	MaxFailedLogins int
	LockoutDuration time.Duration
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	config := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "ecommerce"),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "4444"),
			Host: getEnv("SERVER_HOST", "localhost"),
		},
		Auth: AuthConfig{
			JWTSecret:       getEnv("JWT_SECRET", ""),
			JWTIssuer:       getEnv("JWT_ISSUER", "go-ecommerce"),
			AccessTokenTTL:  getEnvDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTokenTTL: getEnvDuration("JWT_REFRESH_TTL", 720*time.Hour), // 30 days
			MaxFailedLogins: getEnvInt("AUTH_MAX_FAILED_LOGINS", 5),
			LockoutDuration: getEnvDuration("AUTH_LOCKOUT_DURATION", 15*time.Minute),
		},
	}

	// Validate required fields
	if config.Database.Password == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}

	// Fail at startup rather than on the first login attempt. A missing secret
	// is a deployment mistake, and the worst possible fallback is a hardcoded
	// default that silently ships to production.
	if config.Auth.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required (generate one with: openssl rand -base64 48)")
	}

	return config, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvDuration parses a Go duration string such as "15m" or "720h".
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

// getEnvInt parses an integer environment variable.
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

// GetDSN returns the MySQL Data Source Name
func (c *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.DBName,
	)
}
