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

func TestTaskService_CreateTask(t *testing.T) {
	userID := uuid.New().String()

	tests := []struct {
		name          string
		userID        string
		description   string
		category      string
		setupMock     func(*mocks.MockTaskRepository)
		wantErr       bool
		expectedError string
		validateTask  func(*testing.T, *domain.Task)
	}{
		{
			name:        "successful task creation",
			userID:      userID,
			description: "Complete project documentation",
			category:    "Work",
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("CreateTask", mock.MatchedBy(func(task *domain.Task) bool {
					return task.UserID == userID &&
						task.Description == "Complete project documentation" &&
						task.Category == "Work" &&
						!task.Completed &&
						task.ID != "" &&
						task.DeletedAt == nil
				})).Return(nil)
			},
			wantErr: false,
			validateTask: func(t *testing.T, task *domain.Task) {
				assert.Equal(t, userID, task.UserID)
				assert.Equal(t, "Complete project documentation", task.Description)
				assert.Equal(t, "Work", task.Category)
				assert.False(t, task.Completed)
				assert.NotEmpty(t, task.ID)
				assert.Nil(t, task.DeletedAt)
			},
		},
		{
			name:        "successful task creation without category",
			userID:      userID,
			description: "Buy groceries",
			category:    "",
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("CreateTask", mock.MatchedBy(func(task *domain.Task) bool {
					return task.UserID == userID &&
						task.Description == "Buy groceries" &&
						task.Category == "" &&
						!task.Completed
				})).Return(nil)
			},
			wantErr: false,
			validateTask: func(t *testing.T, task *domain.Task) {
				assert.Equal(t, userID, task.UserID)
				assert.Equal(t, "Buy groceries", task.Description)
				assert.Equal(t, "", task.Category)
				assert.False(t, task.Completed)
			},
		},
		{
			name:          "empty user ID",
			userID:        "",
			description:   "Test task",
			category:      "Test",
			setupMock:     func(mockRepo *mocks.MockTaskRepository) {},
			wantErr:       true,
			expectedError: "user ID is required",
		},
		{
			name:          "empty description",
			userID:        userID,
			description:   "",
			category:      "Test",
			setupMock:     func(mockRepo *mocks.MockTaskRepository) {},
			wantErr:       true,
			expectedError: "task description cannot be empty",
		},
		{
			name:          "description with only whitespace",
			userID:        userID,
			description:   "   ",
			category:      "Test",
			setupMock:     func(mockRepo *mocks.MockTaskRepository) {},
			wantErr:       true,
			expectedError: "task description cannot be empty",
		},
		{
			name:          "description too long",
			userID:        userID,
			description:   strings.Repeat("a", 10001),
			category:      "Test",
			setupMock:     func(mockRepo *mocks.MockTaskRepository) {},
			wantErr:       true,
			expectedError: "task description cannot exceed 10000 characters",
		},
		{
			name:        "repository error",
			userID:      userID,
			description: "Valid task",
			category:    "Test",
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("CreateTask", mock.AnythingOfType("*domain.Task")).Return(errors.New("database error"))
			},
			wantErr:       true,
			expectedError: "failed to create task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockTaskRepository(t)
			tt.setupMock(mockRepo)

			service := NewTaskService(mockRepo)
			task, err := service.CreateTask(tt.userID, tt.description, tt.category)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, task)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, task)
				if tt.validateTask != nil {
					tt.validateTask(t, task)
				}
			}
		})
	}
}

