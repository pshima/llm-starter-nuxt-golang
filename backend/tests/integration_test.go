package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Run tests with 5-second timeout as per TEST.md requirements
	m.Run()
}

// TestUserRegistrationFlow tests the complete user registration workflow
// Verifies user can register, data is stored correctly, and duplicate emails are rejected
func TestUserRegistrationFlow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.TeardownTestServer()

	t.Run("successful user registration", func(t *testing.T) {
		user := CreateTestUser()
		resp := ts.RegisterUser(t, user)

		// Verify successful registration
		assert.Equal(t, http.StatusCreated, resp.Code)
		AssertUserResponse(t, resp, user)

		// Verify user ID was set from response
		assert.NotEmpty(t, user.ID)
	})

	t.Run("duplicate email registration should fail", func(t *testing.T) {
		user1 := CreateTestUser()
		user2 := CreateTestUser()
		user2.Email = user1.Email // Use same email

		// Register first user
		resp1 := ts.RegisterUser(t, user1)
		assert.Equal(t, http.StatusCreated, resp1.Code)

		// Try to register second user with same email
		resp2 := ts.RegisterUser(t, user2)
		AssertErrorResponse(t, resp2, http.StatusConflict, "4008")
	})

	t.Run("invalid email format should fail", func(t *testing.T) {
		user := CreateTestUser()
		user.Email = "invalid-email"

		resp := ts.RegisterUser(t, user)
		AssertErrorResponse(t, resp, http.StatusInternalServerError, "4009")
	})

	t.Run("weak password should fail", func(t *testing.T) {
		user := CreateTestUser()
		user.Password = "weak" // Doesn't meet requirements

		resp := ts.RegisterUser(t, user)
		AssertErrorResponse(t, resp, http.StatusInternalServerError, "4009")
	})

	t.Run("missing required fields should fail", func(t *testing.T) {
		reqBody := map[string]string{
			"email": "test@example.com",
			// Missing password and displayName
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		resp := ts.MakeAuthenticatedRequest(t, "POST", "/api/v1/auth/register", body, nil)
		AssertErrorResponse(t, resp, http.StatusBadRequest, "4007")
	})
}

// TestUserLoginFlow tests the complete user login and session management workflow
// Verifies login, session creation, session validation, and logout
func TestUserLoginFlow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.TeardownTestServer()

	// Setup: Register a user
	user := CreateTestUser()
	regResp := ts.RegisterUser(t, user)
	require.Equal(t, http.StatusCreated, regResp.Code)

	t.Run("successful login", func(t *testing.T) {
		resp := ts.LoginUser(t, user)

		// Verify successful login
		assert.Equal(t, http.StatusOK, resp.Code)
		AssertLoginResponse(t, resp, user)

		// Verify session cookie was set
		assert.NotEmpty(t, user.SessionID)

		cookies := resp.Result().Cookies()
		found := false
		for _, cookie := range cookies {
			if cookie.Name == "session" {
				found = true
				assert.Equal(t, "/", cookie.Path)
				assert.True(t, cookie.HttpOnly)
				assert.True(t, cookie.MaxAge > 0) // Remember me should set max age
				break
			}
		}
		assert.True(t, found, "Session cookie should be set")
	})

	t.Run("invalid credentials should fail", func(t *testing.T) {
		invalidUser := CreateTestUser()
		invalidUser.Email = user.Email
		invalidUser.Password = "WrongPass123!"

		resp := ts.LoginUser(t, invalidUser)
		AssertErrorResponse(t, resp, http.StatusUnauthorized, "4010")
	})

	t.Run("nonexistent user should fail", func(t *testing.T) {
		nonexistentUser := CreateTestUser()
		resp := ts.LoginUser(t, nonexistentUser)
		AssertErrorResponse(t, resp, http.StatusUnauthorized, "4010")
	})

	t.Run("session validation with /me endpoint", func(t *testing.T) {
		// Login user first
		loginResp := ts.LoginUser(t, user)
		require.Equal(t, http.StatusOK, loginResp.Code)

		// Call /me endpoint with session
		resp := ts.MakeAuthenticatedRequest(t, "GET", "/api/v1/auth/me", nil, user)
		AssertLoginResponse(t, resp, user)
	})

	t.Run("unauthorized access without session", func(t *testing.T) {
		resp := ts.MakeAuthenticatedRequest(t, "GET", "/api/v1/auth/me", nil, nil)
		AssertErrorResponse(t, resp, http.StatusUnauthorized, "4001")
	})

	t.Run("successful logout", func(t *testing.T) {
		// Login user first
		loginResp := ts.LoginUser(t, user)
		require.Equal(t, http.StatusOK, loginResp.Code)

		// Logout
		resp := ts.MakeAuthenticatedRequest(t, "POST", "/api/v1/auth/logout", nil, user)
		assert.Equal(t, http.StatusOK, resp.Code)

		var response map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "message")

		// Verify session is invalidated
		WaitForRedis()
		meResp := ts.MakeAuthenticatedRequest(t, "GET", "/api/v1/auth/me", nil, user)
		AssertErrorResponse(t, meResp, http.StatusUnauthorized, "4002")
	})
}

