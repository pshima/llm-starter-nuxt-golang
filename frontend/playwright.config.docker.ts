import { defineConfig, devices } from '@playwright/test'

/**
 * Playwright configuration for E2E testing with Docker stack
 * - Headed mode for development visibility
 * - 5 second timeout maximum per test
 * - Screenshots on failure for debugging
 * - Uses existing Docker services instead of starting its own
 */
export default defineConfig({
  testDir: './tests/e2e',
  
  /* Run tests in files in parallel */
  fullyParallel: true,
  
  /* Fail the build on CI if you accidentally left test.only in the source code. */
  forbidOnly: !!process.env.CI,
  
  /* Retry on CI only */
  retries: process.env.CI ? 2 : 0,
  
  /* Opt out of parallel tests on CI. */
  workers: process.env.CI ? 1 : undefined,
  
  /* Reporter to use. See https://playwright.dev/docs/test-reporters */
  reporter: 'html',
  
  /* Shared settings for all the projects below. See https://playwright.dev/docs/api/class-testoptions. */
  use: {
    /* Base URL to use in actions like `await page.goto('/')`. */
    baseURL: 'http://localhost:3000',
    
    /* Collect trace when retrying the failed test. See https://playwright.dev/docs/trace-viewer */
    trace: 'on-first-retry',
    
    /* Take screenshot on failure for debugging */
    screenshot: 'only-on-failure',
    
    /* Record video on failure for debugging */
    video: 'retain-on-failure',
  },
  
  /* Global test timeout - 5 seconds max per test as per requirements */
  timeout: 5000,
  
  /* Expect timeout for assertions */
  expect: {
    timeout: 2000,
  },
  
  /* Configure projects for major browsers */
  projects: [
    {
      name: 'chromium',
      use: { 
        ...devices['Desktop Chrome'],
        /* Use headed mode for development visibility */
        headless: process.env.CI ? true : false,
      },
    },
  ],
  
  /* No webServer configuration - assumes Docker services are already running */
})