package middleware

import (
	"backend/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UserRepository defines the interface needed for authentication middleware
// Provides session validation and user lookup functionality
type UserRepository interface {
	ValidateSession(sessionID string) (bool, error)
	GetSessionUserID(sessionID string) (string, error)
	GetByID(id string) (*domain.User, error)
}

// AuthMiddleware returns a middleware function that validates user sessions
// Checks for valid session cookie and adds user context to the request
func AuthMiddleware(userRepo UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get session cookie
		sessionCookie, err := c.Cookie("session")
		if err != nil || sessionCookie == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
				"code":  "4001",
			})
			c.Abort()
			return
		}

		// Validate session
		isValid, err := userRepo.ValidateSession(sessionCookie)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Session validation failed",
				"code":  "4003",
			})
			c.Abort()
			return
		}

		if !isValid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid session",
				"code":  "4002",
			})
			c.Abort()
			return
		}

		// Get user ID from session
		userID, err := userRepo.GetSessionUserID(sessionCookie)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve user from session",
				"code":  "4005",
			})
			c.Abort()
			return
		}

		// Get user details
		user, err := userRepo.GetByID(userID)
		if err != nil {
			if err == domain.ErrUserNotFound {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "User not found",
					"code":  "4004",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to retrieve user details",
					"code":  "4005",
				})
			}
			c.Abort()
			return
		}

		// Add user context to the request
		c.Set("userID", userID)
		c.Set("user", user)

		// Continue to next handler
		c.Next()
	}
}