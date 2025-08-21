package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"backend/internal/domain"
	"backend/pkg/redis"

	redislib "github.com/redis/go-redis/v9"
)

// UserRepository implements the domain.UserRepository interface
// Provides Redis-based storage for user data and session management
type UserRepository struct {
	client *redis.Client
}

// NewUserRepository creates a new UserRepository instance
// Takes a Redis client and returns a configured user repository
func NewUserRepository(client *redis.Client) *UserRepository {
	return &UserRepository{
		client: client,
	}
}

// Create stores a new user in Redis with email uniqueness validation
// Creates both user data and email index for efficient lookups
func (r *UserRepository) Create(user *domain.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	// Validate required fields
	if strings.TrimSpace(user.Email) == "" {
		return errors.New("email is required")
	}
	if strings.TrimSpace(user.DisplayName) == "" {
		return errors.New("display name is required")
	}
	if strings.TrimSpace(user.Password) == "" {
		return errors.New("password is required")
	}

	ctx := context.Background()

	// Check if email already exists
	emailKey := redis.GenerateKey("user:email", user.Email)
	exists := r.client.Exists(ctx, emailKey).Val()
	if exists > 0 {
		return domain.ErrUserAlreadyExists
	}

	// Create storage representation that includes password
	storageUser := struct {
		ID          string    `json:"id"`
		Email       string    `json:"email"`
		DisplayName string    `json:"display_name"`
		Password    string    `json:"password"` // Include password for storage
		IsAdmin     bool      `json:"is_admin"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Password:    user.Password,
		IsAdmin:     user.IsAdmin,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	// Serialize user data
	userData, err := json.Marshal(storageUser)
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	// Use Redis transaction to ensure atomicity
	pipe := r.client.TxPipeline()

	// Store user data
	userKey := redis.GenerateKey(redis.UserKeyPrefix, user.ID)
	pipe.Set(ctx, userKey, userData, 0)

	// Create email index
	pipe.Set(ctx, emailKey, user.ID, 0)

	// Execute transaction
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by their unique ID
// Returns the user data or ErrUserNotFound if not found
func (r *UserRepository) GetByID(id string) (*domain.User, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("user ID is required")
	}

	ctx := context.Background()
	userKey := redis.GenerateKey(redis.UserKeyPrefix, id)

	userData, err := r.client.Get(ctx, userKey).Result()
	if err != nil {
		if err == redislib.Nil {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Create storage representation for deserialization
	var storageUser struct {
		ID          string    `json:"id"`
		Email       string    `json:"email"`
		DisplayName string    `json:"display_name"`
		Password    string    `json:"password"`
		IsAdmin     bool      `json:"is_admin"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}
	
	err = json.Unmarshal([]byte(userData), &storageUser)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	// Convert back to domain user
	user := &domain.User{
		ID:          storageUser.ID,
		Email:       storageUser.Email,
		DisplayName: storageUser.DisplayName,
		Password:    storageUser.Password,
		IsAdmin:     storageUser.IsAdmin,
		CreatedAt:   storageUser.CreatedAt,
		UpdatedAt:   storageUser.UpdatedAt,
	}

	return user, nil
}

