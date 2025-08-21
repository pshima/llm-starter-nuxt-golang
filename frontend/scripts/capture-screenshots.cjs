const { chromium } = require('playwright');
const fs = require('fs');
const path = require('path');

/**
 * Comprehensive Screenshot Capture System
 * 
 * This script captures screenshots of every page, state, and modal in the application.
 * Run with: node scripts/capture-screenshots.cjs
 * 
 * Features:
 * - Creates test user automatically
 * - Captures all authentication states
 * - Documents all main pages
 * - Shows modals and forms
 * - Generates markdown documentation
 * - Desktop-focused (1920x1080 viewport)
 */

const VIEWPORT_SIZE = { width: 1920, height: 1080 };
const SCREENSHOT_DIR = path.join(__dirname, '..', 'screenshots');
const DOCS_PATH = path.join(__dirname, '..', '..', 'SCREENSHOTS.md');
const FRONTEND_URL = 'http://localhost:3001'; // Changed to port 3001
const BACKEND_URL = 'http://localhost:8080';

class ScreenshotCapture {
  constructor() {
    this.browser = null;
    this.page = null;
    this.testUser = null;
    this.screenshots = [];
  }

  async checkServers() {
    console.log('🔍 Checking server availability...');
    
    try {
      // Check frontend
      const frontendResponse = await fetch(FRONTEND_URL);
      if (!frontendResponse.ok) {
        throw new Error(`Frontend server not responding: ${frontendResponse.status}`);
      }
      console.log(`✅ Frontend server (${FRONTEND_URL}) is running`);
      
      // Check backend
      const backendResponse = await fetch(`${BACKEND_URL}/api/v1/auth/me`);
      // We expect 401 for unauthenticated request, which means server is running
      if (backendResponse.status !== 401) {
        console.log('⚠️  Backend server response unexpected, but server appears to be running');
      } else {
        console.log(`✅ Backend server (${BACKEND_URL}) is running`);
      }
    } catch (error) {
      console.error('❌ Server check failed:', error.message);
      console.error('\nPlease ensure both servers are running:');
      console.error(`  Frontend: npm run dev (should be on ${FRONTEND_URL})`);
      console.error(`  Backend: docker-compose up (should be on ${BACKEND_URL})`);
      throw new Error('Required servers not available');
    }
  }

