## PROJECT OVERVIEW

**IMPORTANT**: This file should be continuously updated as the project evolves. As we make noteworthy decisions that impact the overall project architecture, goals, style, and consistency, this document must be kept current.

## PROJECT PURPOSE & VISION

### Problem Statement
*What problem does this application solve? Who experiences this problem?*

### Solution Overview
*High-level description of how this application solves the problem*

### Target Users
*Primary and secondary user personas, their needs and pain points*

### Success Metrics
*How will we measure if this application is successful? (KPIs, usage metrics, business goals)*

## CORE FEATURES & FUNCTIONALITY

### MVP Features (Phase 1)
*List features required for initial launch with brief descriptions*
- [ ] Feature 1: Description
- [ ] Feature 2: Description

### Future Features (Phase 2+)
*Planned enhancements and additions post-MVP*
- [ ] Feature A: Description
- [ ] Feature B: Description

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
- **[YYYY-MM-DD]** Decision: Rationale

## SUCCESS CRITERIA

### Definition of Done
*What must be true for features to be considered complete*

### Launch Criteria
*Requirements before going to production*

## GLOSSARY

*Project-specific terms and their definitions*
- **Term**: Definition