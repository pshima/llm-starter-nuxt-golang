package services

import (
	"backend/internal/domain"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TaskServiceInterface defines the interface for task business logic operations
// Contains all business rules and validation logic for task operations
type TaskServiceInterface interface {
	CreateTask(userID, description, category string) (*domain.Task, error)
	GetTaskByID(id, userID string) (*domain.Task, error)
	ListTasks(userID string, filters domain.TaskFilters) ([]*domain.Task, error)
	UpdateTaskCompletion(id, userID string, completed bool) (*domain.Task, error)
	SoftDeleteTask(id, userID string) error
	RestoreTask(id, userID string) (*domain.Task, error)
	GetUserCategories(userID string) ([]string, error)
	RenameCategory(userID, oldName, newName string) error
	DeleteCategory(userID, categoryName string) error
}

// TaskService implements task business logic operations
// Handles task creation, management, and category operations with user context validation
type TaskService struct {
	taskRepo TaskRepository
}

// TaskRepository defines the methods needed from the task repository
// This interface ensures loose coupling between service and repository layers
type TaskRepository interface {
	CreateTask(task *domain.Task) error
	GetTaskByID(id string) (*domain.Task, error)
	ListTasks(userID string, filters domain.TaskFilters) ([]*domain.Task, error)
	UpdateTaskCompletion(id string, completed bool) error
	SoftDeleteTask(id string) error
	RestoreTask(id string) error
	GetUserCategories(userID string) ([]string, error)
	RenameCategory(userID, oldName, newName string) error
	DeleteCategory(userID, categoryName string) error
	CleanupExpiredTasks() (int, error)
}

// NewTaskService creates a new instance of TaskService
// Initializes the service with the provided task repository
func NewTaskService(taskRepo TaskRepository) *TaskService {
	return &TaskService{
		taskRepo: taskRepo,
	}
}

// CreateTask creates a new task with validation and user context
// Validates description length and format before creating the task
func (s *TaskService) CreateTask(userID, description, category string) (*domain.Task, error) {
	// Error code 3011: User ID required
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("3011: user ID is required")
	}

	// Error code 3012: Task description validation
	if strings.TrimSpace(description) == "" {
		return nil, fmt.Errorf("3012: task description cannot be empty")
	}

	// Error code 3013: Task description too long
	if len(description) > 10000 {
		return nil, fmt.Errorf("3013: task description cannot exceed 10000 characters")
	}

	// Create new task
	task := &domain.Task{
		ID:          uuid.New().String(),
		UserID:      userID,
		Description: strings.TrimSpace(description),
		Category:    strings.TrimSpace(category),
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		DeletedAt:   nil,
	}

	// Validate task
	if err := task.Validate(); err != nil {
		return nil, fmt.Errorf("3014: %w", err)
	}

	// Save task to repository
	if err := s.taskRepo.CreateTask(task); err != nil {
		return nil, fmt.Errorf("3015: failed to create task: %w", err)
	}

	return task, nil
}

// GetTaskByID retrieves a task by ID with user ownership validation
// Ensures that users can only access their own tasks
func (s *TaskService) GetTaskByID(id, userID string) (*domain.Task, error) {
	// Error code 3011: User ID required
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("3011: user ID is required")
	}

	// Error code 3016: Task ID required
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("3016: task ID is required")
	}

	// Get task from repository
	task, err := s.taskRepo.GetTaskByID(id)
	if err != nil {
		if err == domain.ErrTaskNotFound {
			return nil, fmt.Errorf("3017: task not found")
		}
		return nil, fmt.Errorf("3018: failed to get task: %w", err)
	}

	// Verify task ownership
	if task.UserID != userID {
		return nil, fmt.Errorf("3017: task not found") // Don't reveal that task exists but belongs to another user
	}

	return task, nil
}

// ListTasks retrieves tasks for a user with optional filtering
// Applies user context and validates filter parameters
func (s *TaskService) ListTasks(userID string, filters domain.TaskFilters) ([]*domain.Task, error) {
	// Error code 3011: User ID required
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("3011: user ID is required")
	}

	// Error code 3019: Invalid filters
	if err := filters.Validate(); err != nil {
		return nil, fmt.Errorf("3019: invalid task filters: %w", err)
	}

	// Get tasks from repository
	tasks, err := s.taskRepo.ListTasks(userID, filters)
	if err != nil {
		return nil, fmt.Errorf("3018: failed to list tasks: %w", err)
	}

	return tasks, nil
}

