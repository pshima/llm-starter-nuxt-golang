import { test, expect } from '@playwright/test'
import { 
  createAndLoginUser, 
  ensureLoggedOut 
} from './helpers/auth-helper'
import { 
  goToTasksPage,
  createTask,
  editTask,
  toggleTaskCompletion,
  deleteTask,
  filterTasksByCategory,
  filterTasksByStatus,
  getVisibleTaskCount,
  expectTaskExists,
  expectTaskNotExists,
  expectTaskCompleted,
  createMultipleTasks,
  waitForTaskListToLoad
} from './helpers/task-helper'
import { testTasks, testCategories, invalidData } from './helpers/test-data'

/**
 * E2E Task Management Tests
 * Tests the complete task CRUD operations, filtering, and user workflows
 */

test.describe('Task Management', () => {
  
  test.beforeEach(async ({ page }) => {
    // Ensure clean state
    await ensureLoggedOut(page)
    
    // Create and login a user for each test
    await createAndLoginUser(page)
    
    // Navigate to tasks page
    await goToTasksPage(page)
  })

  test('should create a new task successfully', async ({ page }) => {
    const task = testTasks.simple()
    
    await createTask(page, task)
    
    // Verify task appears in list
    await expectTaskExists(page, task.description)
    
    // Verify task is not completed by default
    await expectTaskCompleted(page, task.description, false)
  })

  test('should edit an existing task', async ({ page }) => {
    const originalTask = testTasks.simple()
    const updatedTask = testTasks.personal()
    
    // Create initial task
    await createTask(page, originalTask)
    
    // Edit the task
    await editTask(page, originalTask.description, updatedTask)
    
    // Verify updated task appears
    await expectTaskExists(page, updatedTask.description)
    await expectTaskNotExists(page, originalTask.description)
  })

  test('should toggle task completion status', async ({ page }) => {
    const task = testTasks.simple()
    
    // Create task
    await createTask(page, task)
    
    // Mark as completed
    await toggleTaskCompletion(page, task.description)
    await expectTaskCompleted(page, task.description, true)
    
    // Mark as incomplete
    await toggleTaskCompletion(page, task.description)
    await expectTaskCompleted(page, task.description, false)
  })

  test('should delete a task successfully', async ({ page }) => {
    const task = testTasks.simple()
    
    // Create task
    await createTask(page, task)
    await expectTaskExists(page, task.description)
    
    // Delete task
    await deleteTask(page, task.description)
    
    // Verify task is removed
    await expectTaskNotExists(page, task.description)
  })

  test('should filter tasks by category', async ({ page }) => {
    const workTask = testTasks.simple() // Default category is 'Work'
    const personalTask = testTasks.personal()
    
    // Create tasks in different categories
    await createTask(page, workTask)
    await createTask(page, personalTask)
    
    // Filter by Work category
    await filterTasksByCategory(page, 'Work')
    await expectTaskExists(page, workTask.description)
    await expectTaskNotExists(page, personalTask.description)
    
    // Filter by Personal category
    await filterTasksByCategory(page, 'Personal')
    await expectTaskExists(page, personalTask.description)
    await expectTaskNotExists(page, workTask.description)
    
    // Show all categories
    await filterTasksByCategory(page, 'All')
    await expectTaskExists(page, workTask.description)
    await expectTaskExists(page, personalTask.description)
  })

  test('should filter tasks by completion status', async ({ page }) => {
    const task1 = testTasks.simple()
    const task2 = testTasks.personal()
    
    // Create tasks
    await createTask(page, task1)
    await createTask(page, task2)
    
    // Mark one task as completed
    await toggleTaskCompletion(page, task1.description)
    
    // Show only incomplete tasks (default)
    await filterTasksByStatus(page, false)
    await expectTaskNotExists(page, task1.description)
    await expectTaskExists(page, task2.description)
    
    // Show completed tasks
    await filterTasksByStatus(page, true)
    await expectTaskExists(page, task1.description)
    await expectTaskExists(page, task2.description)
  })

  test('should handle long task descriptions properly', async ({ page }) => {
    const longTask = testTasks.longDescription()
    
    await createTask(page, longTask)
    
    // Verify task is created and visible
    await expectTaskExists(page, longTask.description)
    
    // Verify the UI handles long text without breaking layout
    const taskElement = page.locator(`[data-testid="task-item"]:has-text("${longTask.description.substring(0, 50)}")`)
    await expect(taskElement).toBeVisible()
  })

  test('should validate task creation form', async ({ page }) => {
    // Click add task button
    await page.click('button[data-testid="add-task-button"]')
    await expect(page.locator('[data-testid="task-modal"]')).toBeVisible()
    
    // Try to submit empty form
    await page.click('button[data-testid="save-task-button"]')
    
    // Should show validation error
    await expect(page.locator('text="Description is required", text="Task description cannot be empty"')).toBeVisible()
    
    // Try with too long description (if there's a limit)
    if (invalidData.task.tooLongDescription) {
      await page.fill('input[name="description"]', invalidData.task.tooLongDescription)
      await page.click('button[data-testid="save-task-button"]')
      
      // Should show validation error for too long description
      await expect(page.locator('text="too long", text="maximum", text="limit"')).toBeVisible()
    }
  })

  test('should show empty state when no tasks exist', async ({ page }) => {
    await waitForTaskListToLoad(page)
    
    // Should show empty state message
    await expect(page.locator('[data-testid="empty-tasks"], text="No tasks found", text="Create your first task"')).toBeVisible()
    
    // Should show add task button
    await expect(page.locator('button[data-testid="add-task-button"]')).toBeVisible()
  })

  test('should handle multiple tasks efficiently', async ({ page }) => {
    // Create multiple tasks
    const tasks = await createMultipleTasks(page, 5)
    
    // Verify all tasks are visible
    for (const task of tasks) {
      await expectTaskExists(page, task.description)
    }
    
    // Verify correct count
    const taskCount = await getVisibleTaskCount(page)
    expect(taskCount).toBe(5)
  })

  test('should preserve task state when navigating away and back', async ({ page }) => {
    const task = testTasks.simple()
    
    // Create and complete task
    await createTask(page, task)
    await toggleTaskCompletion(page, task.description)
    
    // Navigate away to dashboard
    await page.goto('/dashboard')
    
    // Navigate back to tasks
    await goToTasksPage(page)
    
    // Verify task state is preserved
    await expectTaskExists(page, task.description)
    await expectTaskCompleted(page, task.description, true)
  })

  test('should support keyboard navigation in task modal', async ({ page }) => {
    // Open task modal
    await page.click('button[data-testid="add-task-button"]')
    await expect(page.locator('[data-testid="task-modal"]')).toBeVisible()
    
    // Use Tab to navigate between fields
    await page.keyboard.press('Tab')
    const activeElement = page.locator(':focus')
    await expect(activeElement).toHaveAttribute('name', 'description')
    
    // Close modal with Escape key
    await page.keyboard.press('Escape')
    await expect(page.locator('[data-testid="task-modal"]')).not.toBeVisible()
  })

  test('should show loading states during task operations', async ({ page }) => {
    const task = testTasks.simple()
    
    // Open task creation modal
    await page.click('button[data-testid="add-task-button"]')
    await page.fill('input[name="description"]', task.description)
    await page.selectOption('select[name="category"]', task.category)
    
    // Click submit and check for loading state
    await page.click('button[data-testid="save-task-button"]')
    
    // Should show loading indicator
    await expect(page.locator('button[data-testid="save-task-button"]:disabled, text="Saving...", text="Creating..."')).toBeVisible()
  })

  test('should handle task categories properly', async ({ page }) => {
    // Test creating tasks with different categories
    for (const category of testCategories.slice(0, 3)) {
      const task = {
        description: `Task for ${category}`,
        category: category,
      }
      
      await createTask(page, task)
      await expectTaskExists(page, task.description)
      
      // Verify category is displayed correctly
      const taskElement = page.locator(`[data-testid="task-item"]:has-text("${task.description}")`)
      await expect(taskElement.locator(`text="${category}"`)).toBeVisible()
    }
  })
})