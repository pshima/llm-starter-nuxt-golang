# Task Tracker Application

A full-stack task management application built with Go (backend) and Nuxt.js (frontend - coming soon). Features user authentication, task management with categories, and soft delete functionality.

## Features

### Backend (Completed)
- **User Authentication**: Registration, login, logout with session-based authentication
- **Task Management**: Create, read, update completion status, and delete tasks
- **Categories**: User-created categories for organizing tasks
- **Soft Delete**: Tasks are soft-deleted with 7-day recovery window
- **RESTful API**: Clean API design following OpenAPI specification
- **Redis Storage**: All data stored in Redis with efficient data structures
- **Clean Architecture**: Separation of concerns with repositories, services, and handlers
- **Hot Reload**: Development environment with automatic code reloading
- **Comprehensive Testing**: >80% test coverage with unit and integration tests

### Frontend (Coming Soon)
- Nuxt.js with TypeScript
- Tailwind CSS for styling
- Pinia for state management

## Quick Start

### Prerequisites
- Docker and Docker Compose
- Make (optional, for convenience commands)
- Go 1.21+ (for local development without Docker)

### Running the Backend

1. Clone the repository:
```bash
git clone <repository-url>
cd llm-starter-nuxt-golang
```

2. Start the development environment:
```bash
make dev
# or without make:
docker-compose up --build
```

3. The backend will be available at `http://localhost:8080`
   - Health check: `GET http://localhost:8080/health`
   - API documentation: See `/backend/api/openapi.yaml`

4. MailHog web UI is available at `http://localhost:8025` for viewing emails

### Running Tests

Run all tests in Docker:
```bash
make test
# or without make:
docker-compose -f docker-compose.test.yml up --build
```

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login user
- `POST /api/v1/auth/logout` - Logout user
- `GET /api/v1/auth/me` - Get current user info

### Tasks
- `GET /api/v1/tasks` - List all tasks (with filters)
- `POST /api/v1/tasks` - Create new task
- `GET /api/v1/tasks/:id` - Get specific task
- `PUT /api/v1/tasks/:id/complete` - Update task completion
- `DELETE /api/v1/tasks/:id` - Soft delete task
- `POST /api/v1/tasks/:id/restore` - Restore deleted task

### Categories
- `GET /api/v1/categories` - List user's categories
- `PUT /api/v1/categories/:name` - Rename category
- `DELETE /api/v1/categories/:name` - Delete category

## Example API Usage

### Register a User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "Pass123!",
    "displayName": "John Doe"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "Pass123!"
  }' -c cookies.txt
```

### Create a Task
```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "description": "Complete project documentation",
    "category": "work"
  }'
```

### List Tasks
```bash
curl -X GET http://localhost:8080/api/v1/tasks \
  -b cookies.txt
```

## Development

### Project Structure
```
.
├── backend/
│   ├── cmd/server/         # Application entry point
│   ├── internal/
│   │   ├── domain/         # Business entities
│   │   ├── handlers/       # HTTP handlers
│   │   ├── services/       # Business logic
│   │   ├── repositories/   # Data access layer
│   │   └── middleware/     # HTTP middleware
│   ├── pkg/redis/          # Redis client
│   ├── api/                # OpenAPI specification
│   └── tests/              # Integration tests
├── frontend/               # Nuxt.js frontend (coming soon)
├── docker-compose.yml      # Development environment
└── docker-compose.test.yml # Test environment
```

### Make Commands
- `make dev` - Start development environment
- `make test` - Run all tests
- `make logs` - View container logs
- `make redis-cli` - Connect to Redis CLI
- `make mailhog` - Open MailHog web UI
- `make clean` - Clean up containers and volumes

### Environment Variables
Key environment variables (see docker-compose.yml for full list):
- `SERVER_HOST` - Server bind address (default: 0.0.0.0 in Docker)
- `SERVER_PORT` - Server port (default: 8080)
- `REDIS_HOST` - Redis hostname
- `SESSION_SECRET` - Session encryption key
- `SMTP_HOST` - SMTP server for emails

## Architecture Decisions

- **Clean Architecture**: Separation of concerns with clear boundaries between layers
- **Redis Data Structures**: Efficient use of Redis hashes, sets, and sorted sets
- **Session-Based Auth**: 7-day sessions stored in Redis
- **Soft Delete**: Tasks retained for 7 days after deletion for recovery
- **TDD Approach**: All features developed test-first with comprehensive coverage
- **Docker-First**: All development and testing done in containers

## Security Features

- Password requirements: minimum 6 characters, 1 special character, 1 number
- Bcrypt password hashing
- Session-based authentication with HttpOnly cookies
- User data isolation - users can only access their own data
- Input validation and sanitization
- Rate limiting support (1000 requests/minute per IP)

## Testing

The project includes comprehensive testing:
- **Unit Tests**: Repository, service, and handler layers
- **Integration Tests**: Full API workflow testing
- **Test Coverage**: >80% across all packages
- **Test Tools**: Go testing, testify, miniredis for Redis mocking

## Contributing

Please ensure all code follows the project guidelines in CLAUDE.md and includes appropriate tests.

## License

[License information to be added]