func TestTaskService_GetTaskByID(t *testing.T) {
	userID := uuid.New().String()
	taskID := uuid.New().String()
	otherUserID := uuid.New().String()

	testTask := &domain.Task{
		ID:          taskID,
		UserID:      userID,
		Description: "Test task",
		Category:    "Test",
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	otherUserTask := &domain.Task{
		ID:          uuid.New().String(),
		UserID:      otherUserID,
		Description: "Other user's task",
		Category:    "Test",
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tests := []struct {
		name          string
		taskID        string
		userID        string
		setupMock     func(*mocks.MockTaskRepository)
		wantErr       bool
		expectedError string
		validateTask  func(*testing.T, *domain.Task)
	}{
		{
			name:   "successful get task",
			taskID: taskID,
			userID: userID,
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("GetTaskByID", taskID).Return(testTask, nil)
			},
			wantErr: false,
			validateTask: func(t *testing.T, task *domain.Task) {
				assert.Equal(t, testTask.ID, task.ID)
				assert.Equal(t, testTask.UserID, task.UserID)
				assert.Equal(t, testTask.Description, task.Description)
			},
		},
		{
			name:          "empty task ID",
			taskID:        "",
			userID:        userID,
			setupMock:     func(mockRepo *mocks.MockTaskRepository) {},
			wantErr:       true,
			expectedError: "task ID is required",
		},
		{
			name:          "empty user ID",
			taskID:        taskID,
			userID:        "",
			setupMock:     func(mockRepo *mocks.MockTaskRepository) {},
			wantErr:       true,
			expectedError: "user ID is required",
		},
		{
			name:   "task not found",
			taskID: uuid.New().String(),
			userID: userID,
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("GetTaskByID", mock.AnythingOfType("string")).Return(nil, domain.ErrTaskNotFound)
			},
			wantErr:       true,
			expectedError: "task not found",
		},
		{
			name:   "task belongs to different user",
			taskID: otherUserTask.ID,
			userID: userID,
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("GetTaskByID", otherUserTask.ID).Return(otherUserTask, nil)
			},
			wantErr:       true,
			expectedError: "task not found",
		},
		{
			name:   "repository error",
			taskID: taskID,
			userID: userID,
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("GetTaskByID", taskID).Return(nil, errors.New("database error"))
			},
			wantErr:       true,
			expectedError: "failed to get task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockTaskRepository(t)
			tt.setupMock(mockRepo)

			service := NewTaskService(mockRepo)
			task, err := service.GetTaskByID(tt.taskID, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, task)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, task)
				if tt.validateTask != nil {
					tt.validateTask(t, task)
				}
			}
		})
	}
}

func TestTaskService_ListTasks(t *testing.T) {
	userID := uuid.New().String()
	tasks := []*domain.Task{
		{
			ID:          uuid.New().String(),
			UserID:      userID,
			Description: "Task 1",
			Category:    "Work",
			Completed:   false,
		},
		{
			ID:          uuid.New().String(),
			UserID:      userID,
			Description: "Task 2",
			Category:    "Personal",
			Completed:   true,
		},
	}

	tests := []struct {
		name        string
		userID      string
		filters     domain.TaskFilters
		setupMock   func(*mocks.MockTaskRepository)
		wantErr     bool
		expectedErr string
		validateTasks func(*testing.T, []*domain.Task)
	}{
		{
			name:   "successful list tasks",
			userID: userID,
			filters: domain.TaskFilters{
				Limit:  10,
				Offset: 0,
			},
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("ListTasks", userID, mock.MatchedBy(func(filters domain.TaskFilters) bool {
					return filters.Limit == 10 && filters.Offset == 0
				})).Return(tasks, nil)
			},
			wantErr: false,
			validateTasks: func(t *testing.T, returnedTasks []*domain.Task) {
				assert.Len(t, returnedTasks, 2)
				assert.Equal(t, tasks[0].ID, returnedTasks[0].ID)
				assert.Equal(t, tasks[1].ID, returnedTasks[1].ID)
			},
		},
		{
			name:   "list with category filter",
			userID: userID,
			filters: domain.TaskFilters{
				Category: "Work",
				Limit:    10,
				Offset:   0,
			},
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				workTasks := []*domain.Task{tasks[0]}
				mockRepo.On("ListTasks", userID, mock.MatchedBy(func(filters domain.TaskFilters) bool {
					return filters.Category == "Work"
				})).Return(workTasks, nil)
			},
			wantErr: false,
			validateTasks: func(t *testing.T, returnedTasks []*domain.Task) {
				assert.Len(t, returnedTasks, 1)
				assert.Equal(t, "Work", returnedTasks[0].Category)
			},
		},
		{
			name:        "empty user ID",
			userID:      "",
			filters:     domain.TaskFilters{},
			setupMock:   func(mockRepo *mocks.MockTaskRepository) {},
			wantErr:     true,
			expectedErr: "user ID is required",
		},
		{
			name:   "invalid filters",
			userID: userID,
			filters: domain.TaskFilters{
				Limit:  -1,
				Offset: 0,
			},
			setupMock:   func(mockRepo *mocks.MockTaskRepository) {},
			wantErr:     true,
			expectedErr: "invalid task filters",
		},
		{
			name:   "repository error",
			userID: userID,
			filters: domain.TaskFilters{
				Limit:  10,
				Offset: 0,
			},
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("ListTasks", userID, mock.AnythingOfType("domain.TaskFilters")).Return(nil, errors.New("database error"))
			},
			wantErr:     true,
			expectedErr: "failed to list tasks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockTaskRepository(t)
			tt.setupMock(mockRepo)

			service := NewTaskService(mockRepo)
			tasks, err := service.ListTasks(tt.userID, tt.filters)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, tasks)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tasks)
				if tt.validateTasks != nil {
					tt.validateTasks(t, tasks)
				}
			}
		})
	}
}

