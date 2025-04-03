// Test script for the Memory Client API
// This script tests sending messages directly to the Memory Client API

const http = require('http');

// Configuration
const config = {
  apiUrl: 'http://localhost:10010',
  testMessage: {
    role: 'user',
    content: 'This is a test message from the test script'
  }
};

// Function to send a message to the Memory Client API
function sendMessage(message) {
  return new Promise((resolve, reject) => {
    // Prepare the request data
    const data = JSON.stringify(message);
    
    // Set up the request options
    const options = {
      hostname: 'localhost',
      port: 10010,
      path: '/api/message',
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': data.length
      }
    };
    
    // Make the request
    const req = http.request(options, (res) => {
      let responseData = '';
      
      // Collect the response data
      res.on('data', (chunk) => {
        responseData += chunk;
      });
      
      // Process the complete response
      res.on('end', () => {
        console.log(`Status: ${res.statusCode} ${res.statusMessage}`);
        console.log(`Response: ${responseData}`);
        
        if (res.statusCode === 200) {
          console.log('Message sent successfully!');
          resolve(responseData);
        } else {
          console.error('Failed to send message');
          reject(new Error(`HTTP ${res.statusCode}: ${responseData}`));
        }
      });
    });
    
    // Handle request errors
    req.on('error', (error) => {
      console.error('Error sending message:', error.message);
      reject(error);
    });
    
    // Send the request data
    req.write(data);
    req.end();
  });
}

// Main test function
async function runTest() {
  console.log('Testing Memory Client API...');
  console.log(`Sending message to ${config.apiUrl}/api/message`);
  console.log('Message:', config.testMessage);
  
  try {
    const result = await sendMessage(config.testMessage);
    console.log('Test completed successfully');
    return true;
  } catch (error) {
    console.error('Test failed:', error.message);
    return false;
  }
}

// Run the test
runTest();
