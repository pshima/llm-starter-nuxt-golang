const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: false });
  const page = await browser.newPage();
  
  // Generate unique test user
  const timestamp = Date.now();
  const testUser = {
    email: `test${timestamp}@example.com`,
    password: 'Test123!',
    displayName: 'Test User'
  };
  
  console.log('Testing with user:', testUser.email);
  
  // First register a new user
  console.log('\n1. Registering new user...');
  await page.goto('http://localhost:3000/register');
  await page.waitForTimeout(2000);
  
  await page.fill('input[name="displayName"]', testUser.displayName);
  await page.fill('input[name="email"]', testUser.email);
  await page.fill('input[name="password"]', testUser.password);
  
  await page.click('button[type="submit"]');
  await page.waitForTimeout(3000);
  
  // Check if redirected to dashboard
  const urlAfterRegister = page.url();
  console.log('URL after register:', urlAfterRegister);
  
  // Logout
  console.log('\n2. Logging out...');
  const response = await page.goto('http://localhost:8080/api/v1/auth/logout', { method: 'POST' });
  console.log('Logout response:', response?.status());
  
  // Now try to login
  console.log('\n3. Testing login...');
  await page.goto('http://localhost:3000/login');
  await page.waitForTimeout(2000);
  
  await page.fill('input[name="email"]', testUser.email);
  await page.fill('input[name="password"]', testUser.password);
  
  console.log('Submitting login form...');
  await page.click('button[type="submit"]');
  await page.waitForTimeout(3000);
  
  // Check final URL
  const finalUrl = page.url();
  console.log('\n4. Final URL after login:', finalUrl);
  
  if (finalUrl.includes('/dashboard')) {
    console.log('✅ LOGIN SUCCESSFUL - Redirected to dashboard');
  } else if (finalUrl.includes('/login')) {
    console.log('❌ LOGIN FAILED - Still on login page');
    // Take screenshot
    await page.screenshot({ path: 'login-failed.png' });
    console.log('Screenshot saved to login-failed.png');
  } else {
    console.log('🤔 Unexpected URL:', finalUrl);
  }
  
  await browser.close();
})();