func TestTaskService_UpdateTaskCompletion(t *testing.T) {
	userID := uuid.New().String()
	taskID := uuid.New().String()
	otherUserID := uuid.New().String()

	testTask := &domain.Task{
		ID:          taskID,
		UserID:      userID,
		Description: "Test task",
		Category:    "Test",
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	otherUserTask := &domain.Task{
		ID:          uuid.New().String(),
		UserID:      otherUserID,
		Description: "Other user's task",
		Category:    "Test",
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tests := []struct {
		name          string
		taskID        string
		userID        string
		completed     bool
		setupMock     func(*mocks.MockTaskRepository)
		wantErr       bool
		expectedError string
		validateTask  func(*testing.T, *domain.Task)
	}{
		{
			name:      "successful mark as completed",
			taskID:    taskID,
			userID:    userID,
			completed: true,
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("GetTaskByID", taskID).Return(testTask, nil).Once()
				mockRepo.On("UpdateTaskCompletion", taskID, true).Return(nil)
				
				updatedTask := *testTask
				updatedTask.Completed = true
				mockRepo.On("GetTaskByID", taskID).Return(&updatedTask, nil).Once()
			},
			wantErr: false,
			validateTask: func(t *testing.T, task *domain.Task) {
				assert.True(t, task.Completed)
			},
		},
		{
			name:      "successful mark as incomplete",
			taskID:    taskID,
			userID:    userID,
			completed: false,
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				completedTask := *testTask
				completedTask.Completed = true
				mockRepo.On("GetTaskByID", taskID).Return(&completedTask, nil).Once()
				mockRepo.On("UpdateTaskCompletion", taskID, false).Return(nil)
				
				updatedTask := *testTask
				updatedTask.Completed = false
				mockRepo.On("GetTaskByID", taskID).Return(&updatedTask, nil).Once()
			},
			wantErr: false,
			validateTask: func(t *testing.T, task *domain.Task) {
				assert.False(t, task.Completed)
			},
		},
		{
			name:          "empty task ID",
			taskID:        "",
			userID:        userID,
			completed:     true,
			setupMock:     func(mockRepo *mocks.MockTaskRepository) {},
			wantErr:       true,
			expectedError: "task ID is required",
		},
		{
			name:          "empty user ID",
			taskID:        taskID,
			userID:        "",
			completed:     true,
			setupMock:     func(mockRepo *mocks.MockTaskRepository) {},
			wantErr:       true,
			expectedError: "user ID is required",
		},
		{
			name:      "task not found",
			taskID:    uuid.New().String(),
			userID:    userID,
			completed: true,
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("GetTaskByID", mock.AnythingOfType("string")).Return(nil, domain.ErrTaskNotFound)
			},
			wantErr:       true,
			expectedError: "task not found",
		},
		{
			name:      "task belongs to different user",
			taskID:    otherUserTask.ID,
			userID:    userID,
			completed: true,
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("GetTaskByID", otherUserTask.ID).Return(otherUserTask, nil)
			},
			wantErr:       true,
			expectedError: "task not found",
		},
		{
			name:      "repository update error",
			taskID:    taskID,
			userID:    userID,
			completed: true,
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("GetTaskByID", taskID).Return(testTask, nil)
				mockRepo.On("UpdateTaskCompletion", taskID, true).Return(errors.New("database error"))
			},
			wantErr:       true,
			expectedError: "failed to update task completion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockTaskRepository(t)
			tt.setupMock(mockRepo)

			service := NewTaskService(mockRepo)
			task, err := service.UpdateTaskCompletion(tt.taskID, tt.userID, tt.completed)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, task)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, task)
				if tt.validateTask != nil {
					tt.validateTask(t, task)
				}
			}
		})
	}
}

