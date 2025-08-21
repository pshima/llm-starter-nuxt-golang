# Screenshot Capture System

This directory contains scripts for automatically capturing comprehensive screenshots of the Task Tracker application.

## Overview

The screenshot capture system provides a standardized way to document the entire application UI, including:

- All authentication flows (login, register, logout)
- Main application pages (dashboard, tasks)
- Interactive elements (modals, forms, buttons)
- Error states and edge cases
- Multiple UI states (empty, populated, validation errors)

## Quick Start

### Prerequisites

1. **Servers Running**: Ensure both frontend and backend are running:
   ```bash
   # Terminal 1: Backend
   cd backend && docker-compose up
   
   # Terminal 2: Frontend  
   cd frontend && npm run dev
   ```

2. **Clean State**: For consistent screenshots, ensure you have a clean database state (no existing users/tasks)

### Capture Screenshots

```bash
# From the frontend directory
npm run screenshots
```

This will:
- Launch a browser window (visible by default)
- Create a test user automatically  
- Navigate through all application flows
- Capture screenshots of every page/state
- Generate `SCREENSHOTS.md` documentation
- Store all images in `frontend/screenshots/`

### Headless Mode (for CI/automation)

```bash
npm run screenshots:headless
```

## Output

### Screenshots Directory
All screenshots are saved to `frontend/screenshots/` with descriptive filenames:
- `01-homepage-unauthenticated.png`
- `02-login-page-empty.png`
- `03-login-page-validation-errors.png`
- etc.

### Documentation
A comprehensive `SCREENSHOTS.md` file is generated at the project root with:
- Organized sections (Authentication, Main Pages, Task Management, etc.)
- Embedded images with descriptions
- Technical metadata
- Instructions for updating

## Captured Content

### Authentication Flows
- Homepage (unauthenticated)
- Login page (empty, validation errors)  
- Register page (empty, validation errors, filled)
- Post-registration dashboard
- Logout flow

### Main Application
- Dashboard main view
- Tasks page (empty state, with tasks)
- Header navigation
- Error pages (404, etc.)

### Task Management
- Task creation modal (empty, filled)
- Task list (empty, populated)
- Task editing modal
- Category management
- Filters and sorting

### Interactive States
- Form validation errors
- Loading states
- Success/error messages
- Modal animations
- Hover effects

## Configuration

### Viewport Size
Screenshots are captured at **1920x1080** (desktop focus) as defined in the script:

```javascript
const VIEWPORT_SIZE = { width: 1920, height: 1080 };
```

### Browser Settings
- **Default**: Visible browser with slow motion for better screenshots
- **Headless**: Can be enabled via environment variable
- **Full Page**: Screenshots capture the entire page content

### Customization

Edit `scripts/capture-screenshots.cjs` to:
- Add new pages/flows
- Modify viewport size  
- Change screenshot timing
- Add new test scenarios

## Best Practices

### When to Capture
- After significant UI changes
- Before releases/deployments
- When adding new features
- After bug fixes that affect UI
- During design reviews

### Naming Convention
Screenshots follow the pattern: `{order}-{page}-{state}.png`
- `order`: Sequential number (01, 02, etc.)
- `page`: Page or component name
- `state`: UI state description

Examples:
- `01-homepage-unauthenticated.png`
- `12-task-modal-filled.png`  
- `15-tasks-page-with-filters.png`

### Maintenance
- Keep screenshots up to date with UI changes
- Review generated documentation for accuracy
- Clean up old screenshots when UI is redesigned
- Update script when new pages/features are added

## Troubleshooting

### Common Issues

**Browser doesn't launch:**
```bash
# Install Playwright browsers
npx playwright install
```

**Permission errors:**
```bash
# Ensure script is executable
chmod +x scripts/capture-screenshots.cjs
```

**Screenshots appear blank:**
- Verify servers are running on correct ports
- Check for JavaScript errors in browser console
- Ensure test user can be created

**Modal screenshots missing:**
- Modals require timing adjustments
- Increase `waitForTimeout` values in script
- Check modal animation durations

### Debugging

Run with visible browser to debug issues:
```bash
npm run screenshots  # Browser visible by default
```

Add console logging to the script for detailed debugging:
```javascript
console.log('🔍 Debug: Current URL:', this.page.url());
```

## Integration

### CI/CD Pipeline
Add to your deployment pipeline:

```yaml
- name: Capture Screenshots
  run: |
    npm run screenshots:headless
    git add SCREENSHOTS.md frontend/screenshots/
    git commit -m "docs: Update application screenshots"
```

### Documentation Updates
The system automatically updates:
- `SCREENSHOTS.md` with all screenshots and metadata
- Relative paths for proper GitHub/GitLab rendering
- Timestamps and technical details

## Architecture

The screenshot system is built with:
- **Playwright**: Browser automation and screenshot capture
- **Node.js**: Script execution and file management  
- **Markdown Generation**: Automated documentation creation
- **Modular Design**: Easy to extend and maintain

The main class `ScreenshotCapture` handles:
- Browser lifecycle management
- Test user creation
- Navigation and interaction
- Screenshot capture with metadata
- Documentation generation
- Cleanup and error handling

This provides a robust, maintainable solution for keeping application documentation visual and up-to-date.