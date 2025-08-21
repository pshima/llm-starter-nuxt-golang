# Screenshot Capture System

## Overview

This project includes a comprehensive, automated screenshot capture system designed to provide permanent visual documentation of the entire Task Tracker application. The system captures every page, modal, form state, and user interaction flow.

## Key Features

### 🎯 **Complete Coverage**
- **Authentication Flow**: Registration, login, logout, validation errors
- **Main Application**: Dashboard, task management, navigation
- **Interactive Elements**: Modals, dropdowns, forms, buttons
- **Edge Cases**: Empty states, error pages, loading states
- **User Workflows**: Complete user journeys from registration to task management

### 🔄 **Automated & Maintainable**
- **One Command**: `npm run screenshots` captures everything
- **Self-Contained**: Creates test users, navigates through flows automatically
- **Clean State**: Each run starts fresh with new test data
- **Organized Output**: Screenshots organized by feature with descriptive names

### 📝 **Documentation Generation**
- **Automatic Markdown**: Generates `SCREENSHOTS.md` with embedded images
- **Organized Sections**: Groups screenshots by functionality
- **Metadata**: Includes timestamps, technical details, and update instructions
- **GitHub Ready**: Uses relative paths for proper rendering

### 🛠 **Developer Friendly**
- **Visual Mode**: Default browser window for debugging
- **Headless Mode**: For CI/CD and automated runs
- **Error Handling**: Server availability checks and clear error messages
- **Extensible**: Easy to add new pages and capture scenarios

## Quick Start

### Prerequisites
```bash
# Ensure servers are running
docker-compose up -d  # Backend on :8080
cd frontend && npm run dev  # Frontend on :3000
```

### Capture Screenshots
```bash
cd frontend

# Test the system first
npm run screenshots:test

# Capture all screenshots (visible browser)
npm run screenshots

# Or run headless (for automation)
npm run screenshots:headless
```

### Output
- **Screenshots**: `frontend/screenshots/` directory
- **Documentation**: `SCREENSHOTS.md` at project root
- **Organization**: Numbered files with descriptive names

## File Structure

```
frontend/
├── scripts/
│   ├── capture-screenshots.cjs     # Main capture script
│   ├── test-screenshot-system.cjs  # System verification
│   └── README.md                   # Detailed documentation
├── screenshots/                    # Generated screenshots
│   ├── 01-homepage-unauthenticated.png
│   ├── 02-login-page-empty.png
│   ├── 03-login-page-validation-errors.png
│   └── ...
└── package.json                    # npm scripts

SCREENSHOTS.md                      # Generated documentation
```

## Example Screenshots Captured

### Authentication Flow
1. **Homepage** - Unauthenticated state
2. **Login Page** - Empty form
3. **Login Page** - Validation errors  
4. **Registration Page** - Empty form
5. **Registration Page** - Validation errors
6. **Registration Page** - Correctly filled
7. **Dashboard** - Post-registration redirect

### Task Management
8. **Dashboard** - Main view
9. **Tasks Page** - Empty state
10. **Task Modal** - Creation form (empty)
11. **Task Modal** - Creation form (filled)
12. **Tasks Page** - With created tasks
13. **Task Modal** - Edit mode
14. **Tasks Page** - Multiple tasks with categories

### Navigation & States
15. **Header Navigation** - Authenticated user
16. **404 Page** - Error state
17. **Post-logout** - Login page

## Technical Details

### Browser Configuration
- **Viewport**: 1920x1080 (desktop focus)
- **Engine**: Chromium (Playwright)
- **Mode**: Visible by default, headless option available
- **Speed**: Slow motion for better screenshot quality

### Screenshot Quality
- **Full Page**: Captures entire page content, not just viewport
- **High Resolution**: Retina-quality images
- **Stable Timing**: Waits for page load and network idle
- **Animation Handling**: Allows time for modal animations

### Automation Features
- **Test User Creation**: Generates unique test users automatically
- **State Management**: Handles authentication, form filling, navigation
- **Error Recovery**: Robust error handling and cleanup
- **Server Health Checks**: Verifies prerequisites before starting

## Maintenance Workflow

### When to Update Screenshots
- **UI Changes**: After any visual modifications
- **New Features**: When adding pages or components  
- **Before Releases**: To ensure documentation is current
- **Design Reviews**: For stakeholder presentations

### Update Process
1. **Make UI Changes**: Implement your frontend modifications
2. **Start Servers**: Ensure both frontend and backend are running
3. **Run Capture**: Execute `npm run screenshots`
4. **Review Output**: Check `SCREENSHOTS.md` and screenshot files
5. **Commit Changes**: Add both screenshots and markdown to git

### CI/CD Integration
```yaml
# Example GitHub Actions step
- name: Update Screenshots
  run: |
    cd frontend
    npm run screenshots:headless
    git add ../SCREENSHOTS.md screenshots/
    git commit -m "docs: Update application screenshots" || true
```

## Customization

### Adding New Screenshots
Edit `frontend/scripts/capture-screenshots.cjs`:

```javascript
async function captureNewFeature() {
  // Navigate to your new page
  await this.page.goto('http://localhost:3000/new-page');
  
  // Take screenshot
  await this.takeScreenshot('19-new-page.png', 'New Feature Page', 'New Features');
  
  // Interact with UI
  await this.page.click('button');
  await this.takeScreenshot('20-new-page-interaction.png', 'New Feature Interaction', 'New Features');
}

// Add to main run() method
await this.captureNewFeature();
```

### Configuration Options
```javascript
// Modify in capture-screenshots.cjs
const VIEWPORT_SIZE = { width: 1920, height: 1080 };  // Viewport size
const SCREENSHOT_DIR = 'screenshots';                  // Output directory  
const slowMo = 1000;                                  // Animation delay
```

## Benefits

### For Development Teams
- **Visual Regression Testing**: Compare screenshots across versions
- **Documentation Maintenance**: Always up-to-date visual docs
- **Onboarding**: New team members can see complete app flow
- **Quality Assurance**: Verify UI consistency and completeness

### For Stakeholders
- **Design Reviews**: Complete visual overview of application
- **Progress Tracking**: See development progress visually
- **User Experience**: Review complete user journeys
- **Documentation**: Professional visual documentation

### For Users
- **Feature Overview**: See all available functionality
- **Usage Guide**: Visual guide to application features
- **Current State**: Always reflects latest application version

## Support & Troubleshooting

### Common Issues
- **Server Not Running**: System checks for server availability first
- **Permission Errors**: Ensure write access to screenshots directory
- **Browser Issues**: Run `npx playwright install` to install browsers
- **Timing Issues**: Adjust `waitForTimeout` values in script

### Getting Help
1. **Test System**: Run `npm run screenshots:test` first
2. **Visible Mode**: Use `npm run screenshots` to see what's happening
3. **Check Logs**: Browser console and network requests are logged
4. **Script Documentation**: See `frontend/scripts/README.md`

---

This screenshot system provides a permanent, maintainable solution for keeping visual documentation current with your application's evolution. It's designed to be run on-demand when you need updated documentation, ensuring your visual assets never become stale.