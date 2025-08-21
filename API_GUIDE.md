# API Usage Guide

## CRITICAL: Field Naming Convention

**All JSON fields use camelCase naming**. Frontend TypeScript interfaces MUST match these exact field names.

| Go Struct Field | JSON Tag | Example |
|-----------------|----------|---------|
| DisplayName | `displayName` | "John Doe" |
| CreatedAt | `createdAt` | "2024-01-01T00:00:00Z" |
| UpdatedAt | `updatedAt` | "2024-01-01T00:00:00Z" |
| DeletedAt | `deletedAt` | "2024-01-01T00:00:00Z" |
| UserID | `userId` | "user123" |
| RememberMe | `rememberMe` | true |

⚠️ **Common Error**: If you get "Email, password, and display name are required" when all fields appear filled, check that you're using `displayName` not `display_name`.

## Authentication Flow

### Overview

The API uses session-based authentication with HTTP-only cookies. Sessions are stored server-side in Redis with a 7-day TTL.

### Step-by-Step Authentication Implementation

#### 1. User Registration

```bash
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "displayName": "John Doe"
}
```

**Response (201 Created)**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "displayName": "John Doe",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

**Password Requirements**:
- Minimum 6 characters
- At least 1 special character (!@#$%^&*(),.?":{}|<>)
- At least 1 number

#### 2. User Login

```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "rememberMe": true
}
```

**Response (200 OK)**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "displayName": "John Doe",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

**Response Headers**:
```
Set-Cookie: session=abc123def456; Path=/; HttpOnly; SameSite=Lax; Max-Age=604800
```

#### 3. Using Authentication in Subsequent Requests

All protected endpoints require the session cookie:

```bash
# Using curl with cookie jar
curl -b cookies.txt http://localhost:8080/api/v1/tasks

# Using curl with cookie directly
curl -H "Cookie: session=abc123def456" http://localhost:8080/api/v1/tasks
```

#### 4. Logout

```bash
POST /api/v1/auth/logout
Cookie: session=abc123def456
```

**Response (200 OK)**:
```json
{
  "message": "Logged out successfully"
}
```

### Session Management Best Practices

1. **Store cookies securely**: Use cookie jars or secure storage
2. **Handle expiration**: Sessions expire after 7 days
3. **Refresh sessions**: Activity extends session lifetime
4. **Multiple sessions**: Users can have multiple active sessions

### Cookie Handling in Different Clients

#### JavaScript (Fetch API)
```javascript
// Login
const response = await fetch('http://localhost:8080/api/v1/auth/login', {
  method: 'POST',
  credentials: 'include',  // Important: includes cookies
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    email: 'user@example.com',
    password: 'SecurePass123!'
  })
});

// Subsequent requests
const tasks = await fetch('http://localhost:8080/api/v1/tasks', {
  credentials: 'include'  // Sends cookies automatically
});
```

#### Python (requests)
```python
import requests

# Create session object to persist cookies
session = requests.Session()

# Login
login_response = session.post('http://localhost:8080/api/v1/auth/login', 
  json={
    'email': 'user@example.com',
    'password': 'SecurePass123!'
  })

# Subsequent requests use the same session
tasks_response = session.get('http://localhost:8080/api/v1/tasks')
```

#### Go (net/http)
```go
// Create cookie jar
jar, _ := cookiejar.New(nil)
client := &http.Client{Jar: jar}

// Login
loginData := map[string]string{
  "email": "user@example.com",
  "password": "SecurePass123!",
}
jsonData, _ := json.Marshal(loginData)
resp, _ := client.Post("http://localhost:8080/api/v1/auth/login", 
  "application/json", bytes.NewBuffer(jsonData))

// Subsequent requests use the same client
resp, _ = client.Get("http://localhost:8080/api/v1/tasks")
```

## Error Handling

### Error Response Format

All errors follow a consistent format:

```json
{
  "error": "Human-readable error message",
  "code": "4001",
  "details": {
    "field": "additional context"
  }
}
```

### Complete Error Code Reference

#### System Errors (1001-1010)
- `1001`: Failed to load configuration
- `1002`: Failed to start server
- `1003`: Server forced to shutdown
- `1004`: Redis connection failed

#### Repository Errors (2001-2020)
- `2001`: Redis connection error
- `2002`: Data serialization error
- `2003`: Data not found
- `2004`: Duplicate key error
- `2005`: Invalid data format
- `2006`: Transaction failed
- `2007`: TTL setting failed
- `2008`: Key already exists
- `2009`: Permission denied
- `2010`: Operation timeout

#### User Service Errors (3001-3010)
- `3001`: Invalid email format
- `3002`: Invalid display name
- `3003`: Password requirements not met
- `3004`: User creation failed
- `3005`: User already exists
- `3006`: Password hashing failed
- `3007`: User not found
- `3008`: Invalid credentials
- `3009`: Session creation failed
- `3010`: Session not found

#### Task Service Errors (3011-3020)
- `3011`: User ID required
- `3012`: Task description required
- `3013`: Description too long (>10000 chars)
- `3014`: Task validation failed
- `3015`: Task creation failed
- `3016`: Task not found
- `3017`: Permission denied (not owner)
- `3018`: Category not found
- `3019`: Invalid filter parameters
- `3020`: Operation failed

#### API/Handler Errors (4001-4020)
- `4001`: Missing session cookie
- `4002`: Invalid session
- `4003`: Session expired
- `4004`: User not found in session
- `4005`: Authentication required
- `4006`: Invalid request body
- `4007`: Invalid email format
- `4008`: Invalid password format
- `4009`: Missing required fields
- `4010`: Email already registered
- `4011`: Invalid credentials
- `4012`: Invalid task ID
- `4013`: Task not found
- `4014`: Invalid category name
- `4015`: Category not found

### How to Handle Different Error Types

#### Authentication Errors (401)
```javascript
if (response.status === 401) {
  // Session expired or invalid
  // Redirect to login
  window.location.href = '/login';
}
```

#### Validation Errors (400)
```javascript
if (response.status === 400) {
  const error = await response.json();
  if (error.code === '3003') {
    // Password requirements not met
    showError('Password must be at least 6 characters with 1 special character and 1 number');
  }
}
```

#### Not Found Errors (404)
```javascript
if (response.status === 404) {
  const error = await response.json();
  if (error.code === '4013') {
    // Task not found
    showError('Task no longer exists');
  }
}
```

#### Server Errors (500)
```javascript
if (response.status === 500) {
  // Retry with exponential backoff
  await retryWithBackoff(request, maxRetries);
}
```

### Retry Strategies

#### Exponential Backoff
```javascript
async function retryWithBackoff(fn, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await fn();
    } catch (error) {
      if (i === maxRetries - 1) throw error;
      
      const delay = Math.min(1000 * Math.pow(2, i), 10000);
      await new Promise(resolve => setTimeout(resolve, delay));
    }
  }
}
```

#### Idempotent Operations
Safe to retry:
- GET requests
- PUT requests (with same data)
- DELETE requests

Not safe to retry without verification:
- POST requests (might create duplicates)

## Advanced Usage

### Batch Operations

While the API doesn't currently support batch operations natively, you can implement client-side batching:

```javascript
// Batch create tasks
async function batchCreateTasks(tasks) {
  const results = await Promise.allSettled(
    tasks.map(task => 
      fetch('/api/v1/tasks', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(task)
      })
    )
  );
  
  return results.map((result, index) => ({
    task: tasks[index],
    success: result.status === 'fulfilled',
    response: result.value || result.reason
  }));
}
```

### Filtering and Sorting

#### Task Filtering
```bash
# Get tasks by category
GET /api/v1/tasks?category=work

# Get completed tasks
GET /api/v1/tasks?completed=true

# Get incomplete tasks in a category
GET /api/v1/tasks?category=personal&completed=false

# Include deleted tasks (within 7-day window)
GET /api/v1/tasks?includeDeleted=true
```

#### Pagination (Future Enhancement)
```bash
# Pagination parameters (when implemented)
GET /api/v1/tasks?limit=20&offset=0
GET /api/v1/tasks?limit=20&offset=20
```

### Category Management Workflows

#### Complete Category Workflow
```javascript
// 1. Create task with new category
const task = await createTask({
  description: "Review Q4 reports",
  category: "quarterly-review"
});

// 2. List all categories
const categories = await getCategories();
// Returns: ["quarterly-review", "daily", "urgent"]

// 3. Rename category (affects all tasks)
await renameCategory("quarterly-review", "q4-2024");

// 4. Delete category (removes from all tasks)
await deleteCategory("q4-2024");
```

#### Category Best Practices
1. Use lowercase with hyphens for consistency
2. Keep category names short and descriptive
3. Limit number of categories (suggest max 10-15)
4. Periodically clean up unused categories

### Task Lifecycle Management

#### Complete Task Lifecycle
```javascript
// 1. Create task
const task = await createTask({
  description: "Complete project documentation",
  category: "project-alpha"
});

// 2. Update completion status
await updateTaskCompletion(task.id, true);

// 3. Soft delete task
await deleteTask(task.id);

// 4. Task is now in deleted state (not visible in normal list)
const allTasks = await getTasks(); // Won't include deleted task

// 5. View deleted tasks
const tasksWithDeleted = await getTasks({ includeDeleted: true });

// 6. Restore task within 7 days
await restoreTask(task.id);

// 7. After 7 days, task is permanently deleted (automatic)
```

## Rate Limiting

### Current Limits
- 1000 requests per minute per IP address
- No per-endpoint specific limits
- Applies to all endpoints equally

### Handling Rate Limit Errors
```javascript
if (response.status === 429) {
  const retryAfter = response.headers.get('Retry-After') || 60;
  console.log(`Rate limited. Retry after ${retryAfter} seconds`);
  
  // Implement backoff
  setTimeout(() => {
    retryRequest();
  }, retryAfter * 1000);
}
```

### Best Practices to Avoid Rate Limiting
1. Cache responses when possible
2. Batch operations client-side
3. Implement request queuing
4. Use exponential backoff for retries
5. Monitor your request rate

## WebSocket Support (Future)

Currently, the API uses polling for real-time updates. WebSocket support is planned for future versions.

### Current Polling Approach
```javascript
// Poll for updates every 5 seconds
setInterval(async () => {
  const tasks = await getTasks();
  updateUI(tasks);
}, 5000);
```

### Planned WebSocket Events
- Task created/updated/deleted
- Category renamed/deleted
- Real-time collaboration features

## API Versioning

### Current Version
All endpoints are prefixed with `/api/v1/`

### Version Migration (Future)
When v2 is released:
- v1 endpoints will remain available
- Deprecation notices in headers
- Migration guide will be provided
- 6-month deprecation period

### Checking API Version
```bash
GET /api/version

Response:
{
  "version": "1.0.0",
  "apiVersion": "v1",
  "deprecated": false
}
```

## Testing Your Integration

### Using curl for Testing

```bash
# Complete test flow
# 1. Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!","displayName":"Test User"}' \
  -c cookies.txt

# 2. Create task
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"description":"Test task","category":"testing"}'

# 3. List tasks
curl -X GET http://localhost:8080/api/v1/tasks \
  -b cookies.txt | jq

# 4. Update task completion
curl -X PUT http://localhost:8080/api/v1/tasks/{task-id}/complete \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"completed":true}'

# 5. Logout
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -b cookies.txt
```

### Postman Collection

Import this JSON to Postman for a complete collection:

```json
{
  "info": {
    "name": "Task Tracker API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Auth",
      "item": [
        {
          "name": "Register",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"email\": \"{{email}}\",\n  \"password\": \"{{password}}\",\n  \"displayName\": \"{{displayName}}\"\n}"
            },
            "url": {
              "raw": "{{baseUrl}}/api/v1/auth/register",
              "host": ["{{baseUrl}}"],
              "path": ["api", "v1", "auth", "register"]
            }
          }
        }
      ]
    }
  ],
  "variable": [
    {
      "key": "baseUrl",
      "value": "http://localhost:8080",
      "type": "string"
    }
  ]
}
```

### Integration Test Example

```javascript
describe('Task Tracker API Integration', () => {
  let sessionCookie;
  
  beforeAll(async () => {
    // Register and login
    const loginResponse = await fetch('/api/v1/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        email: 'test@example.com',
        password: 'Test123!'
      })
    });
    sessionCookie = loginResponse.headers.get('set-cookie');
  });
  
  test('Complete task workflow', async () => {
    // Create task
    const createResponse = await fetch('/api/v1/tasks', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Cookie': sessionCookie
      },
      body: JSON.stringify({
        description: 'Test task',
        category: 'testing'
      })
    });
    
    expect(createResponse.status).toBe(201);
    const task = await createResponse.json();
    expect(task.id).toBeDefined();
    
    // Complete task
    const completeResponse = await fetch(`/api/v1/tasks/${task.id}/complete`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Cookie': sessionCookie
      },
      body: JSON.stringify({ completed: true })
    });
    
    expect(completeResponse.status).toBe(200);
  });
});
```