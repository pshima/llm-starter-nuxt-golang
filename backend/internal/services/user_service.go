package services

import (
	"backend/internal/domain"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// UserServiceInterface defines the interface for user business logic operations
// Contains all business rules and validation logic for user operations
type UserServiceInterface interface {
	Register(email, displayName, password string) (*domain.User, string, error)
	Login(email, password string) (*domain.User, string, error)
	Logout(sessionID string) error
	GetCurrentUser(sessionID string) (*domain.User, error)
}

// UserService implements user business logic operations
// Handles user registration, authentication, and session management
type UserService struct {
	userRepo UserRepository
}

// UserRepository defines the methods needed from the user repository
// This interface ensures loose coupling between service and repository layers
type UserRepository interface {
	Create(user *domain.User) error
	GetByID(id string) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	Update(user *domain.User) error
	Delete(id string) error
	List(limit, offset int) ([]*domain.User, error)
	CreateSession(userID, sessionID string) error
	ValidateSession(sessionID string) (bool, error)
	GetSessionUserID(sessionID string) (string, error)
	DeleteSession(sessionID string) error
	DeleteAllUserSessions(userID string) error
}

// NewUserService creates a new instance of UserService
// Initializes the service with the provided user repository
func NewUserService(userRepo UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// Register creates a new user account with validation and session creation
// Validates email format, display name, and password requirements before creating user
func (s *UserService) Register(email, displayName, password string) (*domain.User, string, error) {
	// Error code 3001: Invalid email format
	if err := s.validateEmail(email); err != nil {
		return nil, "", domain.ErrInvalidEmail
	}

	// Error code 3002: Display name validation
	if err := s.validateDisplayName(displayName); err != nil {
		return nil, "", fmt.Errorf("3002: %w", err)
	}

	// Error code 3003: Password requirements validation
	if err := domain.ValidatePassword(password); err != nil {
		return nil, "", domain.ErrWeakPassword
	}

	// Check if user already exists (only after all input validation passes)
	existingUser, err := s.userRepo.GetByEmail(email)
	if err != nil && err != domain.ErrUserNotFound {
		return nil, "", fmt.Errorf("3004: failed to check user existence: %w", err)
	}
	if existingUser != nil {
		return nil, "", domain.ErrUserAlreadyExists
	}

	// Create new user
	user := &domain.User{
		ID:          uuid.New().String(),
		Email:       email,
		DisplayName: displayName,
		IsAdmin:     false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Hash password
	if err := user.HashPassword(password); err != nil {
		return nil, "", fmt.Errorf("3006: failed to hash password: %w", err)
	}

	// Save user to repository
	if err := s.userRepo.Create(user); err != nil {
		return nil, "", fmt.Errorf("3007: failed to create user: %w", err)
	}

	// Create session
	sessionID := uuid.New().String()
	if err := s.userRepo.CreateSession(user.ID, sessionID); err != nil {
		return nil, "", fmt.Errorf("3008: failed to create session: %w", err)
	}

	return user, sessionID, nil
}

// Login authenticates a user and creates a new session
// Validates credentials and creates a 7-day session upon successful authentication
func (s *UserService) Login(email, password string) (*domain.User, string, error) {
	// Error code 3009: Required fields validation
	if strings.TrimSpace(email) == "" || strings.TrimSpace(password) == "" {
		return nil, "", fmt.Errorf("3009: email and password are required")
	}

	// Get user by email
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, "", domain.ErrInvalidCredentials
		}
		return nil, "", fmt.Errorf("3004: failed to get user: %w", err)
	}

	// Check password
	if !user.CheckPassword(password) {
		return nil, "", domain.ErrInvalidCredentials
	}

	// Create session
	sessionID := uuid.New().String()
	if err := s.userRepo.CreateSession(user.ID, sessionID); err != nil {
		return nil, "", fmt.Errorf("3008: failed to create session: %w", err)
	}

	return user, sessionID, nil
}

// Logout terminates a user session
// Removes the session from storage to prevent further authentication
func (s *UserService) Logout(sessionID string) error {
	// Error code 3009: Required fields validation
	if strings.TrimSpace(sessionID) == "" {
		return fmt.Errorf("3009: session ID is required")
	}

	// Delete session
	if err := s.userRepo.DeleteSession(sessionID); err != nil {
		return fmt.Errorf("3004: failed to delete session: %w", err)
	}

	return nil
}

// GetCurrentUser retrieves the user associated with a session
// Validates the session and returns the corresponding user information
func (s *UserService) GetCurrentUser(sessionID string) (*domain.User, error) {
	// Error code 3009: Required fields validation
	if strings.TrimSpace(sessionID) == "" {
		return nil, fmt.Errorf("3009: session ID is required")
	}

	// Get user ID from session
	userID, err := s.userRepo.GetSessionUserID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("3010: invalid session: %w", err)
	}

	// Get user by ID
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, fmt.Errorf("3010: user not found")
		}
		return nil, fmt.Errorf("3004: failed to get user: %w", err)
	}

	return user, nil
}

// validateEmail checks if the email format is valid
// Uses regex to validate email format according to basic email rules
func (s *UserService) validateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return errors.New("invalid email format")
	}

	// Basic email validation regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}

	return nil
}

// validateDisplayName checks if the display name meets requirements
// Display name must not be empty and must not exceed 255 characters
func (s *UserService) validateDisplayName(displayName string) error {
	trimmed := strings.TrimSpace(displayName)
	if trimmed == "" {
		return errors.New("display name cannot be empty")
	}

	if len(trimmed) > 255 {
		return errors.New("display name cannot exceed 255 characters")
	}

	return nil
}