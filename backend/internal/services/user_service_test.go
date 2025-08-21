package services

import (
	"backend/internal/domain"
	"backend/internal/mocks"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUserService_Register(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		displayName   string
		password      string
		setupMock     func(*mocks.MockUserRepository)
		wantErr       bool
		expectedError string
		validateUser  func(*testing.T, *domain.User)
	}{
		{
			name:        "successful registration",
			email:       "test@example.com",
			displayName: "Test User",
			password:    "Password123!",
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				// Should check if user exists by email
				mockRepo.On("GetByEmail", "test@example.com").Return(nil, domain.ErrUserNotFound)
				// Should create the user
				mockRepo.On("Create", mock.MatchedBy(func(user *domain.User) bool {
					return user.Email == "test@example.com" &&
						user.DisplayName == "Test User" &&
						user.Password != "" && // Password should be hashed
						user.ID != "" &&
						!user.IsAdmin
				})).Return(nil)
				// Should create a session
				mockRepo.On("CreateSession", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
			},
			wantErr: false,
			validateUser: func(t *testing.T, user *domain.User) {
				assert.Equal(t, "test@example.com", user.Email)
				assert.Equal(t, "Test User", user.DisplayName)
				assert.NotEmpty(t, user.ID)
				assert.False(t, user.IsAdmin)
				assert.NotEmpty(t, user.Password)
				assert.True(t, user.CheckPassword("Password123!"))
			},
		},
		{
			name:          "invalid email format",
			email:         "invalid-email",
			displayName:   "Test User",
			password:      "Password123!",
			setupMock:     func(mockRepo *mocks.MockUserRepository) {},
			wantErr:       true,
			expectedError: "invalid email format",
		},
		{
			name:          "empty email",
			email:         "",
			displayName:   "Test User",
			password:      "Password123!",
			setupMock:     func(mockRepo *mocks.MockUserRepository) {},
			wantErr:       true,
			expectedError: "invalid email format",
		},
		{
			name:          "empty display name",
			email:         "test@example.com",
			displayName:   "",
			password:      "Password123!",
			setupMock:     func(mockRepo *mocks.MockUserRepository) {},
			wantErr:       true,
			expectedError: "display name cannot be empty",
		},
		{
			name:          "display name too long",
			email:         "test@example.com",
			displayName:   strings.Repeat("a", 256),
			password:      "Password123!",
			setupMock:     func(mockRepo *mocks.MockUserRepository) {},
			wantErr:       true,
			expectedError: "display name cannot exceed 255 characters",
		},
		{
			name:          "weak password - too short",
			email:         "test@example.com",
			displayName:   "Test User",
			password:      "P1!",
			setupMock:     func(mockRepo *mocks.MockUserRepository) {},
			wantErr:       true,
			expectedError: "password does not meet requirements",
		},
		{
			name:          "weak password - no special character",
			email:         "test@example.com",
			displayName:   "Test User",
			password:      "Password123",
			setupMock:     func(mockRepo *mocks.MockUserRepository) {},
			wantErr:       true,
			expectedError: "password does not meet requirements",
		},
		{
			name:          "weak password - no number",
			email:         "test@example.com",
			displayName:   "Test User",
			password:      "Password!",
			setupMock:     func(mockRepo *mocks.MockUserRepository) {},
			wantErr:       true,
			expectedError: "password does not meet requirements",
		},
		{
			name:        "user already exists",
			email:       "test@example.com",
			displayName: "Test User",
			password:    "Password123!",
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				existingUser := &domain.User{
					ID:          uuid.New().String(),
					Email:       "test@example.com",
					DisplayName: "Existing User",
				}
				mockRepo.On("GetByEmail", "test@example.com").Return(existingUser, nil)
			},
			wantErr:       true,
			expectedError: "user already exists",
		},
		{
			name:        "repository create error",
			email:       "test@example.com",
			displayName: "Test User",
			password:    "Password123!",
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("GetByEmail", "test@example.com").Return(nil, domain.ErrUserNotFound)
				mockRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(errors.New("database error"))
			},
			wantErr:       true,
			expectedError: "failed to create user",
		},
		{
			name:        "session creation error",
			email:       "test@example.com",
			displayName: "Test User",
			password:    "Password123!",
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("GetByEmail", "test@example.com").Return(nil, domain.ErrUserNotFound)
				mockRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(nil)
				mockRepo.On("CreateSession", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(errors.New("session error"))
			},
			wantErr:       true,
			expectedError: "failed to create session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockUserRepository(t)
			tt.setupMock(mockRepo)

			service := NewUserService(mockRepo)
			user, sessionID, err := service.Register(tt.email, tt.displayName, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, user)
				assert.Empty(t, sessionID)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.NotEmpty(t, sessionID)
				if tt.validateUser != nil {
					tt.validateUser(t, user)
				}
			}
		})
	}
}

