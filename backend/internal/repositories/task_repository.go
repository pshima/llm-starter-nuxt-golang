package repositories

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"backend/internal/domain"
	"backend/pkg/redis"

	redislib "github.com/redis/go-redis/v9"
)

// TaskRepository implements the domain.TaskRepository interface
// Provides Redis-based storage for task data with comprehensive Redis data structures
type TaskRepository struct {
	client *redis.Client
}

// NewTaskRepository creates a new TaskRepository instance
// Takes a Redis client and returns a configured task repository
func NewTaskRepository(client *redis.Client) *TaskRepository {
	return &TaskRepository{
		client: client,
	}
}

// CreateTask stores a new task in Redis with proper data structure management
// Creates task hash and updates all related sets (user tasks, categories, sorted sets)
// Error codes: 2001 (nil task), 2002 (validation error)
func (r *TaskRepository) CreateTask(task *domain.Task) error {
	ctx := context.Background()
	if task == nil {
		return fmt.Errorf("2001: task cannot be nil")
	}

	// Validate task
	if err := task.Validate(); err != nil {
		return fmt.Errorf("2002: %w", err)
	}

	// Serialize task to hash fields
	taskData := map[string]interface{}{
		"id":          task.ID,
		"user_id":     task.UserID,
		"description": task.Description,
		"category":    task.Category,
		"completed":   task.Completed,
		"created_at":  task.CreatedAt.Unix(),
		"updated_at":  task.UpdatedAt.Unix(),
	}

	if task.DeletedAt != nil {
		taskData["deleted_at"] = task.DeletedAt.Unix()
	}

	// Use Redis pipeline for atomic operations
	pipe := r.client.TxPipeline()

	// Store task hash
	taskKey := redis.GenerateKey(redis.TaskKeyPrefix, task.ID)
	pipe.HMSet(ctx, taskKey, taskData)

	// Add to user's active tasks set
	userTasksKey := redis.GenerateKey("user", task.UserID) + ":tasks"
	pipe.SAdd(ctx, userTasksKey, task.ID)

	// Add to user's sorted tasks set (sorted by created timestamp)
	userTasksSortedKey := redis.GenerateKey("user", task.UserID) + ":tasks:sorted"
	pipe.ZAdd(ctx, userTasksSortedKey, redislib.Z{
		Score:  float64(task.CreatedAt.Unix()),
		Member: task.ID,
	})

	// Handle category management if category is not empty
	if strings.TrimSpace(task.Category) != "" {
		// Add category to user's categories set
		userCategoriesKey := redis.GenerateKey("user", task.UserID) + ":categories"
		pipe.SAdd(ctx, userCategoriesKey, task.Category)

		// Add task to category set
		categoryTasksKey := redis.GenerateKey("user", task.UserID) + ":category:" + task.Category
		pipe.SAdd(ctx, categoryTasksKey, task.ID)
	}

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("2002: failed to create task: %w", err)
	}

	return nil
}

// GetTaskByID retrieves a task by ID with access control validation
// Returns task data from Redis hash 
// Error codes: 2003 (not found/access denied), 2004 (invalid ID)
func (r *TaskRepository) GetTaskByID(taskID string) (*domain.Task, error) {
	ctx := context.Background()
	if strings.TrimSpace(taskID) == "" {
		return nil, fmt.Errorf("2004: task ID cannot be empty")
	}

	taskKey := redis.GenerateKey(redis.TaskKeyPrefix, taskID)
	
	// Get task hash data
	taskData, err := r.client.HGetAll(ctx, taskKey).Result()
	if err != nil {
		return nil, fmt.Errorf("2003: failed to get task: %w", err)
	}

	if len(taskData) == 0 {
		return nil, fmt.Errorf("2003: task not found")
	}

	// Note: Access control should be handled at service layer
	// Repository layer focuses on data access only

	// Parse task data
	task, err := r.parseTaskFromHash(taskData)
	if err != nil {
		return nil, fmt.Errorf("2003: failed to parse task: %w", err)
	}

	return task, nil
}

