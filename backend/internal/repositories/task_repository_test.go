package repositories

import (
	"context"
	"testing"
	"time"

	"backend/internal/domain"
	"backend/pkg/redis"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	redislib "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestTaskRepository(t *testing.T) (*TaskRepository, *miniredis.Miniredis) {
	// Create miniredis instance
	s, err := miniredis.Run()
	require.NoError(t, err)

	// Create Redis client
	rdb := redislib.NewClient(&redislib.Options{
		Addr: s.Addr(),
		DB:   0,
	})

	client := &redis.Client{Client: rdb}
	repo := NewTaskRepository(client)

	return repo, s
}

func createTestTask(userID, description, category string) *domain.Task {
	now := time.Now()
	return &domain.Task{
		ID:          uuid.New().String(),
		UserID:      userID,
		Description: description,
		Category:    category,
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   nil,
	}
}

func TestNewTaskRepository(t *testing.T) {
	tests := []struct {
		name   string
		client *redis.Client
		want   bool
	}{
		{
			name:   "should create repository with valid client",
			client: &redis.Client{},
			want:   true,
		},
		{
			name:   "should create repository with nil client",
			client: nil,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewTaskRepository(tt.client)
			assert.NotNil(t, repo)
			assert.Equal(t, tt.client, repo.client)
		})
	}
}

