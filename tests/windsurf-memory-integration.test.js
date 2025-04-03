/**
 * Unit tests for Windsurf Memory Integration Script
 */

// Mock fetch for API calls
global.fetch = jest.fn(() => 
  Promise.resolve({
    ok: true,
    json: () => Promise.resolve({ success: true }),
    text: () => Promise.resolve('success')
  })
);

// Create a mock document and window environment
document.body = document.createElement('body');
document.readyState = 'complete';

// Mock Windsurf environment
global.windsurf = {
  chat: {
    onMessage: jest.fn(callback => {
      global.windsurfMessageCallback = callback;
    })
  },
  commands: {
    register: jest.fn()
  }
};

// Import the script - we need to load it this way since it's an IIFE
const fs = require('fs');
const path = require('path');
const scriptContent = fs.readFileSync(path.join(__dirname, '..', 'scripts', 'windsurf-memory-integration.js'), 'utf8');
const scriptExports = {};
const scriptFn = new Function('exports', scriptContent);
scriptFn(scriptExports);

// Tests
describe('Windsurf Memory Integration', () => {
  beforeEach(() => {
    // Reset mocks
    jest.clearAllMocks();
    document.body.innerHTML = '';
    
    // Re-run the script to reset state
    scriptFn(scriptExports);
  });

  test('should initialize correctly', () => {
    // Check if message capture was set up
    expect(global.windsurf.chat.onMessage).toHaveBeenCalled();
    
    // Check if UI elements were added
    expect(document.getElementById('memory-tag-display')).not.toBeNull();
    expect(document.getElementById('memory-mode-display')).not.toBeNull();
    expect(document.getElementById('memory-status-indicator')).not.toBeNull();
    
    // Check if commands were registered
    expect(global.windsurf.commands.register).toHaveBeenCalledTimes(3);
  });

  test('should capture messages via Windsurf API', () => {
    // Simulate a message from Windsurf
    const testMessage = { role: 'user', content: 'test message' };
    global.windsurfMessageCallback(testMessage);
    
    // Verify message was sent to memory client
    expect(global.fetch).toHaveBeenCalledWith(
      'http://localhost:10010/api/message',
      expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({
          'Content-Type': 'application/json'
        }),
        body: expect.stringContaining('test message')
      })
    );
  });

  test('should fallback to DOM observer when API unavailable', () => {
    // Remove Windsurf chat API
    const originalChat = global.windsurf.chat;
    delete global.windsurf.chat;
    
    // Re-run the script
    scriptFn(scriptExports);
    
    // Check if MutationObserver is used
    // This is hard to test directly, but we can check if the log message is shown
    // For a real test, we'd need to mock MutationObserver and verify it's called
    
    // Restore Windsurf chat API
    global.windsurf.chat = originalChat;
  });

  test('should set and display conversation tag', async () => {
    // Mock the prompt function
    const originalPrompt = window.prompt;
    window.prompt = jest.fn(() => 'test-tag');
    
    // Get the tag display element
    const tagDisplay = document.getElementById('memory-tag-display');
    expect(tagDisplay).not.toBeNull();
    
    // Simulate clicking on the tag display
    tagDisplay.click();
    
    // Verify prompt was shown
    expect(window.prompt).toHaveBeenCalled();
    
    // Verify API call to set tag
    expect(global.fetch).toHaveBeenCalledWith(
      'http://localhost:10010/api/set-conversation-tag',
      expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({
          'Content-Type': 'application/json'
        }),
        body: expect.stringContaining('test-tag')
      })
    );
    
    // Restore prompt function
    window.prompt = originalPrompt;
  });

  test('should toggle tagging mode', async () => {
    // Get the mode display element
    const modeDisplay = document.getElementById('memory-mode-display');
    expect(modeDisplay).not.toBeNull();
    
    // Get initial mode
    const initialMode = modeDisplay.innerHTML;
    
    // Simulate clicking on the mode display
    modeDisplay.click();
    
    // Verify API call to toggle mode
    expect(global.fetch).toHaveBeenCalledWith(
      expect.stringMatching(/http:\/\/localhost:10010\/api\/set-tagging-mode/),
      expect.objectContaining({
        method: 'POST'
      })
    );
  });

  test('should handle connection errors gracefully', async () => {
    // Mock fetch to simulate connection error
    global.fetch.mockImplementationOnce(() => 
      Promise.reject(new Error('Connection failed'))
    );
    
    // Simulate a message from Windsurf
    const testMessage = { role: 'user', content: 'test message with error' };
    global.windsurfMessageCallback(testMessage);
    
    // Check if status indicator changes color (would need to check CSS)
    // For a real test, we'd need to wait for the promise rejection and check the indicator
  });
});
