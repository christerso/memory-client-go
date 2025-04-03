# Windsurf Memory Integration Manual Test Guide

This guide helps you manually test the Windsurf Memory Integration script to ensure it's properly capturing and vectorizing text from Windsurf conversations.

## Prerequisites

1. Memory Client API running on port 10010
2. Windsurf with the memory integration script loaded

## Test Cases

### 1. Basic Message Capture

**Objective**: Verify that messages typed in Windsurf are captured and sent to the Memory Client.

**Steps**:
1. Open Windsurf
2. Type a message in the chat (e.g., "This is a test message")
3. Check the Memory Client logs for incoming message requests
4. Verify the message content matches what was typed

**Expected Result**: The message appears in the Memory Client logs and is stored correctly.

### 2. Tag Management

**Objective**: Verify that conversation tags can be set and retrieved.

**Steps**:
1. Click on the tag display in the bottom right corner of Windsurf
2. Enter a new tag (e.g., "test-conversation")
3. Verify the tag display updates to show the new tag
4. Refresh Windsurf and verify the tag persists

**Expected Result**: The tag is set, displayed, and persists across refreshes.

### 3. Tagging Mode Toggle

**Objective**: Verify that tagging mode can be toggled between automatic and manual.

**Steps**:
1. Click on the mode display in the bottom right corner of Windsurf
2. Verify the mode toggles between "automatic" and "manual"
3. Check that the mode display updates accordingly

**Expected Result**: The tagging mode toggles and the display updates correctly.

### 4. Error Handling

**Objective**: Verify that the script handles errors gracefully.

**Steps**:
1. Stop the Memory Client API service
2. Type a message in Windsurf
3. Verify the status indicator turns red
4. Restart the Memory Client API service
5. Type another message
6. Verify the status indicator turns green

**Expected Result**: The script handles the API being unavailable gracefully and recovers when it becomes available again.

### 5. Message Buffering

**Objective**: Verify that messages are buffered before being sent for vectorization.

**Steps**:
1. Set the tagging mode to "automatic"
2. Type 5 short messages in quick succession
3. Check the Memory Client logs for categorization requests

**Expected Result**: After the 5th message, a categorization request should be sent to the Memory Client.

## Verification Tools

### API Endpoint Testing

You can use curl to test the API endpoints directly:

```bash
# Get current tag
curl http://localhost:10010/api/get-conversation-tag

# Set a tag
curl -X POST -H "Content-Type: application/json" -d '{"tag":"test-tag"}' http://localhost:10010/api/set-conversation-tag

# Send a message
curl -X POST -H "Content-Type: application/json" -d '{"role":"user","content":"test message"}' http://localhost:10010/api/message
```

### Monitoring Memory Client Logs

Monitor the Memory Client logs to see incoming requests:

```bash
# If running as a service
sc query MemoryClientMCPService

# Check logs
tail -f /path/to/memory-client/logs
```

## Troubleshooting

If the integration is not working as expected:

1. Check browser console for JavaScript errors
2. Verify the Memory Client API is running on port 10010
3. Check network requests in browser developer tools
4. Ensure the script is properly loaded in Windsurf
5. Verify the DOM structure matches what the script expects
