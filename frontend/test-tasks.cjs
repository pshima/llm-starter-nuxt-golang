const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: false });
  const page = await browser.newPage();
  
  // Monitor API calls
  page.on('response', response => {
    const url = response.url();
    if (url.includes('/api/')) {
      console.log(`API: ${response.status()} ${url}`);
    }
  });
  
  // Monitor console messages
  page.on('console', msg => {
    if (msg.type() === 'log') {
      const text = msg.text();
      if (text.includes('🔧') || text.includes('📝') || text.includes('✅') || text.includes('❌') || 
          text.includes('📋') || text.includes('📂') || text.includes('🚫') || text.includes('📁') || text.includes('💾')) {
        console.log(`Console: ${text}`);
      }
    }
  });
  
  // First create a user via API and login
  const timestamp = Date.now();
  const testUser = {
    email: `test${timestamp}@example.com`,
    password: 'Test123!',
    displayName: 'Test User'
  };
  
  console.log('Creating test user via API...');
  const response = await fetch('http://localhost:8080/api/v1/auth/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      email: testUser.email,
      password: testUser.password,
      displayName: testUser.displayName
    })
  });
  
  if (response.ok) {
    console.log('✅ User created successfully');
  } else {
    console.log('❌ Failed to create user:', response.status);
    return;
  }
  
  try {
    console.log('\n=== TESTING TASK MANAGEMENT ===');
    
    // 1. Login via form
    console.log('1. Logging in...');
    await page.goto('http://localhost:3000/login');
    await page.waitForLoadState('networkidle');
    
    await page.fill('#email', testUser.email);
    await page.fill('#password', testUser.password);
    await page.click('button[type="submit"]');
    
    // Wait for redirect to dashboard
    await page.waitForURL('**/dashboard');
    console.log('✓ Successfully logged in and redirected to dashboard');
    
    // 2. Navigate to tasks page
    console.log('\n2. Navigating to tasks page...');
    await page.click('a[href="/tasks"]');
    await page.waitForURL('**/tasks');
    await page.waitForLoadState('networkidle');
    console.log('✓ On tasks page');
    
    // 3. Check if "New Task" button exists
    console.log('\n3. Checking task creation UI...');
    const createButton = await page.locator('button:has-text("New Task")').isVisible();
    console.log('  Create Task button visible:', createButton);
    
    if (createButton) {
      // 4. Try to create a task
      console.log('\n4. Creating a new task...');
      await page.click('button:has-text("New Task")');
      
      // Wait for modal to appear (give it more time for transitions)
      await page.waitForTimeout(1000);
      
      // Check if modal elements are present
      const modalExists = await page.locator('[role="dialog"]').count();
      console.log('  Modal elements found:', modalExists);
      
      // Try to find the modal panel specifically
      const modalPanel = await page.locator('[data-testid="modal-panel"]').isVisible();
      console.log('  Modal panel visible:', modalPanel);
      
      if (modalExists > 0) {
        console.log('✓ Task creation modal opened');
      } else {
        console.log('❌ Modal not found');
        await page.screenshot({ path: 'modal-debug.png' });
        return;
      }
      
      // Fill in task details
      console.log('  Filling task description...');
      await page.fill('[data-testid="task-description"]', 'This is a test task created by automation');
      
      // Handle category - skip the dropdown and directly use new category input
      console.log('  Setting up category...');
      
      // Look for the new category input field (should be visible when no categories exist)
      const newCategoryInput = await page.locator('[data-testid="new-category-input"]').isVisible();
      if (newCategoryInput) {
        await page.fill('[data-testid="new-category-input"]', 'Work');
        console.log('  ✓ New category name entered: Work');
      } else {
        console.log('  ! New category input not visible, trying dropdown...');
        await page.click('[data-testid="task-category"] button');
        await page.waitForTimeout(500);
        const categoryOptions = await page.locator('[data-testid="task-category"] [role="option"]').count();
        if (categoryOptions > 0) {
          await page.locator('[data-testid="task-category"] [role="option"]').first().click();
          console.log('  ✓ Category selected from existing');
        } else {
          console.log('  ❌ No category options found');
        }
      }
      
      // Submit the task
      console.log('  Clicking Save button...');
      await page.click('[data-testid="save-button"]');
      
      // Wait for modal to close and task to appear
      await page.waitForSelector('[role="dialog"]', { state: 'hidden' });
      await page.waitForTimeout(1000);
      console.log('  ✓ Modal closed');
      
      // Check if task appears in the list (wait a bit longer for UI to update)
      await page.waitForTimeout(2000);
      
      const taskExists = await page.locator('text=This is a test task created by automation').isVisible();
      console.log('✓ Task created and visible in list:', taskExists);
      
      // Also try to find any task with "test task" in case of truncation  
      const anyTestTask = await page.locator('text=test task').count();
      console.log('  Tasks containing "test task":', anyTestTask);
      
      // Check total task elements on page  
      const totalTasks = await page.locator('[data-testid*="task"]').count();
      console.log('  Total task elements found:', totalTasks);
      
      if (taskExists) {
        console.log('\n5. Testing task interactions...');
        
        // Try to edit the task
        const editButton = await page.locator('[data-testid="edit-task"]').first().isVisible();
        console.log('  Edit button visible:', editButton);
        
        // Try to delete the task
        const deleteButton = await page.locator('[data-testid="delete-task"]').first().isVisible();
        console.log('  Delete button visible:', deleteButton);
        
        // Check task status options
        const statusButton = await page.locator('[data-testid="task-status"]').first().isVisible();
        console.log('  Status button visible:', statusButton);
      }
    } else {
      console.log('❌ New Task button not found - task UI may not be implemented');
    }
    
    await page.screenshot({ path: 'tasks-test-final.png' });
    console.log('\n✅ Task management test completed');
    
  } catch (error) {
    console.error('Test failed:', error.message);
    await page.screenshot({ path: 'tasks-test-error.png' });
  }
  
  await page.waitForTimeout(2000);
  await browser.close();
})();