package tests

import (
	"backend/internal/handlers"
	"backend/internal/middleware"
	"backend/internal/repositories"
	"backend/internal/services"
	"backend/pkg/redis"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServer holds the test server and its dependencies
// Provides a complete test environment with Redis, handlers, and HTTP server
type TestServer struct {
	Router      *gin.Engine
	RedisClient *redis.Client
	MiniRedis   *miniredis.Miniredis
	UserRepo    *repositories.UserRepository
	TaskRepo    *repositories.TaskRepository
	UserService services.UserServiceInterface
	TaskService services.TaskServiceInterface
}

// TestUser represents a test user with credentials
// Used for creating test users and storing their credentials
type TestUser struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Password    string `json:"-"` // Raw password for testing
	SessionID   string `json:"-"` // Session ID after login
}

// TestTask represents a test task
// Used for creating and managing test tasks
type TestTask struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	Description string `json:"description"`
	Category    string `json:"category,omitempty"`
	Completed   bool   `json:"completed"`
}

// SetupTestServer creates a new test server with miniredis
// Returns a fully configured test server ready for integration testing
func SetupTestServer(t *testing.T) *TestServer {
	// Create miniredis instance for testing
	mr, err := miniredis.Run()
	require.NoError(t, err, "Failed to start miniredis")

	// Create Redis client connected to miniredis
	port, err := strconv.Atoi(mr.Port())
	require.NoError(t, err, "Failed to parse miniredis port")

	redisConfig := &redis.Config{
		Host:     "localhost",
		Port:     port,
		Password: "",
		DB:       0,
	}
	
	redisClient, err := redis.NewClient(redisConfig)
	require.NoError(t, err, "Failed to create Redis client")

	// Initialize repositories
	userRepo := repositories.NewUserRepository(redisClient)
	taskRepo := repositories.NewTaskRepository(redisClient)

	// Initialize services
	userService := services.NewUserService(userRepo)
	taskService := services.NewTaskService(taskRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userService)
	taskHandler := handlers.NewTaskHandler(taskService)

	// Initialize middleware
	authMiddleware := middleware.AuthMiddleware(userRepo)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// API version 1 routes group
	v1 := router.Group("/api/v1")
	{
		// Auth routes (no authentication required)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
		}

		// Protected routes (authentication required)
		protected := v1.Group("/")
		protected.Use(authMiddleware)
		{
			// Auth routes that require authentication
			protected.GET("/auth/me", authHandler.Me)
			// Task routes
			protected.GET("/tasks", taskHandler.ListTasks)
			protected.POST("/tasks", taskHandler.CreateTask)
			protected.GET("/tasks/:id", taskHandler.GetTask)
			protected.PUT("/tasks/:id/complete", taskHandler.UpdateTaskCompletion)
			protected.DELETE("/tasks/:id", taskHandler.DeleteTask)
			protected.POST("/tasks/:id/restore", taskHandler.RestoreTask)

			// Category routes
			protected.GET("/categories", taskHandler.GetCategories)
			protected.PUT("/categories/:categoryName", taskHandler.RenameCategory)
			protected.DELETE("/categories/:categoryName", taskHandler.DeleteCategory)
		}
	}

	return &TestServer{
		Router:      router,
		RedisClient: redisClient,
		MiniRedis:   mr,
		UserRepo:    userRepo,
		TaskRepo:    taskRepo,
		UserService: userService,
		TaskService: taskService,
	}
}

// TeardownTestServer cleans up the test server and its resources
// Ensures proper cleanup of Redis connections and miniredis instance
func (ts *TestServer) TeardownTestServer() {
	if ts.RedisClient != nil {
		ts.RedisClient.Close()
	}
	if ts.MiniRedis != nil {
		ts.MiniRedis.Close()
	}
}

// CreateTestUser creates a test user with default values
// Returns a TestUser with generated email, display name, and password
func CreateTestUser() *TestUser {
	id := uuid.New().String()
	return &TestUser{
		ID:          id,
		Email:       fmt.Sprintf("test%s@example.com", id[:8]),
		DisplayName: fmt.Sprintf("Test User %s", id[:8]),
		Password:    "Pass123!", // Meets password requirements
	}
}

