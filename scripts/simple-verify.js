// Simple verification script for the Windsurf memory integration
const fetch = require('node-fetch');

// Test the direct Memory Client API
async function testDirectAPI() {
  console.log("\nTesting direct Memory Client API...");
  try {
    const response = await fetch('http://localhost:10010/api/message', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        role: 'user',
        content: 'Direct test message to Memory Client API'
      })
    });
    
    const responseText = await response.text();
    console.log(`Status: ${response.status} ${response.statusText}`);
    console.log(`Response: ${responseText}`);
    
    if (response.ok) {
      console.log('✅ Direct API test successful');
    } else {
      console.error('❌ Direct API test failed');
    }
  } catch (error) {
    console.error('❌ Error testing direct API:', error.message);
  }
}

// Test the proxy server
async function testProxyServer() {
  console.log("\nTesting proxy server...");
  try {
    const response = await fetch('http://localhost:10011/api/message', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        role: 'assistant',
        content: 'Test message through proxy server'
      })
    });
    
    const responseText = await response.text();
    console.log(`Status: ${response.status} ${response.statusText}`);
    console.log(`Response: ${responseText}`);
    
    if (response.ok) {
      console.log('✅ Proxy server test successful');
    } else {
      console.error('❌ Proxy server test failed');
    }
  } catch (error) {
    console.error('❌ Error testing proxy server:', error.message);
  }
}

// Test setting a conversation tag
async function testSetTag() {
  console.log("\nTesting setting conversation tag...");
  try {
    const tagName = "test-" + Date.now();
    const response = await fetch('http://localhost:10011/api/set-conversation-tag', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        tag: tagName
      })
    });
    
    const responseText = await response.text();
    console.log(`Status: ${response.status} ${response.statusText}`);
    console.log(`Response: ${responseText}`);
    
    if (response.ok) {
      console.log(`✅ Set tag test successful (tag: ${tagName})`);
    } else {
      console.error('❌ Set tag test failed');
    }
  } catch (error) {
    console.error('❌ Error testing set tag:', error.message);
  }
}

// Test getting a conversation tag
async function testGetTag() {
  console.log("\nTesting getting conversation tag...");
  try {
    const response = await fetch('http://localhost:10011/api/get-conversation-tag', {
      method: 'GET'
    });
    
    const responseText = await response.text();
    console.log(`Status: ${response.status} ${response.statusText}`);
    console.log(`Response: ${responseText}`);
    
    if (response.ok) {
      console.log('✅ Get tag test successful');
    } else {
      console.error('❌ Get tag test failed');
    }
  } catch (error) {
    console.error('❌ Error testing get tag:', error.message);
  }
}

// Run the tests one by one
async function runTests() {
  console.log("Running simple verification tests...");
  console.log("===================================");
  
  await testDirectAPI();
  await testProxyServer();
  await testSetTag();
  await testGetTag();
  
  console.log("\nAll tests completed!");
}

runTests();