// ListTasks retrieves tasks for a user with filtering and pagination support
// Supports filtering by category, completion status, deleted status, and pagination
// Error codes: 2005 (invalid filters)
func (r *TaskRepository) ListTasks(userID string, filters domain.TaskFilters) ([]*domain.Task, error) {
	ctx := context.Background()
	// Validate filters
	if err := filters.Validate(); err != nil {
		return nil, fmt.Errorf("2005: %w", err)
	}

	var taskIDs []string
	var err error

	// Get task IDs based on filters
	if filters.IncludeDeleted {
		// Get all tasks (active + deleted)
		taskIDs, err = r.getAllTaskIDs(ctx, userID, filters)
	} else {
		// Get only active tasks
		taskIDs, err = r.getActiveTaskIDs(ctx, userID, filters)
	}

	if err != nil {
		return nil, fmt.Errorf("2005: failed to get task IDs: %w", err)
	}

	// Apply pagination
	start := filters.Offset
	end := len(taskIDs)
	if filters.Limit > 0 && start + filters.Limit < end {
		end = start + filters.Limit
	}

	if start >= len(taskIDs) {
		return []*domain.Task{}, nil
	}

	if start < 0 {
		start = 0
	}

	taskIDs = taskIDs[start:end]

	// Fetch task details
	tasks := make([]*domain.Task, 0, len(taskIDs))
	for _, taskID := range taskIDs {
		task, err := r.GetTaskByID(taskID)
		if err != nil {
			continue // Skip tasks that can't be retrieved
		}

		// Verify task belongs to user
		if task.UserID != userID {
			continue
		}

		// Apply completion filter
		if filters.Completed != nil && task.Completed != *filters.Completed {
			continue
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// UpdateTaskCompletion updates the completion status of a task
// Updates task hash and timestamp 
// Error codes: 2003 (not found), 2004 (invalid ID)
func (r *TaskRepository) UpdateTaskCompletion(taskID string, completed bool) error {
	ctx := context.Background()
	if strings.TrimSpace(taskID) == "" {
		return fmt.Errorf("2004: task ID cannot be empty")
	}

	// Verify task exists
	task, err := r.GetTaskByID(taskID)
	if err != nil {
		return err // Error code already included
	}

	// Update task
	task.Completed = completed
	task.UpdatedAt = time.Now()

	// Update task hash
	taskKey := redis.GenerateKey(redis.TaskKeyPrefix, taskID)
	_, err = r.client.HMSet(ctx, taskKey, map[string]interface{}{
		"completed":  completed,
		"updated_at": task.UpdatedAt.Unix(),
	}).Result()

	if err != nil {
		return fmt.Errorf("2003: failed to update task completion: %w", err)
	}

	return nil
}

// SoftDeleteTask marks a task as deleted by moving it to deleted sorted set
// Removes from active sets and adds to deleted set with expiry tracking
// Error codes: 2003 (not found), 2004 (invalid ID)
func (r *TaskRepository) SoftDeleteTask(taskID string) error {
	ctx := context.Background()
	if strings.TrimSpace(taskID) == "" {
		return fmt.Errorf("2004: task ID cannot be empty")
	}

	// Verify task exists
	task, err := r.GetTaskByID(taskID)
	if err != nil {
		return err // Error code already included
	}

	userID := task.UserID
	now := time.Now()

	// Use pipeline for atomic operations
	pipe := r.client.TxPipeline()

	// Update task hash with deleted timestamp
	taskKey := redis.GenerateKey(redis.TaskKeyPrefix, taskID)
	pipe.HMSet(ctx, taskKey, map[string]interface{}{
		"deleted_at": now.Unix(),
		"updated_at": now.Unix(),
	})

	// Remove from active sets
	userTasksKey := redis.GenerateKey("user", userID) + ":tasks"
	pipe.SRem(ctx, userTasksKey, taskID)

	userTasksSortedKey := redis.GenerateKey("user", userID) + ":tasks:sorted"
	pipe.ZRem(ctx, userTasksSortedKey, taskID)

	// Remove from category set if task has category
	if strings.TrimSpace(task.Category) != "" {
		categoryTasksKey := redis.GenerateKey("user", userID) + ":category:" + task.Category
		pipe.SRem(ctx, categoryTasksKey, taskID)
	}

	// Add to deleted sorted set with deletion timestamp as score
	userDeletedKey := redis.GenerateKey("user", userID) + ":tasks:deleted"
	pipe.ZAdd(ctx, userDeletedKey, redislib.Z{
		Score:  float64(now.Unix()),
		Member: taskID,
	})

	// Note: Redis sorted sets don't have individual expiry, so we rely on CleanupExpiredTasks
	// to periodically clean up expired deleted tasks

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("2003: failed to soft delete task: %w", err)
	}

	return nil
}

// RestoreTask restores a soft-deleted task to active status
// Moves task from deleted set back to active sets and clears deletion timestamp
// Error codes: 2003 (not found), 2004 (invalid ID)
func (r *TaskRepository) RestoreTask(taskID string) error {
	ctx := context.Background()
	if strings.TrimSpace(taskID) == "" {
		return fmt.Errorf("2004: task ID cannot be empty")
	}

	// Verify task exists
	task, err := r.GetTaskByID(taskID)
	if err != nil {
		return err // Error code already included
	}

	userID := task.UserID
	now := time.Now()

	// Use pipeline for atomic operations
	pipe := r.client.TxPipeline()

	// Update task hash to remove deleted timestamp
	taskKey := redis.GenerateKey(redis.TaskKeyPrefix, taskID)
	pipe.HDel(ctx, taskKey, "deleted_at")
	pipe.HMSet(ctx, taskKey, map[string]interface{}{
		"updated_at": now.Unix(),
	})

	// Add back to active sets
	userTasksKey := redis.GenerateKey("user", userID) + ":tasks"
	pipe.SAdd(ctx, userTasksKey, taskID)

	userTasksSortedKey := redis.GenerateKey("user", userID) + ":tasks:sorted"
	pipe.ZAdd(ctx, userTasksSortedKey, redislib.Z{
		Score:  float64(task.CreatedAt.Unix()),
		Member: taskID,
	})

	// Add back to category set if task has category
	if strings.TrimSpace(task.Category) != "" {
		categoryTasksKey := redis.GenerateKey("user", userID) + ":category:" + task.Category
		pipe.SAdd(ctx, categoryTasksKey, taskID)
	}

	// Remove from deleted set
	userDeletedKey := redis.GenerateKey("user", userID) + ":tasks:deleted"
	pipe.ZRem(ctx, userDeletedKey, taskID)

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("2003: failed to restore task: %w", err)
	}

	return nil
}

// GetUserCategories retrieves all unique categories for a user
// Returns sorted list of category names from user's categories set
func (r *TaskRepository) GetUserCategories(userID string) ([]string, error) {
	ctx := context.Background()
	userCategoriesKey := redis.GenerateKey("user", userID) + ":categories"
	
	categories, err := r.client.SMembers(ctx, userCategoriesKey).Result()
	if err != nil {
		return []string{}, nil // Return empty slice instead of error
	}

	// Filter out empty categories
	filteredCategories := make([]string, 0, len(categories))
	for _, category := range categories {
		if strings.TrimSpace(category) != "" {
			filteredCategories = append(filteredCategories, category)
		}
	}

	return filteredCategories, nil
}

// RenameCategory renames a category across all user's tasks
// Updates category name in all tasks and category sets
// Error codes: 2006 (category not found), 2007 (invalid names)
func (r *TaskRepository) RenameCategory(userID, oldName, newName string) error {
	ctx := context.Background()
	if strings.TrimSpace(oldName) == "" || strings.TrimSpace(newName) == "" {
		return fmt.Errorf("2007: category names cannot be empty")
	}

	// Check if old category exists
	userCategoriesKey := redis.GenerateKey("user", userID) + ":categories"
	exists, err := r.client.SIsMember(ctx, userCategoriesKey, oldName).Result()
	if err != nil {
		return fmt.Errorf("2006: failed to check category existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("2006: category not found")
	}

	// Get all tasks in the old category
	oldCategoryKey := redis.GenerateKey("user", userID) + ":category:" + oldName
	taskIDs, err := r.client.SMembers(ctx, oldCategoryKey).Result()
	if err != nil {
		return fmt.Errorf("2006: failed to get tasks in category: %w", err)
	}

	// Use pipeline for atomic operations
	pipe := r.client.TxPipeline()

	// Update category in user's categories set
	pipe.SRem(ctx, userCategoriesKey, oldName)
	pipe.SAdd(ctx, userCategoriesKey, newName)

	// Update each task's category in hash
	for _, taskID := range taskIDs {
		taskKey := redis.GenerateKey(redis.TaskKeyPrefix, taskID)
		pipe.HSet(ctx, taskKey, "category", newName)
		pipe.HSet(ctx, taskKey, "updated_at", time.Now().Unix())
	}

	// Move tasks to new category set
	newCategoryKey := redis.GenerateKey("user", userID) + ":category:" + newName
	for _, taskID := range taskIDs {
		pipe.SAdd(ctx, newCategoryKey, taskID)
	}

	// Remove old category set
	pipe.Del(ctx, oldCategoryKey)

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("2006: failed to rename category: %w", err)
	}

	return nil
}

// DeleteCategory removes a category from all user's tasks
// Sets category to empty string for all tasks in the category
// Error codes: 2006 (category not found), 2007 (invalid name)
func (r *TaskRepository) DeleteCategory(userID, categoryName string) error {
	ctx := context.Background()
	if strings.TrimSpace(categoryName) == "" {
		return fmt.Errorf("2007: category name cannot be empty")
	}

	// Check if category exists
	userCategoriesKey := redis.GenerateKey("user", userID) + ":categories"
	exists, err := r.client.SIsMember(ctx, userCategoriesKey, categoryName).Result()
	if err != nil {
		return fmt.Errorf("2006: failed to check category existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("2006: category not found")
	}

	// Get all tasks in the category
	categoryKey := redis.GenerateKey("user", userID) + ":category:" + categoryName
	taskIDs, err := r.client.SMembers(ctx, categoryKey).Result()
	if err != nil {
		return fmt.Errorf("2006: failed to get tasks in category: %w", err)
	}

	// Use pipeline for atomic operations
	pipe := r.client.TxPipeline()

	// Remove category from user's categories set
	pipe.SRem(ctx, userCategoriesKey, categoryName)

	// Update each task's category to empty string
	for _, taskID := range taskIDs {
		taskKey := redis.GenerateKey(redis.TaskKeyPrefix, taskID)
		pipe.HSet(ctx, taskKey, "category", "")
		pipe.HSet(ctx, taskKey, "updated_at", time.Now().Unix())
	}

	// Remove category set
	pipe.Del(ctx, categoryKey)

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("2006: failed to delete category: %w", err)
	}

	return nil
}

// CleanupExpiredTasks removes tasks that have been soft-deleted for more than 7 days
// Completely removes task hashes and cleans up any remaining references
// Returns the number of tasks cleaned up
func (r *TaskRepository) CleanupExpiredTasks() (int, error) {
	ctx := context.Background()
	// Find all users with deleted tasks
	pattern := "user:*:tasks:deleted"
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to find deleted task sets: %w", err)
	}

	totalCleaned := 0
	expiredTime := time.Now().AddDate(0, 0, -7).Unix() // 7 days ago

	for _, deletedSetKey := range keys {
		// Get expired tasks (deleted more than 7 days ago)
		expiredTasks, err := r.client.ZRangeByScore(ctx, deletedSetKey, &redislib.ZRangeBy{
			Min: "-inf",
			Max: fmt.Sprintf("%d", expiredTime),
		}).Result()

		if err != nil {
			continue // Skip this set on error
		}

		if len(expiredTasks) == 0 {
			continue
		}

		// Use pipeline for cleanup operations
		pipe := r.client.TxPipeline()

		for _, taskID := range expiredTasks {
			// Remove task hash completely
			taskKey := redis.GenerateKey(redis.TaskKeyPrefix, taskID)
			pipe.Del(ctx, taskKey)

			// Remove from deleted set
			pipe.ZRem(ctx, deletedSetKey, taskID)
		}

		// Execute cleanup pipeline
		_, err = pipe.Exec(ctx)
		if err != nil {
			continue // Skip this batch on error
		}

		totalCleaned += len(expiredTasks)
	}

	return totalCleaned, nil
}

