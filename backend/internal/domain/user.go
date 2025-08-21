package domain

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the task tracker system
// Stores user authentication and profile information
type User struct {
	ID          string    `json:"id" redis:"id"`
	Email       string    `json:"email" redis:"email"`
	DisplayName string    `json:"display_name" redis:"display_name"`
	Password    string    `json:"-" redis:"password"` // Never include in JSON responses
	IsAdmin     bool      `json:"is_admin" redis:"is_admin"`
	CreatedAt   time.Time `json:"created_at" redis:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" redis:"updated_at"`
}

// UserRepository defines the interface for user data access operations
// This interface allows for different storage implementations while maintaining clean architecture
type UserRepository interface {
	Create(user *User) error
	GetByID(id string) (*User, error)
	GetByEmail(email string) (*User, error)
	Update(user *User) error
	Delete(id string) error
	List(limit, offset int) ([]*User, error)
}

// UserService defines the interface for user business logic operations
// Contains all business rules and validation logic for user operations
type UserService interface {
	Register(email, displayName, password string) (*User, error)
	Login(email, password string) (*User, error)
	GetProfile(userID string) (*User, error)
	UpdateProfile(userID, displayName string) (*User, error)
	ChangePassword(userID, oldPassword, newPassword string) error
	DeleteUser(userID string) error
	ListUsers(limit, offset int) ([]*User, error)
	CreateAdmin(email, displayName, password string) (*User, error)
}

// Common user-related errors
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrWeakPassword      = errors.New("password does not meet requirements")
	ErrInvalidEmail      = errors.New("invalid email format")
)

// HashPassword hashes a plain text password using bcrypt
// Returns the hashed password or an error if hashing fails
func (u *User) HashPassword(password string) error {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedBytes)
	return nil
}

// CheckPassword compares a plain text password with the stored hash
// Returns true if passwords match, false otherwise
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// ValidatePassword checks if password meets security requirements
// Password must be at least 6 characters with 1 special char and 1 number
func ValidatePassword(password string) error {
	if len(password) < 6 {
		return ErrWeakPassword
	}
	
	hasSpecial := false
	hasNumber := false
	
	for _, char := range password {
		if char >= '0' && char <= '9' {
			hasNumber = true
		} else if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')) {
			hasSpecial = true
		}
	}
	
	if !hasSpecial || !hasNumber {
		return ErrWeakPassword
	}
	
	return nil
}