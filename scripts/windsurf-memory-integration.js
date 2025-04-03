// Windsurf Memory Integration Script
// This script integrates Windsurf with the memory client

(function() {
    // Configuration
    const config = {
        serverUrl: 'http://localhost:10011', // Updated to use the proxy server
        autoCapture: true,
        bufferSize: 5,
        debugMode: false,
        defaultTaggingMode: 'automatic' // Can be 'automatic' or 'manual'
    };

    // State
    let currentTag = '';
    let currentTaggingMode = '';
    let messageBuffer = [];
    let initialized = false;

    // Initialize the integration
    function initialize() {
        if (initialized) return;
        
        log('Initializing Windsurf Memory Integration');
        
        // Check if we're in Windsurf
        if (typeof windsurf === 'undefined') {
            log('Not running in Windsurf environment', 'error');
            return;
        }
        
        // Set up message capture
        setupMessageCapture();
        
        // Add UI elements
        addUIElements();
        
        // Get current tag
        getCurrentTag().then(tag => {
            currentTag = tag;
            updateTagDisplay();
        });
        
        // Get current tagging mode
        getTaggingMode().then(mode => {
            currentTaggingMode = mode;
            updateTaggingModeDisplay();
        });
        
        initialized = true;
        log('Windsurf Memory Integration initialized');
    }

    // Set up message capture
    function setupMessageCapture() {
        // This is a placeholder for the actual Windsurf API
        // The actual implementation would depend on Windsurf's specific API
        if (windsurf.chat && windsurf.chat.onMessage) {
            windsurf.chat.onMessage(message => {
                if (config.autoCapture) {
                    sendMessage(message.role, message.content);
                }
            });
            log('Message capture set up');
        } else {
            // Alternative approach using DOM mutation observer
            log('Using DOM observer for message capture');
            setupDOMObserver();
        }
    }

    // Set up DOM observer as a fallback method
    function setupDOMObserver() {
        // This is a generic approach that might work with different chat interfaces
        const observer = new MutationObserver(mutations => {
            for (const mutation of mutations) {
                if (mutation.type === 'childList' && mutation.addedNodes.length > 0) {
                    for (const node of mutation.addedNodes) {
                        if (node.nodeType === Node.ELEMENT_NODE) {
                            // Look for new message elements
                            // This selector would need to be adjusted for Windsurf's specific DOM structure
                            const messageElements = node.querySelectorAll('.chat-message');
                            
                            for (const msgElement of messageElements) {
                                // Extract role and content from the message element
                                // Again, this would need to be adjusted for Windsurf's specific DOM structure
                                const role = msgElement.classList.contains('user-message') ? 'user' : 'assistant';
                                const content = msgElement.querySelector('.message-content')?.textContent;
                                
                                if (content && !msgElement.dataset.captured) {
                                    sendMessage(role, content);
                                    msgElement.dataset.captured = 'true';
                                }
                            }
                        }
                    }
                }
            }
        });
        
        // Start observing the chat container
        // The selector would need to be adjusted for Windsurf's specific DOM structure
        const chatContainer = document.querySelector('#chat-container');
        if (chatContainer) {
            observer.observe(chatContainer, { childList: true, subtree: true });
            log('DOM observer started');
        } else {
            log('Chat container not found', 'error');
        }
    }

    // Add UI elements
    function addUIElements() {
        // Create tag display element
        const tagDisplay = document.createElement('div');
        tagDisplay.id = 'memory-tag-display';
        tagDisplay.style.cssText = 'position: fixed; bottom: 10px; right: 10px; background: rgba(0,0,0,0.7); color: white; padding: 5px 10px; border-radius: 4px; font-size: 12px; z-index: 1000; cursor: pointer;';
        tagDisplay.innerHTML = 'No tag set';
        tagDisplay.title = 'Click to set conversation tag';
        tagDisplay.onclick = promptForTag;
        document.body.appendChild(tagDisplay);
        
        // Create tagging mode display element
        const modeDisplay = document.createElement('div');
        modeDisplay.id = 'memory-mode-display';
        modeDisplay.style.cssText = 'position: fixed; bottom: 10px; right: 100px; background: rgba(0,0,0,0.7); color: white; padding: 5px 10px; border-radius: 4px; font-size: 12px; z-index: 1000; cursor: pointer;';
        modeDisplay.innerHTML = 'Mode: ?';
        modeDisplay.title = 'Click to toggle tagging mode (automatic/manual)';
        modeDisplay.onclick = toggleTaggingMode;
        document.body.appendChild(modeDisplay);
        
        // Create status indicator
        const statusIndicator = document.createElement('div');
        statusIndicator.id = 'memory-status-indicator';
        statusIndicator.style.cssText = 'position: fixed; bottom: 10px; right: 210px; width: 10px; height: 10px; background: #2ecc71; border-radius: 50%; z-index: 1000;';
        statusIndicator.title = 'Memory client connected';
        document.body.appendChild(statusIndicator);
        
        // Check connection status
        checkConnection().then(connected => {
            if (!connected) {
                statusIndicator.style.background = '#e74c3c';
                statusIndicator.title = 'Memory client disconnected';
            }
        });
        
        // Add command to Windsurf command palette if available
        if (windsurf.commands && windsurf.commands.register) {
            windsurf.commands.register('memory.setTag', promptForTag, 'Set Memory Conversation Tag');
            windsurf.commands.register('memory.captureMessage', promptForMessage, 'Capture Message Manually');
            windsurf.commands.register('memory.toggleTaggingMode', toggleTaggingMode, 'Toggle Tagging Mode (Auto/Manual)');
        }
    }

    // Update tag display
    function updateTagDisplay() {
        const display = document.getElementById('memory-tag-display');
        if (display) {
            display.innerHTML = currentTag ? `Tag: ${currentTag}` : 'No tag set';
        }
    }

    // Update tagging mode display
    function updateTaggingModeDisplay() {
        const display = document.getElementById('memory-mode-display');
        if (display) {
            display.innerHTML = `Mode: ${currentTaggingMode || '?'}`;
            
            // Update color based on mode
            if (currentTaggingMode === 'automatic') {
                display.style.background = 'rgba(46, 204, 113, 0.7)'; // Green for automatic
            } else if (currentTaggingMode === 'manual') {
                display.style.background = 'rgba(52, 152, 219, 0.7)'; // Blue for manual
            } else {
                display.style.background = 'rgba(0, 0, 0, 0.7)'; // Default black
            }
        }
    }

    // Prompt for tag
    function promptForTag() {
        // This would be replaced with Windsurf's native dialog if available
        const tag = prompt('Enter conversation tag:', currentTag);
        if (tag !== null) {
            setConversationTag(tag);
        }
    }

    // Prompt for message
    function promptForMessage() {
        // This would be replaced with Windsurf's native dialog if available
        const role = prompt('Enter message role (user/assistant):', 'user');
        if (role === null) return;
        
        const content = prompt('Enter message content:');
        if (content !== null) {
            sendMessage(role, content);
        }
    }

    // Toggle tagging mode
    function toggleTaggingMode() {
        const newMode = currentTaggingMode === 'automatic' ? 'manual' : 'automatic';
        setTaggingMode(newMode);
    }

    // API functions
    function sendMessage(role, content) {
        if (!config.serverUrl) {
            log('No server URL configured', 'error');
            return Promise.reject(new Error('No server URL configured'));
        }

        // Create a unique ID for the message
        const messageId = generateUUID();
        
        // Create the message object
        const message = {
            role: role,
            content: content,
            id: messageId,
            timestamp: new Date().toISOString()
        };
        
        // Send the message to the server
        return fetch(`${config.serverUrl}/api/message`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(message)
        })
        .then(response => {
            if (!response.ok) {
                return response.text().then(text => {
                    throw new Error(`Failed to send message: ${text}`);
                });
            }
            return response.json();
        })
        .then(data => {
            log(`Message sent successfully`, data);
            return data;
        })
        .catch(error => {
            log(`Error sending message: ${error.message}`, 'error');
            throw error;
        });
    }

    // Generate a UUID v4
    function generateUUID() {
        return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
            const r = Math.random() * 16 | 0;
            const v = c === 'x' ? r : (r & 0x3 | 0x8);
            return v.toString(16);
        });
    }

    async function setConversationTag(tag) {
        try {
            const response = await fetch(`${config.serverUrl}/api/set-conversation-tag`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ tag })
            });

            if (response.ok) {
                currentTag = tag;
                updateTagDisplay();
                log(`Conversation tag set to: ${tag}`);
                flashStatus('#2ecc71'); // Green flash for success
                return true;
            } else {
                const errorText = await response.text();
                log(`Failed to set conversation tag: ${errorText}`, 'error');
                flashStatus('#e74c3c'); // Red flash for error
                return false;
            }
        } catch (error) {
            log(`Error connecting to memory server: ${error.message}`, 'error');
            flashStatus('#e74c3c'); // Red flash for error
            return false;
        }
    }

    async function getCurrentTag() {
        try {
            const response = await fetch(`${config.serverUrl}/api/get-conversation-tag`);
            
            if (response.ok) {
                const data = await response.json();
                log(`Retrieved current tag: ${data.tag || 'none'}`);
                return data.tag;
            } else {
                const errorText = await response.text();
                log(`Failed to get conversation tag: ${errorText}`, 'error');
                return '';
            }
        } catch (error) {
            log(`Error connecting to memory server: ${error.message}`, 'error');
            return '';
        }
    }

    async function setTaggingMode(mode) {
        try {
            const response = await fetch(`${config.serverUrl}/api/set-tagging-mode`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ mode })
            });

            if (response.ok) {
                currentTaggingMode = mode;
                updateTaggingModeDisplay();
                log(`Tagging mode set to: ${mode}`);
                flashStatus('#2ecc71'); // Green flash for success
                return true;
            } else {
                const errorText = await response.text();
                log(`Failed to set tagging mode: ${errorText}`, 'error');
                flashStatus('#e74c3c'); // Red flash for error
                return false;
            }
        } catch (error) {
            log(`Error connecting to memory server: ${error.message}`, 'error');
            flashStatus('#e74c3c'); // Red flash for error
            return false;
        }
    }

    async function getTaggingMode() {
        try {
            const response = await fetch(`${config.serverUrl}/api/get-tagging-mode`);
            
            if (response.ok) {
                const data = await response.json();
                log(`Retrieved current tagging mode: ${data.mode || 'none'}`);
                return data.mode;
            } else {
                const errorText = await response.text();
                log(`Failed to get tagging mode: ${errorText}`, 'error');
                return config.defaultTaggingMode;
            }
        } catch (error) {
            log(`Error connecting to memory server: ${error.message}`, 'error');
            return config.defaultTaggingMode;
        }
    }

    async function checkConnection() {
        try {
            const response = await fetch(`${config.serverUrl}/health`, { method: 'GET' });
            return response.ok;
        } catch (error) {
            log(`Memory server connection check failed: ${error.message}`, 'error');
            return false;
        }
    }

    // Helper functions
    function log(message, level = 'info') {
        if (config.debugMode || level === 'error') {
            console[level === 'error' ? 'error' : 'log'](`[Memory Integration] ${message}`);
        }
    }

    function truncate(str, maxLen) {
        if (str.length <= maxLen) return str;
        return str.substring(0, maxLen) + '...';
    }

    function flashStatus(color) {
        const indicator = document.getElementById('memory-status-indicator');
        if (!indicator) return;
        
        const originalColor = indicator.style.background;
        indicator.style.background = color;
        
        setTimeout(() => {
            indicator.style.background = originalColor;
        }, 500);
    }

    // Initialize when the document is ready
    if (document.readyState === 'complete') {
        initialize();
    } else {
        window.addEventListener('load', initialize);
    }

    // Expose API for external use
    window.windsurfMemory = {
        sendMessage,
        setConversationTag,
        getCurrentTag,
        setTaggingMode,
        getTaggingMode,
        toggleTaggingMode,
        config
    };
})();
