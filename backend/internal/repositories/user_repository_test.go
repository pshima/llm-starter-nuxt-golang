package repositories

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"backend/internal/domain"
	"backend/pkg/redis"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
)

// setupTestRedis creates a miniredis instance for testing
// Returns a Redis client connected to the test instance
func setupTestRedis(t *testing.T) (*redis.Client, func()) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}

	port, err := strconv.Atoi(mr.Port())
	if err != nil {
		mr.Close()
		t.Fatalf("Failed to convert port: %v", err)
	}

	config := &redis.Config{
		Host:     mr.Host(),
		Port:     port,
		Password: "",
		DB:       0,
		PoolSize: 10,
	}

	client, err := redis.NewClient(config)
	if err != nil {
		mr.Close()
		t.Fatalf("Failed to create Redis client: %v", err)
	}

	cleanup := func() {
		client.Close()
		mr.Close()
	}

	return client, cleanup
}

// createTestUser creates a test user with valid data
// Returns a user struct suitable for testing
func createTestUser() *domain.User {
	user := &domain.User{
		ID:          uuid.New().String(),
		Email:       "test@example.com",
		DisplayName: "Test User",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err := user.HashPassword("Password123!")
	if err != nil {
		panic("Failed to hash password: " + err.Error())
	}
	return user
}

func TestUserRepository_CreateUser(t *testing.T) {
	tests := []struct {
		name        string
		user        *domain.User
		setupFunc   func(repo *UserRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name:    "successful user creation",
			user:    createTestUser(),
			wantErr: false,
		},
		{
			name: "duplicate email should fail",
			user: createTestUser(),
			setupFunc: func(repo *UserRepository) {
				// Pre-create a user with the same email
				existingUser := createTestUser()
				existingUser.ID = uuid.New().String() // Different ID but same email
				err := repo.Create(existingUser)
				if err != nil {
					t.Fatalf("Failed to setup test user: %v", err)
				}
			},
			wantErr:     true,
			expectedErr: domain.ErrUserAlreadyExists,
		},
		{
			name: "empty email should fail",
			user: func() *domain.User {
				user := createTestUser()
				user.Email = ""
				return user
			}(),
			wantErr: true,
		},
		{
			name: "empty display name should fail",
			user: func() *domain.User {
				user := createTestUser()
				user.DisplayName = ""
				return user
			}(),
			wantErr: true,
		},
		{
			name: "empty password should fail",
			user: func() *domain.User {
				user := createTestUser()
				user.Password = ""
				return user
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, cleanup := setupTestRedis(t)
			defer cleanup()

			repo := NewUserRepository(client)

			if tt.setupFunc != nil {
				tt.setupFunc(repo)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := repo.Create(tt.user)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.expectedErr != nil && err != tt.expectedErr {
					t.Errorf("Expected error %v but got %v", tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify user was stored correctly
			storedUser, err := repo.GetByID(tt.user.ID)
			if err != nil {
				t.Errorf("Failed to retrieve created user: %v", err)
				return
			}

			if storedUser.Email != tt.user.Email {
				t.Errorf("Expected email %s but got %s", tt.user.Email, storedUser.Email)
			}
			if storedUser.DisplayName != tt.user.DisplayName {
				t.Errorf("Expected display name %s but got %s", tt.user.DisplayName, storedUser.DisplayName)
			}

			// Verify email index was created
			userByEmail, err := repo.GetByEmail(tt.user.Email)
			if err != nil {
				t.Errorf("Failed to retrieve user by email: %v", err)
				return
			}
			if userByEmail.ID != tt.user.ID {
				t.Errorf("Expected user ID %s but got %s", tt.user.ID, userByEmail.ID)
			}

			// Test Redis key structure
			userKey := redis.GenerateKey(redis.UserKeyPrefix, tt.user.ID)
			exists := client.Exists(ctx, userKey).Val()
			if exists != 1 {
				t.Errorf("User key %s should exist in Redis", userKey)
			}

			emailKey := redis.GenerateKey("user:email", tt.user.Email)
			emailExists := client.Exists(ctx, emailKey).Val()
			if emailExists != 1 {
				t.Errorf("Email index key %s should exist in Redis", emailKey)
			}
		})
	}
}

func TestUserRepository_GetUserByID(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		setupFunc   func(repo *UserRepository) *domain.User
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "get existing user",
			userID: "",
			setupFunc: func(repo *UserRepository) *domain.User {
				user := createTestUser()
				err := repo.Create(user)
				if err != nil {
					t.Fatalf("Failed to create test user: %v", err)
				}
				return user
			},
			wantErr: false,
		},
		{
			name:        "get non-existent user",
			userID:      uuid.New().String(),
			wantErr:     true,
			expectedErr: domain.ErrUserNotFound,
		},
		{
			name:    "empty user ID",
			userID:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, cleanup := setupTestRedis(t)
			defer cleanup()

			repo := NewUserRepository(client)

			var expectedUser *domain.User
			if tt.setupFunc != nil {
				expectedUser = tt.setupFunc(repo)
				if tt.userID == "" && expectedUser != nil {
					tt.userID = expectedUser.ID
				}
			}

			user, err := repo.GetByID(tt.userID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.expectedErr != nil && err != tt.expectedErr {
					t.Errorf("Expected error %v but got %v", tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if expectedUser != nil {
				if user.ID != expectedUser.ID {
					t.Errorf("Expected ID %s but got %s", expectedUser.ID, user.ID)
				}
				if user.Email != expectedUser.Email {
					t.Errorf("Expected email %s but got %s", expectedUser.Email, user.Email)
				}
				if user.DisplayName != expectedUser.DisplayName {
					t.Errorf("Expected display name %s but got %s", expectedUser.DisplayName, user.DisplayName)
				}
			}
		})
	}
}

func TestUserRepository_GetUserByEmail(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		setupFunc   func(repo *UserRepository) *domain.User
		wantErr     bool
		expectedErr error
	}{
		{
			name:  "get existing user by email",
			email: "",
			setupFunc: func(repo *UserRepository) *domain.User {
				user := createTestUser()
				err := repo.Create(user)
				if err != nil {
					t.Fatalf("Failed to create test user: %v", err)
				}
				return user
			},
			wantErr: false,
		},
		{
			name:        "get non-existent user by email",
			email:       "nonexistent@example.com",
			wantErr:     true,
			expectedErr: domain.ErrUserNotFound,
		},
		{
			name:    "empty email",
			email:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, cleanup := setupTestRedis(t)
			defer cleanup()

			repo := NewUserRepository(client)

			var expectedUser *domain.User
			if tt.setupFunc != nil {
				expectedUser = tt.setupFunc(repo)
				if tt.email == "" && expectedUser != nil {
					tt.email = expectedUser.Email
				}
			}

			user, err := repo.GetByEmail(tt.email)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.expectedErr != nil && err != tt.expectedErr {
					t.Errorf("Expected error %v but got %v", tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if expectedUser != nil {
				if user.ID != expectedUser.ID {
					t.Errorf("Expected ID %s but got %s", expectedUser.ID, user.ID)
				}
				if user.Email != expectedUser.Email {
					t.Errorf("Expected email %s but got %s", expectedUser.Email, user.Email)
				}
				if user.DisplayName != expectedUser.DisplayName {
					t.Errorf("Expected display name %s but got %s", expectedUser.DisplayName, user.DisplayName)
				}
			}
		})
	}
}

func TestUserRepository_UpdateUser(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(repo *UserRepository) *domain.User
		updateFunc func(user *domain.User) *domain.User
		wantErr   bool
	}{
		{
			name: "successful update display name",
			setupFunc: func(repo *UserRepository) *domain.User {
				user := createTestUser()
				err := repo.Create(user)
				if err != nil {
					t.Fatalf("Failed to create test user: %v", err)
				}
				return user
			},
			updateFunc: func(user *domain.User) *domain.User {
				user.DisplayName = "Updated Name"
				user.UpdatedAt = time.Now()
				return user
			},
			wantErr: false,
		},
		{
			name: "successful password update",
			setupFunc: func(repo *UserRepository) *domain.User {
				user := createTestUser()
				err := repo.Create(user)
				if err != nil {
					t.Fatalf("Failed to create test user: %v", err)
				}
				return user
			},
			updateFunc: func(user *domain.User) *domain.User {
				err := user.HashPassword("NewPassword456!")
				if err != nil {
					t.Fatalf("Failed to hash new password: %v", err)
				}
				user.UpdatedAt = time.Now()
				return user
			},
			wantErr: false,
		},
		{
			name: "update non-existent user should fail",
			setupFunc: func(repo *UserRepository) *domain.User {
				return createTestUser() // Don't actually create in repo
			},
			updateFunc: func(user *domain.User) *domain.User {
				user.DisplayName = "Updated Name"
				return user
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, cleanup := setupTestRedis(t)
			defer cleanup()

			repo := NewUserRepository(client)

			user := tt.setupFunc(repo)
			updatedUser := tt.updateFunc(user)

			err := repo.Update(updatedUser)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify user was updated
			storedUser, err := repo.GetByID(updatedUser.ID)
			if err != nil {
				t.Errorf("Failed to retrieve updated user: %v", err)
				return
			}

			if storedUser.DisplayName != updatedUser.DisplayName {
				t.Errorf("Expected display name %s but got %s", updatedUser.DisplayName, storedUser.DisplayName)
			}

			// Check that UpdatedAt was actually updated
			if !storedUser.UpdatedAt.After(user.CreatedAt) {
				t.Errorf("UpdatedAt should be after CreatedAt")
			}
		})
	}
}

func TestUserRepository_PasswordHashing(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "Password123!",
			wantErr:  false,
		},
		{
			name:     "another valid password",
			password: "MySecret456@",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &domain.User{}
			err := user.HashPassword(tt.password)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify password was hashed (not stored as plain text)
			if user.Password == tt.password {
				t.Errorf("Password should be hashed, not stored as plain text")
			}

			// Verify password can be verified
			if !user.CheckPassword(tt.password) {
				t.Errorf("Password verification failed")
			}

			// Verify wrong password fails
			if user.CheckPassword("wrongpassword") {
				t.Errorf("Wrong password should not verify")
			}
		})
	}
}

