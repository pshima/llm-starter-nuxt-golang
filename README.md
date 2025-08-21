# Task Tracker Application

A full-stack task management application built with Go (backend) and Nuxt.js (frontend - coming soon). Features user authentication, task management with categories, and soft delete functionality.

## Documentation Index

| Document | Purpose | When to Read |
|----------|---------|--------------|
| [README.md](README.md) | Project overview and quick start | First time users |
| [ARCHITECTURE.md](ARCHITECTURE.md) | System design and structure | Before adding features |
| [API_GUIDE.md](API_GUIDE.md) | Complete API reference and examples | When integrating with API |
| [DEPLOYMENT.md](DEPLOYMENT.md) | Production deployment guide | Before going to production |
| [SECURITY.md](SECURITY.md) | Security implementation details | When handling auth or sensitive data |
| [TEST.md](TEST.md) | Testing patterns and strategies | When writing tests |
| [CLAUDE.md](CLAUDE.md) | Development guidelines and conventions | All developers |
| [TASK.md](TASK.md) | Task tracking and progress | Daily development |
| [SCREENSHOTS.md](SCREENSHOTS.md) | Complete UI documentation with screenshots | When reviewing UI/UX |
| [backend/api/openapi.yaml](backend/api/openapi.yaml) | OpenAPI specification | API development |

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

### Frontend (Completed)
- **Nuxt.js with TypeScript**: Modern Vue.js framework with type safety
- **Tailwind CSS**: Utility-first CSS framework for beautiful, responsive design
- **Pinia State Management**: Centralized state with auth and task stores
- **Authentication Pages**: Login and registration with form validation
- **Dashboard**: Welcome page with feature overview and statistics
- **Task Management UI**: Complete CRUD interface with inline editing
- **Category System**: Create and manage task categories
- **Filtering**: View completed and deleted tasks with toggles
- **Responsive Design**: Mobile-first approach with hamburger menu
- **Testing**: Unit tests with Vitest, E2E tests with Playwright

## Quick Start

### Prerequisites
- Docker and Docker Compose
- Make (optional, for convenience commands)
- Node.js 18+ (for local frontend development)
- Go 1.21+ (for local backend development)

### Running the Full Stack

1. Clone the repository:
```bash
git clone <repository-url>
cd llm-starter-nuxt-golang
```

2. Start both backend and frontend with Docker:
```bash
# Start all services (backend, frontend, Redis, MailHog)
docker-compose up --build

# Or run in detached mode
docker-compose up -d --build
```

3. Access the application:
   - **Frontend**: `http://localhost:3000`
   - **Backend API**: `http://localhost:8080`
   - **MailHog**: `http://localhost:8025` (view test emails)

4. Default test user (if seeded):
   - Email: `test@example.com`
   - Password: `Test123!`

### Running the Backend Only

1. Start the backend services:
```bash
make dev
# or without make:
docker-compose up backend redis mailhog
```

2. The backend will be available at `http://localhost:8080`
   - Health check: `GET http://localhost:8080/health`
   - API documentation: See `/backend/api/openapi.yaml`

### Running the Frontend Only

1. Make sure the backend is running (see above)

2. Navigate to the frontend directory:
```bash
cd frontend
```

3. Install dependencies:
```bash
npm install
```

4. Start the development server:
```bash
npm run dev
```

5. Open `http://localhost:3000` in your browser

### Running Tests

#### Backend Tests
Run all backend tests in Docker:
```bash
make test
# or without make:
docker-compose -f docker-compose.test.yml up --build
```

#### Frontend Unit Tests
```bash
cd frontend
npm run test          # Run once
npm run test:watch    # Watch mode
npm run test:coverage # With coverage report
```