func TestTaskRepository_CreateTask(t *testing.T) {
	repo, s := setupTestTaskRepository(t)
	defer s.Close()

	userID := uuid.New().String()

	tests := []struct {
		name    string
		task    *domain.Task
		wantErr bool
		errCode string
	}{
		{
			name:    "should create valid task successfully",
			task:    createTestTask(userID, "Test task", "work"),
			wantErr: false,
		},
		{
			name:    "should fail with nil task",
			task:    nil,
			wantErr: true,
			errCode: "2001",
		},
		{
			name: "should fail with invalid task (empty description)",
			task: &domain.Task{
				ID:          uuid.New().String(),
				UserID:      userID,
				Description: "",
				Category:    "work",
				Completed:   false,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: true,
			errCode: "2002",
		},
		{
			name: "should fail with invalid task (empty user ID)",
			task: &domain.Task{
				ID:          uuid.New().String(),
				UserID:      "",
				Description: "Test task",
				Category:    "work",
				Completed:   false,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: true,
			errCode: "2002",
		},
		{
			name: "should create task with empty category",
			task: &domain.Task{
				ID:          uuid.New().String(),
				UserID:      userID,
				Description: "Test task without category",
				Category:    "",
				Completed:   false,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := repo.CreateTask(tt.task)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != "" {
					assert.Contains(t, err.Error(), tt.errCode)
				}
				return
			}

			assert.NoError(t, err)

			// Verify task was stored in hash
			ctx = context.Background()
			taskKey := redis.GenerateKey(redis.TaskKeyPrefix, tt.task.ID)
			exists := repo.client.Exists(ctx, taskKey).Val()
			assert.Equal(t, int64(1), exists)

			// Verify task was added to user's task set
			userTasksKey := redis.GenerateKey("user", tt.task.UserID) + ":tasks"
			isMember := repo.client.SIsMember(ctx, userTasksKey, tt.task.ID).Val()
			assert.True(t, isMember)

			// Verify task was added to sorted set with timestamp
			userTasksSortedKey := redis.GenerateKey("user", tt.task.UserID) + ":tasks:sorted"
			score := repo.client.ZScore(ctx, userTasksSortedKey, tt.task.ID).Val()
			assert.Greater(t, score, float64(0))

			// Verify category was added if not empty
			if tt.task.Category != "" {
				userCategoriesKey := redis.GenerateKey("user", tt.task.UserID) + ":categories"
				isMember := repo.client.SIsMember(ctx, userCategoriesKey, tt.task.Category).Val()
				assert.True(t, isMember)

				categoryTasksKey := redis.GenerateKey("user", tt.task.UserID) + ":category:" + tt.task.Category
				isMember = repo.client.SIsMember(ctx, categoryTasksKey, tt.task.ID).Val()
				assert.True(t, isMember)
			}
		})
	}
}

func TestTaskRepository_GetTaskByID(t *testing.T) {
	repo, s := setupTestTaskRepository(t)
	defer s.Close()

	userID := uuid.New().String()

	// Create a test task
	task := createTestTask(userID, "Test task", "work")
	err := repo.CreateTask(task)
	require.NoError(t, err)

	tests := []struct {
		name       string
		taskID     string
		wantErr    bool
		errCode    string
		wantTask   *domain.Task
	}{
		{
			name:       "should get existing task",
			taskID:     task.ID,
			wantErr:    false,
			wantTask:   task,
		},
		{
			name:       "should fail with empty task ID",
			taskID:     "",
			wantErr:    true,
			errCode:    "2004",
		},
		{
			name:       "should fail with non-existent task ID",
			taskID:     uuid.New().String(),
			wantErr:    true,
			errCode:    "2003",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetTaskByID(tt.taskID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errCode != "" {
					assert.Contains(t, err.Error(), tt.errCode)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.wantTask.ID, result.ID)
			assert.Equal(t, tt.wantTask.UserID, result.UserID)
			assert.Equal(t, tt.wantTask.Description, result.Description)
			assert.Equal(t, tt.wantTask.Category, result.Category)
			assert.Equal(t, tt.wantTask.Completed, result.Completed)
		})
	}
}

func TestTaskRepository_ListTasks(t *testing.T) {
	repo, s := setupTestTaskRepository(t)
	defer s.Close()

	userID := uuid.New().String()
	otherUserID := uuid.New().String()

	// Create test tasks
	task1 := createTestTask(userID, "Task 1", "work")
	task2 := createTestTask(userID, "Task 2", "personal")
	task3 := createTestTask(userID, "Task 3", "work")
	task4 := createTestTask(otherUserID, "Other user task", "work")

	// Mark task2 as completed
	task2.MarkCompleted()

	// Soft delete task3
	task3.SoftDelete()

	// Create tasks
	require.NoError(t, repo.CreateTask(task1))
	require.NoError(t, repo.CreateTask(task2))
	require.NoError(t, repo.CreateTask(task3))
	require.NoError(t, repo.CreateTask(task4))

	// Update task2 completion status
	require.NoError(t, repo.UpdateTaskCompletion(task2.ID, true))

	// Soft delete task3
	require.NoError(t, repo.SoftDeleteTask(task3.ID))

	tests := []struct {
		name        string
		userID      string
		filters     domain.TaskFilters
		wantCount   int
		wantTaskIDs []string
		wantErr     bool
		errCode     string
	}{
		{
			name:        "should list all active tasks for user",
			userID:      userID,
			filters:     domain.TaskFilters{},
			wantCount:   2,
			wantTaskIDs: []string{task1.ID, task2.ID},
			wantErr:     false,
		},
		{
			name:   "should filter by category",
			userID: userID,
			filters: domain.TaskFilters{
				Category: "work",
			},
			wantCount:   1,
			wantTaskIDs: []string{task1.ID},
			wantErr:     false,
		},
		{
			name:   "should filter by completed status",
			userID: userID,
			filters: domain.TaskFilters{
				Completed: boolPtr(true),
			},
			wantCount:   1,
			wantTaskIDs: []string{task2.ID},
			wantErr:     false,
		},
		{
			name:   "should include deleted tasks when requested",
			userID: userID,
			filters: domain.TaskFilters{
				IncludeDeleted: true,
			},
			wantCount:   3,
			wantTaskIDs: []string{task1.ID, task2.ID, task3.ID},
			wantErr:     false,
		},
		{
			name:   "should apply limit",
			userID: userID,
			filters: domain.TaskFilters{
				Limit: 1,
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:   "should apply offset",
			userID: userID,
			filters: domain.TaskFilters{
				Offset: 1,
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:   "should return empty list for user with no tasks",
			userID: uuid.New().String(),
			filters: domain.TaskFilters{},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:   "should fail with invalid filters",
			userID: userID,
			filters: domain.TaskFilters{
				Limit: -1,
			},
			wantErr: true,
			errCode: "2005",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.ListTasks(tt.userID, tt.filters)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errCode != "" {
					assert.Contains(t, err.Error(), tt.errCode)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.wantCount, len(result))

			if len(tt.wantTaskIDs) > 0 {
				resultIDs := make([]string, len(result))
				for i, task := range result {
					resultIDs[i] = task.ID
				}
				for _, expectedID := range tt.wantTaskIDs {
					assert.Contains(t, resultIDs, expectedID)
				}
			}

			// Verify tasks are sorted by CreatedAt desc
			if len(result) > 1 {
				for i := 1; i < len(result); i++ {
					assert.True(t, result[i-1].CreatedAt.After(result[i].CreatedAt) || 
						result[i-1].CreatedAt.Equal(result[i].CreatedAt))
				}
			}
		})
	}
}

func TestTaskRepository_UpdateTaskCompletion(t *testing.T) {
	repo, s := setupTestTaskRepository(t)
	defer s.Close()

	userID := uuid.New().String()

	// Create a test task
	task := createTestTask(userID, "Test task", "work")
	err := repo.CreateTask(task)
	require.NoError(t, err)

	tests := []struct {
		name      string
		taskID    string
		completed bool
		wantErr   bool
		errCode   string
	}{
		{
			name:      "should update task completion to true",
			taskID:    task.ID,
			completed: true,
			wantErr:   false,
		},
		{
			name:      "should update task completion to false",
			taskID:    task.ID,
			completed: false,
			wantErr:   false,
		},
		{
			name:      "should fail with non-existent task",
			taskID:    uuid.New().String(),
			completed: true,
			wantErr:   true,
			errCode:   "2003",
		},
		{
			name:      "should fail with empty task ID",
			taskID:    "",
			completed: true,
			wantErr:   true,
			errCode:   "2004",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.UpdateTaskCompletion(tt.taskID, tt.completed)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != "" {
					assert.Contains(t, err.Error(), tt.errCode)
				}
				return
			}

			assert.NoError(t, err)

			// Verify the task was updated
			updatedTask, err := repo.GetTaskByID(tt.taskID)
			assert.NoError(t, err)
			assert.Equal(t, tt.completed, updatedTask.Completed)
			
			// Verify UpdatedAt was modified (allowing for second-precision due to Redis storage)
			assert.False(t, updatedTask.UpdatedAt.Before(task.UpdatedAt.Truncate(time.Second)))
		})
	}
}