func TestUserRepository_Sessions(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(repo *UserRepository) (*domain.User, string)
		testFunc    func(repo *UserRepository, userID, sessionID string) error
		description string
	}{
		{
			name: "create session",
			setupFunc: func(repo *UserRepository) (*domain.User, string) {
				user := createTestUser()
				err := repo.Create(user)
				if err != nil {
					t.Fatalf("Failed to create test user: %v", err)
				}
				sessionID := uuid.New().String()
				return user, sessionID
			},
			testFunc: func(repo *UserRepository, userID, sessionID string) error {
				return repo.CreateSession(userID, sessionID)
			},
			description: "should create session successfully",
		},
		{
			name: "validate existing session",
			setupFunc: func(repo *UserRepository) (*domain.User, string) {
				user := createTestUser()
				err := repo.Create(user)
				if err != nil {
					t.Fatalf("Failed to create test user: %v", err)
				}
				sessionID := uuid.New().String()
				err = repo.CreateSession(user.ID, sessionID)
				if err != nil {
					t.Fatalf("Failed to create session: %v", err)
				}
				return user, sessionID
			},
			testFunc: func(repo *UserRepository, userID, sessionID string) error {
				valid, err := repo.ValidateSession(sessionID)
				if err != nil {
					return err
				}
				if !valid {
					t.Errorf("Session should be valid")
				}
				return nil
			},
			description: "should validate existing session",
		},
		{
			name: "delete session",
			setupFunc: func(repo *UserRepository) (*domain.User, string) {
				user := createTestUser()
				err := repo.Create(user)
				if err != nil {
					t.Fatalf("Failed to create test user: %v", err)
				}
				sessionID := uuid.New().String()
				err = repo.CreateSession(user.ID, sessionID)
				if err != nil {
					t.Fatalf("Failed to create session: %v", err)
				}
				return user, sessionID
			},
			testFunc: func(repo *UserRepository, userID, sessionID string) error {
				err := repo.DeleteSession(sessionID)
				if err != nil {
					return err
				}
				
				// Verify session is no longer valid
				valid, err := repo.ValidateSession(sessionID)
				if err != nil {
					return err
				}
				if valid {
					t.Errorf("Session should not be valid after deletion")
				}
				return nil
			},
			description: "should delete session successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, cleanup := setupTestRedis(t)
			defer cleanup()

			repo := NewUserRepository(client)

			user, sessionID := tt.setupFunc(repo)
			err := tt.testFunc(repo, user.ID, sessionID)

			if err != nil {
				t.Errorf("Test failed: %v", err)
			}
		})
	}
}

