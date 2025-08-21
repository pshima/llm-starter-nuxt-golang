const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ 
    headless: false,
    slowMo: 500 // Slow down actions so we can see what's happening
  });
  const context = await browser.newContext();
  const page = await context.newPage();
  
  // Generate unique test user
  const timestamp = Date.now();
  const testUser = {
    email: `test${timestamp}@example.com`,
    password: 'Test123!',
    displayName: 'Test User ' + timestamp
  };
  
  console.log('========================================');
  console.log('FULL AUTHENTICATION FLOW TEST');
  console.log('========================================');
  console.log('Test user:', testUser.email);
  console.log('');
  
  try {
    // Step 1: Visit home page
    console.log('Step 1: Visiting home page...');
    await page.goto('http://localhost:3000');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: 'flow-1-homepage.png' });
    console.log('✓ Homepage loaded');
    
    // Step 2: Navigate to registration
    console.log('\nStep 2: Navigating to registration page...');
    await page.click('text=Sign Up');
    await page.waitForURL('**/register');
    await page.screenshot({ path: 'flow-2-register.png' });
    console.log('✓ On registration page');
    
    // Step 3: Fill registration form
    console.log('\nStep 3: Filling registration form...');
    console.log('  Filling display name...');
    await page.fill('#displayName', testUser.displayName);
    console.log('  Filling email...');
    await page.fill('#email', testUser.email);
    console.log('  Filling password...');
    await page.fill('#password', testUser.password);
    await page.screenshot({ path: 'flow-3-register-filled.png' });
    console.log('✓ Form filled');
    
    // Step 4: Submit registration
    console.log('\nStep 4: Submitting registration...');
    await page.click('button:has-text("Create Account")');
    
    // Wait for navigation or error
    await Promise.race([
      page.waitForURL('**/dashboard', { timeout: 5000 }),
      page.waitForSelector('text=already exists', { timeout: 5000 }).catch(() => null),
      page.waitForSelector('.text-red-600', { timeout: 5000 }).catch(() => null)
    ]);
    
    await page.screenshot({ path: 'flow-4-after-register.png' });
    
    const currentUrl = page.url();
    if (currentUrl.includes('/dashboard')) {
      console.log('✓ Registration successful - redirected to dashboard');
    } else {
      console.log('✗ Registration may have failed, still on:', currentUrl);
      const errorText = await page.textContent('.text-red-600').catch(() => null);
      if (errorText) console.log('  Error:', errorText);
    }
    
    // Step 5: Check if we're logged in
    console.log('\nStep 5: Verifying login status...');
    if (currentUrl.includes('/dashboard')) {
      const welcomeText = await page.textContent('h1').catch(() => '');
      console.log('✓ On dashboard, welcome text:', welcomeText);
      
      // Step 6: Logout
      console.log('\nStep 6: Logging out...');
      // Look for logout button in header or profile dropdown
      const logoutButton = await page.locator('button:has-text("Logout"), a:has-text("Logout")').first();
      if (await logoutButton.isVisible()) {
        await logoutButton.click();
        await page.waitForLoadState('networkidle');
        console.log('✓ Logged out');
      } else {
        // Try profile dropdown
        await page.click('[data-testid="profile-menu"], button:has-text("' + testUser.displayName + '")').catch(() => {
          console.log('  Could not find logout button');
        });
      }
    }
    
    // Step 7: Test login
    console.log('\nStep 7: Testing login flow...');
    await page.goto('http://localhost:3000/login');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: 'flow-5-login-page.png' });
    console.log('✓ On login page');
    
    // Step 8: Fill login form
    console.log('\nStep 8: Filling login form...');
    console.log('  Filling email...');
    await page.fill('#email', testUser.email);
    console.log('  Filling password...');
    await page.fill('#password', testUser.password);
    await page.screenshot({ path: 'flow-6-login-filled.png' });
    console.log('✓ Login form filled');
    
    // Step 9: Submit login
    console.log('\nStep 9: Submitting login...');
    await page.click('button:has-text("Sign in"), button:has-text("Sign In"), button:has-text("Login")');
    
    // Wait for navigation
    await Promise.race([
      page.waitForURL('**/dashboard', { timeout: 5000 }),
      page.waitForSelector('.text-red-600', { timeout: 5000 }).catch(() => null)
    ]);
    
    await page.screenshot({ path: 'flow-7-after-login.png' });
    
    const finalUrl = page.url();
    if (finalUrl.includes('/dashboard')) {
      console.log('✓ Login successful - redirected to dashboard');
    } else {
      console.log('✗ Login failed, still on:', finalUrl);
      const errorText = await page.textContent('.text-red-600').catch(() => null);
      if (errorText) console.log('  Error:', errorText);
    }
    
    // Final summary
    console.log('\n========================================');
    console.log('TEST SUMMARY');
    console.log('========================================');
    console.log('Registration:', currentUrl.includes('/dashboard') ? '✓ PASSED' : '✗ FAILED');
    console.log('Login:', finalUrl.includes('/dashboard') ? '✓ PASSED' : '✗ FAILED');
    console.log('\nScreenshots saved:');
    console.log('  - flow-1-homepage.png');
    console.log('  - flow-2-register.png');
    console.log('  - flow-3-register-filled.png');
    console.log('  - flow-4-after-register.png');
    console.log('  - flow-5-login-page.png');
    console.log('  - flow-6-login-filled.png');
    console.log('  - flow-7-after-login.png');
    
  } catch (error) {
    console.error('\n✗ Test failed with error:', error.message);
    await page.screenshot({ path: 'flow-error.png' });
    console.log('Error screenshot saved to flow-error.png');
  }
  
  // Keep browser open for 5 seconds to see final state
  await page.waitForTimeout(5000);
  
  await browser.close();
  console.log('\n✓ Test complete');
})();