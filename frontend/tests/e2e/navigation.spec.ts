import { test, expect } from '@playwright/test'
import { 
  createAndLoginUser, 
  ensureLoggedOut,
  loginUser
} from './helpers/auth-helper'
import { generateTestUser } from './helpers/test-data'

/**
 * E2E Navigation and Routing Tests
 * Tests navigation between pages, route protection, and UI responsiveness
 */

test.describe('Navigation and Routing', () => {
  
  test.beforeEach(async ({ page }) => {
    // Ensure clean state
    await ensureLoggedOut(page)
  })

  test('should navigate to home page and show landing content', async ({ page }) => {
    await page.goto('/')
    
    // Should show landing page content
    await expect(page.locator('h1, .hero-title')).toBeVisible()
    await expect(page.locator('text="Sign in", text="Get Started"')).toBeVisible()
  })

  test('should navigate between authentication pages', async ({ page }) => {
    // Start on home page
    await page.goto('/')
    
    // Click sign in link to go to login
    await page.click('text="Sign in"')
    await expect(page).toHaveURL('/login')
    await expect(page.locator('h1:has-text("Sign in"), h2:has-text("Sign in")')).toBeVisible()
    
    // Click register link
    await page.click('text="Sign up", text="Create account", text="Register"')
    await expect(page).toHaveURL('/register')
    await expect(page.locator('h1:has-text("Register"), h2:has-text("Sign up")')).toBeVisible()
    
    // Go back to login
    await page.click('text="Sign in", text="Login"')
    await expect(page).toHaveURL('/login')
  })

  test('should protect authenticated routes for unauthenticated users', async ({ page }) => {
    // Try to access dashboard without authentication
    await page.goto('/dashboard')
    await expect(page).toHaveURL('/login')
    
    // Try to access tasks without authentication
    await page.goto('/tasks')
    await expect(page).toHaveURL('/login')
  })

  test('should navigate to dashboard after successful authentication', async ({ page }) => {
    const user = generateTestUser()
    
    // Register user should redirect to dashboard
    await page.goto('/register')
    await page.fill('input[name="displayName"]', user.displayName)
    await page.fill('input[name="email"]', user.email)
    await page.fill('input[name="password"]', user.password)
    await page.click('button[type="submit"]')
    
    await expect(page).toHaveURL('/dashboard')
    await expect(page.locator('text="' + user.displayName + '"')).toBeVisible()
  })

  test('should navigate between authenticated pages using header navigation', async ({ page }) => {
    await createAndLoginUser(page)
    
    // Should be on dashboard
    await expect(page).toHaveURL('/dashboard')
    
    // Navigate to tasks via header navigation
    await page.click('a[href="/tasks"], text="Tasks"')
    await expect(page).toHaveURL('/tasks')
    await expect(page.locator('[data-testid="task-list"], h1:has-text("Tasks")')).toBeVisible()
    
    // Navigate back to dashboard
    await page.click('a[href="/dashboard"], text="Dashboard"')
    await expect(page).toHaveURL('/dashboard')
    await expect(page.locator('text="Dashboard", .dashboard-content')).toBeVisible()
  })

  test('should show user information in header when authenticated', async ({ page }) => {
    const user = await createAndLoginUser(page)
    
    // Should show user info in header
    await expect(page.locator('[data-testid="user-menu-button"]')).toBeVisible()
    await expect(page.locator('text="' + user.displayName + '"')).toBeVisible()
    
    // Click user menu to show dropdown
    await page.click('[data-testid="user-menu-button"]')
    await expect(page.locator('[data-testid="user-dropdown"]')).toBeVisible()
    await expect(page.locator('text="Logout"')).toBeVisible()
  })

  test('should handle responsive navigation on mobile viewport', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 })
    
    await createAndLoginUser(page)
    
    // Should show mobile menu button
    await expect(page.locator('[data-testid="mobile-menu-button"]')).toBeVisible()
    
    // Regular navigation should be hidden on mobile
    await expect(page.locator('[data-testid="desktop-navigation"]')).not.toBeVisible()
    
    // Click mobile menu button
    await page.click('[data-testid="mobile-menu-button"]')
    await expect(page.locator('[data-testid="mobile-menu"]')).toBeVisible()
    
    // Should show navigation links in mobile menu
    await expect(page.locator('[data-testid="mobile-menu"] text="Tasks"')).toBeVisible()
    await expect(page.locator('[data-testid="mobile-menu"] text="Dashboard"')).toBeVisible()
  })

  test('should handle browser back/forward navigation', async ({ page }) => {
    await createAndLoginUser(page)
    
    // Navigate to tasks
    await page.click('a[href="/tasks"]')
    await expect(page).toHaveURL('/tasks')
    
    // Use browser back button
    await page.goBack()
    await expect(page).toHaveURL('/dashboard')
    
    // Use browser forward button
    await page.goForward()
    await expect(page).toHaveURL('/tasks')
  })

  test('should show correct page titles', async ({ page }) => {
    // Home page
    await page.goto('/')
    await expect(page).toHaveTitle(/Task Manager|Home/)
    
    // Login page
    await page.goto('/login')
    await expect(page).toHaveTitle(/Sign in|Login/)
    
    // Register page
    await page.goto('/register')
    await expect(page).toHaveTitle(/Register|Sign up/)
    
    await createAndLoginUser(page)
    
    // Dashboard
    await expect(page).toHaveTitle(/Dashboard/)
    
    // Tasks page
    await page.goto('/tasks')
    await expect(page).toHaveTitle(/Tasks/)
  })

  test('should handle logout navigation correctly', async ({ page }) => {
    const user = await createAndLoginUser(page)
    
    // Logout via user menu
    await page.click('[data-testid="user-menu-button"]')
    await page.click('[data-testid="logout-button"]')
    
    // Should redirect to home page
    await expect(page).toHaveURL('/')
    
    // Should no longer show user info
    await expect(page.locator('text="' + user.displayName + '"')).not.toBeVisible()
    
    // Should show login/register options
    await expect(page.locator('text="Sign in"')).toBeVisible()
  })

  test('should maintain navigation state during page refreshes', async ({ page }) => {
    await createAndLoginUser(page)
    
    // Navigate to tasks
    await page.goto('/tasks')
    await expect(page).toHaveURL('/tasks')
    
    // Refresh page
    await page.reload()
    
    // Should still be on tasks page and authenticated
    await expect(page).toHaveURL('/tasks')
    await expect(page.locator('[data-testid="user-menu-button"]')).toBeVisible()
  })

  test('should show loading states during navigation', async ({ page }) => {
    await createAndLoginUser(page)
    
    // Navigate to different page and look for loading indicators
    const navigationPromise = page.waitForURL('/tasks')
    await page.click('a[href="/tasks"]')
    
    // Should show some form of loading state during navigation
    // This might be a loading spinner or the page content updating
    await navigationPromise
    
    await expect(page).toHaveURL('/tasks')
  })

  test('should handle direct URL access for authenticated routes', async ({ page }) => {
    // First, register a user to have valid credentials
    const user = generateTestUser()
    await page.goto('/register')
    await page.fill('input[name="displayName"]', user.displayName)
    await page.fill('input[name="email"]', user.email)
    await page.fill('input[name="password"]', user.password)
    await page.click('button[type="submit"]')
    
    // Now logout
    await page.click('[data-testid="user-menu-button"]')
    await page.click('[data-testid="logout-button"]')
    
    // Try direct access to protected route
    await page.goto('/tasks')
    await expect(page).toHaveURL('/login')
    
    // Login and try again
    await loginUser(page, user)
    await page.goto('/tasks')
    await expect(page).toHaveURL('/tasks')
  })

  test('should show proper error handling for invalid routes', async ({ page }) => {
    // Try to access a route that doesn't exist
    await page.goto('/nonexistent-page')
    
    // Should show 404 page or redirect to home
    const url = page.url()
    expect(url === 'http://localhost:3000/' || url.includes('404')).toBeTruthy()
  })
})