// RegisterUser registers a test user through the API
// Returns the HTTP response and sets the user ID from the response
func (ts *TestServer) RegisterUser(t *testing.T, user *TestUser) *httptest.ResponseRecorder {
	reqBody := map[string]string{
		"email":       user.Email,
		"password":    user.Password,
		"displayName": user.DisplayName,
	}
	
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	resp := httptest.NewRecorder()
	ts.Router.ServeHTTP(resp, req)

	// Extract user ID from response if successful
	if resp.Code == http.StatusCreated {
		var response map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &response)
		require.NoError(t, err)
		user.ID = response["id"].(string)
	}

	return resp
}

// LoginUser logs in a test user through the API
// Returns the HTTP response and extracts the session cookie
func (ts *TestServer) LoginUser(t *testing.T, user *TestUser) *httptest.ResponseRecorder {
	reqBody := map[string]interface{}{
		"email":      user.Email,
		"password":   user.Password,
		"rememberMe": true,
	}
	
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	resp := httptest.NewRecorder()
	ts.Router.ServeHTTP(resp, req)

	// Extract session cookie if successful
	if resp.Code == http.StatusOK {
		cookies := resp.Result().Cookies()
		for _, cookie := range cookies {
			if cookie.Name == "session" {
				user.SessionID = cookie.Value
				break
			}
		}
	}

	return resp
}

// CreateTestTask creates a test task with default values
// Returns a TestTask with generated description and category
func CreateTestTask(userID string) *TestTask {
	id := uuid.New().String()
	return &TestTask{
		ID:          id,
		UserID:      userID,
		Description: fmt.Sprintf("Test task %s", id[:8]),
		Category:    "test",
		Completed:   false,
	}
}

// CreateTaskWithAuth creates a task through the API with authentication
// Returns the HTTP response and sets the task ID from the response
func (ts *TestServer) CreateTaskWithAuth(t *testing.T, user *TestUser, task *TestTask) *httptest.ResponseRecorder {
	reqBody := map[string]string{
		"description": task.Description,
		"category":    task.Category,
	}
	
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: user.SessionID,
	})
	
	resp := httptest.NewRecorder()
	ts.Router.ServeHTTP(resp, req)

	// Extract task ID from response if successful
	if resp.Code == http.StatusCreated {
		var response map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &response)
		require.NoError(t, err)
		task.ID = response["id"].(string)
	}

	return resp
}

// MakeAuthenticatedRequest makes an HTTP request with session authentication
// Adds the session cookie to the request and returns the response recorder
func (ts *TestServer) MakeAuthenticatedRequest(t *testing.T, method, path string, body []byte, user *TestUser) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	
	if user != nil && user.SessionID != "" {
		req.AddCookie(&http.Cookie{
			Name:  "session",
			Value: user.SessionID,
		})
	}
	
	resp := httptest.NewRecorder()
	ts.Router.ServeHTTP(resp, req)
	return resp
}

// AssertErrorResponse checks that the response contains an error with expected code
// Verifies error response format matches OpenAPI spec
func AssertErrorResponse(t *testing.T, resp *httptest.ResponseRecorder, expectedStatus int, expectedCode string) {
	assert.Equal(t, expectedStatus, resp.Code)
	
	var errorResp map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	
	assert.Contains(t, errorResp, "error")
	assert.Contains(t, errorResp, "code")
	if expectedCode != "" {
		assert.Equal(t, expectedCode, errorResp["code"])
	}
}

// AssertSuccessResponse checks that the response indicates success
// Verifies successful response format and status code
func AssertSuccessResponse(t *testing.T, resp *httptest.ResponseRecorder, expectedStatus int) {
	assert.Equal(t, expectedStatus, resp.Code)
	
	// Ensure response body is valid JSON
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)
}

// AssertUserResponse checks that the response contains a valid user object for registration
// Verifies user response format matches OpenAPI spec for 201 Created
func AssertUserResponse(t *testing.T, resp *httptest.ResponseRecorder, expectedUser *TestUser) {
	AssertSuccessResponse(t, resp, http.StatusCreated)
	
	var user map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &user)
	require.NoError(t, err)
	
	assert.Contains(t, user, "id")
	assert.Contains(t, user, "email")
	assert.Contains(t, user, "displayName") // Use camelCase as in actual response
	assert.Contains(t, user, "createdAt")   // Use camelCase as in actual response
	assert.Contains(t, user, "updatedAt")   // Use camelCase as in actual response
	assert.NotContains(t, user, "password") // Password should never be in response
	
	if expectedUser != nil {
		assert.Equal(t, expectedUser.Email, user["email"])
		assert.Equal(t, expectedUser.DisplayName, user["displayName"])
	}
}