func TestUserService_Login(t *testing.T) {
	// Create a test user with hashed password
	testUser := &domain.User{
		ID:          uuid.New().String(),
		Email:       "test@example.com",
		DisplayName: "Test User",
		IsAdmin:     false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	testUser.HashPassword("Password123!")

	tests := []struct {
		name          string
		email         string
		password      string
		setupMock     func(*mocks.MockUserRepository)
		wantErr       bool
		expectedError string
		validateUser  func(*testing.T, *domain.User)
	}{
		{
			name:     "successful login",
			email:    "test@example.com",
			password: "Password123!",
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("GetByEmail", "test@example.com").Return(testUser, nil)
				mockRepo.On("CreateSession", testUser.ID, mock.AnythingOfType("string")).Return(nil)
			},
			wantErr: false,
			validateUser: func(t *testing.T, user *domain.User) {
				assert.Equal(t, testUser.ID, user.ID)
				assert.Equal(t, testUser.Email, user.Email)
				assert.Equal(t, testUser.DisplayName, user.DisplayName)
			},
		},
		{
			name:          "empty email",
			email:         "",
			password:      "Password123!",
			setupMock:     func(mockRepo *mocks.MockUserRepository) {},
			wantErr:       true,
			expectedError: "email and password are required",
		},
		{
			name:          "empty password",
			email:         "test@example.com",
			password:      "",
			setupMock:     func(mockRepo *mocks.MockUserRepository) {},
			wantErr:       true,
			expectedError: "email and password are required",
		},
		{
			name:     "user not found",
			email:    "nonexistent@example.com",
			password: "Password123!",
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("GetByEmail", "nonexistent@example.com").Return(nil, domain.ErrUserNotFound)
			},
			wantErr:       true,
			expectedError: "invalid credentials",
		},
		{
			name:     "incorrect password",
			email:    "test@example.com",
			password: "WrongPassword!",
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("GetByEmail", "test@example.com").Return(testUser, nil)
			},
			wantErr:       true,
			expectedError: "invalid credentials",
		},
		{
			name:     "repository error",
			email:    "test@example.com",
			password: "Password123!",
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("GetByEmail", "test@example.com").Return(nil, errors.New("database error"))
			},
			wantErr:       true,
			expectedError: "failed to get user",
		},
		{
			name:     "session creation error",
			email:    "test@example.com",
			password: "Password123!",
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("GetByEmail", "test@example.com").Return(testUser, nil)
				mockRepo.On("CreateSession", testUser.ID, mock.AnythingOfType("string")).Return(errors.New("session error"))
			},
			wantErr:       true,
			expectedError: "failed to create session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockUserRepository(t)
			tt.setupMock(mockRepo)

			service := NewUserService(mockRepo)
			user, sessionID, err := service.Login(tt.email, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, user)
				assert.Empty(t, sessionID)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.NotEmpty(t, sessionID)
				if tt.validateUser != nil {
					tt.validateUser(t, user)
				}
			}
		})
	}
}

