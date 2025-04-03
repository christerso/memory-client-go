// Simple script to test the memory client API

const fetch = require('node-fetch');

async function testMemoryAPI() {
  try {
    // Test the proxy server
    const response = await fetch('http://localhost:10011/api/message', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        role: 'user',
        content: 'Test message through proxy server'
      })
    });

    let responseText = '';
    try {
      responseText = await response.text();
      console.log(`Status: ${response.status} ${response.statusText}`);
      console.log(`Response: ${responseText}`);
    } catch (error) {
      console.error('Error reading response:', error);
    }
    
    if (response.ok) {
      console.log('Message sent successfully!');
    } else {
      console.error('Failed to send message');
    }
  } catch (error) {
    console.error('Error:', error.message);
  }
}

testMemoryAPI();
