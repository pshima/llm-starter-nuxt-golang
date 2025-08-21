package handlers

import (
	"backend/internal/domain"
	"backend/internal/mocks"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTaskHandler_CreateTask(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		requestBody    interface{}
		mockResponse   *domain.Task
		mockError      error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:   "Successful task creation",
			userID: "user-123",
			requestBody: map[string]interface{}{
				"description": "Complete project documentation",
				"category":    "work",
			},
			mockResponse: &domain.Task{
				ID:          "task-123",
				UserID:      "user-123",
				Description: "Complete project documentation",
				Category:    "work",
				Completed:   false,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockError:      nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "Missing description",
			userID: "user-123",
			requestBody: map[string]interface{}{
				"category": "work",
			},
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "4015",
		},
		{
			name:   "Invalid JSON",
			userID: "user-123",
			requestBody: `{"description": "invalid-json"`,
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "4006",
		},
		{
			name:   "Service error",
			userID: "user-123",
			requestBody: map[string]interface{}{
				"description": "Complete project documentation",
				"category":    "work",
			},
			mockResponse:   nil,
			mockError:      errors.New("service error"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "4016",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(mocks.MockTaskService)

			if tt.mockError == nil && tt.expectedStatus == http.StatusCreated {
				mockService.On("CreateTask", tt.userID, "Complete project documentation", "work").
					Return(tt.mockResponse, nil)
			} else if tt.mockError != nil {
				mockService.On("CreateTask", tt.userID, "Complete project documentation", "work").
					Return((*domain.Task)(nil), tt.mockError)
			}

			// Create handler
			handler := NewTaskHandler(mockService)

			// Setup request
			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest("POST", "/tasks", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			// Set user context (simulating auth middleware)
			c.Set("userID", tt.userID)

			// Execute
			handler.CreateTask(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedCode != "" {
				assert.Contains(t, w.Body.String(), tt.expectedCode)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestTaskHandler_GetTask(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		taskID         string
		mockResponse   *domain.Task
		mockError      error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:   "Successful get task",
			userID: "user-123",
			taskID: "task-123",
			mockResponse: &domain.Task{
				ID:          "task-123",
				UserID:      "user-123",
				Description: "Complete project documentation",
				Category:    "work",
				Completed:   false,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Task not found",
			userID:         "user-123",
			taskID:         "nonexistent-task",
			mockResponse:   nil,
			mockError:      domain.ErrTaskNotFound,
			expectedStatus: http.StatusNotFound,
			expectedCode:   "4017",
		},
		{
			name:           "Service error",
			userID:         "user-123",
			taskID:         "task-123",
			mockResponse:   nil,
			mockError:      errors.New("service error"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "4018",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(mocks.MockTaskService)
			mockService.On("GetTaskByID", tt.taskID, tt.userID).Return(tt.mockResponse, tt.mockError)

			// Create handler
			handler := NewTaskHandler(mockService)

			// Setup request
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest("GET", "/tasks/"+tt.taskID, nil)
			c.Request = req
			c.Params = []gin.Param{{Key: "id", Value: tt.taskID}}

			// Set user context (simulating auth middleware)
			c.Set("userID", tt.userID)

			// Execute
			handler.GetTask(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedCode != "" {
				assert.Contains(t, w.Body.String(), tt.expectedCode)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestTaskHandler_ListTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		queryParams    map[string]string
		expectedFilters domain.TaskFilters
		mockResponse   []*domain.Task
		mockError      error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:   "Successful list tasks with no filters",
			userID: "user-123",
			queryParams: map[string]string{},
			expectedFilters: domain.TaskFilters{
				Limit:  100,
				Offset: 0,
			},
			mockResponse: []*domain.Task{
				{
					ID:          "task-1",
					UserID:      "user-123",
					Description: "Task 1",
					Category:    "work",
					Completed:   false,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:   "List tasks with category filter",
			userID: "user-123",
			queryParams: map[string]string{
				"category": "work",
			},
			expectedFilters: domain.TaskFilters{
				Category: "work",
				Limit:    100,
				Offset:   0,
			},
			mockResponse:   []*domain.Task{},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:   "List tasks with completed filter",
			userID: "user-123",
			queryParams: map[string]string{
				"completed": "true",
			},
			expectedFilters: domain.TaskFilters{
				Completed: func() *bool { b := true; return &b }(),
				Limit:     100,
				Offset:    0,
			},
			mockResponse:   []*domain.Task{},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:   "List tasks with includeDeleted filter",
			userID: "user-123",
			queryParams: map[string]string{
				"includeDeleted": "true",
			},
			expectedFilters: domain.TaskFilters{
				IncludeDeleted: true,
				Limit:          100,
				Offset:         0,
			},
			mockResponse:   []*domain.Task{},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Service error",
			userID: "user-123",
			queryParams: map[string]string{},
			expectedFilters: domain.TaskFilters{
				Limit:  100,
				Offset: 0,
			},
			mockResponse:   nil,
			mockError:      errors.New("service error"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "4019",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(mocks.MockTaskService)
			mockService.On("ListTasks", tt.userID, tt.expectedFilters).Return(tt.mockResponse, tt.mockError)

			// Create handler
			handler := NewTaskHandler(mockService)

			// Setup request with query parameters
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest("GET", "/tasks", nil)
			
			// Add query parameters
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()
			
			c.Request = req

			// Set user context (simulating auth middleware)
			c.Set("userID", tt.userID)

			// Execute
			handler.ListTasks(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedCode != "" {
				assert.Contains(t, w.Body.String(), tt.expectedCode)
			}

			// For successful responses, check the response structure
			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "tasks")
				assert.Contains(t, response, "total")
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestTaskHandler_UpdateTaskCompletion(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		taskID         string
		requestBody    interface{}
		mockResponse   *domain.Task
		mockError      error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:   "Successful task completion update",
			userID: "user-123",
			taskID: "task-123",
			requestBody: map[string]interface{}{
				"completed": true,
			},
			mockResponse: &domain.Task{
				ID:          "task-123",
				UserID:      "user-123",
				Description: "Complete project documentation",
				Category:    "work",
				Completed:   true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Invalid JSON",
			userID: "user-123",
			taskID: "task-123",
			requestBody: `{"completed": invalid-json}`,
			mockResponse:   nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "4006",
		},
		{
			name:   "Task not found",
			userID: "user-123",
			taskID: "nonexistent-task",
			requestBody: map[string]interface{}{
				"completed": true,
			},
			mockResponse:   nil,
			mockError:      domain.ErrTaskNotFound,
			expectedStatus: http.StatusNotFound,
			expectedCode:   "4017",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(mocks.MockTaskService)

			if tt.mockError == nil && tt.expectedStatus == http.StatusOK {
				mockService.On("UpdateTaskCompletion", tt.taskID, tt.userID, true).
					Return(tt.mockResponse, nil)
			} else if tt.mockError != nil {
				mockService.On("UpdateTaskCompletion", tt.taskID, tt.userID, true).
					Return((*domain.Task)(nil), tt.mockError)
			}

			// Create handler
			handler := NewTaskHandler(mockService)

			// Setup request
			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest("PUT", "/tasks/"+tt.taskID+"/complete", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req
			c.Params = []gin.Param{{Key: "id", Value: tt.taskID}}

			// Set user context (simulating auth middleware)
			c.Set("userID", tt.userID)

			// Execute
			handler.UpdateTaskCompletion(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedCode != "" {
				assert.Contains(t, w.Body.String(), tt.expectedCode)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestTaskHandler_DeleteTask(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		taskID         string
		mockError      error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Successful task deletion",
			userID:         "user-123",
			taskID:         "task-123",
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Task not found",
			userID:         "user-123",
			taskID:         "nonexistent-task",
			mockError:      domain.ErrTaskNotFound,
			expectedStatus: http.StatusNotFound,
			expectedCode:   "4017",
		},
		{
			name:           "Service error",
			userID:         "user-123",
			taskID:         "task-123",
			mockError:      errors.New("service error"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "4020",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(mocks.MockTaskService)
			mockService.On("SoftDeleteTask", tt.taskID, tt.userID).Return(tt.mockError)

			// Create handler
			handler := NewTaskHandler(mockService)

			// Setup request
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest("DELETE", "/tasks/"+tt.taskID, nil)
			c.Request = req
			c.Params = []gin.Param{{Key: "id", Value: tt.taskID}}

			// Set user context (simulating auth middleware)
			c.Set("userID", tt.userID)

			// Execute
			handler.DeleteTask(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedCode != "" {
				assert.Contains(t, w.Body.String(), tt.expectedCode)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestTaskHandler_RestoreTask(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		taskID         string
		mockResponse   *domain.Task
		mockError      error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:   "Successful task restoration",
			userID: "user-123",
			taskID: "task-123",
			mockResponse: &domain.Task{
				ID:          "task-123",
				UserID:      "user-123",
				Description: "Complete project documentation",
				Category:    "work",
				Completed:   false,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				DeletedAt:   nil,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Task not found",
			userID:         "user-123",
			taskID:         "nonexistent-task",
			mockResponse:   nil,
			mockError:      domain.ErrTaskNotFound,
			expectedStatus: http.StatusNotFound,
			expectedCode:   "4017",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(mocks.MockTaskService)
			mockService.On("RestoreTask", tt.taskID, tt.userID).Return(tt.mockResponse, tt.mockError)

			// Create handler
			handler := NewTaskHandler(mockService)

			// Setup request
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest("POST", "/tasks/"+tt.taskID+"/restore", nil)
			c.Request = req
			c.Params = []gin.Param{{Key: "id", Value: tt.taskID}}

			// Set user context (simulating auth middleware)
			c.Set("userID", tt.userID)

			// Execute
			handler.RestoreTask(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedCode != "" {
				assert.Contains(t, w.Body.String(), tt.expectedCode)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestTaskHandler_GetCategories(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		mockResponse   []string
		mockError      error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Successful get categories",
			userID:         "user-123",
			mockResponse:   []string{"work", "personal", "health"},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Empty categories list",
			userID:         "user-123",
			mockResponse:   []string{},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Service error",
			userID:         "user-123",
			mockResponse:   nil,
			mockError:      errors.New("service error"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "4019",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(mocks.MockTaskService)
			mockService.On("GetUserCategories", tt.userID).Return(tt.mockResponse, tt.mockError)

			// Create handler
			handler := NewTaskHandler(mockService)

			// Setup request
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest("GET", "/categories", nil)
			c.Request = req

			// Set user context (simulating auth middleware)
			c.Set("userID", tt.userID)

			// Execute
			handler.GetCategories(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedCode != "" {
				assert.Contains(t, w.Body.String(), tt.expectedCode)
			}

			// For successful responses, check the response structure
			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "categories")
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestTaskHandler_RenameCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		categoryName   string
		requestBody    interface{}
		mockError      error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:         "Successful category rename",
			userID:       "user-123",
			categoryName: "work",
			requestBody: map[string]interface{}{
				"newName": "projects",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:         "Category not found",
			userID:       "user-123",
			categoryName: "nonexistent",
			requestBody: map[string]interface{}{
				"newName": "projects",
			},
			mockError:      domain.ErrCategoryNotFound,
			expectedStatus: http.StatusNotFound,
			expectedCode:   "4017",
		},
		{
			name:         "Invalid JSON",
			userID:       "user-123",
			categoryName: "work",
			requestBody:  `{"newName": invalid-json}`,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "4006",
		},
		{
			name:         "Missing new name",
			userID:       "user-123",
			categoryName: "work",
			requestBody: map[string]interface{}{
				"oldName": "work",
			},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "4015",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(mocks.MockTaskService)

			if tt.mockError == nil && tt.expectedStatus == http.StatusOK {
				mockService.On("RenameCategory", tt.userID, tt.categoryName, "projects").Return(nil)
			} else if tt.mockError != nil {
				mockService.On("RenameCategory", tt.userID, tt.categoryName, "projects").Return(tt.mockError)
			}

			// Create handler
			handler := NewTaskHandler(mockService)

			// Setup request
			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest("PUT", "/categories/"+tt.categoryName, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req
			c.Params = []gin.Param{{Key: "categoryName", Value: tt.categoryName}}

			// Set user context (simulating auth middleware)
			c.Set("userID", tt.userID)

			// Execute
			handler.RenameCategory(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedCode != "" {
				assert.Contains(t, w.Body.String(), tt.expectedCode)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestTaskHandler_DeleteCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         string
		categoryName   string
		mockError      error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Successful category deletion",
			userID:         "user-123",
			categoryName:   "work",
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Category not found",
			userID:         "user-123",
			categoryName:   "nonexistent",
			mockError:      domain.ErrCategoryNotFound,
			expectedStatus: http.StatusNotFound,
			expectedCode:   "4017",
		},
		{
			name:           "Service error",
			userID:         "user-123",
			categoryName:   "work",
			mockError:      errors.New("service error"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "4019",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := new(mocks.MockTaskService)
			mockService.On("DeleteCategory", tt.userID, tt.categoryName).Return(tt.mockError)

			// Create handler
			handler := NewTaskHandler(mockService)

			// Setup request
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest("DELETE", "/categories/"+tt.categoryName, nil)
			c.Request = req
			c.Params = []gin.Param{{Key: "categoryName", Value: tt.categoryName}}

			// Set user context (simulating auth middleware)
			c.Set("userID", tt.userID)

			// Execute
			handler.DeleteCategory(c)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedCode != "" {
				assert.Contains(t, w.Body.String(), tt.expectedCode)
			}

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}
