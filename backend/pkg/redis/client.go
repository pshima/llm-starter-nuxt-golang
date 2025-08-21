package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps the Redis client with additional functionality
// Provides connection management and health checking capabilities
type Client struct {
	*redis.Client
	config *Config
}

// Config holds Redis connection configuration
// Contains all settings needed to establish and maintain Redis connection
type Config struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
	PoolSize int    `json:"pool_size"`
}

// DefaultConfig returns a default Redis configuration
// Uses standard Redis settings suitable for local development
func DefaultConfig() *Config {
	return &Config{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
		PoolSize: 10,
	}
}

// NewClient creates a new Redis client with the provided configuration
// Establishes connection and performs initial health check
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
		PoolSize: config.PoolSize,
	})

	client := &Client{
		Client: rdb,
		config: config,
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}

// Close closes the Redis connection
// Should be called when shutting down the application
func (c *Client) Close() error {
	return c.Client.Close()
}

// HealthCheck performs a health check on the Redis connection
// Returns nil if Redis is responding, error otherwise
func (c *Client) HealthCheck(ctx context.Context) error {
	return c.Ping(ctx).Err()
}

// GetConfig returns the current Redis configuration
// Useful for debugging and monitoring purposes
func (c *Client) GetConfig() *Config {
	return c.config
}

// GenerateKey creates a Redis key with the given prefix and identifier
// Provides consistent key naming across the application
func GenerateKey(prefix, id string) string {
	return fmt.Sprintf("%s:%s", prefix, id)
}

// GenerateSetKey creates a Redis key for a set with the given prefix
// Used for maintaining collections of related items
func GenerateSetKey(prefix string) string {
	return fmt.Sprintf("%s:set", prefix)
}

// Redis key prefixes for different domain objects
const (
	UserKeyPrefix    = "user"
	TaskKeyPrefix    = "task"
	SessionKeyPrefix = "session"
	AdminKeyPrefix   = "admin"
)