func TestTaskRepository_SoftDeleteTask(t *testing.T) {
	repo, s := setupTestTaskRepository(t)
	defer s.Close()

	userID := uuid.New().String()

	// Create a test task
	task := createTestTask(userID, "Test task", "work")
	err := repo.CreateTask(task)
	require.NoError(t, err)

	tests := []struct {
		name    string
		taskID  string
		wantErr bool
		errCode string
	}{
		{
			name:    "should soft delete task successfully",
			taskID:  task.ID,
			wantErr: false,
		},
		{
			name:    "should fail with non-existent task",
			taskID:  uuid.New().String(),
			wantErr: true,
			errCode: "2003",
		},
		{
			name:    "should fail with empty task ID",
			taskID:  "",
			wantErr: true,
			errCode: "2004",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.SoftDeleteTask(tt.taskID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != "" {
					assert.Contains(t, err.Error(), tt.errCode)
				}
				return
			}

			assert.NoError(t, err)

			// Get the task to find its user ID for verification
			updatedTask, err := repo.GetTaskByID(tt.taskID)
			assert.NoError(t, err)
			userID := updatedTask.UserID

			// Verify task was moved to deleted sorted set
			ctx := context.Background()
			userDeletedKey := redis.GenerateKey("user", userID) + ":tasks:deleted"
			score := repo.client.ZScore(ctx, userDeletedKey, tt.taskID).Val()
			assert.Greater(t, score, float64(0))

			// Verify task was removed from active sets
			userTasksKey := redis.GenerateKey("user", userID) + ":tasks"
			isMember := repo.client.SIsMember(ctx, userTasksKey, tt.taskID).Val()
			assert.False(t, isMember)

			userTasksSortedKey := redis.GenerateKey("user", userID) + ":tasks:sorted"
			_, err = repo.client.ZScore(ctx, userTasksSortedKey, tt.taskID).Result()
			assert.Error(t, err) // Should return error because task is not in sorted set

			// Verify task is marked as deleted but still exists in hash
			assert.True(t, updatedTask.IsDeleted())
		})
	}
}