#### Frontend E2E Tests
```bash
cd frontend

# Run with browser UI (development)
npm run test:e2e

# Run headless (CI)
npm run test:e2e:headless

# Run against Docker stack
npm run test:e2e:docker

# Interactive UI mode
npm run test:e2e:ui
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
├── frontend/
│   ├── components/         # Vue components
│   │   ├── AppHeader.vue   # Navigation header
│   │   ├── TaskItem.vue    # Individual task display
│   │   ├── TaskList.vue    # Task list container
│   │   ├── TaskModal.vue   # Task creation/edit modal
│   │   └── TaskFilters.vue # Filter controls
│   ├── layouts/            # Page layouts
│   │   ├── default.vue     # Authenticated layout
│   │   └── auth.vue        # Public/auth layout
│   ├── pages/              # Route pages
│   │   ├── index.vue       # Welcome page
│   │   ├── login.vue       # Login page
│   │   ├── register.vue    # Registration page
│   │   ├── dashboard.vue   # User dashboard
│   │   └── tasks.vue       # Task management
│   ├── stores/             # Pinia stores
│   │   ├── auth.ts         # Authentication state
│   │   └── tasks.ts        # Task management state
│   ├── middleware/         # Route middleware
│   ├── composables/        # Shared composables
│   ├── types/              # TypeScript types
│   └── tests/              # Tests
│       ├── unit/           # Unit tests
│       └── e2e/            # E2E tests
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

## Frontend Import Rules - CRITICAL

**NEVER use the `~` alias in imports - it's broken in Nuxt 4/Docker environments**

```typescript
// ❌ WRONG - Will cause "Failed to resolve import" errors
import { useAuthStore } from '~/stores/auth'
import { useApi } from '~/composables/useApi'

// ✅ CORRECT - Use relative paths
import { useAuthStore } from '../stores/auth'
import { useApi } from '../composables/useApi'
import type { User } from '../types'

// ✅ CORRECT - Use #app for Nuxt utilities
import { navigateTo } from '#app'
```

## Frontend-Backend Field Naming Convention

**ALWAYS use camelCase for JSON fields to match the backend**

```typescript
// ❌ WRONG - Backend expects camelCase
interface RegisterRequest {
  display_name: string  // Will cause "field required" errors
  created_at: string
}

// ✅ CORRECT - Matches backend JSON tags
interface RegisterRequest {
  displayName: string   // Matches Go: `json:"displayName"`
  createdAt: string     // Matches Go: `json:"createdAt"`
}
```

Common field mappings:
- `displayName` (NOT display_name)
- `createdAt` (NOT created_at)
- `updatedAt` (NOT updated_at)
- `userId` (NOT user_id)

## Testing

The project includes comprehensive testing for both backend and frontend:

### Backend Testing
- **Unit Tests**: Repository, service, and handler layers
- **Integration Tests**: Full API workflow testing
- **Test Coverage**: >80% across all packages
- **Test Tools**: Go testing, testify, miniredis for Redis mocking

### Frontend Testing
- **Unit Tests**: Components and pages with Vitest
- **E2E Tests**: Full user flows with Playwright
- **Test Coverage**: Components, stores, and composables
- **Test Patterns**: Page Object Model for maintainable E2E tests
- **Browser Testing**: Tests run in headed mode for visibility during development

## Screenshots & Documentation

The project includes an automated screenshot capture system for comprehensive UI documentation.

### Visual Documentation
- **Complete UI Coverage**: All pages, forms, modals, and states
- **Up-to-date Screenshots**: Easily regenerated when UI changes
- **Organized Documentation**: Generated `SCREENSHOTS.md` with embedded images
- **Desktop Focused**: 1920x1080 viewport for consistent documentation

### Capturing Screenshots
```bash
# Ensure both servers are running
docker-compose up -d  # Backend on :8080
cd frontend && npm run dev  # Frontend on :3000

# Capture all screenshots (visible browser)
cd frontend
npm run screenshots

