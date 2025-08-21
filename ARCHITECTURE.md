# Architecture Documentation

## Backend Structure

### Clean Architecture Layers

The backend follows clean architecture principles with clear separation of concerns:

```
┌─────────────────────────────────────────────────┐
│                   HTTP Layer                    │
│              (Handlers & Middleware)            │
├─────────────────────────────────────────────────┤
│                 Service Layer                   │
│            (Business Logic & Rules)             │
├─────────────────────────────────────────────────┤
│               Repository Layer                  │
│              (Data Access Logic)                │
├─────────────────────────────────────────────────┤
│                 Domain Layer                    │
│            (Entities & Interfaces)              │
├─────────────────────────────────────────────────┤
│                Infrastructure                   │
│              (Redis, External APIs)             │
└─────────────────────────────────────────────────┘
```

### Dependency Flow

Dependencies flow inward, following the Dependency Inversion Principle:

- **Handlers** depend on Service interfaces (not concrete implementations)
- **Services** depend on Repository interfaces (not concrete implementations)
- **Repositories** implement domain interfaces
- **Domain** has no dependencies on other layers

### Layer Responsibilities

#### Domain Layer (`/internal/domain`)
- Defines core business entities (User, Task)
- Defines repository and service interfaces
- Contains business rules and validation logic
- No external dependencies

#### Repository Layer (`/internal/repositories`)
- Implements data access logic
- Handles Redis operations
- Manages data serialization/deserialization
- Implements domain repository interfaces

#### Service Layer (`/internal/services`)
- Implements business logic
- Orchestrates repository operations
- Handles complex validations
- Implements domain service interfaces

#### Handler Layer (`/internal/handlers`)
- Handles HTTP requests/responses
- Validates request payloads
- Transforms data for API responses
- Manages HTTP status codes and errors

#### Middleware Layer (`/internal/middleware`)
- Authentication/authorization
- Request logging
- Error handling
- CORS configuration

## Redis Schema

### Key Patterns

All Redis keys follow a hierarchical naming convention:

```
{entity}:{identifier}:{sub-entity}
```

### User Data

```
# User hash - stores user details
user:{userID}
  Fields: id, email, displayName, passwordHash, createdAt, updatedAt
  Type: Hash
  TTL: None (permanent)

# Email to user ID mapping
user:email:{email}
  Value: userID
  Type: String
  TTL: None (permanent)

# Session data
sessions:{sessionID}
  Fields: userID, createdAt, lastAccessed
  Type: Hash
  TTL: 7 days (604800 seconds)
```

### Task Data

```
# Task hash - stores task details
task:{taskID}
  Fields: id, userID, description, category, completed, createdAt, updatedAt, deletedAt
  Type: Hash
  TTL: None for active tasks

# User's active tasks
user:{userID}:tasks
  Values: Set of taskIDs
  Type: Set
  TTL: None

# User's tasks sorted by creation date
user:{userID}:tasks:sorted
  Values: taskIDs with timestamp scores
  Type: Sorted Set
  TTL: None

# User's deleted tasks
user:{userID}:tasks:deleted
  Values: taskIDs with deletion timestamp scores
  Type: Sorted Set
  TTL: Individual members expire after 7 days

# User's categories
user:{userID}:categories
  Values: Set of category names
  Type: Set
  TTL: None

# Tasks in a specific category
user:{userID}:category:{categoryName}
  Values: Set of taskIDs
  Type: Set
  TTL: None
```

### Data Type Choices

- **Hash**: Used for structured data (users, tasks) for efficient field access
- **Set**: Used for unique collections (task IDs, categories)
- **Sorted Set**: Used for time-ordered data (tasks by creation date, deleted tasks)
- **String**: Used for simple mappings (email to user ID)

### Indexing Strategies

1. **Primary Indexes**: Direct access by ID (user:{id}, task:{id})
2. **Secondary Indexes**: Email lookup (user:email:{email})
3. **Time-based Indexes**: Sorted sets with timestamp scores
4. **Category Indexes**: Sets for filtering by category

## Adding New Features

### Example: Adding Task Priority

1. **Update Domain Model** (`/internal/domain/task.go`):
```go
type Task struct {
    // ... existing fields
    Priority int `json:"priority"` // 1-5, higher is more important
}
```

2. **Update Repository** (`/internal/repositories/task_repository.go`):
```go
// Add priority to Redis hash
hash["priority"] = strconv.Itoa(task.Priority)

// Create priority index
user:{userID}:tasks:priority:{priority} -> Set of taskIDs
```

3. **Update Service** (`/internal/services/task_service.go`):
```go
func (s *taskService) CreateTask(userID, description, category string, priority int) (*domain.Task, error) {
    // Validate priority (1-5)
    if priority < 1 || priority > 5 {
        return nil, ErrInvalidPriority
    }
    // ... rest of implementation
}
```

4. **Update Handler** (`/internal/handlers/task_handler.go`):
```go
type CreateTaskRequest struct {
    Description string `json:"description" binding:"required,max=10000"`
    Category    string `json:"category,omitempty"`
    Priority    int    `json:"priority,omitempty"`
}
```

5. **Update Tests**:
- Add test cases for priority validation
- Test priority filtering
- Update integration tests

### Best Practices for New Features

1. **Start with Domain**: Define the business logic first
2. **Design Redis Schema**: Plan key structures and indexes
3. **Write Tests First**: Follow TDD approach
4. **Update OpenAPI Spec**: Document API changes
5. **Consider Migration**: Handle existing data if needed
6. **Update Documentation**: Keep README and API docs current

## Performance Considerations

### Redis Optimization

1. **Pipeline Operations**: Use pipelines for multiple operations
2. **Avoid KEYS Command**: Use sets/sorted sets for lookups
3. **TTL Management**: Set appropriate TTLs to prevent memory bloat
4. **Connection Pooling**: Configure pool size based on load

### Query Patterns

1. **Batch Reads**: Use MGET for multiple hash reads
2. **Sorted Set Queries**: Use ZRANGE with limits for pagination
3. **Index Maintenance**: Update indexes atomically with data

### Caching Strategy

Currently, Redis serves as both the primary database and cache. For scaling:

1. **Read Replicas**: Add Redis replicas for read scaling
2. **Sharding**: Partition users across Redis instances
3. **Memory Management**: Monitor memory usage and eviction policies

## Error Handling

### Error Code Allocation

Each layer has its own error code range:

- **1xxx**: System/Infrastructure errors
- **2xxx**: Repository/Data layer errors
- **3xxx**: Service/Business logic errors
- **4xxx**: Handler/API errors

### Error Propagation

Errors bubble up through layers with context:

```
Repository Error (2001: Redis connection failed)
    ↓
Service wraps it (3015: Task creation failed)
    ↓
Handler returns HTTP 500 with error code
```

## Security Architecture

### Authentication Flow

1. User provides email/password
2. Handler validates input format
3. Service verifies credentials
4. Repository checks password hash
5. Session created with 7-day TTL
6. Session ID returned as HTTP-only cookie

### Authorization

- All task operations verify user ownership
- Session validation on every protected route
- User isolation enforced at repository level

### Data Protection

- Passwords hashed with bcrypt (cost 10)
- Sessions stored server-side only
- No sensitive data in cookies
- Input validation at every layer