func TestTaskService_SoftDeleteTask(t *testing.T) {
	userID := uuid.New().String()
	taskID := uuid.New().String()
	otherUserID := uuid.New().String()

	testTask := &domain.Task{
		ID:          taskID,
		UserID:      userID,
		Description: "Test task",
		Category:    "Test",
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	otherUserTask := &domain.Task{
		ID:          uuid.New().String(),
		UserID:      otherUserID,
		Description: "Other user's task",
		Category:    "Test",
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tests := []struct {
		name          string
		taskID        string
		userID        string
		setupMock     func(*mocks.MockTaskRepository)
		wantErr       bool
		expectedError string
	}{
		{
			name:   "successful soft delete",
			taskID: taskID,
			userID: userID,
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("GetTaskByID", taskID).Return(testTask, nil)
				mockRepo.On("SoftDeleteTask", taskID).Return(nil)
			},
			wantErr: false,
		},
		{
			name:          "empty task ID",
			taskID:        "",
			userID:        userID,
			setupMock:     func(mockRepo *mocks.MockTaskRepository) {},
			wantErr:       true,
			expectedError: "task ID is required",
		},
		{
			name:          "empty user ID",
			taskID:        taskID,
			userID:        "",
			setupMock:     func(mockRepo *mocks.MockTaskRepository) {},
			wantErr:       true,
			expectedError: "user ID is required",
		},
		{
			name:   "task not found",
			taskID: uuid.New().String(),
			userID: userID,
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("GetTaskByID", mock.AnythingOfType("string")).Return(nil, domain.ErrTaskNotFound)
			},
			wantErr:       true,
			expectedError: "task not found",
		},
		{
			name:   "task belongs to different user",
			taskID: otherUserTask.ID,
			userID: userID,
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("GetTaskByID", otherUserTask.ID).Return(otherUserTask, nil)
			},
			wantErr:       true,
			expectedError: "task not found",
		},
		{
			name:   "repository delete error",
			taskID: taskID,
			userID: userID,
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("GetTaskByID", taskID).Return(testTask, nil)
				mockRepo.On("SoftDeleteTask", taskID).Return(errors.New("database error"))
			},
			wantErr:       true,
			expectedError: "failed to delete task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockTaskRepository(t)
			tt.setupMock(mockRepo)

			service := NewTaskService(mockRepo)
			err := service.SoftDeleteTask(tt.taskID, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskService_RestoreTask(t *testing.T) {
	userID := uuid.New().String()
	taskID := uuid.New().String()
	otherUserID := uuid.New().String()

	now := time.Now()
	deletedTask := &domain.Task{
		ID:          taskID,
		UserID:      userID,
		Description: "Deleted task",
		Category:    "Test",
		Completed:   false,
		CreatedAt:   now.Add(-time.Hour),
		UpdatedAt:   now.Add(-time.Minute),
		DeletedAt:   &now,
	}

	oldDeletedTask := &domain.Task{
		ID:          uuid.New().String(),
		UserID:      userID,
		Description: "Old deleted task",
		Category:    "Test",
		Completed:   false,
		CreatedAt:   now.Add(-10 * 24 * time.Hour), // 10 days ago
		UpdatedAt:   now.Add(-10 * 24 * time.Hour),
		DeletedAt:   &[]time.Time{now.Add(-8 * 24 * time.Hour)}[0], // 8 days ago
	}

	otherUserTask := &domain.Task{
		ID:          uuid.New().String(),
		UserID:      otherUserID,
		Description: "Other user's deleted task",
		Category:    "Test",
		Completed:   false,
		CreatedAt:   now.Add(-time.Hour),
		UpdatedAt:   now.Add(-time.Minute),
		DeletedAt:   &now,
	}

	tests := []struct {
		name          string
		taskID        string
		userID        string
		setupMock     func(*mocks.MockTaskRepository)
		wantErr       bool
		expectedError string
		validateTask  func(*testing.T, *domain.Task)
	}{
		{
			name:   "successful restore",
			taskID: taskID,
			userID: userID,
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("GetTaskByID", taskID).Return(deletedTask, nil).Once()
				mockRepo.On("RestoreTask", taskID).Return(nil)
				
				restoredTask := *deletedTask
				restoredTask.DeletedAt = nil
				mockRepo.On("GetTaskByID", taskID).Return(&restoredTask, nil).Once()
			},
			wantErr: false,
			validateTask: func(t *testing.T, task *domain.Task) {
				assert.Nil(t, task.DeletedAt)
			},
		},
		{
			name:          "empty task ID",
			taskID:        "",
			userID:        userID,
			setupMock:     func(mockRepo *mocks.MockTaskRepository) {},
			wantErr:       true,
			expectedError: "task ID is required",
		},
		{
			name:          "empty user ID",
			taskID:        taskID,
			userID:        "",
			setupMock:     func(mockRepo *mocks.MockTaskRepository) {},
			wantErr:       true,
			expectedError: "user ID is required",
		},
		{
			name:   "task not found",
			taskID: uuid.New().String(),
			userID: userID,
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("GetTaskByID", mock.AnythingOfType("string")).Return(nil, domain.ErrTaskNotFound)
			},
			wantErr:       true,
			expectedError: "task not found",
		},
		{
			name:   "task belongs to different user",
			taskID: otherUserTask.ID,
			userID: userID,
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("GetTaskByID", otherUserTask.ID).Return(otherUserTask, nil)
			},
			wantErr:       true,
			expectedError: "task not found",
		},
		{
			name:   "task not deleted",
			taskID: taskID,
			userID: userID,
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				activeTask := *deletedTask
				activeTask.DeletedAt = nil
				mockRepo.On("GetTaskByID", taskID).Return(&activeTask, nil)
			},
			wantErr:       true,
			expectedError: "task is not deleted",
		},
		{
			name:   "task deleted more than 7 days ago",
			taskID: oldDeletedTask.ID,
			userID: userID,
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("GetTaskByID", oldDeletedTask.ID).Return(oldDeletedTask, nil)
			},
			wantErr:       true,
			expectedError: "task cannot be restored after 7 days",
		},
		{
			name:   "repository restore error",
			taskID: taskID,
			userID: userID,
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("GetTaskByID", taskID).Return(deletedTask, nil)
				mockRepo.On("RestoreTask", taskID).Return(errors.New("database error"))
			},
			wantErr:       true,
			expectedError: "failed to restore task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockTaskRepository(t)
			tt.setupMock(mockRepo)

			service := NewTaskService(mockRepo)
			task, err := service.RestoreTask(tt.taskID, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, task)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, task)
				if tt.validateTask != nil {
					tt.validateTask(t, task)
				}
			}
		})
	}
}