func TestUserRepository_SessionTTL(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	repo := NewUserRepository(client)

	// Create a user
	user := createTestUser()
	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a session
	sessionID := uuid.New().String()
	err = repo.CreateSession(user.ID, sessionID)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Check TTL is set correctly (7 days = 604800 seconds)
	ctx := context.Background()
	sessionKey := redis.GenerateKey(redis.SessionKeyPrefix, sessionID)
	ttl := client.TTL(ctx, sessionKey).Val()

	expectedTTL := 7 * 24 * time.Hour
	tolerance := 1 * time.Minute // Allow for some timing variation

	if ttl < expectedTTL-tolerance || ttl > expectedTTL+tolerance {
		t.Errorf("Expected TTL around %v but got %v", expectedTTL, ttl)
	}
}

func TestUserRepository_ValidateNonExistentSession(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	repo := NewUserRepository(client)

	// Test validating a session that doesn't exist
	sessionID := uuid.New().String()
	valid, err := repo.ValidateSession(sessionID)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if valid {
		t.Errorf("Non-existent session should not be valid")
	}
}

func TestUserRepository_ErrorCases(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(repo *UserRepository) error
		wantErr  bool
	}{
		{
			name: "create session for non-existent user",
			testFunc: func(repo *UserRepository) error {
				sessionID := uuid.New().String()
				userID := uuid.New().String()
				return repo.CreateSession(userID, sessionID)
			},
			wantErr: false, // Should still create session even if user doesn't exist
		},
		{
			name: "delete non-existent session",
			testFunc: func(repo *UserRepository) error {
				sessionID := uuid.New().String()
				return repo.DeleteSession(sessionID)
			},
			wantErr: false, // Should not error when deleting non-existent session
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, cleanup := setupTestRedis(t)
			defer cleanup()

			repo := NewUserRepository(client)

			err := tt.testFunc(repo)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// Benchmark tests to ensure performance requirements
func BenchmarkUserRepository_CreateUser(b *testing.B) {
	client, cleanup := setupTestRedis(&testing.T{})
	defer cleanup()

	repo := NewUserRepository(client)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := &domain.User{
			ID:          uuid.New().String(),
			Email:       "bench" + uuid.New().String() + "@example.com",
			DisplayName: "Bench User",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		user.HashPassword("Password123!")
		repo.Create(user)
	}
}

func BenchmarkUserRepository_GetUserByID(b *testing.B) {
	client, cleanup := setupTestRedis(&testing.T{})
	defer cleanup()

	repo := NewUserRepository(client)

	// Setup: create a user
	user := createTestUser()
	repo.Create(user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.GetByID(user.ID)
	}
}

func BenchmarkUserRepository_ValidateSession(b *testing.B) {
	client, cleanup := setupTestRedis(&testing.T{})
	defer cleanup()

	repo := NewUserRepository(client)

	// Setup: create user and session
	user := createTestUser()
	repo.Create(user)
	sessionID := uuid.New().String()
	repo.CreateSession(user.ID, sessionID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.ValidateSession(sessionID)
	}
}

func TestUserRepository_Delete(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(repo *UserRepository) *domain.User
		userID    string
		wantErr   bool
	}{
		{
			name: "delete existing user",
			setupFunc: func(repo *UserRepository) *domain.User {
				user := createTestUser()
				err := repo.Create(user)
				if err != nil {
					t.Fatalf("Failed to create test user: %v", err)
				}
				return user
			},
			wantErr: false,
		},
		{
			name:    "delete non-existent user",
			userID:  uuid.New().String(),
			wantErr: true,
		},
		{
			name:    "delete with empty ID",
			userID:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, cleanup := setupTestRedis(t)
			defer cleanup()

			repo := NewUserRepository(client)

			var user *domain.User
			if tt.setupFunc != nil {
				user = tt.setupFunc(repo)
				if tt.userID == "" {
					tt.userID = user.ID
				}
			}

			err := repo.Delete(tt.userID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify user was deleted
			_, err = repo.GetByID(tt.userID)
			if err != domain.ErrUserNotFound {
				t.Errorf("Expected user to be deleted")
			}

			// Verify email index was deleted
			_, err = repo.GetByEmail(user.Email)
			if err != domain.ErrUserNotFound {
				t.Errorf("Expected email index to be deleted")
			}
		})
	}
}

func TestUserRepository_List(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(repo *UserRepository) []*domain.User
		limit     int
		offset    int
		wantCount int
	}{
		{
			name: "list users with pagination",
			setupFunc: func(repo *UserRepository) []*domain.User {
				users := make([]*domain.User, 3)
				for i := 0; i < 3; i++ {
					user := &domain.User{
						ID:          uuid.New().String(),
						Email:       "test" + uuid.New().String() + "@example.com",
						DisplayName: "Test User " + uuid.New().String(),
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					}
					user.HashPassword("Password123!")
					repo.Create(user)
					users[i] = user
				}
				return users
			},
			limit:     2,
			offset:    0,
			wantCount: 2,
		},
		{
			name: "list users with offset",
			setupFunc: func(repo *UserRepository) []*domain.User {
				users := make([]*domain.User, 3)
				for i := 0; i < 3; i++ {
					user := &domain.User{
						ID:          uuid.New().String(),
						Email:       "test" + uuid.New().String() + "@example.com",
						DisplayName: "Test User " + uuid.New().String(),
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					}
					user.HashPassword("Password123!")
					repo.Create(user)
					users[i] = user
				}
				return users
			},
			limit:     5,
			offset:    1,
			wantCount: 2,
		},
		{
			name:      "list with no users",
			limit:     10,
			offset:    0,
			wantCount: 0,
		},
		{
			name: "list with offset beyond range",
			setupFunc: func(repo *UserRepository) []*domain.User {
				user := createTestUser()
				repo.Create(user)
				return []*domain.User{user}
			},
			limit:     10,
			offset:    5,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, cleanup := setupTestRedis(t)
			defer cleanup()

			repo := NewUserRepository(client)

			if tt.setupFunc != nil {
				tt.setupFunc(repo)
			}

			users, err := repo.List(tt.limit, tt.offset)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(users) != tt.wantCount {
				t.Errorf("Expected %d users but got %d", tt.wantCount, len(users))
			}
		})
	}
}