// GetByEmail retrieves a user by their email address
// Uses email index for efficient lookup
func (r *UserRepository) GetByEmail(email string) (*domain.User, error) {
	if strings.TrimSpace(email) == "" {
		return nil, errors.New("email is required")
	}

	ctx := context.Background()
	emailKey := redis.GenerateKey("user:email", email)

	userID, err := r.client.Get(ctx, emailKey).Result()
	if err != nil {
		if err == redislib.Nil {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return r.GetByID(userID)
}

// Update modifies an existing user's data
// Validates user exists before updating
func (r *UserRepository) Update(user *domain.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	// Check if user exists
	_, err := r.GetByID(user.ID)
	if err != nil {
		return err
	}

	// Create storage representation that includes password
	storageUser := struct {
		ID          string    `json:"id"`
		Email       string    `json:"email"`
		DisplayName string    `json:"display_name"`
		Password    string    `json:"password"` // Include password for storage
		IsAdmin     bool      `json:"is_admin"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Password:    user.Password,
		IsAdmin:     user.IsAdmin,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	// Serialize updated user data
	userData, err := json.Marshal(storageUser)
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	ctx := context.Background()
	userKey := redis.GenerateKey(redis.UserKeyPrefix, user.ID)

	err = r.client.Set(ctx, userKey, userData, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Delete removes a user and their associated data
// Cleans up user data and email index
func (r *UserRepository) Delete(id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("user ID is required")
	}

	// Get user to find email for cleanup
	user, err := r.GetByID(id)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Use transaction to ensure atomicity
	pipe := r.client.TxPipeline()

	// Delete user data
	userKey := redis.GenerateKey(redis.UserKeyPrefix, id)
	pipe.Del(ctx, userKey)

	// Delete email index
	emailKey := redis.GenerateKey("user:email", user.Email)
	pipe.Del(ctx, emailKey)

	// Execute transaction
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// List retrieves a paginated list of users
// Returns users with limit and offset for pagination
func (r *UserRepository) List(limit, offset int) ([]*domain.User, error) {
	// This is a simplified implementation
	// In a real application, you might use Redis sets or sorted sets for better pagination
	ctx := context.Background()

	// Get all user keys
	pattern := redis.GenerateKey(redis.UserKeyPrefix, "*")
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user keys: %w", err)
	}

	// Apply pagination
	start := offset
	end := offset + limit
	if start >= len(keys) {
		return []*domain.User{}, nil
	}
	if end > len(keys) {
		end = len(keys)
	}

	users := make([]*domain.User, 0, end-start)
	for i := start; i < end; i++ {
		userData, err := r.client.Get(ctx, keys[i]).Result()
		if err != nil {
			continue // Skip invalid entries
		}

		var user domain.User
		err = json.Unmarshal([]byte(userData), &user)
		if err != nil {
			continue // Skip invalid entries
		}

		users = append(users, &user)
	}

	return users, nil
}

// CreateSession creates a new user session with 7-day TTL
// Stores session data in Redis with automatic expiration
func (r *UserRepository) CreateSession(userID, sessionID string) error {
	if strings.TrimSpace(userID) == "" {
		return errors.New("user ID is required")
	}
	if strings.TrimSpace(sessionID) == "" {
		return errors.New("session ID is required")
	}

	ctx := context.Background()
	sessionKey := redis.GenerateKey(redis.SessionKeyPrefix, sessionID)

	// Create session data
	sessionData := map[string]interface{}{
		"user_id":    userID,
		"created_at": time.Now().Unix(),
	}

	sessionJSON, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	// Set session with 7-day TTL
	ttl := 7 * 24 * time.Hour
	err = r.client.Set(ctx, sessionKey, sessionJSON, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// ValidateSession checks if a session exists and is valid
// Returns true if session exists and hasn't expired
func (r *UserRepository) ValidateSession(sessionID string) (bool, error) {
	if strings.TrimSpace(sessionID) == "" {
		return false, errors.New("session ID is required")
	}

	ctx := context.Background()
	sessionKey := redis.GenerateKey(redis.SessionKeyPrefix, sessionID)

	exists := r.client.Exists(ctx, sessionKey).Val()
	return exists > 0, nil
}

// GetSessionUserID retrieves the user ID associated with a session
// Returns the user ID or error if session doesn't exist
func (r *UserRepository) GetSessionUserID(sessionID string) (string, error) {
	if strings.TrimSpace(sessionID) == "" {
		return "", errors.New("session ID is required")
	}

	ctx := context.Background()
	sessionKey := redis.GenerateKey(redis.SessionKeyPrefix, sessionID)

	sessionData, err := r.client.Get(ctx, sessionKey).Result()
	if err != nil {
		if err == redislib.Nil {
			return "", errors.New("session not found")
		}
		return "", fmt.Errorf("failed to get session: %w", err)
	}

	var session map[string]interface{}
	err = json.Unmarshal([]byte(sessionData), &session)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	userID, ok := session["user_id"].(string)
	if !ok {
		return "", errors.New("invalid session data")
	}

	return userID, nil
}

// DeleteSession removes a session from Redis
// Used for logout functionality
func (r *UserRepository) DeleteSession(sessionID string) error {
	if strings.TrimSpace(sessionID) == "" {
		return errors.New("session ID is required")
	}

	ctx := context.Background()
	sessionKey := redis.GenerateKey(redis.SessionKeyPrefix, sessionID)

	err := r.client.Del(ctx, sessionKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// DeleteAllUserSessions removes all sessions for a specific user
// Useful for logout from all devices functionality
func (r *UserRepository) DeleteAllUserSessions(userID string) error {
	if strings.TrimSpace(userID) == "" {
		return errors.New("user ID is required")
	}

	ctx := context.Background()

	// Find all sessions for this user
	pattern := redis.GenerateKey(redis.SessionKeyPrefix, "*")
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get session keys: %w", err)
	}

	// Check each session to see if it belongs to the user
	var sessionsToDelete []string
	for _, key := range keys {
		sessionData, err := r.client.Get(ctx, key).Result()
		if err != nil {
			continue // Skip invalid sessions
		}

		var session map[string]interface{}
		err = json.Unmarshal([]byte(sessionData), &session)
		if err != nil {
			continue // Skip invalid sessions
		}

		if sessionUserID, ok := session["user_id"].(string); ok && sessionUserID == userID {
			sessionsToDelete = append(sessionsToDelete, key)
		}
	}

	// Delete all matching sessions
	if len(sessionsToDelete) > 0 {
		err = r.client.Del(ctx, sessionsToDelete...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete user sessions: %w", err)
		}
	}

	return nil
}