func TestTaskRepository_RestoreTask(t *testing.T) {
	repo, s := setupTestTaskRepository(t)
	defer s.Close()

	userID := uuid.New().String()

	// Create and soft delete a test task
	task := createTestTask(userID, "Test task", "work")
	err := repo.CreateTask(task)
	require.NoError(t, err)
	err = repo.SoftDeleteTask(task.ID)
	require.NoError(t, err)

	tests := []struct {
		name    string
		taskID  string
		wantErr bool
		errCode string
	}{
		{
			name:    "should restore soft deleted task successfully",
			taskID:  task.ID,
			wantErr: false,
		},
		{
			name:    "should fail with non-existent task",
			taskID:  uuid.New().String(),
			wantErr: true,
			errCode: "2003",
		},
		{
			name:    "should fail with empty task ID",
			taskID:  "",
			wantErr: true,
			errCode: "2004",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.RestoreTask(tt.taskID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != "" {
					assert.Contains(t, err.Error(), tt.errCode)
				}
				return
			}

			assert.NoError(t, err)

			// Get the task to find its user ID for verification
			restoredTask, err := repo.GetTaskByID(tt.taskID)
			assert.NoError(t, err)
			userID := restoredTask.UserID

			// Verify task was restored to active sets
			ctx := context.Background()
			userTasksKey := redis.GenerateKey("user", userID) + ":tasks"
			isMember := repo.client.SIsMember(ctx, userTasksKey, tt.taskID).Val()
			assert.True(t, isMember)

			userTasksSortedKey := redis.GenerateKey("user", userID) + ":tasks:sorted"
			score := repo.client.ZScore(ctx, userTasksSortedKey, tt.taskID).Val()
			assert.Greater(t, score, float64(0))

			// Verify task was removed from deleted set
			userDeletedKey := redis.GenerateKey("user", userID) + ":tasks:deleted"
			_, err = repo.client.ZScore(ctx, userDeletedKey, tt.taskID).Result()
			assert.Error(t, err) // Should return error because task is not in deleted set

			// Verify task is not marked as deleted
			assert.False(t, restoredTask.IsDeleted())
		})
	}
}