// AssertLoginResponse checks that the response contains a valid user object for login
// Verifies user response format matches OpenAPI spec for 200 OK
func AssertLoginResponse(t *testing.T, resp *httptest.ResponseRecorder, expectedUser *TestUser) {
	AssertSuccessResponse(t, resp, http.StatusOK)
	
	var user map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &user)
	require.NoError(t, err)
	
	assert.Contains(t, user, "id")
	assert.Contains(t, user, "email")
	assert.Contains(t, user, "displayName")
	assert.Contains(t, user, "createdAt")
	assert.Contains(t, user, "updatedAt")
	assert.NotContains(t, user, "password")
	
	if expectedUser != nil {
		assert.Equal(t, expectedUser.Email, user["email"])
		assert.Equal(t, expectedUser.DisplayName, user["displayName"])
	}
}

// AssertTaskResponse checks that the response contains a valid task object
// Verifies task response format matches OpenAPI spec
func AssertTaskResponse(t *testing.T, resp *httptest.ResponseRecorder, expectedTask *TestTask) {
	var task map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &task)
	require.NoError(t, err)
	
	assert.Contains(t, task, "id")
	assert.Contains(t, task, "userId")     // Use camelCase as in actual response
	assert.Contains(t, task, "description")
	assert.Contains(t, task, "completed")
	assert.Contains(t, task, "createdAt")  // Use camelCase as in actual response
	assert.Contains(t, task, "updatedAt")  // Use camelCase as in actual response
	
	if expectedTask != nil {
		assert.Equal(t, expectedTask.Description, task["description"])
		assert.Equal(t, expectedTask.UserID, task["userId"])
		assert.Equal(t, expectedTask.Completed, task["completed"])
		
		if expectedTask.Category != "" {
			assert.Equal(t, expectedTask.Category, task["category"])
		}
	}
}

// AssertTaskListResponse checks that the response contains a valid task list
// Verifies task list response format matches OpenAPI spec
func AssertTaskListResponse(t *testing.T, resp *httptest.ResponseRecorder, expectedCount int) {
	AssertSuccessResponse(t, resp, http.StatusOK)
	
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Contains(t, response, "tasks")
	assert.Contains(t, response, "total")
	
	tasks, ok := response["tasks"].([]interface{})
	require.True(t, ok)
	
	if expectedCount >= 0 {
		assert.Equal(t, expectedCount, len(tasks))
		assert.Equal(t, float64(expectedCount), response["total"])
	}
	
	// Check each task has required fields
	for _, taskInterface := range tasks {
		task, ok := taskInterface.(map[string]interface{})
		require.True(t, ok)
		
		assert.Contains(t, task, "id")
		assert.Contains(t, task, "userId")     // Use camelCase
		assert.Contains(t, task, "description")
		assert.Contains(t, task, "completed")
		assert.Contains(t, task, "createdAt")  // Use camelCase
		assert.Contains(t, task, "updatedAt")  // Use camelCase
	}
}

// AssertCategoryListResponse checks that the response contains a valid category list
// Verifies category list response format matches OpenAPI spec
func AssertCategoryListResponse(t *testing.T, resp *httptest.ResponseRecorder, expectedCategories []string) {
	AssertSuccessResponse(t, resp, http.StatusOK)
	
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Contains(t, response, "categories")
	
	categories, ok := response["categories"].([]interface{})
	require.True(t, ok)
	
	if expectedCategories != nil {
		assert.Equal(t, len(expectedCategories), len(categories))
		
		// Check each category has required fields
		for _, categoryInterface := range categories {
			category, ok := categoryInterface.(map[string]interface{})
			require.True(t, ok)
			
			assert.Contains(t, category, "name")
			assert.Contains(t, category, "taskCount")
		}
	}
}

// SeedMultipleTasks creates multiple test tasks for a user
// Returns a slice of TestTask objects for bulk operations
func (ts *TestServer) SeedMultipleTasks(t *testing.T, user *TestUser, count int) []*TestTask {
	tasks := make([]*TestTask, count)
	
	for i := 0; i < count; i++ {
		task := CreateTestTask(user.ID)
		task.Category = fmt.Sprintf("category%d", i%3) // Distribute across 3 categories
		task.Completed = i%2 == 0                      // Half completed, half not
		
		resp := ts.CreateTaskWithAuth(t, user, task)
		require.Equal(t, http.StatusCreated, resp.Code)
		
		tasks[i] = task
	}
	
	return tasks
}

// WaitForRedis waits for Redis operations to complete
// Ensures data consistency in tests by adding small delay
func WaitForRedis() {
	time.Sleep(10 * time.Millisecond)
}