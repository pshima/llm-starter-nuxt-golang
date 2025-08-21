const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: false });
  const page = await browser.newPage();
  
  const timestamp = Date.now();
  const testUser = {
    email: `test${timestamp}@example.com`,
    password: 'Test123!',
    displayName: 'Test User'
  };
  
  console.log('Testing registration redirect...');
  console.log('User:', testUser.email);
  
  // Monitor network requests
  page.on('response', response => {
    const url = response.url();
    if (url.includes('/auth/')) {
      console.log(`API Response: ${response.status()} ${url}`);
    }
  });
  
  // Monitor console messages
  page.on('console', msg => {
    if (msg.type() === 'error') {
      console.log('Console Error:', msg.text());
    }
  });
  
  try {
    // Go to register page
    await page.goto('http://localhost:3000/register');
    await page.waitForLoadState('networkidle');
    console.log('On register page');
    
    // Fill form
    await page.fill('#displayName', testUser.displayName);
    await page.fill('#email', testUser.email);
    await page.fill('#password', testUser.password);
    console.log('Form filled');
    
    // Submit and watch what happens
    console.log('Submitting form...');
    await page.click('button:has-text("Create Account")');
    
    // Wait for any navigation and track it
    await page.waitForTimeout(3000);
    
    const finalUrl = page.url();
    console.log('\nFinal URL:', finalUrl);
    
    if (finalUrl.includes('/dashboard')) {
      console.log('✅ SUCCESS: Redirected to dashboard');
    } else if (finalUrl.includes('/login')) {
      console.log('❌ ISSUE: Redirected to login instead of dashboard');
    } else {
      console.log('❓ UNEXPECTED: Redirected to', finalUrl);
    }
    
    // Take screenshot
    await page.screenshot({ path: 'register-final.png' });
    console.log('Screenshot saved');
    
  } catch (error) {
    console.error('Test failed:', error.message);
    await page.screenshot({ path: 'register-error.png' });
  }
  
  await page.waitForTimeout(2000);
  await browser.close();
})();