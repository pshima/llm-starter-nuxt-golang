## The most important rule to always be followed and never to be broken.
- Do not make assumptions.  If you need the answer to something never guess, ask me.  If I have not answered a question, re prompt me until I answer it.  If you are in the middle of a task and have a question, at the next stopping point, ask me to clarify.  Review things with me early and often.

## Required - should NEVER be skipped or worked around.
- When completing a task, update README.md if it warrants an update with a goal of keeping the project's documentation up to date.
- Check `TASK.md` before starting a new task.  If the task is not listed, add it with a brief description and today's date.
- When working on a task that requires code in the application to be written, ALWAYS start with the tests before writing any code.  We refer to this as TDD, Test Driven Development.
- All backend services MUST have an openapi spec that is defined.  Whenever adding a new api, immedaitely update or add to the openapi spec.  Add this as a task immediately after implementation.  
- All testing and development should be done with Docker and test helper scripts.
- Use docker-compose.test.yml for test environment configuration.

## At the start of a conversation, always do these things
- Read `OVERVIEW.md` to understand the projects architecture, goals, style and consistency.  As we make noteworth decisions that impact the overall project, make sure `OVERVIEW.md` is kept up to date.
- Read `TEST.md` to understand how we do testing for the project.
- Read `AGENTS.md` to understand the agent configuration and which agents to utilize.
- Check port availability (3000 for frontend, 8080 for backend) before starting services. If ports are in use, stop and ask what to do.
- Ask me questions about what is unclear.
- We will often run with multiple agents, at the start of the conversation ensure you are utilizing all agents as specified in AGENTS.md.

# Code
- Build modules and components and build an architecture and codebase 
- Do not make up or imagine any libraries that do not exist.  If we import or utilize a library it needs to be a real library.
- Hot reload should be enabled for development.  We should not be rebuilding and restarting contianers to see changes.  Prioritize this.

## Philosophical
- Simple is better than complicated
- I always know best, do not assume that you know better than I do.

## Task Completion
- **Mark completed tasks in `TASK.md`** immediately after finishing them.
- Add new sub-tasks or TODOs discovered during development to `TASK.md` under a “Discovered During Work” section.

## Error Handling and Debugging
- Debugging is a critical part of software development, always create the necessary basic debugging abilities
- Logging should be enabled in the beginning of development and use clear and concise error codes.
- Each error should be assigned a globally unique code across the entire application in an incrementing numerical format.  Where there is the need for multiple error codes using the same integer, use increasing alphabet codes such as A, B, C (e.g., 1001A, 1001B, 1002, etc.).  
- Implement proper error handling and provide feedback to the user

## API Development
- Use appropriate HTTP status codes and error messages
- Validate all input parameters before processing
- Return consistent response formats

## Code Organization
- Follow clean architecture principles - handlers → services → models
- Keep business logic in services, not handlers
- Use dependency injection for testability
- Separate concerns into appropriate packages

## Code Guidelines
- Add code comments for all functions (1-3 lines)
- Use meaningful variable and function names
- Keep functions small and focused (< 50 lines preferred)
- Handle errors explicitly, don't ignore them
- Use early returns to reduce nesting

## Testing
- Write unit tests for all new functions BEFORE writing the function and update for existing functions as necessary.  
- Maintain code coverage above 80%
- Use table-driven tests for multiple scenarios
- Mock external dependencies in unit tests

## Security
- Validate and sanitize all user input
- Use parameterized queries (if adding database)
- Never log sensitive information
- Keep dependencies updated
- Use security best practices where it makes sense, always confirm when you make critical security decisions
- Use thread safe operations when appropriate

## Documentation
- Update README.md for user-facing features
- Update CLAUDE.md for development changes
- Keep OpenAPI spec in sync with code
- Document complex algorithms inline
- Update SECURITY.md for security changes
- All functions or major code definitions should have code comments that are 1-3 lines that describe what it does in plain english as well as why it is needed


## Database & Infrastructure
- **Database**: Redis (for both data storage and caching)
- **Sessions**: Redis-based session storage
- **Backend Port**: 8080
- **Frontend Port**: 3000
- **Session Duration**: 7 days
- **Remember Me**: Enabled (users stay logged in across browser sessions)

## Authentication
- **Type**: Session-based authentication
- **Methods**: 
  - Email/Password registration
  - Google OAuth integration (use placeholder configuration for development)
- **Password Requirements**: 
  - Minimum 6 characters
  - Must contain at least 1 special character
  - Must contain at least 1 number
- **Username**: Email address is the username (no separate username field)
- **Display Name**: Separate display name for showing in app (set during registration)
- **Unique Constraints**: Email addresses must be unique across the system

## Email Configuration
- **Service**: SMTP (configurable)
- **Development**: MailHog for local testing (included in Docker setup)
- **From Address**: noreply@no.reply.com
- **Invite System**: Unique links sent via email

## Admin Interface
- **Route**: /admin
- **Default Credentials**: 
  - Username: admin
  - Password: admin
  - Note: Admin account is auto-created on first application install when no admin exists
- **Admin Levels**: Single admin level (but multiple admin users allowed)
- **Capabilities**:
  - Developed as the app develops, this list should be updated.

## Technical Decisions
- **Backend Framework**: Go with Gin
- **Frontend Framework**: Nuxt.js with TypeScript
- **State Management**: Pinia for frontend state management
- **UI Styling**: Tailwind CSS (use wherever possible)
- **Redis Library**: Best available Go Redis library
- **Redis Data Types**: Traditional Redis data types (hashes, sets, lists, sorted sets, strings) - NOT Redis JSON. Use Redis hashes for objects, sets/lists for collections (specific patterns TBD based on project requirements)
- **Real-time Updates**: Polling (no websockets)
- **Rate Limiting**: IP-based, 1000 requests per minute per IP (storage method TBD)
- **Error Codes**: Sensible format (numerical incrementing as per CLAUDE.md)
- **Analytics**: Google Analytics with configurable ID (use dummy ID for development, e.g., 'UA-XXXXXXXXX-X')
- **Language**: English only (no i18n)
- **Data Retention**: Keep all data indefinitely
- **UI Testing**: Always use Playwright with browser UI (not headless) during development
- **Ongoing Development Workflows**: Always utilize hot reload capability and docker
- **Go Hot Reload**: Use Air for Go backend hot reload during development
- **Ports**: Backend: 8080, Frontend: 3000 (confirm availability before starting)
- **Domain**: Use localhost initially, configure custom domain when specified
- **Google OAuth**: Use placeholder configuration for development (to be replaced with actual credentials)

## Project Structure
- **Backend**: /backend directory
- **Frontend**: /frontend directory
- **Architecture**: Clean architecture pattern for backend
- **API Versioning**: v1 (internal versioning, hidden from users until v2 is developed)

## How to develop new code step by step
- Understand the problem first, ask questions, get clarifications.  Do not make assumptions.
- Share an outline/plan for what you intend to do, get confirmation of that plan before moving any further.
- When developing your plan, move component by component or break down the task into smaller sub tasks.
- Write the tests first, then write the code that passes the tests.
- When you think you are complete, run the full tests and make sure they pass
- When all the tests pass, ask yourself, is this work of the quality an expert in this area would agree with?  If you are unsure, ask questions.