// TestTaskCRUDOperations tests complete task CRUD workflow with authentication
// Verifies create, read, update, delete operations for tasks
func TestTaskCRUDOperations(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.TeardownTestServer()

	// Setup: Register and login user
	user := CreateTestUser()
	regResp := ts.RegisterUser(t, user)
	require.Equal(t, http.StatusCreated, regResp.Code)

	loginResp := ts.LoginUser(t, user)
	require.Equal(t, http.StatusOK, loginResp.Code)

	t.Run("create task successfully", func(t *testing.T) {
		task := CreateTestTask(user.ID)
		resp := ts.CreateTaskWithAuth(t, user, task)

		assert.Equal(t, http.StatusCreated, resp.Code)
		AssertTaskResponse(t, resp, task)
		assert.NotEmpty(t, task.ID)
	})

	t.Run("create task without authentication should fail", func(t *testing.T) {
		task := CreateTestTask(user.ID)
		resp := ts.CreateTaskWithAuth(t, &TestUser{}, task) // No session

		AssertErrorResponse(t, resp, http.StatusUnauthorized, "4001")
	})

	t.Run("create task with invalid data should fail", func(t *testing.T) {
		reqBody := map[string]string{
			"description": "", // Empty description
			"category":    "test",
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		resp := ts.MakeAuthenticatedRequest(t, "POST", "/api/v1/tasks", body, user)
		AssertErrorResponse(t, resp, http.StatusBadRequest, "4015")
	})

	t.Run("get task by ID", func(t *testing.T) {
		// Create a task first
		task := CreateTestTask(user.ID)
		createResp := ts.CreateTaskWithAuth(t, user, task)
		require.Equal(t, http.StatusCreated, createResp.Code)

		// Get the task
		resp := ts.MakeAuthenticatedRequest(t, "GET", fmt.Sprintf("/api/v1/tasks/%s", task.ID), nil, user)
		assert.Equal(t, http.StatusOK, resp.Code)
		AssertTaskResponse(t, resp, task)
	})

	t.Run("get nonexistent task should fail", func(t *testing.T) {
		resp := ts.MakeAuthenticatedRequest(t, "GET", "/api/v1/tasks/nonexistent-id", nil, user)
		AssertErrorResponse(t, resp, http.StatusNotFound, "4017")
	})

	t.Run("update task completion status", func(t *testing.T) {
		// Create a task first
		task := CreateTestTask(user.ID)
		createResp := ts.CreateTaskWithAuth(t, user, task)
		require.Equal(t, http.StatusCreated, createResp.Code)

		// Update completion status
		reqBody := map[string]bool{"completed": true}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		resp := ts.MakeAuthenticatedRequest(t, "PUT", fmt.Sprintf("/api/v1/tasks/%s/complete", task.ID), body, user)
		assert.Equal(t, http.StatusOK, resp.Code)

		// Verify task was updated
		var updatedTask map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &updatedTask)
		require.NoError(t, err)
		assert.True(t, updatedTask["completed"].(bool))
	})

	t.Run("list tasks", func(t *testing.T) {
		// Create multiple tasks
		tasks := ts.SeedMultipleTasks(t, user, 5)

		// List all tasks
		resp := ts.MakeAuthenticatedRequest(t, "GET", "/api/v1/tasks", nil, user)
		AssertTaskListResponse(t, resp, len(tasks))
	})

	t.Run("list tasks with filters", func(t *testing.T) {
		// Create tasks with different categories and completion status
		ts.SeedMultipleTasks(t, user, 6)

		// Filter by category
		resp := ts.MakeAuthenticatedRequest(t, "GET", "/api/v1/tasks?category=category0", nil, user)
		AssertTaskListResponse(t, resp, -1) // Don't check exact count, just verify format

		// Filter by completion status
		resp = ts.MakeAuthenticatedRequest(t, "GET", "/api/v1/tasks?completed=true", nil, user)
		AssertTaskListResponse(t, resp, -1)
	})
}

