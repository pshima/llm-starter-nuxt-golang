const { chromium } = require('playwright');
const fs = require('fs');
const path = require('path');

/**
 * Quick test to verify the screenshot system is working
 * This captures just a few key screenshots to test the infrastructure
 */

async function testScreenshotSystem() {
  console.log('🧪 Testing screenshot system...');

  let browser;
  try {
    // Check if servers are accessible
    console.log('🔍 Checking servers...');
    const frontendResponse = await fetch('http://localhost:3000');
    if (!frontendResponse.ok) {
      throw new Error('Frontend not accessible at localhost:3000');
    }
    console.log('✅ Frontend accessible');

    const backendResponse = await fetch('http://localhost:8080/api/v1/auth/me');
    if (backendResponse.status === 401) {
      console.log('✅ Backend accessible');
    } else {
      console.log('⚠️  Backend responding but with unexpected status');
    }

    // Test browser automation
    browser = await chromium.launch({ headless: true });
    const page = await browser.newPage();
    await page.setViewportSize({ width: 1920, height: 1080 });

    // Test screenshot directory
    const screenshotDir = path.join(__dirname, '..', 'screenshots', 'test');
    if (!fs.existsSync(screenshotDir)) {
      fs.mkdirSync(screenshotDir, { recursive: true });
    }

    // Take a simple test screenshot
    await page.goto('http://localhost:3000');
    await page.waitForLoadState('networkidle');
    const screenshotPath = path.join(screenshotDir, 'test-homepage.png');
    await page.screenshot({ path: screenshotPath, fullPage: true });
    
    console.log('✅ Test screenshot captured:', screenshotPath);
    
    // Verify file was created
    if (fs.existsSync(screenshotPath)) {
      const stats = fs.statSync(screenshotPath);
      console.log(`✅ Screenshot file size: ${stats.size} bytes`);
    } else {
      throw new Error('Screenshot file was not created');
    }

    console.log('🎉 Screenshot system test passed!');
    console.log('\nYou can now run the full screenshot capture:');
    console.log('  npm run screenshots');

  } catch (error) {
    console.error('❌ Screenshot system test failed:', error.message);
    console.error('\nTroubleshooting:');
    console.error('1. Ensure frontend is running: npm run dev');
    console.error('2. Ensure backend is running: docker-compose up');
    console.error('3. Install Playwright browsers: npx playwright install');
    process.exit(1);
  } finally {
    if (browser) {
      await browser.close();
    }
  }
}

testScreenshotSystem();