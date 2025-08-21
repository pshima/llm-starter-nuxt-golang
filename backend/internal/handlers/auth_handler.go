package handlers

import (
	"backend/internal/domain"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// UserService defines the interface for user business logic operations
// Contains methods needed for authentication handlers
type UserService interface {
	Register(email, displayName, password string) (*domain.User, string, error)
	Login(email, password string) (*domain.User, string, error)
	Logout(sessionID string) error
	GetCurrentUser(sessionID string) (*domain.User, error)
}

// AuthHandler handles authentication-related HTTP requests
// Provides endpoints for user registration, login, logout, and profile retrieval
type AuthHandler struct {
	userService UserService
}

// NewAuthHandler creates a new instance of AuthHandler
// Initializes the handler with the provided user service
func NewAuthHandler(userService UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// RegisterRequest represents the request payload for user registration
type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
}

// LoginRequest represents the request payload for user login
type LoginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	RememberMe bool   `json:"rememberMe"`
}

// UserResponse represents the response payload for user data
type UserResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Register handles user registration requests
// Creates a new user account and starts a session
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest

	// Parse JSON first without binding validation
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON format",
			"code":  "4006",
		})
		return
	}

	// Validate required fields explicitly
	if req.Email == "" || req.Password == "" || req.DisplayName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email, password, and display name are required",
			"code":  "4007",
		})
		return
	}

	// Call service to register user
	user, sessionID, err := h.userService.Register(req.Email, req.DisplayName, req.Password)
	if err != nil {
		if err == domain.ErrUserAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Email already exists",
				"code":  "4008",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Registration failed",
				"code":  "4009",
			})
		}
		return
	}

	// Set session cookie
	h.setSessionCookie(c, sessionID)

	// Return user data
	c.JSON(http.StatusCreated, &UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	})
}

// Login handles user login requests
// Authenticates user and creates a session
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	// Parse JSON first without binding validation
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON format",
			"code":  "4006",
		})
		return
	}

	// Validate required fields explicitly
	if req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email and password are required",
			"code":  "4007",
		})
		return
	}

	// Call service to authenticate user
	user, sessionID, err := h.userService.Login(req.Email, req.Password)
	if err != nil {
		if err == domain.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid credentials",
				"code":  "4010",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Login failed",
				"code":  "4009",
			})
		}
		return
	}

	// Set session cookie
	h.setSessionCookie(c, sessionID)

	// Return user data
	c.JSON(http.StatusOK, &UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	})
}

// Logout handles user logout requests
// Invalidates the current session
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get session cookie
	sessionID, err := c.Cookie("session")
	if err != nil || sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No active session found",
			"code":  "4011",
		})
		return
	}

	// Call service to logout user
	if err := h.userService.Logout(sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Logout failed",
			"code":  "4012",
		})
		return
	}

	// Clear session cookie
	c.SetCookie("session", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// Me handles requests to get current user information
// Returns the profile of the authenticated user
func (h *AuthHandler) Me(c *gin.Context) {
	// Get session cookie
	sessionID, err := c.Cookie("session")
	if err != nil || sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No active session found",
			"code":  "4013",
		})
		return
	}

	// Call service to get current user
	user, err := h.userService.GetCurrentUser(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve user information",
			"code":  "4014",
		})
		return
	}

	// Return user data
	c.JSON(http.StatusOK, &UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	})
}

// setSessionCookie sets the session cookie with proper security settings
// Configures cookie for 7-day expiration with security flags
func (h *AuthHandler) setSessionCookie(c *gin.Context, sessionID string) {
	// Set session cookie for 7 days (604800 seconds)
	// Note: SameSite=Lax is the default for modern browsers
	c.SetCookie(
		"session",      // name
		sessionID,      // value
		604800,         // maxAge (7 days in seconds)
		"/",            // path
		"",             // domain (empty for current domain)
		false,          // secure (set to true in production with HTTPS)
		true,           // httpOnly
	)
}