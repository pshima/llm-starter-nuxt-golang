import { Page, expect } from '@playwright/test'
import { TestTask, testTasks } from './test-data'

/**
 * Task helper functions for E2E tests
 * Provides reusable functions for task CRUD operations
 */

/**
 * Navigate to tasks page and wait for it to load
 */
export async function goToTasksPage(page: Page): Promise<void> {
  await page.goto('/tasks')
  await page.waitForLoadState('networkidle')
  
  // Wait for task list to be visible
  await expect(page.locator('[data-testid="task-list"]')).toBeVisible()
}

/**
 * Create a new task through the UI
 */
export async function createTask(page: Page, task: TestTask): Promise<void> {
  // Click add task button
  await page.click('button[data-testid="add-task-button"]')
  
  // Wait for modal to open
  await expect(page.locator('[data-testid="task-modal"]')).toBeVisible()
  
  // Fill task form
  await page.fill('input[name="description"]', task.description)
  await page.selectOption('select[name="category"]', task.category)
  
  // Submit form
  await page.click('button[data-testid="save-task-button"]')
  
  // Wait for modal to close
  await expect(page.locator('[data-testid="task-modal"]')).not.toBeVisible()
  
  // Verify task appears in list
  await expect(page.locator('text="' + task.description + '"')).toBeVisible()
}

/**
 * Edit an existing task through the UI
 */
export async function editTask(page: Page, originalDescription: string, newTask: TestTask): Promise<void> {
  // Find and click edit button for the task
  const taskRow = page.locator(`[data-testid="task-item"]:has-text("${originalDescription}")`)
  await taskRow.locator('button[data-testid="edit-task-button"]').click()
  
  // Wait for modal to open
  await expect(page.locator('[data-testid="task-modal"]')).toBeVisible()
  
  // Clear and fill new data
  await page.fill('input[name="description"]', '')
  await page.fill('input[name="description"]', newTask.description)
  await page.selectOption('select[name="category"]', newTask.category)
  
  // Submit form
  await page.click('button[data-testid="save-task-button"]')
  
  // Wait for modal to close
  await expect(page.locator('[data-testid="task-modal"]')).not.toBeVisible()
  
  // Verify updated task appears
  await expect(page.locator('text="' + newTask.description + '"')).toBeVisible()
}

/**
 * Toggle task completion status
 */
export async function toggleTaskCompletion(page: Page, taskDescription: string): Promise<void> {
  const taskRow = page.locator(`[data-testid="task-item"]:has-text("${taskDescription}")`)
  await taskRow.locator('input[type="checkbox"]').click()
  
  // Wait a moment for the state to update
  await page.waitForTimeout(500)
}

/**
 * Delete a task through the UI
 */
export async function deleteTask(page: Page, taskDescription: string): Promise<void> {
  // Find and click delete button for the task
  const taskRow = page.locator(`[data-testid="task-item"]:has-text("${taskDescription}")`)
  await taskRow.locator('button[data-testid="delete-task-button"]').click()
  
  // Confirm deletion if there's a confirmation dialog
  const confirmButton = page.locator('button:has-text("Delete"), button:has-text("Confirm")')
  if (await confirmButton.isVisible()) {
    await confirmButton.click()
  }
  
  // Verify task is removed from list
  await expect(page.locator('text="' + taskDescription + '"')).not.toBeVisible()
}

/**
 * Filter tasks by category
 */
export async function filterTasksByCategory(page: Page, category: string): Promise<void> {
  await page.selectOption('[data-testid="category-filter"]', category)
  await page.waitForTimeout(500) // Wait for filter to apply
}

/**
 * Filter tasks by completion status
 */
export async function filterTasksByStatus(page: Page, showCompleted: boolean): Promise<void> {
  const checkbox = page.locator('[data-testid="show-completed-filter"]')
  
  const isChecked = await checkbox.isChecked()
  if (isChecked !== showCompleted) {
    await checkbox.click()
  }
  
  await page.waitForTimeout(500) // Wait for filter to apply
}

/**
 * Get count of visible tasks
 */
export async function getVisibleTaskCount(page: Page): Promise<number> {
  const tasks = page.locator('[data-testid="task-item"]')
  return await tasks.count()
}

/**
 * Verify task exists in list
 */
export async function expectTaskExists(page: Page, taskDescription: string): Promise<void> {
  await expect(page.locator('text="' + taskDescription + '"')).toBeVisible()
}

/**
 * Verify task does not exist in list
 */
export async function expectTaskNotExists(page: Page, taskDescription: string): Promise<void> {
  await expect(page.locator('text="' + taskDescription + '"')).not.toBeVisible()
}

/**
 * Verify task completion status
 */
export async function expectTaskCompleted(page: Page, taskDescription: string, completed: boolean): Promise<void> {
  const taskRow = page.locator(`[data-testid="task-item"]:has-text("${taskDescription}")`)
  const checkbox = taskRow.locator('input[type="checkbox"]')
  
  if (completed) {
    await expect(checkbox).toBeChecked()
  } else {
    await expect(checkbox).not.toBeChecked()
  }
}

/**
 * Create multiple test tasks
 */
export async function createMultipleTasks(page: Page, count: number = 3): Promise<TestTask[]> {
  const tasks: TestTask[] = []
  
  for (let i = 0; i < count; i++) {
    const task = testTasks.simple()
    await createTask(page, task)
    tasks.push(task)
  }
  
  return tasks
}

/**
 * Wait for task list to load completely
 */
export async function waitForTaskListToLoad(page: Page): Promise<void> {
  await page.waitForLoadState('networkidle')
  
  // Wait for either tasks to load or empty state to show
  await Promise.race([
    page.locator('[data-testid="task-item"]').first().waitFor(),
    page.locator('[data-testid="empty-tasks"]').waitFor(),
  ]).catch(() => {
    // If neither appears, just continue - might be a loading state
  })
}