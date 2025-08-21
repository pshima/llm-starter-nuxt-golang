/**
 * Test data generators for E2E tests
 * Provides consistent test data across all E2E tests
 */

export interface TestUser {
  displayName: string
  email: string
  password: string
}

export interface TestTask {
  description: string
  category: string
}

/**
 * Generate a unique test user with timestamp to avoid conflicts
 */
export function generateTestUser(): TestUser {
  const timestamp = Date.now()
  return {
    displayName: `Test User ${timestamp}`,
    email: `test${timestamp}@example.com`,
    password: 'TestPass123!',
  }
}

/**
 * Generate test users for specific scenarios
 */
export const testUsers = {
  valid: (): TestUser => generateTestUser(),
  withSpecialChars: (): TestUser => ({
    ...generateTestUser(),
    displayName: 'Test User @#$%',
  }),
}

/**
 * Generate test tasks for different scenarios
 */
export const testTasks = {
  simple: (): TestTask => ({
    description: `Test task ${Date.now()}`,
    category: 'Work',
  }),
  personal: (): TestTask => ({
    description: `Personal task ${Date.now()}`,
    category: 'Personal',
  }),
  longDescription: (): TestTask => ({
    description: `This is a very long task description that should test the UI layout and ensure it handles long text properly without breaking the design ${Date.now()}`,
    category: 'Work',
  }),
}

/**
 * Test categories for task organization
 */
export const testCategories = [
  'Work',
  'Personal',
  'Shopping',
  'Health',
  'Education',
]

/**
 * Invalid data for negative testing
 */
export const invalidData = {
  user: {
    invalidEmail: 'notanemail',
    shortPassword: '123',
    passwordWithoutSpecial: 'TestPass123',
    passwordWithoutNumber: 'TestPass!',
    emptyDisplayName: '',
  },
  task: {
    emptyDescription: '',
    tooLongDescription: 'x'.repeat(1001), // Assuming 1000 char limit
  },
}

/**
 * API endpoints for reference in tests
 */
export const endpoints = {
  auth: {
    register: '/api/v1/auth/register',
    login: '/api/v1/auth/login',
    logout: '/api/v1/auth/logout',
    me: '/api/v1/auth/me',
  },
  tasks: {
    list: '/api/v1/tasks',
    create: '/api/v1/tasks',
    update: (id: string) => `/api/v1/tasks/${id}`,
    delete: (id: string) => `/api/v1/tasks/${id}`,
    restore: (id: string) => `/api/v1/tasks/${id}/restore`,
  },
}