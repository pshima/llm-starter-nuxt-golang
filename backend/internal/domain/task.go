package domain

import (
	"errors"
	"strings"
	"time"
)

// Task represents a task in the task tracker system
// Contains all information needed to track and manage individual tasks
type Task struct {
	ID          string     `json:"id" redis:"id"`
	UserID      string     `json:"user_id" redis:"user_id"`
	Description string     `json:"description" redis:"description"`
	Category    string     `json:"category,omitempty" redis:"category"`
	Completed   bool       `json:"completed" redis:"completed"`
	CreatedAt   time.Time  `json:"created_at" redis:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" redis:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" redis:"deleted_at"`
}

// TaskFilters represents filtering options for task queries
// Used to filter tasks by various criteria in list operations
type TaskFilters struct {
	Category       string `json:"category,omitempty"`
	Completed      *bool  `json:"completed,omitempty"`
	IncludeDeleted bool   `json:"include_deleted"`
	Limit          int    `json:"limit"`
	Offset         int    `json:"offset"`
}

// TaskRepository defines the interface for task data access operations
// This interface allows for different storage implementations while maintaining clean architecture
type TaskRepository interface {
	CreateTask(task *Task) error
	GetTaskByID(id string) (*Task, error)
	ListTasks(userID string, filters TaskFilters) ([]*Task, error)
	UpdateTaskCompletion(id string, completed bool) error
	SoftDeleteTask(id string) error
	RestoreTask(id string) error
	GetUserCategories(userID string) ([]string, error)
	RenameCategory(userID, oldName, newName string) error
	DeleteCategory(userID, categoryName string) error
	CleanupExpiredTasks() (int, error)
}

// TaskService defines the interface for task business logic operations
// Contains all business rules and validation logic for task operations
type TaskService interface {
	CreateTask(userID, description, category string) (*Task, error)
	GetTaskByID(id string) (*Task, error)
	ListTasks(userID string, filters TaskFilters) ([]*Task, error)
	UpdateTaskCompletion(id string, completed bool) (*Task, error)
	SoftDeleteTask(id string) error
	RestoreTask(id string) (*Task, error)
	GetUserCategories(userID string) ([]string, error)
	RenameCategory(userID, oldName, newName string) error
	DeleteCategory(userID, categoryName string) error
}

// Common task-related errors
var (
	ErrTaskNotFound            = errors.New("task not found")
	ErrTaskInvalidDescription  = errors.New("task description cannot be empty")
	ErrTaskDescriptionTooLong  = errors.New("task description cannot exceed 10000 characters")
	ErrTaskInvalidUser         = errors.New("invalid user for task")
	ErrTaskInvalidID           = errors.New("task ID cannot be empty")
	ErrCategoryNotFound        = errors.New("category not found")
	ErrInvalidFilters          = errors.New("invalid task filters")
)

// Validate checks if the task has valid data
// Returns error if task data is invalid according to business rules
func (t *Task) Validate() error {
	if strings.TrimSpace(t.ID) == "" {
		return ErrTaskInvalidID
	}
	if strings.TrimSpace(t.UserID) == "" {
		return ErrTaskInvalidUser
	}
	if strings.TrimSpace(t.Description) == "" {
		return ErrTaskInvalidDescription
	}
	if len(t.Description) > 10000 {
		return ErrTaskDescriptionTooLong
	}
	return nil
}

// MarkCompleted marks the task as completed and updates the timestamp
// Updates the task's completion status and modified time
func (t *Task) MarkCompleted() {
	t.Completed = true
	t.UpdatedAt = time.Now()
}

// MarkIncomplete marks the task as incomplete and updates the timestamp
// Updates the task's completion status and modified time
func (t *Task) MarkIncomplete() {
	t.Completed = false
	t.UpdatedAt = time.Now()
}

// SoftDelete marks the task as deleted without removing it from storage
// Sets the DeletedAt timestamp and updates the modified time
func (t *Task) SoftDelete() {
	now := time.Now()
	t.DeletedAt = &now
	t.UpdatedAt = now
}

// Restore restores a soft-deleted task
// Clears the DeletedAt timestamp and updates the modified time
func (t *Task) Restore() {
	t.DeletedAt = nil
	t.UpdatedAt = time.Now()
}

// IsDeleted checks if the task is soft-deleted
// Returns true if the task has a DeletedAt timestamp
func (t *Task) IsDeleted() bool {
	return t.DeletedAt != nil
}

// Validate checks if the task filters have valid values
// Returns error if filter values are invalid
func (f *TaskFilters) Validate() error {
	if f.Limit < 0 || f.Limit > 1000 {
		return errors.New("limit must be between 0 and 1000")
	}
	if f.Offset < 0 {
		return errors.New("offset must be non-negative")
	}
	return nil
}