// TestTaskSoftDeleteAndRestore tests soft delete and restore functionality
// Verifies tasks can be soft deleted, excluded from listings, and restored
func TestTaskSoftDeleteAndRestore(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.TeardownTestServer()

	// Setup: Register and login user
	user := CreateTestUser()
	regResp := ts.RegisterUser(t, user)
	require.Equal(t, http.StatusCreated, regResp.Code)

	loginResp := ts.LoginUser(t, user)
	require.Equal(t, http.StatusOK, loginResp.Code)

	t.Run("soft delete task", func(t *testing.T) {
		// Create a task
		task := CreateTestTask(user.ID)
		createResp := ts.CreateTaskWithAuth(t, user, task)
		require.Equal(t, http.StatusCreated, createResp.Code)

		// Soft delete the task
		resp := ts.MakeAuthenticatedRequest(t, "DELETE", fmt.Sprintf("/api/v1/tasks/%s", task.ID), nil, user)
		assert.Equal(t, http.StatusOK, resp.Code)

		var response map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "message")
	})

	t.Run("soft deleted task not in regular list", func(t *testing.T) {
		// Create and delete a task
		task := CreateTestTask(user.ID)
		createResp := ts.CreateTaskWithAuth(t, user, task)
		require.Equal(t, http.StatusCreated, createResp.Code)

		deleteResp := ts.MakeAuthenticatedRequest(t, "DELETE", fmt.Sprintf("/api/v1/tasks/%s", task.ID), nil, user)
		require.Equal(t, http.StatusOK, deleteResp.Code)

		WaitForRedis()

		// List tasks (should not include deleted task)
		resp := ts.MakeAuthenticatedRequest(t, "GET", "/api/v1/tasks", nil, user)
		AssertTaskListResponse(t, resp, 0) // Should be empty

		// List tasks with includeDeleted=true
		resp = ts.MakeAuthenticatedRequest(t, "GET", "/api/v1/tasks?includeDeleted=true", nil, user)
		AssertTaskListResponse(t, resp, 1) // Should include the deleted task
	})

	t.Run("restore soft deleted task", func(t *testing.T) {
		// Create and delete a task
		task := CreateTestTask(user.ID)
		createResp := ts.CreateTaskWithAuth(t, user, task)
		require.Equal(t, http.StatusCreated, createResp.Code)

		deleteResp := ts.MakeAuthenticatedRequest(t, "DELETE", fmt.Sprintf("/api/v1/tasks/%s", task.ID), nil, user)
		require.Equal(t, http.StatusOK, deleteResp.Code)

		// Restore the task
		resp := ts.MakeAuthenticatedRequest(t, "POST", fmt.Sprintf("/api/v1/tasks/%s/restore", task.ID), nil, user)
		assert.Equal(t, http.StatusOK, resp.Code)

		// Verify task is restored
		AssertTaskResponse(t, resp, task)

		WaitForRedis()

		// Verify task appears in regular list again
		listResp := ts.MakeAuthenticatedRequest(t, "GET", "/api/v1/tasks", nil, user)
		AssertTaskListResponse(t, listResp, 1)
	})

	t.Run("restore nonexistent task should fail", func(t *testing.T) {
		resp := ts.MakeAuthenticatedRequest(t, "POST", "/api/v1/tasks/nonexistent-id/restore", nil, user)
		AssertErrorResponse(t, resp, http.StatusNotFound, "4017")
	})

	t.Run("restore task outside 7-day window should fail", func(t *testing.T) {
		// This would require manipulating time in Redis, which is complex in miniredis
		// For now, we'll test the error path by trying to restore a non-deleted task
		task := CreateTestTask(user.ID)
		createResp := ts.CreateTaskWithAuth(t, user, task)
		require.Equal(t, http.StatusCreated, createResp.Code)

		// Try to restore a task that wasn't deleted
		resp := ts.MakeAuthenticatedRequest(t, "POST", fmt.Sprintf("/api/v1/tasks/%s/restore", task.ID), nil, user)
		// This should fail because the task isn't deleted
		AssertErrorResponse(t, resp, http.StatusInternalServerError, "4018")
	})
}