func TestTaskService_GetUserCategories(t *testing.T) {
	userID := uuid.New().String()
	categories := []string{"Work", "Personal", "Shopping"}

	tests := []struct {
		name              string
		userID            string
		setupMock         func(*mocks.MockTaskRepository)
		wantErr           bool
		expectedError     string
		validateCategories func(*testing.T, []string)
	}{
		{
			name:   "successful get categories",
			userID: userID,
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("GetUserCategories", userID).Return(categories, nil)
			},
			wantErr: false,
			validateCategories: func(t *testing.T, returnedCategories []string) {
				assert.Equal(t, categories, returnedCategories)
				assert.Len(t, returnedCategories, 3)
			},
		},
		{
			name:          "empty user ID",
			userID:        "",
			setupMock:     func(mockRepo *mocks.MockTaskRepository) {},
			wantErr:       true,
			expectedError: "user ID is required",
		},
		{
			name:   "repository error",
			userID: userID,
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("GetUserCategories", userID).Return(nil, errors.New("database error"))
			},
			wantErr:       true,
			expectedError: "failed to get user categories",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockTaskRepository(t)
			tt.setupMock(mockRepo)

			service := NewTaskService(mockRepo)
			categories, err := service.GetUserCategories(tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, categories)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, categories)
				if tt.validateCategories != nil {
					tt.validateCategories(t, categories)
				}
			}
		})
	}
}

