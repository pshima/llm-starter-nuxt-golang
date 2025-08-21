const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: false });
  const page = await browser.newPage();
  
  // First create a user and login to get session cookie
  const timestamp = Date.now();
  const testUser = {
    email: `test${timestamp}@example.com`,
    password: 'Test123!',
    displayName: 'Test User'
  };
  
  console.log('Creating test user via API...');
  const registerResponse = await fetch('http://localhost:8080/api/v1/auth/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      email: testUser.email,
      password: testUser.password,
      displayName: testUser.displayName
    })
  });
  
  if (registerResponse.ok) {
    console.log('✅ User created successfully');
    
    // Get the session cookie from the response
    const setCookieHeader = registerResponse.headers.get('set-cookie');
    console.log('Set-Cookie header:', setCookieHeader);
    
    if (setCookieHeader) {
      // Parse the session cookie
      const sessionMatch = setCookieHeader.match(/session=([^;]+)/);
      if (sessionMatch) {
        const sessionCookie = sessionMatch[1];
        console.log('Session cookie extracted:', sessionCookie);
        
        // Now try to create a task via API
        console.log('\n=== TESTING DIRECT TASK CREATION VIA API ===');
        
        const taskData = {
          description: 'Test task created directly via API',
          category: 'Work'
        };
        
        console.log('Creating task via API:', taskData);
        const createTaskResponse = await fetch('http://localhost:8080/api/v1/tasks', {
          method: 'POST',
          headers: { 
            'Content-Type': 'application/json',
            'Cookie': `session=${sessionCookie}`
          },
          body: JSON.stringify(taskData)
        });
        
        if (createTaskResponse.ok) {
          const createdTask = await createTaskResponse.json();
          console.log('✅ Task created successfully:', createdTask);
          
          // Now check if we can fetch the tasks back
          console.log('\nFetching tasks list...');
          const tasksResponse = await fetch('http://localhost:8080/api/v1/tasks', {
            headers: { 'Cookie': `session=${sessionCookie}` }
          });
          
          if (tasksResponse.ok) {
            const tasks = await tasksResponse.json();
            console.log('✅ Tasks fetched:', tasks);
            
            if (tasks.tasks && tasks.tasks.length > 0) {
              console.log('🎉 SUCCESS: Task creation and retrieval working!');
            } else {
              console.log('❌ No tasks found in response');
            }
          } else {
            console.log('❌ Failed to fetch tasks:', tasksResponse.status);
          }
          
        } else {
          const errorData = await createTaskResponse.text();
          console.log('❌ Failed to create task:', createTaskResponse.status, errorData);
        }
      } else {
        console.log('❌ Could not extract session cookie');
      }
    } else {
      console.log('❌ No Set-Cookie header in response');
    }
  } else {
    console.log('❌ Failed to create user:', registerResponse.status);
    return;
  }
  
  await browser.close();
})();