// UpdateTaskCompletion updates the completion status of a task
// Validates user ownership before allowing the update
func (s *TaskService) UpdateTaskCompletion(id, userID string, completed bool) (*domain.Task, error) {
	// Error code 3011: User ID required
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("3011: user ID is required")
	}

	// Error code 3016: Task ID required
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("3016: task ID is required")
	}

	// Verify task exists and user owns it
	_, err := s.GetTaskByID(id, userID)
	if err != nil {
		return nil, err // Error already has proper code from GetTaskByID
	}

	// Update task completion in repository
	if err := s.taskRepo.UpdateTaskCompletion(id, completed); err != nil {
		return nil, fmt.Errorf("3020: failed to update task completion: %w", err)
	}

	// Return updated task
	updatedTask, err := s.taskRepo.GetTaskByID(id)
	if err != nil {
		return nil, fmt.Errorf("3018: failed to get updated task: %w", err)
	}

	return updatedTask, nil
}

// SoftDeleteTask marks a task as deleted without removing it from storage
// Validates user ownership before allowing the deletion
func (s *TaskService) SoftDeleteTask(id, userID string) error {
	// Error code 3011: User ID required
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("3011: user ID is required")
	}

	// Error code 3016: Task ID required
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("3016: task ID is required")
	}

	// Verify task exists and user owns it
	_, err := s.GetTaskByID(id, userID)
	if err != nil {
		return err // Error already has proper code from GetTaskByID
	}

	// Soft delete task in repository
	if err := s.taskRepo.SoftDeleteTask(id); err != nil {
		return fmt.Errorf("3020: failed to delete task: %w", err)
	}

	return nil
}

// RestoreTask restores a soft-deleted task if within the 7-day window
// Validates user ownership and deletion window before restoring
func (s *TaskService) RestoreTask(id, userID string) (*domain.Task, error) {
	// Error code 3011: User ID required
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("3011: user ID is required")
	}

	// Error code 3016: Task ID required
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("3016: task ID is required")
	}

	// Get task from repository (this includes deleted tasks)
	task, err := s.taskRepo.GetTaskByID(id)
	if err != nil {
		if err == domain.ErrTaskNotFound {
			return nil, fmt.Errorf("3017: task not found")
		}
		return nil, fmt.Errorf("3018: failed to get task: %w", err)
	}

	// Verify task ownership
	if task.UserID != userID {
		return nil, fmt.Errorf("3017: task not found") // Don't reveal that task exists but belongs to another user
	}

	// Check if task is actually deleted
	if task.DeletedAt == nil {
		return nil, fmt.Errorf("3017: task is not deleted")
	}

	// Check if deletion is within 7-day window
	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)
	if task.DeletedAt.Before(sevenDaysAgo) {
		return nil, fmt.Errorf("3017: task cannot be restored after 7 days")
	}

	// Restore task in repository
	if err := s.taskRepo.RestoreTask(id); err != nil {
		return nil, fmt.Errorf("3020: failed to restore task: %w", err)
	}

	// Return restored task
	restoredTask, err := s.taskRepo.GetTaskByID(id)
	if err != nil {
		return nil, fmt.Errorf("3018: failed to get restored task: %w", err)
	}

	return restoredTask, nil
}

// GetUserCategories retrieves all categories used by a user's tasks
// Returns a list of unique category names for the specified user
func (s *TaskService) GetUserCategories(userID string) ([]string, error) {
	// Error code 3011: User ID required
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("3011: user ID is required")
	}

	// Get categories from repository
	categories, err := s.taskRepo.GetUserCategories(userID)
	if err != nil {
		return nil, fmt.Errorf("3018: failed to get user categories: %w", err)
	}

	return categories, nil
}

// RenameCategory renames a category across all of a user's tasks
// Validates category names and ensures they are different
func (s *TaskService) RenameCategory(userID, oldName, newName string) error {
	// Error code 3011: User ID required
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("3011: user ID is required")
	}

	// Error code 3012: Category names validation
	oldName = strings.TrimSpace(oldName)
	newName = strings.TrimSpace(newName)
	if oldName == "" || newName == "" {
		return fmt.Errorf("3012: category names cannot be empty")
	}

	// Error code 3014: Same names validation
	if oldName == newName {
		return fmt.Errorf("3014: new category name must be different")
	}

	// Rename category in repository
	if err := s.taskRepo.RenameCategory(userID, oldName, newName); err != nil {
		return fmt.Errorf("3020: failed to rename category: %w", err)
	}

	return nil
}

// DeleteCategory removes a category from all of a user's tasks
// Sets the category field to empty string for affected tasks
func (s *TaskService) DeleteCategory(userID, categoryName string) error {
	// Error code 3011: User ID required
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("3011: user ID is required")
	}

	// Error code 3012: Category name validation
	categoryName = strings.TrimSpace(categoryName)
	if categoryName == "" {
		return fmt.Errorf("3012: category name cannot be empty")
	}

	// Delete category in repository
	if err := s.taskRepo.DeleteCategory(userID, categoryName); err != nil {
		return fmt.Errorf("3020: failed to delete category: %w", err)
	}

	return nil
}