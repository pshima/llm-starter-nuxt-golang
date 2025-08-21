package handlers

import (
	"backend/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TaskService defines the interface for task business logic operations
// Contains methods needed for task and category handlers
type TaskService interface {
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

// TaskHandler handles task and category-related HTTP requests
// Provides endpoints for task management and category operations
type TaskHandler struct {
	taskService TaskService
}

// NewTaskHandler creates a new instance of TaskHandler
// Initializes the handler with the provided task service
func NewTaskHandler(taskService TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

// CreateTaskRequest represents the request payload for creating a task
type CreateTaskRequest struct {
	Description string `json:"description"`
	Category    string `json:"category"`
}

// UpdateTaskCompletionRequest represents the request payload for updating task completion
type UpdateTaskCompletionRequest struct {
	Completed bool `json:"completed"`
}

// RenameCategoryRequest represents the request payload for renaming a category
type RenameCategoryRequest struct {
	NewName string `json:"newName"`
}

// TaskResponse represents the response payload for task data
type TaskResponse struct {
	ID          string `json:"id"`
	UserID      string `json:"userId"`
	Description string `json:"description"`
	Category    string `json:"category,omitempty"`
	Completed   bool   `json:"completed"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	DeletedAt   string `json:"deletedAt,omitempty"`
}

// CategoryInfo represents category information with task count
type CategoryInfo struct {
	Name      string `json:"name"`
	TaskCount int    `json:"taskCount"`
}

// CreateTask handles task creation requests
// Creates a new task for the authenticated user
func (h *TaskHandler) CreateTask(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"code":  "4001",
		})
		return
	}

	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON format",
			"code":  "4006",
		})
		return
	}

	// Validate required fields
	if req.Description == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task description is required",
			"code":  "4015",
		})
		return
	}

	// Create task
	task, err := h.taskService.CreateTask(userID.(string), req.Description, req.Category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create task",
			"code":  "4016",
		})
		return
	}

	// Return task response
	c.JSON(http.StatusCreated, h.taskToResponse(task))
}

// GetTask handles requests to get a specific task by ID
// Returns task details for the authenticated user
func (h *TaskHandler) GetTask(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"code":  "4001",
		})
		return
	}

	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task ID is required",
			"code":  "4015",
		})
		return
	}

	task, err := h.taskService.GetTaskByID(taskID, userID.(string))
	if err != nil {
		if err == domain.ErrTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Task not found",
				"code":  "4017",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve task",
				"code":  "4018",
			})
		}
		return
	}

	c.JSON(http.StatusOK, h.taskToResponse(task))
}

// ListTasks handles requests to list tasks with optional filters
// Returns paginated list of tasks for the authenticated user
func (h *TaskHandler) ListTasks(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"code":  "4001",
		})
		return
	}

	// Parse filters from query parameters
	filters := domain.TaskFilters{
		Category:       c.Query("category"),
		IncludeDeleted: c.Query("includeDeleted") == "true",
		Limit:          100, // Default limit
		Offset:         0,   // Default offset
	}

	// Parse completed filter
	if completed := c.Query("completed"); completed != "" {
		if completed == "true" {
			filters.Completed = &[]bool{true}[0]
		} else if completed == "false" {
			filters.Completed = &[]bool{false}[0]
		}
	}

	// Get tasks from service
	tasks, err := h.taskService.ListTasks(userID.(string), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve tasks",
			"code":  "4019",
		})
		return
	}

	// Convert to response format
	taskResponses := make([]TaskResponse, len(tasks))
	for i, task := range tasks {
		taskResponses[i] = h.taskToResponse(task)
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": taskResponses,
		"total": len(taskResponses),
	})
}

// UpdateTaskCompletion handles requests to update task completion status
// Toggles task completion for the authenticated user
func (h *TaskHandler) UpdateTaskCompletion(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"code":  "4001",
		})
		return
	}

	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task ID is required",
			"code":  "4015",
		})
		return
	}

	var req UpdateTaskCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON format",
			"code":  "4006",
		})
		return
	}

	// Update task completion
	task, err := h.taskService.UpdateTaskCompletion(taskID, userID.(string), req.Completed)
	if err != nil {
		if err == domain.ErrTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Task not found",
				"code":  "4017",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update task",
				"code":  "4018",
			})
		}
		return
	}

	c.JSON(http.StatusOK, h.taskToResponse(task))
}

// DeleteTask handles requests to soft delete a task
// Marks the task as deleted without removing it from storage
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"code":  "4001",
		})
		return
	}

	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task ID is required",
			"code":  "4015",
		})
		return
	}

	err := h.taskService.SoftDeleteTask(taskID, userID.(string))
	if err != nil {
		if err == domain.ErrTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Task not found",
				"code":  "4017",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete task",
				"code":  "4020",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task deleted successfully",
	})
}

// RestoreTask handles requests to restore a soft-deleted task
// Restores a previously deleted task for the authenticated user
func (h *TaskHandler) RestoreTask(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"code":  "4001",
		})
		return
	}

	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Task ID is required",
			"code":  "4015",
		})
		return
	}

	task, err := h.taskService.RestoreTask(taskID, userID.(string))
	if err != nil {
		if err == domain.ErrTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Task not found",
				"code":  "4017",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to restore task",
				"code":  "4018",
			})
		}
		return
	}

	c.JSON(http.StatusOK, h.taskToResponse(task))
}

// GetCategories handles requests to get user categories
// Returns list of categories used by the authenticated user
func (h *TaskHandler) GetCategories(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"code":  "4001",
		})
		return
	}

	categories, err := h.taskService.GetUserCategories(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve categories",
			"code":  "4019",
		})
		return
	}

	// Convert to category info format (for now, just return names without counts)
	categoryInfos := make([]CategoryInfo, len(categories))
	for i, category := range categories {
		categoryInfos[i] = CategoryInfo{
			Name:      category,
			TaskCount: 0, // TODO: Add task count if needed
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": categoryInfos,
	})
}

// RenameCategory handles requests to rename a category
// Updates category name for all tasks of the authenticated user
func (h *TaskHandler) RenameCategory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"code":  "4001",
		})
		return
	}

	categoryName := c.Param("categoryName")
	if categoryName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Category name is required",
			"code":  "4015",
		})
		return
	}

	var req RenameCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON format",
			"code":  "4006",
		})
		return
	}

	if req.NewName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "New category name is required",
			"code":  "4015",
		})
		return
	}

	err := h.taskService.RenameCategory(userID.(string), categoryName, req.NewName)
	if err != nil {
		if err == domain.ErrCategoryNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Category not found",
				"code":  "4017",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to rename category",
				"code":  "4019",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Category renamed successfully",
		"tasksUpdated": 0, // TODO: Return actual count if needed
	})
}

// DeleteCategory handles requests to delete a category
// Removes category from all tasks of the authenticated user
func (h *TaskHandler) DeleteCategory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"code":  "4001",
		})
		return
	}

	categoryName := c.Param("categoryName")
	if categoryName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Category name is required",
			"code":  "4015",
		})
		return
	}

	err := h.taskService.DeleteCategory(userID.(string), categoryName)
	if err != nil {
		if err == domain.ErrCategoryNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Category not found",
				"code":  "4017",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete category",
				"code":  "4019",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Category deleted successfully",
		"tasksUpdated": 0, // TODO: Return actual count if needed
	})
}

// taskToResponse converts a domain Task to TaskResponse
// Handles proper formatting of timestamps and optional fields
func (h *TaskHandler) taskToResponse(task *domain.Task) TaskResponse {
	response := TaskResponse{
		ID:          task.ID,
		UserID:      task.UserID,
		Description: task.Description,
		Category:    task.Category,
		Completed:   task.Completed,
		CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if task.DeletedAt != nil {
		response.DeletedAt = task.DeletedAt.Format("2006-01-02T15:04:05Z")
	}

	return response
}