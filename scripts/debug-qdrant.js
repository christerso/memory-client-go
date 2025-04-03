// Debug script for Qdrant integration
const fetch = require('node-fetch');

// Generate a proper UUID
function generateUUID() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0;
    const v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

async function testQdrantDirectly() {
  try {
    console.log("Testing Qdrant integration directly...");
    
    // Generate a UUID for the point
    const pointId = generateUUID();
    console.log(`Using point ID: ${pointId}`);
    
    // Create a test point
    const point = {
      id: pointId,
      vector: Array(384).fill(0).map(() => Math.random()), // Random 384-dim vector
      payload: {
        role: "user",
        content: "Test message for Qdrant",
        timestamp: new Date().toISOString(),
        metadata: {},
        tags: ["test"]
      }
    };
    
    // Create the request body
    const requestBody = {
      points: [point],
      ids: [pointId]
    };
    
    // Send the request directly to Qdrant
    console.log("Sending request to Qdrant...");
    const response = await fetch('http://localhost:6333/collections/conversation_memory/points', {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(requestBody)
    });
    
    const responseData = await response.text();
    console.log(`Status: ${response.status} ${response.statusText}`);
    console.log(`Response: ${responseData}`);
    
    if (response.ok) {
      console.log('✅ Successfully added point to Qdrant');
    } else {
      console.error('❌ Failed to add point to Qdrant');
    }
  } catch (error) {
    console.error('Error:', error.message);
  }
}

// Run the test
testQdrantDirectly();