func TestUserRepository_GetSessionUserID(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(repo *UserRepository) (string, string)
		sessionID string
		wantErr   bool
	}{
		{
			name: "get user ID from valid session",
			setupFunc: func(repo *UserRepository) (string, string) {
				user := createTestUser()
				repo.Create(user)
				sessionID := uuid.New().String()
				repo.CreateSession(user.ID, sessionID)
				return user.ID, sessionID
			},
			wantErr: false,
		},
		{
			name:      "get user ID from non-existent session",
			sessionID: uuid.New().String(),
			wantErr:   true,
		},
		{
			name:      "empty session ID",
			sessionID: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, cleanup := setupTestRedis(t)
			defer cleanup()

			repo := NewUserRepository(client)

			var expectedUserID string
			if tt.setupFunc != nil {
				userID, sessionID := tt.setupFunc(repo)
				expectedUserID = userID
				if tt.sessionID == "" {
					tt.sessionID = sessionID
				}
			}

			userID, err := repo.GetSessionUserID(tt.sessionID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if userID != expectedUserID {
				t.Errorf("Expected user ID %s but got %s", expectedUserID, userID)
			}
		})
	}
}

func TestUserRepository_DeleteAllUserSessions(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(repo *UserRepository) string
		userID    string
		wantErr   bool
	}{
		{
			name: "delete all sessions for user",
			setupFunc: func(repo *UserRepository) string {
				user := createTestUser()
				repo.Create(user)

				// Create multiple sessions for the user
				for i := 0; i < 3; i++ {
					sessionID := uuid.New().String()
					repo.CreateSession(user.ID, sessionID)
				}

				return user.ID
			},
			wantErr: false,
		},
		{
			name:   "delete sessions for user with no sessions",
			userID: uuid.New().String(),
			wantErr: false,
		},
		{
			name:    "empty user ID",
			userID:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, cleanup := setupTestRedis(t)
			defer cleanup()

			repo := NewUserRepository(client)

			if tt.setupFunc != nil {
				userID := tt.setupFunc(repo)
				if tt.userID == "" {
					tt.userID = userID
				}
			}

			err := repo.DeleteAllUserSessions(tt.userID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify no sessions exist for the user
			ctx := context.Background()
			pattern := redis.GenerateKey(redis.SessionKeyPrefix, "*")
			keys, _ := client.Keys(ctx, pattern).Result()

			for _, key := range keys {
				sessionData, err := client.Get(ctx, key).Result()
				if err != nil {
					continue
				}

				var session map[string]interface{}
				json.Unmarshal([]byte(sessionData), &session)
				if sessionUserID, ok := session["user_id"].(string); ok && sessionUserID == tt.userID {
					t.Errorf("Found session that should have been deleted: %s", key)
				}
			}
		})
	}
}

func TestUserRepository_NilUserHandling(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	repo := NewUserRepository(client)

	// Test Create with nil user
	err := repo.Create(nil)
	if err == nil {
		t.Errorf("Expected error when creating nil user")
	}

	// Test Update with nil user
	err = repo.Update(nil)
	if err == nil {
		t.Errorf("Expected error when updating nil user")
	}
}