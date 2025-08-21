## PROJECT OVERVIEW

**IMPORTANT**: This file should be continuously updated as the project evolves. As we make noteworthy decisions that impact the overall project architecture, goals, style, and consistency, this document must be kept current.

## PROJECT PURPOSE & VISION

### Problem Statement
Individuals struggle to organize and track their daily tasks effectively, leading to decreased productivity and forgotten responsibilities.

### Solution Overview
A simple, fast, and intuitive task management system that allows users to quickly capture tasks, organize them by categories, and track completion status with the ability to recover accidentally deleted items.

### Target Users
- **Primary**: Individual professionals managing personal and work tasks
- **Secondary**: Students organizing assignments and study tasks
- **Needs**: Quick task entry, categorization, reliable storage, simple interface

### Success Metrics
- User retention rate (target: 60% after 30 days)
- Average tasks created per user per week (target: 20+)
- Task completion rate (target: 70%)
- System uptime (target: 99.9%)

### Current Status (2025-08-21)
**MVP COMPLETE** - Both backend and frontend are fully implemented with:
- Full authentication system (register, login, logout)
- Complete task management (CRUD operations)
- Category system for organization
- Soft delete with 7-day recovery
- Beautiful, responsive UI with Tailwind CSS
- Comprehensive testing (unit and E2E)
- Docker development environment with hot reload
- Ready for production deployment

## CORE FEATURES & FUNCTIONALITY

### MVP Features (Phase 1 - COMPLETED)
**Backend (Completed)**
- [x] User Authentication: Registration, login, logout with sessions
- [x] Task CRUD: Create, read, update, delete tasks
- [x] Categories: User-created categories for organization
- [x] Soft Delete: 7-day recovery window for deleted tasks
- [x] API Documentation: Complete OpenAPI specification

**Frontend (Completed)**
- [x] Login/Register pages with form validation
- [x] Task list view with filtering (completed/deleted tasks)
- [x] Task creation and editing forms with inline editing
- [x] Category management interface (create, rename, delete)
- [x] Responsive design for mobile and desktop
- [x] Dashboard with welcome content and navigation
- [x] Profile dropdown with user info and logout
- [x] Unit and E2E testing infrastructure

### Future Features (Phase 2+)
- [ ] Task Due Dates: Add deadlines with reminders
- [ ] Task Priority: High/medium/low priority levels
- [ ] Task Search: Full-text search across all tasks
- [ ] Bulk Operations: Select and update multiple tasks
- [ ] Task Templates: Reusable task templates
- [ ] Data Export: Export tasks to CSV/JSON
- [ ] Dark Mode: Theme switching support
- [ ] Collaborative Lists: Share task lists with others
- [ ] Recurring Tasks: Daily/weekly/monthly task patterns
- [ ] Task Attachments: File uploads for tasks

### Non-Functional Requirements
- **Performance**: Response time targets, concurrent user limits
- **Security**: Authentication requirements, data protection needs
- **Scalability**: Expected growth, user/data volume projections
- **Availability**: Uptime requirements, maintenance windows
- **Compliance**: Regulatory requirements (GDPR, HIPAA, etc.)

## USER FLOWS & JOURNEYS

### Primary User Flow
*Step-by-step journey of the main user interaction*
1. User action/decision point
2. System response
3. Next step...

### Secondary Flows
*Additional important user journeys*

## DATA MODEL & BUSINESS LOGIC

### Core Entities
*Main data objects and their relationships*
- **Entity 1**: Properties, relationships, constraints
- **Entity 2**: Properties, relationships, constraints

### Business Rules
*Key business logic and validation rules*
- Rule 1: Description and implementation notes
- Rule 2: Description and implementation notes

### State Management
*How application state is managed and synchronized*

## ARCHITECTURE

- **Backend**: Go with Gin framework, Clean Architecture pattern
- **Frontend**: Nuxt.js with TypeScript and Tailwind CSS
- **State Management**: Pinia for Nuxt.js frontend state management
- **Database**: Redis for all data storage and session management
- **Testing**: TDD approach with miniredis for backend, Vitest for frontend, Playwright for E2E (browser UI mode)
- **Development**: Docker-based with hot reload (Air for Go, native for Nuxt)
- **Email**: SMTP with MailHog for development
- **Authentication**: Session-based with email/password and Google OAuth (placeholder credentials for development)
- **API**: RESTful with internal v1 versioning
- **Rate Limiting**: 1000 requests per minute per IP
- **Ports**: Backend: 8080, Frontend: 3000 (Claude should confirm availability before starting)
- **Domain**: To be configured when needed (use localhost for initial development)

## API DESIGN

### Endpoint Structure
*RESTful API endpoints and their purposes*
- `GET /api/v1/resource` - Description
- `POST /api/v1/resource` - Description

### Authentication Flow
*How users authenticate and maintain sessions*

### Error Handling Strategy
*Standardized error responses and codes*

## UI/UX DESIGN

### Design System
*Colors, typography, spacing, component library*

### Page Structure
*List of main pages/routes and their purposes*
- `/` - Homepage: Description
- `/dashboard` - User Dashboard: Description

### Responsive Breakpoints
*Mobile, tablet, desktop considerations*

## INTEGRATIONS & EXTERNAL SERVICES

### Third-Party Services
*External APIs, services, and tools*
- **Service Name**: Purpose, API limits, cost considerations

### Internal Integrations
*How different parts of the system communicate*

## DEPLOYMENT & OPERATIONS

### Environments
- **Development**: Local setup, tools, access
- **Staging**: Testing environment, data
- **Production**: Live environment, scaling

### Release Strategy
*How features move from development to production*

### Monitoring & Alerts
*What we monitor and when we get alerted*

## DOCUMENTATION

- CLAUDE.md - Development instructions and conventions
- TEST.md - Testing guidelines and practices
- TASK.md - Active and completed tasks tracking
- AGENTS.md - Agent configuration for Claude Code
- README.md - User-facing documentation

## CONSTRAINTS & LIMITATIONS

### Technical Constraints
*Known technical limitations or decisions*

### Business Constraints
*Budget, timeline, resource limitations*

### Assumptions
*Key assumptions we're making about users, technology, or business*

## RISKS & MITIGATIONS

### Technical Risks
- **Risk**: Description | **Mitigation**: How we address it

### Business Risks
- **Risk**: Description | **Mitigation**: How we address it

## DECISION LOG

### Major Decisions
*Record of important architectural and business decisions with dates*
- **[2025-08-21]** Frontend Implementation Complete: Nuxt.js frontend with TypeScript, Tailwind CSS, and Pinia state management fully implemented
- **[2025-08-21]** Testing Strategy: TDD approach with Vitest for unit tests and Playwright for E2E tests (browser UI mode for development)
- **[2025-08-21]** Pinia Store Imports: Explicit imports required for Pinia stores - auto-imports not reliable
- **[2025-08-21]** UI Design: Gradient backgrounds, responsive mobile-first design with Headless UI components
- **[2025-08-21]** Docker Integration: Full stack runs in Docker with hot reload for both frontend and backend

## SUCCESS CRITERIA

### Definition of Done
*What must be true for features to be considered complete*

### Launch Criteria
*Requirements before going to production*

## GLOSSARY

*Project-specific terms and their definitions*
- **Term**: Definition