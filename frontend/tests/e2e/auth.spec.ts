import { test, expect } from '@playwright/test'
import { 
  registerUser, 
  loginUser, 
  logoutUser, 
  fillRegistrationForm, 
  fillLoginForm,
  expectValidationError,
  ensureLoggedOut 
} from './helpers/auth-helper'
import { generateTestUser, invalidData } from './helpers/test-data'

/**
 * E2E Authentication Tests
 * Tests the complete user authentication flows including registration,
 * login, logout, and form validation
 */

test.describe('User Authentication', () => {
  
  test.beforeEach(async ({ page }) => {
    // Ensure user is logged out before each test
    await ensureLoggedOut(page)
  })

  test('should register a new user successfully', async ({ page }) => {
    const user = generateTestUser()
    
    await registerUser(page, user)
    
    // Verify we're on dashboard with user info
    await expect(page).toHaveURL('/dashboard')
    await expect(page.locator('text="' + user.displayName + '"')).toBeVisible()
  })

  test('should login existing user successfully', async ({ page }) => {
    const user = generateTestUser()
    
    // First register the user
    await registerUser(page, user)
    
    // Logout
    await logoutUser(page)
    
    // Then login
    await loginUser(page, user)
    
    // Verify successful login
    await expect(page).toHaveURL('/dashboard')
    await expect(page.locator('text="' + user.displayName + '"')).toBeVisible()
  })

  test('should logout user successfully', async ({ page }) => {
    const user = generateTestUser()
    
    // Register and login user
    await registerUser(page, user)
    
    // Logout
    await logoutUser(page)
    
    // Verify logout
    await expect(page).toHaveURL('/')
    await expect(page.locator('text="Sign in"')).toBeVisible()
  })

  test('should validate registration form fields', async ({ page }) => {
    // Test empty display name
    await fillRegistrationForm(page, {
      displayName: invalidData.user.emptyDisplayName,
      email: 'test@example.com',
      password: 'TestPass123!',
    })
    await page.click('button[type="submit"]')
    await expectValidationError(page, 'Display name is required')

    // Test invalid email
    await fillRegistrationForm(page, {
      displayName: 'Test User',
      email: invalidData.user.invalidEmail,
      password: 'TestPass123!',
    })
    await page.click('button[type="submit"]')
    await expectValidationError(page, 'Please enter a valid email address')

    // Test short password
    await fillRegistrationForm(page, {
      displayName: 'Test User',
      email: 'test@example.com',
      password: invalidData.user.shortPassword,
    })
    await page.click('button[type="submit"]')
    await expectValidationError(page, 'Password must be at least 6 characters')

    // Test password without special character
    await fillRegistrationForm(page, {
      displayName: 'Test User',
      email: 'test@example.com',
      password: invalidData.user.passwordWithoutSpecial,
    })
    await page.click('button[type="submit"]')
    await expectValidationError(page, 'Password must contain at least 1 special character')

    // Test password without number
    await fillRegistrationForm(page, {
      displayName: 'Test User',
      email: 'test@example.com',
      password: invalidData.user.passwordWithoutNumber,
    })
    await page.click('button[type="submit"]')
    await expectValidationError(page, 'Password must contain at least 1 number')
  })

  test('should validate login form fields', async ({ page }) => {
    // Test invalid email
    await fillLoginForm(page, {
      displayName: '',
      email: invalidData.user.invalidEmail,
      password: 'TestPass123!',
    })
    await page.click('button[type="submit"]')
    await expectValidationError(page, 'Please enter a valid email address')

    // Test empty password
    await fillLoginForm(page, {
      displayName: '',
      email: 'test@example.com',
      password: '',
    })
    await page.click('button[type="submit"]')
    await expectValidationError(page, 'Password is required')
  })

  test('should show error for invalid login credentials', async ({ page }) => {
    await fillLoginForm(page, {
      displayName: '',
      email: 'nonexistent@example.com',
      password: 'WrongPassword123!',
    })
    await page.click('button[type="submit"]')
    
    // Should show error message (exact message depends on backend implementation)
    await expect(page.locator('.error-message, .alert-error, [data-testid="error-message"]')).toBeVisible()
  })

  test('should prevent duplicate user registration', async ({ page }) => {
    const user = generateTestUser()
    
    // Register user first time
    await registerUser(page, user)
    await logoutUser(page)
    
    // Try to register same user again
    await fillRegistrationForm(page, user)
    await page.click('button[type="submit"]')
    
    // Should show error about duplicate email
    await expect(page.locator('.error-message, .alert-error, [data-testid="error-message"]')).toBeVisible()
    await expect(page.locator('text="already exists", text="already registered"')).toBeVisible()
  })

  test('should redirect authenticated users away from auth pages', async ({ page }) => {
    const user = generateTestUser()
    
    // Register and login user
    await registerUser(page, user)
    
    // Try to go to login page
    await page.goto('/login')
    await expect(page).toHaveURL('/dashboard')
    
    // Try to go to register page
    await page.goto('/register')
    await expect(page).toHaveURL('/dashboard')
  })

  test('should show loading state during form submission', async ({ page }) => {
    const user = generateTestUser()
    
    await page.goto('/register')
    await page.fill('input[name="displayName"]', user.displayName)
    await page.fill('input[name="email"]', user.email)
    await page.fill('input[name="password"]', user.password)
    
    // Click submit and immediately check for loading state
    await page.click('button[type="submit"]')
    
    // Should show loading indicator (button disabled or loading text)
    await expect(page.locator('button[type="submit"]:disabled, text="Registering...", text="Loading..."')).toBeVisible()
  })

  test('should persist session across browser refresh', async ({ page }) => {
    const user = generateTestUser()
    
    // Register user
    await registerUser(page, user)
    
    // Refresh page
    await page.reload()
    
    // Should still be logged in
    await expect(page).toHaveURL('/dashboard')
    await expect(page.locator('text="' + user.displayName + '"')).toBeVisible()
  })
})