  async createPermanentTestUser() {
    // Use a fixed test user for screenshots
    this.testUser = {
      email: 'screenshots@tasktracker.local',
      password: 'Screenshots123!',
      displayName: 'Screenshot Test User'
    };

    console.log('👤 Creating permanent test user:', this.testUser.email);

    try {
      // Try to create the test user (will fail if already exists, which is fine)
      const response = await fetch(`${BACKEND_URL}/api/v1/auth/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: this.testUser.email,
          password: this.testUser.password,
          displayName: this.testUser.displayName
        })
      });

      if (response.ok) {
        console.log('✅ Test user created successfully');
      } else if (response.status === 409) {
        console.log('✅ Test user already exists, continuing...');
      } else {
        console.log('⚠️  Unexpected response creating test user:', response.status);
      }
    } catch (error) {
      console.log('⚠️  Error creating test user (continuing anyway):', error.message);
    }
  }

  async authenticateUser() {
    console.log('🔐 Authenticating test user for protected screenshots...');
    
    try {
      // Navigate to login page
      await this.page.goto(`${FRONTEND_URL}/login`);
      await this.page.waitForLoadState('networkidle');

      // Fill login form
      await this.page.fill('#email', this.testUser.email);
      await this.page.fill('#password', this.testUser.password);
      
      // Submit form
      await this.page.waitForSelector('button[type="submit"]', { state: 'visible', timeout: 5000 });
      await this.page.click('button[type="submit"]');
      
      // Wait for successful redirect (should go to dashboard)
      try {
        await this.page.waitForURL('**/dashboard', { timeout: 10000 });
        console.log('✅ Authentication successful - logged in to dashboard');
        return true;
      } catch (redirectError) {
        // If redirect fails, check if we're still on login page with errors
        const currentUrl = this.page.url();
        if (currentUrl.includes('/login')) {
          console.log('❌ Authentication failed - still on login page');
          return false;
        } else {
          console.log('✅ Authentication successful - redirected to:', currentUrl);
          return true;
        }
      }
    } catch (error) {
      console.error('❌ Authentication error:', error.message);
      return false;
    }
  }

  async setup() {
    // Check servers first
    await this.checkServers();
    
    // Ensure screenshots directory exists
    if (!fs.existsSync(SCREENSHOT_DIR)) {
      fs.mkdirSync(SCREENSHOT_DIR, { recursive: true });
    }

    // Launch browser
    const headless = process.env.HEADLESS === 'true';
    this.browser = await chromium.launch({ 
      headless: headless,
      slowMo: headless ? 500 : 1000 // Faster in headless mode
    });
    
    console.log(`🌐 Browser mode: ${headless ? 'Headless' : 'Visible'}`);
    
    this.page = await this.browser.newPage();
    await this.page.setViewportSize(VIEWPORT_SIZE);
    
    // Log console messages for debugging
    this.page.on('console', msg => {
      if (msg.type() === 'error') {
        console.log('🔴 Browser Console Error:', msg.text());
      }
    });
    
    // Log page errors
    this.page.on('pageerror', error => {
      console.log('🔴 Page Error:', error.message);
    });
    
    // Log failed requests
    this.page.on('requestfailed', request => {
      console.log('🔴 Request Failed:', request.url(), '-', request.failure().errorText);
    });

    // Create and authenticate permanent test user
    await this.createPermanentTestUser();

    console.log('🔧 Setup complete');
  }

  async takeScreenshot(filename, description, section = 'General', requiresAuth = false) {
    const screenshotPath = path.join(SCREENSHOT_DIR, filename);
    await this.page.screenshot({ 
      path: screenshotPath,
      fullPage: true 
    });
    
    this.screenshots.push({
      filename,
      description,
      section,
      requiresAuth,
      timestamp: new Date().toISOString()
    });
    
    const authIcon = requiresAuth ? '🔒' : '🌐';
    console.log(`📸 Captured: ${filename} - ${description} ${authIcon}`);
    
    // Small delay to ensure UI is stable
    await this.page.waitForTimeout(500);
  }

  async captureAuthenticationFlows() {
    console.log('\n📋 === CAPTURING AUTHENTICATION FLOWS ===');

    // 1. Home/Landing page (unauthenticated)
    await this.page.goto(`${FRONTEND_URL}`, { waitUntil: 'domcontentloaded' });
    // Wait for Vue app to render
    try {
      await this.page.waitForSelector('h1', { timeout: 10000 }); // Wait for heading
      await this.page.waitForTimeout(2000); // Additional wait for full render
    } catch (e) {
      console.log('⚠️  Homepage may not have rendered fully');
    }
    await this.takeScreenshot('01-homepage-unauthenticated.png', 'Homepage - Unauthenticated State', 'Authentication', false);

    // 2. Login page
    await this.page.goto(`${FRONTEND_URL}/login`, { waitUntil: 'domcontentloaded' });
    // Wait for Vue app to render by waiting for a specific element
    try {
      await this.page.waitForSelector('h1', { timeout: 10000 }); // Wait for any h1 heading
      await this.page.waitForTimeout(2000); // Additional wait for full render
    } catch (e) {
      console.log('⚠️  Login page may not have rendered fully');
    }
    await this.takeScreenshot('02-login-page-empty.png', 'Login Page - Empty Form', 'Authentication', false);

    // 3. Login page with validation errors - try clicking without filling form
    try {
      // Wait for the button to be visible and clickable
      await this.page.waitForSelector('button[type="submit"]', { state: 'visible', timeout: 5000 });
      await this.page.click('button[type="submit"]');
      await this.page.waitForTimeout(1000);
      await this.takeScreenshot('03-login-page-validation-errors.png', 'Login Page - Validation Errors', 'Authentication', false);
    } catch (error) {
      console.log('⚠️  Could not trigger validation errors on login page:', error.message);
      await this.takeScreenshot('03-login-page-validation-errors.png', 'Login Page - Validation Errors', 'Authentication', false);
    }

    // 4. Register page
    await this.page.goto(`${FRONTEND_URL}/register`);
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(2000); // Wait for hydration
    await this.takeScreenshot('04-register-page-empty.png', 'Register Page - Empty Form', 'Authentication', false);

    // 5. Register page with validation errors
    try {
      await this.page.waitForSelector('button[type="submit"]', { state: 'visible', timeout: 5000 });
      await this.page.click('button[type="submit"]');
      await this.page.waitForTimeout(1000);
      await this.takeScreenshot('05-register-page-validation-errors.png', 'Register Page - Validation Errors', 'Authentication', false);
    } catch (error) {
      console.log('⚠️  Could not trigger validation errors on register page:', error.message);
      await this.takeScreenshot('05-register-page-validation-errors.png', 'Register Page - Validation Errors', 'Authentication', false);
    }

    // 6. Register page filled out correctly (using a temp user for demo)
    try {
      const tempUser = {
        displayName: 'Demo User',
        email: 'demo@example.com',
        password: 'Demo123!'
      };
      
      await this.page.waitForSelector('#displayName', { state: 'visible', timeout: 5000 });
      await this.page.fill('#displayName', tempUser.displayName);
      await this.page.fill('#email', tempUser.email);
      await this.page.fill('#password', tempUser.password);
      await this.takeScreenshot('06-register-page-filled.png', 'Register Page - Form Filled Correctly', 'Authentication', false);
    } catch (error) {
      console.log('⚠️  Could not fill register form:', error.message);
      await this.takeScreenshot('06-register-page-filled.png', 'Register Page - Current State', 'Authentication', false);
    }

    // 7. Now authenticate with our permanent test user for subsequent screenshots
    const authSuccess = await this.authenticateUser();
    if (authSuccess) {
      await this.takeScreenshot('07-registration-success-dashboard.png', 'Registration Success - Redirected to Dashboard', 'Authentication', true);
    } else {
      console.log('⚠️  Authentication failed, skipping authenticated screenshots');
    }
    
    return authSuccess;
  }

  async captureMainPages(isAuthenticated = false) {
    console.log('\n📋 === CAPTURING MAIN APPLICATION PAGES ===');

    if (!isAuthenticated) {
      console.log('⚠️  Skipping authenticated pages - user not logged in');
      return false;
    }

    // 1. Dashboard page
    await this.page.goto(`${FRONTEND_URL}/dashboard`);
    await this.page.waitForLoadState('networkidle');
    await this.takeScreenshot('08-dashboard-main.png', 'Dashboard - Main View', 'Main Pages', true);

    // 2. Tasks page (empty state)
    await this.page.goto(`${FRONTEND_URL}/tasks`);
    await this.page.waitForLoadState('networkidle');
    await this.takeScreenshot('09-tasks-page-empty.png', 'Tasks Page - Empty State', 'Main Pages', true);

    // 3. Header navigation
    await this.takeScreenshot('10-header-navigation.png', 'Header Navigation - Logged In State', 'Main Pages', true);
    
    return true;
  }

  async captureTaskManagement(isAuthenticated = false) {
    console.log('\n📋 === CAPTURING TASK MANAGEMENT UI ===');

    if (!isAuthenticated) {
      console.log('⚠️  Skipping task management - user not authenticated');
      return false;
    }

    // Ensure we're on tasks page
    await this.page.goto(`${FRONTEND_URL}/tasks`);
    await this.page.waitForLoadState('networkidle');

    // Check if we're still authenticated by looking for login heading specifically
    const isOnLoginPage = await this.page.locator('h1:has-text("Login")').isVisible();
    if (isOnLoginPage) {
      console.log('⚠️  Redirected to login - session expired, re-authenticating...');
      const reauth = await this.authenticateUser();
      if (!reauth) {
        console.log('❌ Re-authentication failed, skipping task management');
        return false;
      }
      // Navigate back to tasks
      await this.page.goto(`${FRONTEND_URL}/tasks`);
      await this.page.waitForLoadState('networkidle');
    }

    // 1. Look for "New Task" button
    const newTaskButton = await this.page.locator('button:has-text("New Task")').isVisible();
    if (!newTaskButton) {
      // Try alternative selectors
      const newTaskAlt = await this.page.locator('[data-testid="new-task-button"]').isVisible();
      
      if (!newTaskAlt) {
        console.log('⚠️  New Task button not found, taking debug screenshot');
        await this.takeScreenshot('11-tasks-page-debug.png', 'Tasks Page - No New Task Button Found', 'Task Management', true);
        return false;
      }
      
      await this.page.click('[data-testid="new-task-button"]');
    } else {
      await this.page.click('button:has-text("New Task")');
    }
    
    await this.page.waitForTimeout(1500); // Wait for modal animation
    await this.takeScreenshot('11-task-modal-empty.png', 'Task Creation Modal - Empty Form', 'Task Management', true);

    // 2. Fill out the form
    await this.page.fill('[data-testid="task-description"]', 'Complete project documentation and review all components for final deployment');
    
    // Handle category (either select existing or create new)
    const newCategoryInput = await this.page.locator('[data-testid="new-category-input"]').isVisible();
    if (newCategoryInput) {
      await this.page.fill('[data-testid="new-category-input"]', 'Documentation');
    } else {
      // Try to select from dropdown if categories exist
      try {
        await this.page.click('[data-testid="task-category"] button');
        await this.page.waitForTimeout(500);
        const categoryOptions = await this.page.locator('[data-testid="task-category"] [role="option"]').count();
        if (categoryOptions > 0) {
          await this.page.locator('[data-testid="task-category"] [role="option"]').first().click();
        }
      } catch (error) {
        console.log('⚠️  Category selection failed, continuing...');
      }
    }

    await this.takeScreenshot('12-task-modal-filled.png', 'Task Creation Modal - Form Filled', 'Task Management', true);

    // 3. Submit the form
    try {
      await this.page.click('[data-testid="save-button"]');
      await this.page.waitForTimeout(3000); // Wait for task creation and modal close
      
      // Check if modal is still visible
      const modalStillVisible = await this.page.locator('[role="dialog"]').isVisible().catch(() => false);
      if (!modalStillVisible) {
        await this.takeScreenshot('13-tasks-page-with-task.png', 'Tasks Page - With Created Task', 'Task Management', true);

        // 4. Create a few more tasks for variety
        await this.createSampleTask('Review user authentication flow', 'Security');
        await this.createSampleTask('Implement task filtering functionality', 'Development');
        
        await this.page.waitForTimeout(2000);
        await this.takeScreenshot('14-tasks-page-multiple-tasks.png', 'Tasks Page - Multiple Tasks', 'Task Management', true);
      }
    } catch (error) {
      console.log('⚠️  Task creation failed:', error.message);
    }
    
    return true;
  }

  async createSampleTask(description, category) {
    await this.page.click('button:has-text("New Task")');
    await this.page.waitForTimeout(1000);
    await this.page.fill('[data-testid="task-description"]', description);
    
    const newCategoryInput = await this.page.locator('[data-testid="new-category-input"]').isVisible();
    if (newCategoryInput) {
      await this.page.fill('[data-testid="new-category-input"]', category);
    }
    
    await this.page.click('[data-testid="save-button"]');
    await this.page.waitForSelector('[role="dialog"]', { state: 'hidden' });
    await this.page.waitForTimeout(1000);
  }

  async captureErrorStates(isAuthenticated = false) {
    console.log('\n📋 === CAPTURING ERROR STATES ===');

    // 1. 404 page (if exists)
    await this.page.goto(`${FRONTEND_URL}/non-existent-page`);
    await this.page.waitForLoadState('networkidle');
    await this.takeScreenshot('17-404-page.png', '404 Error Page', 'Error States', false);

    // 2. Logout process (only if authenticated)
    if (isAuthenticated) {
      await this.page.goto(`${FRONTEND_URL}/dashboard`);
      await this.page.waitForLoadState('networkidle');
      
      // Look for logout button in header
      const logoutButton = await this.page.locator('text=Logout').first();
      if (await logoutButton.isVisible()) {
        await logoutButton.click();
        await this.page.waitForURL('**/login');
        await this.page.waitForLoadState('networkidle');
        await this.takeScreenshot('18-post-logout-login.png', 'Post-Logout Login Page', 'Authentication', false);
      }
    }
  }

  async generateMarkdown() {
    console.log('\n📋 === GENERATING DOCUMENTATION ===');

    const sections = {};
    
    // Group screenshots by section
    this.screenshots.forEach(screenshot => {
      if (!sections[screenshot.section]) {
        sections[screenshot.section] = [];
      }
      sections[screenshot.section].push(screenshot);
    });

    let markdown = `# Application Screenshots

*Generated on: ${new Date().toLocaleString()}*

This document provides a comprehensive visual overview of the Task Tracker application, including all pages, states, and user interactions.

## Overview

The Task Tracker is a full-stack web application built with:
- **Frontend**: Nuxt.js 4, Vue 3, TypeScript, Tailwind CSS
- **Backend**: Go with Gin framework
- **Database**: Redis for session and data storage
- **Authentication**: Session-based with HTTP-only cookies

---

`;

    // Generate sections
    Object.keys(sections).sort().forEach(sectionName => {
      markdown += `## ${sectionName}\n\n`;
      
      sections[sectionName].forEach(screenshot => {
        const relativePath = `frontend/screenshots/${screenshot.filename}`;
        markdown += `### ${screenshot.description}\n\n`;
        markdown += `![${screenshot.description}](${relativePath})\n\n`;
        markdown += `*Captured: ${new Date(screenshot.timestamp).toLocaleString()}*\n\n`;
        markdown += '---\n\n';
      });
    });

    // Add footer
    markdown += `## Technical Notes

### Screenshot Details
- **Viewport**: ${VIEWPORT_SIZE.width}x${VIEWPORT_SIZE.height} (Desktop)
- **Browser**: Chromium (Playwright)
- **Total Screenshots**: ${this.screenshots.length}
- **Sections Covered**: ${Object.keys(sections).length}

### How to Update Screenshots

To regenerate these screenshots:

\`\`\`bash
cd frontend

# Test the system first
npm run screenshots:test

# Capture all screenshots (visible browser)
npm run screenshots

# Or run in headless mode (for CI/automation)
npm run screenshots:headless
\`\`\`

### Available Commands

- \`npm run screenshots\` - Interactive mode with visible browser
- \`npm run screenshots:headless\` - Automated mode for CI/CD
- \`npm run screenshots:test\` - Quick system verification

### When to Update Screenshots

Update screenshots after:
- **UI Changes**: Any visual modifications to pages or components
- **New Features**: Adding new pages, modals, or functionality  
- **Bug Fixes**: Fixing visual issues or validation problems
- **Before Releases**: Ensure documentation reflects current state
- **Design Reviews**: For stakeholder presentations

### Prerequisites
- Frontend server running on ${FRONTEND_URL}
- Backend server running on ${BACKEND_URL}
- Redis server running
- All dependencies installed (\`npm install\` in frontend directory)

---

*This documentation is automatically generated and should be updated whenever significant UI changes are made.*
`;

    fs.writeFileSync(DOCS_PATH, markdown);
    console.log(`📝 Documentation generated: ${DOCS_PATH}`);
  }

  async cleanup() {
    if (this.browser) {
      await this.browser.close();
    }
    console.log('🧹 Cleanup complete');
  }

  async run() {
    try {
      console.log('🚀 Starting comprehensive screenshot capture...\n');
      
      await this.setup();
      
      // Capture authentication flows and track authentication state
      const isAuthenticated = await this.captureAuthenticationFlows();
      
      // Capture authenticated pages if login was successful
      await this.captureMainPages(isAuthenticated);
      await this.captureTaskManagement(isAuthenticated);
      await this.captureErrorStates(isAuthenticated);
      
      await this.generateMarkdown();
      
      console.log(`\n✅ Screenshot capture complete!`);
      console.log(`📸 Captured ${this.screenshots.length} screenshots`);
      console.log(`📝 Documentation: ${DOCS_PATH}`);
      console.log(`🗂️  Screenshots: ${SCREENSHOT_DIR}`);
      
    } catch (error) {
      console.error('❌ Screenshot capture failed:', error);
      throw error;
    } finally {
      await this.cleanup();
    }
  }
}

// Run the capture if this script is executed directly
if (require.main === module) {
  const capture = new ScreenshotCapture();
  capture.run().catch(console.error);
}

module.exports = ScreenshotCapture;