func TestTaskService_RenameCategory(t *testing.T) {
	userID := uuid.New().String()

	tests := []struct {
		name          string
		userID        string
		oldName       string
		newName       string
		setupMock     func(*mocks.MockTaskRepository)
		wantErr       bool
		expectedError string
	}{
		{
			name:    "successful rename category",
			userID:  userID,
			oldName: "Work",
			newName: "Professional",
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("RenameCategory", userID, "Work", "Professional").Return(nil)
			},
			wantErr: false,
		},
		{
			name:          "empty user ID",
			userID:        "",
			oldName:       "Work",
			newName:       "Professional",
			setupMock:     func(mockRepo *mocks.MockTaskRepository) {},
			wantErr:       true,
			expectedError: "user ID is required",
		},
		{
			name:          "empty old name",
			userID:        userID,
			oldName:       "",
			newName:       "Professional",
			setupMock:     func(mockRepo *mocks.MockTaskRepository) {},
			wantErr:       true,
			expectedError: "category names cannot be empty",
		},
		{
			name:          "empty new name",
			userID:        userID,
			oldName:       "Work",
			newName:       "",
			setupMock:     func(mockRepo *mocks.MockTaskRepository) {},
			wantErr:       true,
			expectedError: "category names cannot be empty",
		},
		{
			name:          "same old and new names",
			userID:        userID,
			oldName:       "Work",
			newName:       "Work",
			setupMock:     func(mockRepo *mocks.MockTaskRepository) {},
			wantErr:       true,
			expectedError: "new category name must be different",
		},
		{
			name:    "repository error",
			userID:  userID,
			oldName: "Work",
			newName: "Professional",
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("RenameCategory", userID, "Work", "Professional").Return(errors.New("database error"))
			},
			wantErr:       true,
			expectedError: "failed to rename category",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockTaskRepository(t)
			tt.setupMock(mockRepo)

			service := NewTaskService(mockRepo)
			err := service.RenameCategory(tt.userID, tt.oldName, tt.newName)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskService_DeleteCategory(t *testing.T) {
	userID := uuid.New().String()

	tests := []struct {
		name          string
		userID        string
		categoryName  string
		setupMock     func(*mocks.MockTaskRepository)
		wantErr       bool
		expectedError string
	}{
		{
			name:         "successful delete category",
			userID:       userID,
			categoryName: "Work",
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("DeleteCategory", userID, "Work").Return(nil)
			},
			wantErr: false,
		},
		{
			name:          "empty user ID",
			userID:        "",
			categoryName:  "Work",
			setupMock:     func(mockRepo *mocks.MockTaskRepository) {},
			wantErr:       true,
			expectedError: "user ID is required",
		},
		{
			name:          "empty category name",
			userID:        userID,
			categoryName:  "",
			setupMock:     func(mockRepo *mocks.MockTaskRepository) {},
			wantErr:       true,
			expectedError: "category name cannot be empty",
		},
		{
			name:         "repository error",
			userID:       userID,
			categoryName: "Work",
			setupMock: func(mockRepo *mocks.MockTaskRepository) {
				mockRepo.On("DeleteCategory", userID, "Work").Return(errors.New("database error"))
			},
			wantErr:       true,
			expectedError: "failed to delete category",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockTaskRepository(t)
			tt.setupMock(mockRepo)

			service := NewTaskService(mockRepo)
			err := service.DeleteCategory(tt.userID, tt.categoryName)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskService_ErrorCodes(t *testing.T) {
	// Test that the service returns proper error codes
	mockRepo := mocks.NewMockTaskRepository(t)
	service := NewTaskService(mockRepo)

	// Test error code 3011 - User ID required
	_, err := service.CreateTask("", "Valid description", "Category")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "3011")

	// Test error code 3012 - Task description validation
	_, err = service.CreateTask(uuid.New().String(), "", "Category")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "3012")

	// Test error code 3013 - Task description too long
	_, err = service.CreateTask(uuid.New().String(), strings.Repeat("a", 10001), "Category")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "3013")
}

func BenchmarkTaskService_CreateTask(b *testing.B) {
	mockRepo := mocks.NewMockTaskRepository(b)
	mockRepo.On("CreateTask", mock.AnythingOfType("*domain.Task")).Return(nil)

	service := NewTaskService(mockRepo)
	userID := uuid.New().String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		description := fmt.Sprintf("Task %d", i)
		service.CreateTask(userID, description, "Test")
	}
}

func BenchmarkTaskService_GetTaskByID(b *testing.B) {
	userID := uuid.New().String()
	taskID := uuid.New().String()
	
	testTask := &domain.Task{
		ID:          taskID,
		UserID:      userID,
		Description: "Test task",
		Category:    "Test",
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo := mocks.NewMockTaskRepository(b)
	mockRepo.On("GetTaskByID", taskID).Return(testTask, nil)

	service := NewTaskService(mockRepo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GetTaskByID(taskID, userID)
	}
}