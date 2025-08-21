package config

import (
	"os"
	"strconv"

	"backend/pkg/redis"
)

// Config holds all application configuration
// Contains settings for server, database, and external services
type Config struct {
	Server   ServerConfig `json:"server"`
	Redis    RedisConfig  `json:"redis"`
	Session  SessionConfig `json:"session"`
	Email    EmailConfig  `json:"email"`
	Security SecurityConfig `json:"security"`
}

// ServerConfig contains HTTP server configuration
// Defines port, timeouts, and other server-related settings
type ServerConfig struct {
	Port         int    `json:"port"`
	Host         string `json:"host"`
	ReadTimeout  int    `json:"read_timeout"`  // seconds
	WriteTimeout int    `json:"write_timeout"` // seconds
	Environment  string `json:"environment"`   // development, staging, production
}

// RedisConfig contains Redis database configuration
// Embeds the redis package Config for consistency
type RedisConfig struct {
	*redis.Config
}

// SessionConfig contains session management configuration
// Defines session duration and security settings
type SessionConfig struct {
	Duration int    `json:"duration"` // days
	SecretKey string `json:"secret_key"`
	Secure    bool   `json:"secure"`
	HTTPOnly  bool   `json:"http_only"`
}

// EmailConfig contains SMTP email configuration
// Used for sending registration confirmations and notifications
type EmailConfig struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	FromAddress  string `json:"from_address"`
	FromName     string `json:"from_name"`
}

// SecurityConfig contains security-related settings
// Defines rate limiting and other security measures
type SecurityConfig struct {
	RateLimit    int  `json:"rate_limit"`     // requests per minute per IP
	EnableCORS   bool `json:"enable_cors"`
	AllowedOrigins []string `json:"allowed_origins"`
}

// Load creates a new configuration from environment variables
// Uses sensible defaults when environment variables are not set
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			Host:         getEnv("SERVER_HOST", "localhost"),
			ReadTimeout:  getEnvAsInt("SERVER_READ_TIMEOUT", 30),
			WriteTimeout: getEnvAsInt("SERVER_WRITE_TIMEOUT", 30),
			Environment:  getEnv("ENVIRONMENT", "development"),
		},
		Redis: RedisConfig{
			Config: &redis.Config{
				Host:     getEnv("REDIS_HOST", "localhost"),
				Port:     getEnvAsInt("REDIS_PORT", 6379),
				Password: getEnv("REDIS_PASSWORD", ""),
				DB:       getEnvAsInt("REDIS_DB", 0),
				PoolSize: getEnvAsInt("REDIS_POOL_SIZE", 10),
			},
		},
		Session: SessionConfig{
			Duration:  getEnvAsInt("SESSION_DURATION", 7),
			SecretKey: getEnv("SESSION_SECRET", "your-secret-key-change-in-production"),
			Secure:    getEnvAsBool("SESSION_SECURE", false),
			HTTPOnly:  getEnvAsBool("SESSION_HTTP_ONLY", true),
		},
		Email: EmailConfig{
			SMTPHost:     getEnv("SMTP_HOST", "localhost"),
			SMTPPort:     getEnvAsInt("SMTP_PORT", 1025), // MailHog default
			SMTPUsername: getEnv("SMTP_USERNAME", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			FromAddress:  getEnv("EMAIL_FROM_ADDRESS", "noreply@no.reply.com"),
			FromName:     getEnv("EMAIL_FROM_NAME", "Task Tracker"),
		},
		Security: SecurityConfig{
			RateLimit:      getEnvAsInt("RATE_LIMIT", 1000),
			EnableCORS:     getEnvAsBool("ENABLE_CORS", true),
			AllowedOrigins: []string{getEnv("FRONTEND_URL", "http://localhost:3000")},
		},
	}
}

// IsDevelopment returns true if running in development mode
// Used to enable development-specific features and logging
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

// IsProduction returns true if running in production mode
// Used to enable production-specific optimizations and security
func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}

// getEnv gets an environment variable with a fallback default value
// Returns the default if the environment variable is not set
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer with a fallback default
// Returns the default if the environment variable is not set or invalid
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool gets an environment variable as boolean with a fallback default
// Returns the default if the environment variable is not set or invalid
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}