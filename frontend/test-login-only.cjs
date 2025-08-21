const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: false });
  const page = await browser.newPage();
  
  // First create a user via API
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
  
  // Monitor console for errors
  page.on('console', msg => {
    if (msg.type() === 'error') {
      console.log('Console Error:', msg.text());
    } else if (msg.text().includes('📝') || msg.text().includes('✅') || msg.text().includes('👤')) {
      console.log('Auth Log:', msg.text());
    }
  });
  
  // Monitor API calls
  page.on('response', response => {
    const url = response.url();
    if (url.includes('/auth/')) {
      console.log(`API: ${response.status()} ${url}`);
    }
  });
  
  try {
    console.log('\n=== TESTING LOGIN FLOW ===');
    
    // 1. Go to login page
    console.log('1. Going to login page...');
    await page.goto('http://localhost:3000/login');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: 'login-test-1-page.png' });
    console.log('✓ On login page');
    
    // 2. Check if form fields exist
    console.log('\n2. Checking form fields...');
    const emailField = await page.locator('#email').isVisible();
    const passwordField = await page.locator('#password').isVisible();
    const submitButton = await page.locator('button[type="submit"]').isVisible();
    
    console.log('  Email field visible:', emailField);
    console.log('  Password field visible:', passwordField);
    console.log('  Submit button visible:', submitButton);
    
    if (!emailField || !passwordField || !submitButton) {
      console.log('❌ Form fields not found!');
      return;
    }
    
    // 3. Fill form with debug info
    console.log('\n3. Filling form...');
    console.log('  Filling email:', testUser.email);
    await page.fill('#email', testUser.email);
    
    console.log('  Filling password...');
    await page.fill('#password', testUser.password);
    
    await page.screenshot({ path: 'login-test-2-filled.png' });
    console.log('✓ Form filled');
    
    // 4. Submit form
    console.log('\n4. Submitting form...');
    await page.click('button[type="submit"]');
    
    // Wait and see what happens
    await page.waitForTimeout(3000);
    
    const finalUrl = page.url();
    console.log('\n5. Results:');
    console.log('  Final URL:', finalUrl);
    
    if (finalUrl.includes('/dashboard')) {
      console.log('  ✅ SUCCESS: Redirected to dashboard');
    } else if (finalUrl.includes('/login')) {
      console.log('  ❌ FAILED: Still on login page');
      
      // Check for error messages
      const errorMessages = await page.locator('.text-red-600').allTextContents();
      if (errorMessages.length > 0) {
        console.log('  Error messages:', errorMessages);
      }
    } else {
      console.log('  ❓ UNEXPECTED: Redirected to', finalUrl);
    }
    
    await page.screenshot({ path: 'login-test-3-final.png' });
    
  } catch (error) {
    console.error('Test failed:', error.message);
    await page.screenshot({ path: 'login-test-error.png' });
  }
  
  await page.waitForTimeout(2000);
  await browser.close();
})();