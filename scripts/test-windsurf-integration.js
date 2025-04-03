// Test script for Windsurf memory integration
const fetch = require('node-fetch');

async function testWindsurfIntegration() {
  try {
    console.log("Testing Windsurf memory integration...");
    
    // Test sending a user message
    console.log("\n1. Testing user message:");
    await testSendMessage('user', 'This is a test user message from Windsurf');
    
    // Test sending an assistant message
    console.log("\n2. Testing assistant message:");
    await testSendMessage('assistant', 'This is a test assistant response from Windsurf');
    
    // Test setting a conversation tag
    console.log("\n3. Testing setting conversation tag:");
    await testSetConversationTag('test-integration');
    
    // Test getting the conversation tag
    console.log("\n4. Testing getting conversation tag:");
    await testGetConversationTag();
    
    console.log("\nAll tests completed!");
  } catch (error) {
    console.error("Test failed:", error);
  }
}

async function testSendMessage(role, content) {
  try {
    const response = await fetch('http://localhost:10011/api/message', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        role: role,
        content: content,
        id: generateUUID()
      })
    });
    
    const data = await response.text();
    console.log(`Status: ${response.status} ${response.statusText}`);
    console.log(`Response: ${data}`);
    
    if (response.ok) {
      console.log(`✅ Successfully sent ${role} message`);
    } else {
      console.error(`❌ Failed to send ${role} message`);
    }
  } catch (error) {
    console.error(`❌ Error sending ${role} message:`, error.message);
  }
}

async function testSetConversationTag(tag) {
  try {
    const response = await fetch('http://localhost:10011/api/set-conversation-tag', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        tag: tag
      })
    });
    
    const data = await response.text();
    console.log(`Status: ${response.status} ${response.statusText}`);
    console.log(`Response: ${data}`);
    
    if (response.ok) {
      console.log(`✅ Successfully set conversation tag to "${tag}"`);
    } else {
      console.error(`❌ Failed to set conversation tag`);
    }
  } catch (error) {
    console.error(`❌ Error setting conversation tag:`, error.message);
  }
}

async function testGetConversationTag() {
  try {
    const response = await fetch('http://localhost:10011/api/get-conversation-tag', {
      method: 'GET'
    });
    
    const data = await response.text();
    console.log(`Status: ${response.status} ${response.statusText}`);
    console.log(`Response: ${data}`);
    
    if (response.ok) {
      console.log(`✅ Successfully retrieved conversation tag`);
    } else {
      console.error(`❌ Failed to retrieve conversation tag`);
    }
  } catch (error) {
    console.error(`❌ Error retrieving conversation tag:`, error.message);
  }
}

// Generate a UUID for messages
function generateUUID() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0;
    const v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

// Run the tests
testWindsurfIntegration();