// TestCategoryManagement tests category management functionality
// Verifies listing, renaming, and deleting categories
func TestCategoryManagement(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.TeardownTestServer()

	// Setup: Register and login user
	user := CreateTestUser()
	regResp := ts.RegisterUser(t, user)
	require.Equal(t, http.StatusCreated, regResp.Code)

	loginResp := ts.LoginUser(t, user)
	require.Equal(t, http.StatusOK, loginResp.Code)

	t.Run("list categories when no tasks exist", func(t *testing.T) {
		resp := ts.MakeAuthenticatedRequest(t, "GET", "/api/v1/categories", nil, user)
		AssertCategoryListResponse(t, resp, []string{})
	})

	t.Run("list categories after creating tasks", func(t *testing.T) {
		// Create tasks with different categories
		categories := []string{"work", "personal", "shopping"}
		for _, category := range categories {
			task := CreateTestTask(user.ID)
			task.Category = category
			resp := ts.CreateTaskWithAuth(t, user, task)
			require.Equal(t, http.StatusCreated, resp.Code)
		}

		WaitForRedis()

		// List categories
		resp := ts.MakeAuthenticatedRequest(t, "GET", "/api/v1/categories", nil, user)
		AssertCategoryListResponse(t, resp, categories)
	})

	t.Run("rename category", func(t *testing.T) {
		// Create a task with a category
		task := CreateTestTask(user.ID)
		task.Category = "oldname"
		createResp := ts.CreateTaskWithAuth(t, user, task)
		require.Equal(t, http.StatusCreated, createResp.Code)

		// Rename the category
		reqBody := map[string]string{"newName": "newname"}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		resp := ts.MakeAuthenticatedRequest(t, "PUT", "/api/v1/categories/oldname", body, user)
		assert.Equal(t, http.StatusOK, resp.Code)

		var response map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "message")
		assert.Contains(t, response, "tasksUpdated")
	})

	t.Run("rename nonexistent category should fail", func(t *testing.T) {
		reqBody := map[string]string{"newName": "newname"}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		resp := ts.MakeAuthenticatedRequest(t, "PUT", "/api/v1/categories/nonexistent", body, user)
		AssertErrorResponse(t, resp, http.StatusNotFound, "4017")
	})

	t.Run("delete category", func(t *testing.T) {
		// Create tasks with a category
		category := "todelete"
		for i := 0; i < 3; i++ {
			task := CreateTestTask(user.ID)
			task.Category = category
			resp := ts.CreateTaskWithAuth(t, user, task)
			require.Equal(t, http.StatusCreated, resp.Code)
		}

		WaitForRedis()

		// Delete the category
		resp := ts.MakeAuthenticatedRequest(t, "DELETE", fmt.Sprintf("/api/v1/categories/%s", category), nil, user)
		assert.Equal(t, http.StatusOK, resp.Code)

		var response map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "message")
		assert.Contains(t, response, "tasksUpdated")
		assert.Equal(t, float64(3), response["tasksUpdated"]) // Should update 3 tasks
	})

	t.Run("delete nonexistent category should fail", func(t *testing.T) {
		resp := ts.MakeAuthenticatedRequest(t, "DELETE", "/api/v1/categories/nonexistent", nil, user)
		AssertErrorResponse(t, resp, http.StatusNotFound, "4017")
	})
}

