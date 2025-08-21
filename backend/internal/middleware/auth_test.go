package middleware

import (
	"backend/internal/domain"
	"backend/internal/mocks"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		sessionCookie      string
		mockSessionValid   bool
		mockUserID         string
		mockSessionError   error
		mockGetUserError   error
		mockUser           *domain.User
		expectedStatus     int
		expectedErrorCode  string
		expectUserInCtx    bool
	}{
		{
			name:               "Valid session with existing user",
			sessionCookie:      "valid-session-123",
			mockSessionValid:   true,
			mockUserID:         "user-123",
			mockSessionError:   nil,
			mockGetUserError:   nil,
			mockUser: &domain.User{
				ID:          "user-123",
				Email:       "test@example.com",
				DisplayName: "Test User",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			expectedStatus:  http.StatusOK,
			expectUserInCtx: true,
		},
		{
			name:               "Missing session cookie",
			sessionCookie:      "",
			mockSessionValid:   false,
			mockUserID:         "",
			mockSessionError:   nil,
			mockGetUserError:   nil,
			mockUser:           nil,
			expectedStatus:     http.StatusUnauthorized,
			expectedErrorCode:  "4001",
			expectUserInCtx:    false,
		},
		{
			name:               "Invalid session",
			sessionCookie:      "invalid-session",
			mockSessionValid:   false,
			mockUserID:         "",
			mockSessionError:   nil,
			mockGetUserError:   nil,
			mockUser:           nil,
			expectedStatus:     http.StatusUnauthorized,
			expectedErrorCode:  "4002",
			expectUserInCtx:    false,
		},
		{
			name:               "Session validation error",
			sessionCookie:      "error-session",
			mockSessionValid:   false,
			mockUserID:         "",
			mockSessionError:   errors.New("redis connection failed"),
			mockGetUserError:   nil,
			mockUser:           nil,
			expectedStatus:     http.StatusInternalServerError,
			expectedErrorCode:  "4003",
			expectUserInCtx:    false,
		},
		{
			name:               "Valid session but user not found",
			sessionCookie:      "orphaned-session",
			mockSessionValid:   true,
			mockUserID:         "missing-user-123",
			mockSessionError:   nil,
			mockGetUserError:   domain.ErrUserNotFound,
			mockUser:           nil,
			expectedStatus:     http.StatusUnauthorized,
			expectedErrorCode:  "4004",
			expectUserInCtx:    false,
		},
		{
			name:               "Valid session but user fetch error",
			sessionCookie:      "error-user-session",
			mockSessionValid:   true,
			mockUserID:         "error-user-123",
			mockSessionError:   nil,
			mockGetUserError:   errors.New("database error"),
			mockUser:           nil,
			expectedStatus:     http.StatusInternalServerError,
			expectedErrorCode:  "4005",
			expectUserInCtx:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock repository
			mockRepo := new(mocks.MockUserRepository)
			
			if tt.sessionCookie != "" {
				mockRepo.On("ValidateSession", tt.sessionCookie).Return(tt.mockSessionValid, tt.mockSessionError)
				
				if tt.mockSessionValid && tt.mockSessionError == nil {
					mockRepo.On("GetSessionUserID", tt.sessionCookie).Return(tt.mockUserID, nil)
					mockRepo.On("GetByID", tt.mockUserID).Return(tt.mockUser, tt.mockGetUserError)
				}
			}

			// Create middleware
			authMiddleware := AuthMiddleware(mockRepo)

			// Setup test request
			w := httptest.NewRecorder()
			_, router := gin.CreateTestContext(w)

			// Add test endpoint that checks if user is in context
			router.Use(authMiddleware)
			router.GET("/test", func(c *gin.Context) {
				if tt.expectUserInCtx {
					userID, exists := c.Get("userID")
					assert.True(t, exists)
					assert.Equal(t, tt.mockUserID, userID)
					
					user, exists := c.Get("user")
					assert.True(t, exists)
					assert.Equal(t, tt.mockUser, user)
				} else {
					_, exists := c.Get("userID")
					assert.False(t, exists)
					
					_, exists = c.Get("user")
					assert.False(t, exists)
				}
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Create request with session cookie if provided
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.sessionCookie != "" {
				req.AddCookie(&http.Cookie{
					Name:  "session",
					Value: tt.sessionCookie,
				})
			}

			// Execute request
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedErrorCode != "" {
				assert.Contains(t, w.Body.String(), tt.expectedErrorCode)
			}

			// Verify all expectations
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthMiddleware_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mock repository
	mockRepo := new(mocks.MockUserRepository)
	
	// Setup test data
	userID := "test-user-123"
	sessionID := "valid-session-456"
	user := &domain.User{
		ID:          userID,
		Email:       "integration@test.com",
		DisplayName: "Integration Test User",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Setup mock expectations
	mockRepo.On("ValidateSession", sessionID).Return(true, nil)
	mockRepo.On("GetSessionUserID", sessionID).Return(userID, nil)
	mockRepo.On("GetByID", userID).Return(user, nil)

	// Create middleware and router
	authMiddleware := AuthMiddleware(mockRepo)
	router := gin.New()
	router.Use(authMiddleware)

	// Add protected endpoint
	router.GET("/protected", func(c *gin.Context) {
		userID, exists := c.Get("userID")
		assert.True(t, exists)
		assert.Equal(t, "test-user-123", userID)

		userObj, exists := c.Get("user")
		assert.True(t, exists)
		assert.Equal(t, user, userObj)

		c.JSON(http.StatusOK, gin.H{
			"message": "Access granted",
			"user_id": userID,
		})
	})

	// Create request with valid session
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: sessionID,
	})

	// Execute request
	router.ServeHTTP(w, req)

	// Assert successful access
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Access granted")
	assert.Contains(t, w.Body.String(), userID)

	// Verify all expectations
	mockRepo.AssertExpectations(t)
}