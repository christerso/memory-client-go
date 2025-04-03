// Comprehensive test script for the Windsurf memory integration
const fetch = require('node-fetch');

async function verifyIntegration() {
  try {
    console.log("Verifying Windsurf memory integration...");
    console.log("======================================");
    
    // Step 1: Test direct Memory Client API
    console.log("\n1. Testing direct Memory Client API:");
    const directResult = await testDirectAPI();
    
    // Step 2: Test proxy server
    console.log("\n2. Testing proxy server:");
    const proxyResult = await testProxyServer();
    
    // Step 3: Test conversation tagging
    console.log("\n3. Testing conversation tagging:");
    const tagResult = await testConversationTag();
    
    // Summary
    console.log("\n======================================");
    console.log("Integration verification summary:");
    console.log(`- Direct Memory Client API: ${directResult ? '✅ PASSED' : '❌ FAILED'}`);
    console.log(`- Proxy Server: ${proxyResult ? '✅ PASSED' : '❌ FAILED'}`);
    console.log(`- Conversation Tagging: ${tagResult ? '✅ PASSED' : '❌ FAILED'}`);
    
    if (directResult && proxyResult && tagResult) {
      console.log("\n✅ SUCCESS: All tests passed! The Windsurf memory integration is working correctly.");
    } else {
      console.log("\n❌ FAILURE: Some tests failed. Please check the logs for details.");
    }
  } catch (error) {
    console.error('Error during verification:', error);
  }
}

async function testDirectAPI() {
  try {
    console.log("  Testing direct message to Memory Client API...");
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
    console.log(`  Status: ${response.status} ${response.statusText}`);
    console.log(`  Response: ${responseText}`);
    
    if (response.ok) {
      console.log('  ✅ Direct API test successful');
      return true;
    } else {
      console.error('  ❌ Direct API test failed');
      return false;
    }
  } catch (error) {
    console.error('  ❌ Error testing direct API:', error.message);
    return false;
  }
}

async function testProxyServer() {
  try {
    console.log("  Testing message through proxy server...");
    const response = await fetch('http://localhost:10011/api/message', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        role: 'assistant',
        content: 'Test message through proxy server',
        id: generateUUID()
      })
    });
    
    const responseText = await response.text();
    console.log(`  Status: ${response.status} ${response.statusText}`);
    console.log(`  Response: ${responseText}`);
    
    if (response.ok) {
      console.log('  ✅ Proxy server test successful');
      return true;
    } else {
      console.error('  ❌ Proxy server test failed');
      return false;
    }
  } catch (error) {
    console.error('  ❌ Error testing proxy server:', error.message);
    return false;
  }
}

async function testConversationTag() {
  try {
    // Set a test tag
    console.log("  Setting conversation tag...");
    const tagName = "test-" + Date.now();
    
    const setResponse = await fetch('http://localhost:10011/api/set-conversation-tag', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        tag: tagName
      })
    });
    
    const setResponseText = await setResponse.text();
    console.log(`  Set tag status: ${setResponse.status} ${setResponse.statusText}`);
    console.log(`  Set tag response: ${setResponseText}`);
    
    if (!setResponse.ok) {
      console.error('  ❌ Failed to set conversation tag');
      return false;
    }
    
    // Get the tag to verify
    console.log("  Getting conversation tag...");
    const getResponse = await fetch('http://localhost:10011/api/get-conversation-tag', {
      method: 'GET'
    });
    
    const getResponseText = await getResponse.text();
    console.log(`  Get tag status: ${getResponse.status} ${getResponse.statusText}`);
    console.log(`  Get tag response: ${getResponseText}`);
    
    if (getResponse.ok) {
      try {
        const tagData = JSON.parse(getResponseText);
        if (tagData.tag === tagName) {
          console.log(`  ✅ Conversation tag test successful (tag: ${tagName})`);
          return true;
        } else {
          console.error(`  ❌ Tag mismatch: expected "${tagName}", got "${tagData.tag}"`);
          return false;
        }
      } catch (e) {
        console.error('  ❌ Failed to parse tag response');
        return false;
      }
    } else {
      console.error('  ❌ Failed to get conversation tag');
      return false;
    }
  } catch (error) {
    console.error('  ❌ Error testing conversation tag:', error.message);
    return false;
  }
}

// Generate a UUID v4
function generateUUID() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0;
    const v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

// Run the verification
verifyIntegration();