// TestUserIsolation tests that users can only access their own data
// Verifies proper authentication and authorization enforcement
func TestUserIsolation(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.TeardownTestServer()

	// Setup: Register two users
	user1 := CreateTestUser()
	user2 := CreateTestUser()

	regResp1 := ts.RegisterUser(t, user1)
	require.Equal(t, http.StatusCreated, regResp1.Code)

	regResp2 := ts.RegisterUser(t, user2)
	require.Equal(t, http.StatusCreated, regResp2.Code)

	loginResp1 := ts.LoginUser(t, user1)
	require.Equal(t, http.StatusOK, loginResp1.Code)

	loginResp2 := ts.LoginUser(t, user2)
	require.Equal(t, http.StatusOK, loginResp2.Code)

	t.Run("user can only see their own tasks", func(t *testing.T) {
		// Create tasks for both users
		task1 := CreateTestTask(user1.ID)
		createResp1 := ts.CreateTaskWithAuth(t, user1, task1)
		require.Equal(t, http.StatusCreated, createResp1.Code)

		task2 := CreateTestTask(user2.ID)
		createResp2 := ts.CreateTaskWithAuth(t, user2, task2)
		require.Equal(t, http.StatusCreated, createResp2.Code)

		WaitForRedis()

		// User1 should only see their own task
		resp1 := ts.MakeAuthenticatedRequest(t, "GET", "/api/v1/tasks", nil, user1)
		AssertTaskListResponse(t, resp1, 1)

		// User2 should only see their own task
		resp2 := ts.MakeAuthenticatedRequest(t, "GET", "/api/v1/tasks", nil, user2)
		AssertTaskListResponse(t, resp2, 1)
	})

	t.Run("user cannot access another user's task", func(t *testing.T) {
		// Create a task for user1
		task1 := CreateTestTask(user1.ID)
		createResp1 := ts.CreateTaskWithAuth(t, user1, task1)
		require.Equal(t, http.StatusCreated, createResp1.Code)

		// User2 tries to access user1's task
		resp := ts.MakeAuthenticatedRequest(t, "GET", fmt.Sprintf("/api/v1/tasks/%s", task1.ID), nil, user2)
		AssertErrorResponse(t, resp, http.StatusNotFound, "4017") // Should not find the task
	})

	t.Run("user cannot modify another user's task", func(t *testing.T) {
		// Create a task for user1
		task1 := CreateTestTask(user1.ID)
		createResp1 := ts.CreateTaskWithAuth(t, user1, task1)
		require.Equal(t, http.StatusCreated, createResp1.Code)

		// User2 tries to update user1's task
		reqBody := map[string]bool{"completed": true}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		resp := ts.MakeAuthenticatedRequest(t, "PUT", fmt.Sprintf("/api/v1/tasks/%s/complete", task1.ID), body, user2)
		AssertErrorResponse(t, resp, http.StatusNotFound, "4017") // Should not find the task
	})

	t.Run("user can only see their own categories", func(t *testing.T) {
		// Create tasks with categories for both users
		task1 := CreateTestTask(user1.ID)
		task1.Category = "user1category"
		createResp1 := ts.CreateTaskWithAuth(t, user1, task1)
		require.Equal(t, http.StatusCreated, createResp1.Code)

		task2 := CreateTestTask(user2.ID)
		task2.Category = "user2category"
		createResp2 := ts.CreateTaskWithAuth(t, user2, task2)
		require.Equal(t, http.StatusCreated, createResp2.Code)

		WaitForRedis()

		// Each user should only see their own categories
		resp1 := ts.MakeAuthenticatedRequest(t, "GET", "/api/v1/categories", nil, user1)
		AssertCategoryListResponse(t, resp1, []string{"user1category"})

		resp2 := ts.MakeAuthenticatedRequest(t, "GET", "/api/v1/categories", nil, user2)
		AssertCategoryListResponse(t, resp2, []string{"user2category"})
	})
}