func TestTaskRepository_GetUserCategories(t *testing.T) {
	repo, s := setupTestTaskRepository(t)
	defer s.Close()

	userID := uuid.New().String()
	otherUserID := uuid.New().String()

	// Create tasks with different categories
	task1 := createTestTask(userID, "Task 1", "work")
	task2 := createTestTask(userID, "Task 2", "personal")
	task3 := createTestTask(userID, "Task 3", "work") // Duplicate category
	task4 := createTestTask(otherUserID, "Other task", "work")

	require.NoError(t, repo.CreateTask(task1))
	require.NoError(t, repo.CreateTask(task2))
	require.NoError(t, repo.CreateTask(task3))
	require.NoError(t, repo.CreateTask(task4))

	tests := []struct {
		name           string
		userID         string
		wantCategories []string
		wantErr        bool
		errCode        string
	}{
		{
			name:           "should get unique categories for user",
			userID:         userID,
			wantCategories: []string{"work", "personal"},
			wantErr:        false,
		},
		{
			name:           "should get categories for other user",
			userID:         otherUserID,
			wantCategories: []string{"work"},
			wantErr:        false,
		},
		{
			name:           "should return empty list for user with no tasks",
			userID:         uuid.New().String(),
			wantCategories: []string{},
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetUserCategories(tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errCode != "" {
					assert.Contains(t, err.Error(), tt.errCode)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, len(tt.wantCategories), len(result))

			for _, expectedCategory := range tt.wantCategories {
				assert.Contains(t, result, expectedCategory)
			}
		})
	}
}

func TestTaskRepository_RenameCategory(t *testing.T) {
	repo, s := setupTestTaskRepository(t)
	defer s.Close()

	userID := uuid.New().String()
	otherUserID := uuid.New().String()

	// Create tasks with categories
	task1 := createTestTask(userID, "Task 1", "work")
	task2 := createTestTask(userID, "Task 2", "work")
	task3 := createTestTask(userID, "Task 3", "personal")
	task4 := createTestTask(otherUserID, "Other task", "work")

	require.NoError(t, repo.CreateTask(task1))
	require.NoError(t, repo.CreateTask(task2))
	require.NoError(t, repo.CreateTask(task3))
	require.NoError(t, repo.CreateTask(task4))

	tests := []struct {
		name    string
		userID  string
		oldName string
		newName string
		wantErr bool
		errCode string
	}{
		{
			name:    "should rename category successfully",
			userID:  userID,
			oldName: "work",
			newName: "office",
			wantErr: false,
		},
		{
			name:    "should fail with non-existent category",
			userID:  userID,
			oldName: "nonexistent",
			newName: "new",
			wantErr: true,
			errCode: "2006",
		},
		{
			name:    "should fail with empty old name",
			userID:  userID,
			oldName: "",
			newName: "new",
			wantErr: true,
			errCode: "2007",
		},
		{
			name:    "should fail with empty new name",
			userID:  userID,
			oldName: "personal",
			newName: "",
			wantErr: true,
			errCode: "2007",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.RenameCategory(tt.userID, tt.oldName, tt.newName)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != "" {
					assert.Contains(t, err.Error(), tt.errCode)
				}
				return
			}

			assert.NoError(t, err)

			// Verify category was renamed in user's categories set
			categories, err := repo.GetUserCategories(tt.userID)
			assert.NoError(t, err)
			assert.Contains(t, categories, tt.newName)
			assert.NotContains(t, categories, tt.oldName)

			// Verify tasks with the old category now have the new category
			tasks, err := repo.ListTasks(tt.userID, domain.TaskFilters{Category: tt.newName})
			assert.NoError(t, err)
			assert.Greater(t, len(tasks), 0)

			for _, task := range tasks {
				assert.Equal(t, tt.newName, task.Category)
			}

			// Verify old category set is empty/removed
			ctx := context.Background()
			oldCategoryKey := redis.GenerateKey("user", tt.userID) + ":category:" + tt.oldName
			count := repo.client.SCard(ctx, oldCategoryKey).Val()
			assert.Equal(t, int64(0), count)
		})
	}
}