// Helper methods

// parseTaskFromHash converts Redis hash data to Task struct
func (r *TaskRepository) parseTaskFromHash(data map[string]string) (*domain.Task, error) {
	task := &domain.Task{
		ID:          data["id"],
		UserID:      data["user_id"],
		Description: data["description"],
		Category:    data["category"],
	}

	// Parse boolean completed field
	if data["completed"] == "true" || data["completed"] == "1" {
		task.Completed = true
	}

	// Parse timestamps
	if createdAtStr := data["created_at"]; createdAtStr != "" {
		if createdAt, err := parseUnixTimestamp(createdAtStr); err == nil {
			task.CreatedAt = createdAt
		}
	}

	if updatedAtStr := data["updated_at"]; updatedAtStr != "" {
		if updatedAt, err := parseUnixTimestamp(updatedAtStr); err == nil {
			task.UpdatedAt = updatedAt
		}
	}

	if deletedAtStr := data["deleted_at"]; deletedAtStr != "" {
		if deletedAt, err := parseUnixTimestamp(deletedAtStr); err == nil {
			task.DeletedAt = &deletedAt
		}
	}

	return task, nil
}

// getActiveTaskIDs retrieves task IDs from active sets with optional category filtering
func (r *TaskRepository) getActiveTaskIDs(ctx context.Context, userID string, filters domain.TaskFilters) ([]string, error) {
	var taskIDs []string
	var err error

	if strings.TrimSpace(filters.Category) != "" {
		// Get tasks from specific category
		categoryKey := redis.GenerateKey("user", userID) + ":category:" + filters.Category
		taskIDs, err = r.client.SMembers(ctx, categoryKey).Result()
	} else {
		// Get all active tasks sorted by creation time (descending)
		userTasksSortedKey := redis.GenerateKey("user", userID) + ":tasks:sorted"
		taskIDs, err = r.client.ZRevRange(ctx, userTasksSortedKey, 0, -1).Result()
	}

	return taskIDs, err
}

// getAllTaskIDs retrieves task IDs from both active and deleted sets
func (r *TaskRepository) getAllTaskIDs(ctx context.Context, userID string, filters domain.TaskFilters) ([]string, error) {
	// Get active tasks
	activeIDs, err := r.getActiveTaskIDs(ctx, userID, filters)
	if err != nil {
		return nil, err
	}

	// Get deleted tasks
	userDeletedKey := redis.GenerateKey("user", userID) + ":tasks:deleted"
	deletedIDs, err := r.client.ZRevRange(ctx, userDeletedKey, 0, -1).Result()
	if err != nil {
		return activeIDs, nil // Return active tasks if deleted query fails
	}

	// Combine and maintain order (active first, then deleted)
	allIDs := make([]string, 0, len(activeIDs)+len(deletedIDs))
	allIDs = append(allIDs, activeIDs...)
	allIDs = append(allIDs, deletedIDs...)

	return allIDs, nil
}

// parseUnixTimestamp converts unix timestamp string to time.Time
func parseUnixTimestamp(timestampStr string) (time.Time, error) {
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(timestamp, 0), nil
}