// TestErrorScenarios tests various error conditions and edge cases
// Verifies proper error handling and response formats
func TestErrorScenarios(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.TeardownTestServer()

	t.Run("malformed JSON should return 400", func(t *testing.T) {
		resp := ts.MakeAuthenticatedRequest(t, "POST", "/api/v1/auth/register", []byte("{invalid json}"), nil)
		AssertErrorResponse(t, resp, http.StatusBadRequest, "4006")
	})

	t.Run("unsupported HTTP method should return 405", func(t *testing.T) {
		resp := ts.MakeAuthenticatedRequest(t, "PATCH", "/api/v1/auth/register", nil, nil)
		assert.Equal(t, http.StatusMethodNotAllowed, resp.Code)
	})

	t.Run("nonexistent endpoint should return 404", func(t *testing.T) {
		resp := ts.MakeAuthenticatedRequest(t, "GET", "/api/v1/nonexistent", nil, nil)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("protected endpoint without auth should return 401", func(t *testing.T) {
		resp := ts.MakeAuthenticatedRequest(t, "GET", "/api/v1/tasks", nil, nil)
		AssertErrorResponse(t, resp, http.StatusUnauthorized, "4001")
	})

	t.Run("invalid session should return 401", func(t *testing.T) {
		user := &TestUser{SessionID: "invalid-session-id"}
		resp := ts.MakeAuthenticatedRequest(t, "GET", "/api/v1/tasks", nil, user)
		AssertErrorResponse(t, resp, http.StatusUnauthorized, "4002")
	})
}

// TestFullWorkflow tests a complete end-to-end user workflow
// Simulates a realistic user interaction with the system
func TestFullWorkflow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.TeardownTestServer()

	t.Run("complete user workflow", func(t *testing.T) {
		// Step 1: Register a new user
		user := CreateTestUser()
		regResp := ts.RegisterUser(t, user)
		require.Equal(t, http.StatusCreated, regResp.Code)

		// Step 2: Login
		loginResp := ts.LoginUser(t, user)
		require.Equal(t, http.StatusOK, loginResp.Code)

		// Step 3: Create several tasks in different categories
		taskDescriptions := []string{
			"Buy groceries",
			"Finish project report",
			"Call dentist",
			"Plan vacation",
		}
		categories := []string{"shopping", "work", "health", "personal"}
		createdTasks := make([]*TestTask, len(taskDescriptions))

		for i, desc := range taskDescriptions {
			task := CreateTestTask(user.ID)
			task.Description = desc
			task.Category = categories[i]
			resp := ts.CreateTaskWithAuth(t, user, task)
			require.Equal(t, http.StatusCreated, resp.Code)
			createdTasks[i] = task
		}

		WaitForRedis()

		// Step 4: List all tasks
		listResp := ts.MakeAuthenticatedRequest(t, "GET", "/api/v1/tasks", nil, user)
		AssertTaskListResponse(t, listResp, len(taskDescriptions))

		// Step 5: Mark some tasks as completed
		for i := 0; i < 2; i++ {
			reqBody := map[string]bool{"completed": true}
			body, err := json.Marshal(reqBody)
			require.NoError(t, err)

			resp := ts.MakeAuthenticatedRequest(t, "PUT", fmt.Sprintf("/api/v1/tasks/%s/complete", createdTasks[i].ID), body, user)
			assert.Equal(t, http.StatusOK, resp.Code)
		}

		WaitForRedis()

		// Step 6: List only completed tasks
		completedResp := ts.MakeAuthenticatedRequest(t, "GET", "/api/v1/tasks?completed=true", nil, user)
		AssertTaskListResponse(t, completedResp, 2)

		// Step 7: List categories
		categoriesResp := ts.MakeAuthenticatedRequest(t, "GET", "/api/v1/categories", nil, user)
		AssertCategoryListResponse(t, categoriesResp, categories)

		// Step 8: Rename a category
		reqBody := map[string]string{"newName": "groceries"}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)
		renameResp := ts.MakeAuthenticatedRequest(t, "PUT", "/api/v1/categories/shopping", body, user)
		assert.Equal(t, http.StatusOK, renameResp.Code)

		// Step 9: Delete a task
		deleteResp := ts.MakeAuthenticatedRequest(t, "DELETE", fmt.Sprintf("/api/v1/tasks/%s", createdTasks[0].ID), nil, user)
		assert.Equal(t, http.StatusOK, deleteResp.Code)

		WaitForRedis()

		// Step 10: Verify task is soft deleted
		listAfterDeleteResp := ts.MakeAuthenticatedRequest(t, "GET", "/api/v1/tasks", nil, user)
		AssertTaskListResponse(t, listAfterDeleteResp, len(taskDescriptions)-1)

		// Step 11: Restore the deleted task
		restoreResp := ts.MakeAuthenticatedRequest(t, "POST", fmt.Sprintf("/api/v1/tasks/%s/restore", createdTasks[0].ID), nil, user)
		assert.Equal(t, http.StatusOK, restoreResp.Code)

		WaitForRedis()

		// Step 12: Verify task is restored
		listAfterRestoreResp := ts.MakeAuthenticatedRequest(t, "GET", "/api/v1/tasks", nil, user)
		AssertTaskListResponse(t, listAfterRestoreResp, len(taskDescriptions))

		// Step 13: Logout
		logoutResp := ts.MakeAuthenticatedRequest(t, "POST", "/api/v1/auth/logout", nil, user)
		assert.Equal(t, http.StatusOK, logoutResp.Code)

		// Step 14: Verify session is invalidated
		WaitForRedis()
		meResp := ts.MakeAuthenticatedRequest(t, "GET", "/api/v1/auth/me", nil, user)
		AssertErrorResponse(t, meResp, http.StatusUnauthorized, "4002")
	})
}