import { test, expect } from '@playwright/test'

/**
 * Smoke tests to verify Playwright setup and basic application loading
 */

test.describe('Smoke Tests', () => {
  
  test('should load the home page successfully', async ({ page }) => {
    await page.goto('/')
    
    // Should load without errors
    await expect(page).toHaveTitle(/Task Manager|Home|Nuxt/)
    
    // Should show some content
    await expect(page.locator('body')).toBeVisible()
  })

  test('should load the login page successfully', async ({ page }) => {
    await page.goto('/login')
    
    // Should load without errors
    await expect(page).toHaveTitle(/Sign in|Login|Nuxt/)
    
    // Should show login form
    await expect(page.locator('input[name="email"], input[type="email"]')).toBeVisible()
    await expect(page.locator('input[name="password"], input[type="password"]')).toBeVisible()
    await expect(page.locator('button[type="submit"]')).toBeVisible()
  })

  test('should load the register page successfully', async ({ page }) => {
    await page.goto('/register')
    
    // Should load without errors
    await expect(page).toHaveTitle(/Register|Sign up|Nuxt/)
    
    // Should show registration form
    await expect(page.locator('input[name="displayName"], input[name="display_name"]')).toBeVisible()
    await expect(page.locator('input[name="email"], input[type="email"]')).toBeVisible()
    await expect(page.locator('input[name="password"], input[type="password"]')).toBeVisible()
    await expect(page.locator('button[type="submit"]')).toBeVisible()
  })
})