func TestUserService_Logout(t *testing.T) {
	tests := []struct {
		name          string
		sessionID     string
		setupMock     func(*mocks.MockUserRepository)
		wantErr       bool
		expectedError string
	}{
		{
			name:      "successful logout",
			sessionID: uuid.New().String(),
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("DeleteSession", mock.AnythingOfType("string")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:          "empty session ID",
			sessionID:     "",
			setupMock:     func(mockRepo *mocks.MockUserRepository) {},
			wantErr:       true,
			expectedError: "session ID is required",
		},
		{
			name:      "repository error",
			sessionID: uuid.New().String(),
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("DeleteSession", mock.AnythingOfType("string")).Return(errors.New("database error"))
			},
			wantErr:       true,
			expectedError: "failed to delete session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockUserRepository(t)
			tt.setupMock(mockRepo)

			service := NewUserService(mockRepo)
			err := service.Logout(tt.sessionID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserService_GetCurrentUser(t *testing.T) {
	testUser := &domain.User{
		ID:          uuid.New().String(),
		Email:       "test@example.com",
		DisplayName: "Test User",
		IsAdmin:     false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tests := []struct {
		name          string
		sessionID     string
		setupMock     func(*mocks.MockUserRepository)
		wantErr       bool
		expectedError string
		validateUser  func(*testing.T, *domain.User)
	}{
		{
			name:      "successful get current user",
			sessionID: uuid.New().String(),
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("GetSessionUserID", mock.AnythingOfType("string")).Return(testUser.ID, nil)
				mockRepo.On("GetByID", testUser.ID).Return(testUser, nil)
			},
			wantErr: false,
			validateUser: func(t *testing.T, user *domain.User) {
				assert.Equal(t, testUser.ID, user.ID)
				assert.Equal(t, testUser.Email, user.Email)
				assert.Equal(t, testUser.DisplayName, user.DisplayName)
			},
		},
		{
			name:          "empty session ID",
			sessionID:     "",
			setupMock:     func(mockRepo *mocks.MockUserRepository) {},
			wantErr:       true,
			expectedError: "session ID is required",
		},
		{
			name:      "session not found",
			sessionID: uuid.New().String(),
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("GetSessionUserID", mock.AnythingOfType("string")).Return("", errors.New("session not found"))
			},
			wantErr:       true,
			expectedError: "invalid session",
		},
		{
			name:      "user not found",
			sessionID: uuid.New().String(),
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("GetSessionUserID", mock.AnythingOfType("string")).Return(testUser.ID, nil)
				mockRepo.On("GetByID", testUser.ID).Return(nil, domain.ErrUserNotFound)
			},
			wantErr:       true,
			expectedError: "user not found",
		},
		{
			name:      "repository error getting user",
			sessionID: uuid.New().String(),
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("GetSessionUserID", mock.AnythingOfType("string")).Return(testUser.ID, nil)
				mockRepo.On("GetByID", testUser.ID).Return(nil, errors.New("database error"))
			},
			wantErr:       true,
			expectedError: "failed to get user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockUserRepository(t)
			tt.setupMock(mockRepo)

			service := NewUserService(mockRepo)
			user, err := service.GetCurrentUser(tt.sessionID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				if tt.validateUser != nil {
					tt.validateUser(t, user)
				}
			}
		})
	}
}

func TestUserService_ValidateEmail(t *testing.T) {
	service := &UserService{}

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid email", "test@example.com", false},
		{"valid email with subdomain", "user@mail.example.com", false},
		{"valid email with plus", "test+tag@example.com", false},
		{"valid email with dash", "test-user@example.com", false},
		{"empty email", "", true},
		{"email without @", "testexample.com", true},
		{"email without domain", "test@", true},
		{"email without local part", "@example.com", true},
		{"email with spaces", "test @example.com", true},
		{"email with invalid characters", "test<>@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateEmail(tt.email)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserService_ErrorCodes(t *testing.T) {
	// Test that the service returns proper error codes
	mockRepo := mocks.NewMockUserRepository(t)
	service := NewUserService(mockRepo)

	// Test error code 3001 - Invalid email format
	_, _, err := service.Register("invalid-email", "Test User", "Password123!")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "3001")

	// Test error code 3002 - Display name validation
	_, _, err = service.Register("test@example.com", "", "Password123!")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "3002")

	// Test error code 3003 - Password requirements
	_, _, err = service.Register("test@example.com", "Test User", "weak")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "3003")
}

func BenchmarkUserService_Register(b *testing.B) {
	mockRepo := mocks.NewMockUserRepository(b)
	mockRepo.On("GetByEmail", mock.AnythingOfType("string")).Return(nil, domain.ErrUserNotFound)
	mockRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(nil)
	mockRepo.On("CreateSession", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)

	service := NewUserService(mockRepo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		email := fmt.Sprintf("test%d@example.com", i)
		service.Register(email, "Test User", "Password123!")
	}
}

func BenchmarkUserService_Login(b *testing.B) {
	testUser := &domain.User{
		ID:          uuid.New().String(),
		Email:       "test@example.com",
		DisplayName: "Test User",
	}
	testUser.HashPassword("Password123!")

	mockRepo := mocks.NewMockUserRepository(b)
	mockRepo.On("GetByEmail", "test@example.com").Return(testUser, nil)
	mockRepo.On("CreateSession", testUser.ID, mock.AnythingOfType("string")).Return(nil)

	service := NewUserService(mockRepo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.Login("test@example.com", "Password123!")
	}
}