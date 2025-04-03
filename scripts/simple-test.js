// Simple test script for the Memory Client API
const fetch = require('node-fetch');

async function testMemoryClientAPI() {
  try {
    console.log("Testing Memory Client API...");
    
    // Test sending a message
    const response = await fetch('http://localhost:10010/api/message', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        role: 'user',
        content: 'Simple test message for Memory Client API'
      })
    });
    
    const responseText = await response.text();
    console.log(`Status: ${response.status} ${response.statusText}`);
    console.log(`Response: ${responseText}`);
    
    if (response.ok) {
      console.log('✅ Success! The Memory Client API is working correctly.');
    } else {
      console.error('❌ Failed to send message to Memory Client API.');
    }
  } catch (error) {
    console.error('Error:', error.message);
  }
}

// Run the test
testMemoryClientAPI();
