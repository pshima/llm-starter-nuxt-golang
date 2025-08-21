const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: false });
  const page = await browser.newPage();
  
  // Enable console logging
  page.on('console', msg => console.log('Browser:', msg.type(), msg.text()));
  page.on('pageerror', error => console.log('Page error:', error.message));
  page.on('requestfailed', request => console.log('Failed:', request.url()));
  
  console.log('Navigating to homepage...');
  await page.goto('http://localhost:3001', { waitUntil: 'networkidle' });
  
  console.log('Waiting 5 seconds...');
  await page.waitForTimeout(5000);
  
  // Check what's on the page
  const title = await page.title();
  console.log('Title:', title);
  
  const bodyText = await page.textContent('body');
  console.log('Body text:', bodyText.substring(0, 200));
  
  // Take a screenshot
  await page.screenshot({ path: 'test-homepage.png' });
  console.log('Screenshot saved to test-homepage.png');
  
  await browser.close();
})();