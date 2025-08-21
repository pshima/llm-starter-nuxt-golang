package handlers

import (
	"backend/internal/domain"
	"backend/internal/mocks"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockResponse   *domain.User
		mockSessionID  string
		mockError      error
		expectedStatus int
		expectedCode   string
		checkCookie    bool
	}{
		{
			name: "Successful registration",
			requestBody: map[string]interface{}{
				"email":       "test@example.com",
				"password":    "Test123!",
				"displayName": "Test User",
			},
			mockResponse: &domain.User{
				ID:          "user-123",
				Email:       "test@example.com",
				DisplayName: "Test User",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockSessionID:  "session-123",
			mockError:      nil,
			expectedStatus: http.StatusCreated,
			checkCookie:    true,
		},
		{
			name: "Invalid JSON",
			requestBody: `{"email": "invalid-json"`,
			mockResponse:   nil,
			mockSessionID:  "",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "4006",
			checkCookie:    false,
		},
		{
			name: "Missing email",
			requestBody: map[string]interface{}{
				"password":    "Test123!",
				"displayName": "Test User",
			},
			mockResponse:   nil,
			mockSessionID:  "",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "4007",
			checkCookie:    false,
		},
		{
			name: "Missing password",
			requestBody: map[string]interface{}{
				"email":       "test@example.com",
				"displayName": "Test User",
			},
			mockResponse:   nil,
			mockSessionID:  "",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "4007",
			checkCookie:    false,
		},
		{
			name: "Missing display name",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "Test123!",
			},
			mockResponse:   nil,
			mockSessionID:  "",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "4007",
			checkCookie:    false,
		},
		{
			name: "Email already exists",
			requestBody: map[string]interface{}{
				"email":       "existing@example.com",
				"password":    "Test123!",
				"displayName": "Test User",
			},
			mockResponse:   nil,
			mockSessionID:  "",
			mockError:      domain.ErrUserAlreadyExists,
			expectedStatus: http.StatusConflict,
			expectedCode:   "4008",
			checkCookie:    false,
		},
		{
			name: "Service error",
			requestBody: map[string]interface{}{
				"email":       "test@example.com",
				"password":    "Test123!",
				"displayName": "Test User",
			},
			mockResponse:   nil,
			mockSessionID:  "",
			mockError:      errors.New("internal service error"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "4009",
			checkCookie:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(mocks.MockUserService)

			if tt.mockError == nil && tt.expectedStatus == http.StatusCreated {
				mockService.On("Register", "test@example.com", "Test User", "Test123!").
					Return(tt.mockResponse, tt.mockSessionID, nil)
			} else if tt.mockError != nil {
				// Extract email from the request body for proper mock setup
				if reqMap, ok := tt.requestBody.(map[string]interface{}); ok {
					email := reqMap["email"].(string)
					displayName := reqMap["displayName"].(string)
					password := reqMap["password"].(string)
					mockService.On("Register", email, displayName, password).
						Return((*domain.User)(nil), "", tt.mockError)
				}
			} else if tt.expectedStatus == http.StatusBadRequest && tt.expectedCode == "4007" {
				// Don't set up mock for validation errors
			}

			// Create handler
			handler := NewAuthHandler(mockService)

			// Setup request
			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			// Execute
			handler.Register(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedCode != "" {
				assert.Contains(t, w.Body.String(), tt.expectedCode)
			}

			if tt.checkCookie {
				cookies := w.Result().Cookies()
				var sessionCookie *http.Cookie
				for _, cookie := range cookies {
					if cookie.Name == "session" {
						sessionCookie = cookie
						break
					}
				}
				assert.NotNil(t, sessionCookie)
				assert.Equal(t, tt.mockSessionID, sessionCookie.Value)
				assert.True(t, sessionCookie.HttpOnly)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockResponse   *domain.User
		mockSessionID  string
		mockError      error
		expectedStatus int
		expectedCode   string
		checkCookie    bool
	}{
		{
			name: "Successful login",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "Test123!",
			},
			mockResponse: &domain.User{
				ID:          "user-123",
				Email:       "test@example.com",
				DisplayName: "Test User",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockSessionID:  "session-123",
			mockError:      nil,
			expectedStatus: http.StatusOK,
			checkCookie:    true,
		},
		{
			name: "Invalid credentials",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "wrongpassword",
			},
			mockResponse:   nil,
			mockSessionID:  "",
			mockError:      domain.ErrInvalidCredentials,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "4010",
			checkCookie:    false,
		},
		{
			name: "Missing email",
			requestBody: map[string]interface{}{
				"password": "Test123!",
			},
			mockResponse:   nil,
			mockSessionID:  "",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "4007",
			checkCookie:    false,
		},
		{
			name: "Invalid JSON",
			requestBody: `{"email": "invalid-json"`,
			mockResponse:   nil,
			mockSessionID:  "",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "4006",
			checkCookie:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(mocks.MockUserService)

			if tt.mockError == nil && tt.expectedStatus == http.StatusOK {
				mockService.On("Login", "test@example.com", "Test123!").
					Return(tt.mockResponse, tt.mockSessionID, nil)
			} else if tt.mockError != nil {
				// Extract email and password from the request body for proper mock setup
				if reqMap, ok := tt.requestBody.(map[string]interface{}); ok {
					email := reqMap["email"].(string)
					password := reqMap["password"].(string)
					mockService.On("Login", email, password).
						Return((*domain.User)(nil), "", tt.mockError)
				}
			}

			// Create handler
			handler := NewAuthHandler(mockService)

			// Setup request
			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			// Execute
			handler.Login(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedCode != "" {
				assert.Contains(t, w.Body.String(), tt.expectedCode)
			}

			if tt.checkCookie {
				cookies := w.Result().Cookies()
				var sessionCookie *http.Cookie
				for _, cookie := range cookies {
					if cookie.Name == "session" {
						sessionCookie = cookie
						break
					}
				}
				assert.NotNil(t, sessionCookie)
				assert.Equal(t, tt.mockSessionID, sessionCookie.Value)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		sessionCookie  string
		mockError      error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Successful logout",
			sessionCookie:  "valid-session-123",
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing session cookie",
			sessionCookie:  "",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "4011",
		},
		{
			name:           "Service error during logout",
			sessionCookie:  "valid-session-123",
			mockError:      errors.New("logout failed"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "4012",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(mocks.MockUserService)

			if tt.sessionCookie != "" {
				mockService.On("Logout", tt.sessionCookie).Return(tt.mockError)
			}

			// Create handler
			handler := NewAuthHandler(mockService)

			// Setup request
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest("POST", "/auth/logout", nil)
			if tt.sessionCookie != "" {
				req.AddCookie(&http.Cookie{
					Name:  "session",
					Value: tt.sessionCookie,
				})
			}
			c.Request = req

			// Execute
			handler.Logout(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedCode != "" {
				assert.Contains(t, w.Body.String(), tt.expectedCode)
			}

			// Check that cookie is cleared on successful logout
			if tt.expectedStatus == http.StatusOK {
				cookies := w.Result().Cookies()
				var sessionCookie *http.Cookie
				for _, cookie := range cookies {
					if cookie.Name == "session" {
						sessionCookie = cookie
						break
					}
				}
				assert.NotNil(t, sessionCookie)
				assert.Equal(t, "", sessionCookie.Value)
				assert.Equal(t, -1, sessionCookie.MaxAge)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Me(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		sessionCookie  string
		mockResponse   *domain.User
		mockError      error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:          "Successful get current user",
			sessionCookie: "valid-session-123",
			mockResponse: &domain.User{
				ID:          "user-123",
				Email:       "test@example.com",
				DisplayName: "Test User",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing session cookie",
			sessionCookie:  "",
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "4013",
		},
		{
			name:           "Service error",
			sessionCookie:  "valid-session-123",
			mockResponse:   nil,
			mockError:      errors.New("user fetch failed"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "4014",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(mocks.MockUserService)

			if tt.sessionCookie != "" {
				mockService.On("GetCurrentUser", tt.sessionCookie).Return(tt.mockResponse, tt.mockError)
			}

			// Create handler
			handler := NewAuthHandler(mockService)

			// Setup request
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest("GET", "/auth/me", nil)
			if tt.sessionCookie != "" {
				req.AddCookie(&http.Cookie{
					Name:  "session",
					Value: tt.sessionCookie,
				})
			}
			c.Request = req

			// Execute
			handler.Me(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedCode != "" {
				assert.Contains(t, w.Body.String(), tt.expectedCode)
			}

			if tt.expectedStatus == http.StatusOK && tt.mockResponse != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResponse.ID, response["id"])
				assert.Equal(t, tt.mockResponse.Email, response["email"])
				assert.Equal(t, tt.mockResponse.DisplayName, response["displayName"])
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}