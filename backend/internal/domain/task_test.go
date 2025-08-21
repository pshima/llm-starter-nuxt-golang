package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTask_Validation(t *testing.T) {
	tests := []struct {
		name          string
		task          Task
		expectedError string
	}{
		{
			name: "valid task",
			task: Task{
				ID:          uuid.New().String(),
				UserID:      uuid.New().String(),
				Description: "Test task description",
				Category:    "Work",
				Completed:   false,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			expectedError: "",
		},
		{
			name: "empty description",
			task: Task{
				ID:        uuid.New().String(),
				UserID:    uuid.New().String(),
				Category:  "Work",
				Completed: false,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectedError: "task description cannot be empty",
		},
		{
			name: "description too long",
			task: Task{
				ID:          uuid.New().String(),
				UserID:      uuid.New().String(),
				Description: string(make([]byte, 10001)), // Over 10000 chars
				Category:    "Work",
				Completed:   false,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			expectedError: "task description cannot exceed 10000 characters",
		},
		{
			name: "empty user ID",
			task: Task{
				ID:          uuid.New().String(),
				Description: "Test task description",
				Category:    "Work",
				Completed:   false,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			expectedError: "invalid user for task",
		},
		{
			name: "empty task ID",
			task: Task{
				UserID:      uuid.New().String(),
				Description: "Test task description",
				Category:    "Work",
				Completed:   false,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			expectedError: "task ID cannot be empty",
		},
		{
			name: "whitespace-only task ID",
			task: Task{
				ID:          "   ",
				UserID:      uuid.New().String(),
				Description: "Test task description",
				Category:    "Work",
				Completed:   false,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			expectedError: "task ID cannot be empty",
		},
		{
			name: "whitespace-only user ID",
			task: Task{
				ID:          uuid.New().String(),
				UserID:      "   ",
				Description: "Test task description",
				Category:    "Work",
				Completed:   false,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			expectedError: "invalid user for task",
		},
		{
			name: "whitespace-only description",
			task: Task{
				ID:          uuid.New().String(),
				UserID:      uuid.New().String(),
				Description: "   ",
				Category:    "Work",
				Completed:   false,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			expectedError: "task description cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}

func TestTask_MarkCompleted(t *testing.T) {
	task := &Task{
		ID:          uuid.New().String(),
		UserID:      uuid.New().String(),
		Description: "Test task",
		Category:    "Work",
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	beforeUpdate := time.Now()
	task.MarkCompleted()

	assert.True(t, task.Completed)
	assert.True(t, task.UpdatedAt.After(beforeUpdate) || task.UpdatedAt.Equal(beforeUpdate))
}

func TestTask_MarkIncomplete(t *testing.T) {
	task := &Task{
		ID:          uuid.New().String(),
		UserID:      uuid.New().String(),
		Description: "Test task",
		Category:    "Work",
		Completed:   true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	beforeUpdate := time.Now()
	task.MarkIncomplete()

	assert.False(t, task.Completed)
	assert.True(t, task.UpdatedAt.After(beforeUpdate) || task.UpdatedAt.Equal(beforeUpdate))
}

func TestTask_SoftDelete(t *testing.T) {
	task := &Task{
		ID:          uuid.New().String(),
		UserID:      uuid.New().String(),
		Description: "Test task",
		Category:    "Work",
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	beforeDelete := time.Now()
	task.SoftDelete()

	assert.NotNil(t, task.DeletedAt)
	assert.True(t, task.DeletedAt.After(beforeDelete) || task.DeletedAt.Equal(beforeDelete))
	assert.True(t, task.UpdatedAt.After(beforeDelete) || task.UpdatedAt.Equal(beforeDelete))
}

func TestTask_Restore(t *testing.T) {
	deletedAt := time.Now()
	task := &Task{
		ID:          uuid.New().String(),
		UserID:      uuid.New().String(),
		Description: "Test task",
		Category:    "Work",
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		DeletedAt:   &deletedAt,
	}

	beforeRestore := time.Now()
	task.Restore()

	assert.Nil(t, task.DeletedAt)
	assert.True(t, task.UpdatedAt.After(beforeRestore) || task.UpdatedAt.Equal(beforeRestore))
}

func TestTask_IsDeleted(t *testing.T) {
	tests := []struct {
		name      string
		deletedAt *time.Time
		expected  bool
	}{
		{
			name:      "not deleted",
			deletedAt: nil,
			expected:  false,
		},
		{
			name: "deleted",
			deletedAt: func() *time.Time {
				t := time.Now()
				return &t
			}(),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{
				ID:          uuid.New().String(),
				UserID:      uuid.New().String(),
				Description: "Test task",
				DeletedAt:   tt.deletedAt,
			}

			assert.Equal(t, tt.expected, task.IsDeleted())
		})
	}
}

// Test interface compliance
func TestTaskRepository_InterfaceCompliance(t *testing.T) {
	// This test ensures the interface methods are properly defined
	var _ TaskRepository = (*mockTaskRepository)(nil)
}

func TestTaskService_InterfaceCompliance(t *testing.T) {
	// This test ensures the interface methods are properly defined
	var _ TaskService = (*mockTaskService)(nil)
}

// Test error constants
func TestErrorConstants(t *testing.T) {
	errors := []error{
		ErrTaskNotFound,
		ErrTaskInvalidDescription,
		ErrTaskDescriptionTooLong,
		ErrTaskInvalidUser,
		ErrTaskInvalidID,
		ErrCategoryNotFound,
		ErrInvalidFilters,
	}

	for _, err := range errors {
		assert.NotNil(t, err)
		assert.NotEmpty(t, err.Error())
	}
}

// Test with various task states
func TestTaskStates(t *testing.T) {
	baseTask := Task{
		ID:          uuid.New().String(),
		UserID:      uuid.New().String(),
		Description: "Test task",
		Category:    "Work",
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	t.Run("new task is not deleted", func(t *testing.T) {
		task := baseTask
		assert.False(t, task.IsDeleted())
		assert.False(t, task.Completed)
	})

	t.Run("complete and then incomplete", func(t *testing.T) {
		task := baseTask
		
		// Mark as completed
		task.MarkCompleted()
		assert.True(t, task.Completed)
		
		// Mark as incomplete
		task.MarkIncomplete()
		assert.False(t, task.Completed)
	})

	t.Run("soft delete and restore", func(t *testing.T) {
		task := baseTask
		
		// Soft delete
		task.SoftDelete()
		assert.True(t, task.IsDeleted())
		assert.NotNil(t, task.DeletedAt)
		
		// Restore
		task.Restore()
		assert.False(t, task.IsDeleted())
		assert.Nil(t, task.DeletedAt)
	})
}

// Mock implementations for interface compliance testing
type mockTaskRepository struct{}

func (m *mockTaskRepository) CreateTask(task *Task) error                          { return nil }
func (m *mockTaskRepository) GetTaskByID(id string) (*Task, error)                { return nil, nil }
func (m *mockTaskRepository) ListTasks(userID string, filters TaskFilters) ([]*Task, error) { return nil, nil }
func (m *mockTaskRepository) UpdateTaskCompletion(id string, completed bool) error { return nil }
func (m *mockTaskRepository) SoftDeleteTask(id string) error                      { return nil }
func (m *mockTaskRepository) RestoreTask(id string) error                         { return nil }
func (m *mockTaskRepository) GetUserCategories(userID string) ([]string, error)   { return nil, nil }
func (m *mockTaskRepository) RenameCategory(userID, oldName, newName string) error { return nil }
func (m *mockTaskRepository) DeleteCategory(userID, categoryName string) error     { return nil }
func (m *mockTaskRepository) CleanupExpiredTasks() (int, error)                    { return 0, nil }

type mockTaskService struct{}

func (m *mockTaskService) CreateTask(userID, description, category string) (*Task, error) { return nil, nil }
func (m *mockTaskService) GetTaskByID(id string) (*Task, error)                          { return nil, nil }
func (m *mockTaskService) ListTasks(userID string, filters TaskFilters) ([]*Task, error) { return nil, nil }
func (m *mockTaskService) UpdateTaskCompletion(id string, completed bool) (*Task, error) { return nil, nil }
func (m *mockTaskService) SoftDeleteTask(id string) error                                { return nil }
func (m *mockTaskService) RestoreTask(id string) (*Task, error)                          { return nil, nil }
func (m *mockTaskService) GetUserCategories(userID string) ([]string, error)             { return nil, nil }
func (m *mockTaskService) RenameCategory(userID, oldName, newName string) error           { return nil }
func (m *mockTaskService) DeleteCategory(userID, categoryName string) error               { return nil }

// Test TaskFilters default values and edge cases
func TestTaskFilters_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		filters TaskFilters
		isValid bool
	}{
		{
			name: "zero limit and offset",
			filters: TaskFilters{
				Limit:  0,
				Offset: 0,
			},
			isValid: true,
		},
		{
			name: "maximum limit",
			filters: TaskFilters{
				Limit:  1000,
				Offset: 0,
			},
			isValid: true,
		},
		{
			name: "limit exactly over maximum",
			filters: TaskFilters{
				Limit:  1001,
				Offset: 0,
			},
			isValid: false,
		},
		{
			name: "completed filter true",
			filters: TaskFilters{
				Completed: &[]bool{true}[0],
				Limit:     10,
				Offset:    0,
			},
			isValid: true,
		},
		{
			name: "completed filter false",
			filters: TaskFilters{
				Completed: &[]bool{false}[0],
				Limit:     10,
				Offset:    0,
			},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filters.Validate()
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// Test TaskFilters validation
func TestTaskFilters_Validation(t *testing.T) {
	tests := []struct {
		name    string
		filters TaskFilters
		isValid bool
	}{
		{
			name: "valid filters with all fields",
			filters: TaskFilters{
				Category:        "Work",
				Completed:       &[]bool{true}[0],
				IncludeDeleted:  false,
				Limit:          20,
				Offset:         0,
			},
			isValid: true,
		},
		{
			name: "valid filters with minimal fields",
			filters: TaskFilters{
				Limit:  10,
				Offset: 0,
			},
			isValid: true,
		},
		{
			name: "invalid limit too large",
			filters: TaskFilters{
				Limit:  1001,
				Offset: 0,
			},
			isValid: false,
		},
		{
			name: "invalid negative offset",
			filters: TaskFilters{
				Limit:  10,
				Offset: -1,
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filters.Validate()
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}