func TestTaskRepository_DeleteCategory(t *testing.T) {
	repo, s := setupTestTaskRepository(t)
	defer s.Close()

	userID := uuid.New().String()

	// Create tasks with categories
	task1 := createTestTask(userID, "Task 1", "work")
	task2 := createTestTask(userID, "Task 2", "work")
	task3 := createTestTask(userID, "Task 3", "personal")

	require.NoError(t, repo.CreateTask(task1))
	require.NoError(t, repo.CreateTask(task2))
	require.NoError(t, repo.CreateTask(task3))

	tests := []struct {
		name         string
		userID       string
		categoryName string
		wantErr      bool
		errCode      string
	}{
		{
			name:         "should delete category successfully",
			userID:       userID,
			categoryName: "work",
			wantErr:      false,
		},
		{
			name:         "should fail with non-existent category",
			userID:       userID,
			categoryName: "nonexistent",
			wantErr:      true,
			errCode:      "2006",
		},
		{
			name:         "should fail with empty category name",
			userID:       userID,
			categoryName: "",
			wantErr:      true,
			errCode:      "2007",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.DeleteCategory(tt.userID, tt.categoryName)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != "" {
					assert.Contains(t, err.Error(), tt.errCode)
				}
				return
			}

			assert.NoError(t, err)

			// Verify category was removed from user's categories set
			categories, err := repo.GetUserCategories(tt.userID)
			assert.NoError(t, err)
			assert.NotContains(t, categories, tt.categoryName)

			// Verify tasks with the deleted category now have empty category
			tasks, err := repo.ListTasks(tt.userID, domain.TaskFilters{})
			assert.NoError(t, err)

			for _, task := range tasks {
				if task.ID == task1.ID || task.ID == task2.ID {
					assert.Equal(t, "", task.Category)
				}
			}

			// Verify category set is removed
			ctx := context.Background()
			categoryKey := redis.GenerateKey("user", tt.userID) + ":category:" + tt.categoryName
			count := repo.client.SCard(ctx, categoryKey).Val()
			assert.Equal(t, int64(0), count)
		})
	}
}

func TestTaskRepository_CleanupExpiredTasks(t *testing.T) {
	repo, s := setupTestTaskRepository(t)
	defer s.Close()

	userID := uuid.New().String()
	ctx := context.Background()

	// Create and immediately delete tasks to simulate expired tasks
	oldTask := createTestTask(userID, "Old task", "work")
	recentTask := createTestTask(userID, "Recent task", "personal")

	require.NoError(t, repo.CreateTask(oldTask))
	require.NoError(t, repo.CreateTask(recentTask))

	// Soft delete both tasks
	require.NoError(t, repo.SoftDeleteTask(oldTask.ID))
	require.NoError(t, repo.SoftDeleteTask(recentTask.ID))

	// Manually set the old task's deletion time to more than 7 days ago
	ctx = context.Background()
	userDeletedKey := redis.GenerateKey("user", userID) + ":tasks:deleted"
	expiredTime := time.Now().AddDate(0, 0, -8).Unix() // 8 days ago
	repo.client.ZAdd(ctx, userDeletedKey, redislib.Z{
		Score:  float64(expiredTime),
		Member: oldTask.ID,
	})

	tests := []struct {
		name        string
		wantCleaned int
		wantErr     bool
		errCode     string
	}{
		{
			name:        "should cleanup expired tasks",
			wantCleaned: 1,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanedCount, err := repo.CleanupExpiredTasks()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != "" {
					assert.Contains(t, err.Error(), tt.errCode)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCleaned, cleanedCount)

			// Verify old task was completely removed
			ctx := context.Background()
			taskKey := redis.GenerateKey(redis.TaskKeyPrefix, oldTask.ID)
			exists := repo.client.Exists(ctx, taskKey).Val()
			assert.Equal(t, int64(0), exists)

			// Verify recent task still exists in deleted set
			score := repo.client.ZScore(ctx, userDeletedKey, recentTask.ID).Val()
			assert.Greater(t, score, float64(0))

			// Verify recent task hash still exists
			recentTaskKey := redis.GenerateKey(redis.TaskKeyPrefix, recentTask.ID)
			exists = repo.client.Exists(ctx, recentTaskKey).Val()
			assert.Equal(t, int64(1), exists)
		})
	}
}

// Helper function to create bool pointer
func boolPtr(b bool) *bool {
	return &b
}