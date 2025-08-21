import { Page, expect } from '@playwright/test'
import { TestUser, generateTestUser } from './test-data'

/**
 * Authentication helper functions for E2E tests
 * Provides reusable functions for login, registration, and logout
 */

/**
 * Register a new user through the UI
 */
export async function registerUser(page: Page, user: TestUser): Promise<void> {
  await page.goto('/register')
  
  // Fill registration form
  await page.fill('input[name="displayName"]', user.displayName)
  await page.fill('input[name="email"]', user.email)
  await page.fill('input[name="password"]', user.password)
  
  // Submit form
  await page.click('button[type="submit"]')
  
  // Wait for redirect to dashboard
  await page.waitForURL('/dashboard')
  
  // Verify user is logged in
  await expect(page.locator('text="' + user.displayName + '"')).toBeVisible()
}

/**
 * Login an existing user through the UI
 */
export async function loginUser(page: Page, user: TestUser): Promise<void> {
  await page.goto('/login')
  
  // Fill login form
  await page.fill('input[name="email"]', user.email)
  await page.fill('input[name="password"]', user.password)
  
  // Submit form
  await page.click('button[type="submit"]')
  
  // Wait for redirect to dashboard
  await page.waitForURL('/dashboard')
  
  // Verify user is logged in
  await expect(page.locator('text="' + user.displayName + '"')).toBeVisible()
}

/**
 * Logout the current user through the UI
 */
export async function logoutUser(page: Page): Promise<void> {
  // Click user profile dropdown
  await page.click('button[data-testid="user-menu-button"]')
  
  // Click logout button
  await page.click('button[data-testid="logout-button"]')
  
  // Wait for redirect to home page
  await page.waitForURL('/')
  
  // Verify user is logged out
  await expect(page.locator('text="Sign in"')).toBeVisible()
}

/**
 * Register and login a new user in one step
 */
export async function createAndLoginUser(page: Page, user?: TestUser): Promise<TestUser> {
  const testUser = user || generateTestUser()
  await registerUser(page, testUser)
  return testUser
}

/**
 * Check if user is currently logged in
 */
export async function isLoggedIn(page: Page): Promise<boolean> {
  try {
    // Look for user menu button which indicates logged in state
    await page.locator('button[data-testid="user-menu-button"]').waitFor({ timeout: 1000 })
    return true
  } catch {
    return false
  }
}

/**
 * Ensure user is logged in, register and login if not
 */
export async function ensureLoggedIn(page: Page, user?: TestUser): Promise<TestUser> {
  if (await isLoggedIn(page)) {
    // Return a placeholder user if already logged in
    return { displayName: 'Current User', email: 'current@example.com', password: '' }
  }
  
  const testUser = user || generateTestUser()
  await registerUser(page, testUser)
  return testUser
}

/**
 * Ensure user is logged out
 */
export async function ensureLoggedOut(page: Page): Promise<void> {
  if (await isLoggedIn(page)) {
    await logoutUser(page)
  }
}

/**
 * Assert that form validation error is shown
 */
export async function expectValidationError(page: Page, message: string): Promise<void> {
  await expect(page.locator('text="' + message + '"')).toBeVisible()
}

/**
 * Assert that API error is shown
 */
export async function expectApiError(page: Page, message: string): Promise<void> {
  await expect(page.locator('.error-message, .alert-error, [data-testid="error-message"]')).toContainText(message)
}

/**
 * Fill registration form without submitting
 */
export async function fillRegistrationForm(page: Page, user: TestUser): Promise<void> {
  await page.goto('/register')
  await page.fill('input[name="displayName"]', user.displayName)
  await page.fill('input[name="email"]', user.email)
  await page.fill('input[name="password"]', user.password)
}

/**
 * Fill login form without submitting
 */
export async function fillLoginForm(page: Page, user: TestUser): Promise<void> {
  await page.goto('/login')
  await page.fill('input[name="email"]', user.email)
  await page.fill('input[name="password"]', user.password)
}