# Or run in headless mode
npm run screenshots:headless
```

### What Gets Captured
- **Authentication Flow**: Login, register, validation states
- **Main Pages**: Dashboard, tasks, navigation
- **Interactive Elements**: Modals, forms, dropdowns
- **Task Management**: Creation, editing, filtering, categories
- **Error States**: 404 pages, validation errors

### Output
- **Screenshots**: Stored in `frontend/screenshots/` with descriptive names
- **Documentation**: Auto-generated `SCREENSHOTS.md` at project root
- **Organization**: Grouped by feature (Authentication, Task Management, etc.)

### Maintenance
Run screenshot capture after:
- Significant UI changes
- New feature additions  
- Before releases
- During design reviews

See `frontend/scripts/README.md` for detailed usage instructions.

## Troubleshooting

### Common Docker Issues

**Problem**: "Empty reply from server" when accessing endpoints
- **Solution**: Ensure `SERVER_HOST=0.0.0.0` is set in docker-compose.yml

**Problem**: "Cannot connect to Redis"
- **Solution**: Check Redis container is running: `docker-compose ps`
- **Solution**: Verify Redis host in environment variables

**Problem**: Port 8080 already in use
- **Solution**: Stop conflicting service: `lsof -i :8080` then kill the process
- **Solution**: Or change port in docker-compose.yml and environment variables

**Problem**: Changes not reflecting in Docker
- **Solution**: Rebuild containers: `docker-compose down && docker-compose up --build`
- **Solution**: Check Air hot reload logs: `docker-compose logs backend`

### Frontend Issues

**Problem**: Pinia store not found or undefined errors
- **Solution**: Always use explicit imports for Pinia stores:
```typescript
// REQUIRED: Explicit import
import { useAuthStore } from '~/stores/auth'
import { useTaskStore } from '~/stores/tasks'

// This will NOT work reliably with auto-imports
const authStore = useAuthStore() // May fail without explicit import
```

**Problem**: API calls failing with CORS errors
- **Solution**: Ensure backend is running on port 8080
- **Solution**: Check `nuxt.config.ts` has correct API URL configuration

**Problem**: Hot reload not working in frontend
- **Solution**: Ensure volume mounting is correct in docker-compose.yml
- **Solution**: Try restarting the frontend container

### Development Tips

#### Frontend Development

1. **Pinia Stores**: Always use explicit imports for stores
```typescript
// Correct
import { useAuthStore } from '~/stores/auth'

// Won't work reliably with auto-imports
const authStore = useAuthStore() // May fail
```

2. **API Calls**: Use the `useApi` composable for consistent error handling
```typescript
const { $api } = useNuxtApp()
const response = await $api('/tasks', { method: 'POST', body: data })
```

3. **Testing Components**: Use the test utilities in `frontend/tests/utils/`
```typescript
import { renderWithClient } from '~/tests/utils/test-utils'
```

4. **Hot Reload**: Frontend changes reflect immediately, no restart needed

#### Adding New Endpoints

1. Define the endpoint in `/backend/api/openapi.yaml`
2. Add domain model if needed in `/backend/internal/domain`
3. Implement repository method in `/backend/internal/repositories`
4. Add service logic in `/backend/internal/services`
5. Create handler in `/backend/internal/handlers`
6. Add tests at each layer
7. Update route in `/backend/cmd/server/main.go`

#### Testing Strategies

- **Unit Tests**: Run specific package tests: `go test ./internal/services/...`
- **Integration Tests**: Run with Docker: `make test`
- **Manual Testing**: Use curl or Postman with cookie storage
- **Coverage Report**: `go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out`

#### Debugging with Docker

```bash
# View real-time logs
docker-compose logs -f backend

# Access Redis CLI
docker exec -it task-tracker-redis redis-cli

# Check Redis keys for a user
docker exec -it task-tracker-redis redis-cli KEYS "user:*"

# Execute commands in backend container
docker exec -it task-tracker-backend sh
```

### Performance Notes

#### Redis Query Patterns

- **Efficient**: Using sets for collections, sorted sets for time-ordering
- **Avoid**: KEYS command in production (use SCAN instead)
- **Indexes**: Secondary indexes maintained for email lookup and categories
- **Pagination**: Use ZRANGE with LIMIT for sorted sets

#### Optimization Opportunities

1. **Caching**: Add caching layer for frequently accessed data
2. **Batch Operations**: Use Redis pipelines for multiple operations
3. **Connection Pooling**: Adjust pool size based on load
4. **Query Optimization**: Use partial key matching with SCAN
5. **Memory Management**: Set appropriate TTLs and eviction policies

## Contributing

Please ensure all code follows the project guidelines in CLAUDE.md and includes appropriate tests.

## License

[License information to be added]