// Windsurf Standalone Memory Integration
// This script provides memory functionality directly in the browser without requiring the Memory Client API

(function() {
    // Configuration
    const config = {
        autoCapture: true,
        bufferSize: 5,
        debugMode: false,
        storageKey: 'windsurf-memory',
        maxMessages: 1000 // Maximum number of messages to store
    };

    // State
    let messageStore = [];
    let currentTag = '';
    let taggingMode = 'automatic';
    let initialized = false;

    // Initialize the integration
    function initialize() {
        if (initialized) return;
        
        console.log('Initializing Windsurf Standalone Memory Integration');
        
        // Check if we're in Windsurf
        if (typeof windsurf === 'undefined') {
            console.log('Not running in Windsurf environment');
            return;
        }
        
        // Load stored messages
        loadFromLocalStorage();
        
        // Set up message capture
        setupMessageCapture();
        
        // Add UI elements
        addUIElements();
        
        initialized = true;
        console.log('Windsurf Standalone Memory Integration initialized');
        console.log(`Loaded ${messageStore.length} messages from storage`);
    }

    // Set up message capture
    function setupMessageCapture() {
        // This is a placeholder for the actual Windsurf API
        if (windsurf.chat && windsurf.chat.onMessage) {
            windsurf.chat.onMessage(message => {
                if (config.autoCapture) {
                    storeMessage(message.role, message.content);
                }
            });
            console.log('Message capture set up');
        } else {
            // Alternative approach using DOM mutation observer
            console.log('Using DOM observer for message capture');
            setupDOMObserver();
        }
    }

    // Set up DOM observer as a fallback method
    function setupDOMObserver() {
        const observer = new MutationObserver(mutations => {
            for (const mutation of mutations) {
                if (mutation.type === 'childList' && mutation.addedNodes.length > 0) {
                    for (const node of mutation.addedNodes) {
                        if (node.nodeType === Node.ELEMENT_NODE) {
                            // Look for new message elements
                            const messageElements = node.querySelectorAll('.chat-message');
                            
                            for (const msgElement of messageElements) {
                                // Extract role and content from the message element
                                const role = msgElement.classList.contains('user-message') ? 'user' : 'assistant';
                                const content = msgElement.querySelector('.message-content')?.textContent;
                                
                                if (content && !msgElement.dataset.captured) {
                                    storeMessage(role, content);
                                    msgElement.dataset.captured = 'true';
                                }
                            }
                        }
                    }
                }
            }
        });
        
        // Start observing the chat container
        const chatContainer = document.querySelector('#chat-container');
        if (chatContainer) {
            observer.observe(chatContainer, { childList: true, subtree: true });
            console.log('DOM observer started');
        } else {
            console.log('Chat container not found');
        }
    }

    // Add UI elements
    function addUIElements() {
        // Create tag display element
        const tagDisplay = document.createElement('div');
        tagDisplay.id = 'memory-tag-display';
        tagDisplay.style.cssText = 'position: fixed; bottom: 10px; right: 10px; background: rgba(0,0,0,0.7); color: white; padding: 5px 10px; border-radius: 4px; font-size: 12px; z-index: 1000; cursor: pointer;';
        tagDisplay.innerHTML = currentTag ? `Tag: ${currentTag}` : 'No tag set';
        tagDisplay.title = 'Click to set conversation tag';
        tagDisplay.onclick = promptForTag;
        document.body.appendChild(tagDisplay);
        
        // Create tagging mode display element
        const modeDisplay = document.createElement('div');
        modeDisplay.id = 'memory-mode-display';
        modeDisplay.style.cssText = 'position: fixed; bottom: 10px; right: 100px; background: rgba(0,0,0,0.7); color: white; padding: 5px 10px; border-radius: 4px; font-size: 12px; z-index: 1000; cursor: pointer;';
        modeDisplay.innerHTML = `Mode: ${taggingMode}`;
        modeDisplay.title = 'Click to toggle tagging mode (automatic/manual)';
        modeDisplay.onclick = toggleTaggingMode;
        document.body.appendChild(modeDisplay);
        
        // Create message count indicator
        const countIndicator = document.createElement('div');
        countIndicator.id = 'memory-count-indicator';
        countIndicator.style.cssText = 'position: fixed; bottom: 10px; right: 210px; background: rgba(0,0,0,0.7); color: white; padding: 5px 10px; border-radius: 4px; font-size: 12px; z-index: 1000; cursor: pointer;';
        countIndicator.innerHTML = `Messages: ${messageStore.length}`;
        countIndicator.title = 'Click to view stored messages';
        countIndicator.onclick = showStoredMessages;
        document.body.appendChild(countIndicator);
        
        // Add command to Windsurf command palette if available
        if (windsurf.commands && windsurf.commands.register) {
            windsurf.commands.register('memory.setTag', promptForTag, 'Set Memory Conversation Tag');
            windsurf.commands.register('memory.captureMessage', promptForMessage, 'Capture Message Manually');
            windsurf.commands.register('memory.toggleTaggingMode', toggleTaggingMode, 'Toggle Tagging Mode (Auto/Manual)');
            windsurf.commands.register('memory.viewMessages', showStoredMessages, 'View Stored Messages');
            windsurf.commands.register('memory.exportMessages', exportMessages, 'Export Stored Messages');
            windsurf.commands.register('memory.clearMessages', clearMessages, 'Clear Stored Messages');
        }
    }

    // Store a message
    function storeMessage(role, content) {
        const message = {
            id: generateId(),
            role: role,
            content: content,
            timestamp: new Date().toISOString(),
            tags: currentTag ? [currentTag] : []
        };
        
        // Add to message store
        messageStore.push(message);
        
        // Trim if exceeding max size
        if (messageStore.length > config.maxMessages) {
            messageStore.shift();
        }
        
        // Save to local storage
        saveToLocalStorage();
        
        // Update message count display
        updateMessageCount();
        
        // Analyze and tag if in automatic mode
        if (taggingMode === 'automatic' && messageStore.length % config.bufferSize === 0) {
            analyzeAndTagMessages();
        }
        
        return message;
    }

    // Generate a unique ID
    function generateId() {
        return Date.now().toString(36) + Math.random().toString(36).substr(2, 5);
    }

    // Save messages to local storage
    function saveToLocalStorage() {
        try {
            localStorage.setItem(config.storageKey, JSON.stringify({
                messages: messageStore,
                currentTag: currentTag,
                taggingMode: taggingMode
            }));
        } catch (error) {
            console.error('Error saving to local storage:', error);
        }
    }

    // Load messages from local storage
    function loadFromLocalStorage() {
        try {
            const data = localStorage.getItem(config.storageKey);
            if (data) {
                const parsed = JSON.parse(data);
                messageStore = parsed.messages || [];
                currentTag = parsed.currentTag || '';
                taggingMode = parsed.taggingMode || 'automatic';
            }
        } catch (error) {
            console.error('Error loading from local storage:', error);
        }
    }

    // Update message count display
    function updateMessageCount() {
        const countIndicator = document.getElementById('memory-count-indicator');
        if (countIndicator) {
            countIndicator.innerHTML = `Messages: ${messageStore.length}`;
        }
    }

    // Prompt for tag
    function promptForTag() {
        const tag = prompt('Enter conversation tag:', currentTag);
        if (tag !== null) {
            currentTag = tag;
            updateTagDisplay();
            saveToLocalStorage();
        }
    }

    // Update tag display
    function updateTagDisplay() {
        const display = document.getElementById('memory-tag-display');
        if (display) {
            display.innerHTML = currentTag ? `Tag: ${currentTag}` : 'No tag set';
        }
    }

    // Toggle tagging mode
    function toggleTaggingMode() {
        taggingMode = taggingMode === 'automatic' ? 'manual' : 'automatic';
        updateTaggingModeDisplay();
        saveToLocalStorage();
    }

    // Update tagging mode display
    function updateTaggingModeDisplay() {
        const display = document.getElementById('memory-mode-display');
        if (display) {
            display.innerHTML = `Mode: ${taggingMode}`;
            
            // Update color based on mode
            if (taggingMode === 'automatic') {
                display.style.background = 'rgba(46, 204, 113, 0.7)'; // Green for automatic
            } else {
                display.style.background = 'rgba(52, 152, 219, 0.7)'; // Blue for manual
            }
        }
    }

    // Prompt for message
    function promptForMessage() {
        const role = prompt('Enter message role (user/assistant):', 'user');
        if (role === null) return;
        
        const content = prompt('Enter message content:');
        if (content !== null) {
            storeMessage(role, content);
        }
    }

    // Show stored messages
    function showStoredMessages() {
        // Create a modal to display messages
        const modal = document.createElement('div');
        modal.style.cssText = 'position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.8); z-index: 2000; display: flex; justify-content: center; align-items: center;';
        
        // Create modal content
        const content = document.createElement('div');
        content.style.cssText = 'background: #1e1e1e; color: #e0e0e0; width: 80%; height: 80%; padding: 20px; border-radius: 8px; overflow: auto; position: relative;';
        
        // Add close button
        const closeButton = document.createElement('button');
        closeButton.innerHTML = '×';
        closeButton.style.cssText = 'position: absolute; top: 10px; right: 10px; background: none; border: none; color: #e0e0e0; font-size: 24px; cursor: pointer;';
        closeButton.onclick = () => document.body.removeChild(modal);
        content.appendChild(closeButton);
        
        // Add title
        const title = document.createElement('h2');
        title.innerHTML = 'Stored Messages';
        title.style.cssText = 'margin-top: 0; color: #4dabf7;';
        content.appendChild(title);
        
        // Add tag info
        const tagInfo = document.createElement('p');
        tagInfo.innerHTML = `Current tag: <strong>${currentTag || 'None'}</strong> | Mode: <strong>${taggingMode}</strong> | Total: <strong>${messageStore.length}</strong> messages`;
        content.appendChild(tagInfo);
        
        // Add export button
        const exportButton = document.createElement('button');
        exportButton.innerHTML = 'Export Messages';
        exportButton.style.cssText = 'background: #4dabf7; color: white; border: none; padding: 8px 16px; border-radius: 4px; margin-right: 10px; cursor: pointer;';
        exportButton.onclick = exportMessages;
        content.appendChild(exportButton);
        
        // Add clear button
        const clearButton = document.createElement('button');
        clearButton.innerHTML = 'Clear All Messages';
        clearButton.style.cssText = 'background: #e74c3c; color: white; border: none; padding: 8px 16px; border-radius: 4px; cursor: pointer;';
        clearButton.onclick = () => {
            if (confirm('Are you sure you want to clear all stored messages?')) {
                clearMessages();
                document.body.removeChild(modal);
            }
        };
        content.appendChild(clearButton);
        
        // Add message list
        const messageList = document.createElement('div');
        messageList.style.cssText = 'margin-top: 20px; border-top: 1px solid #333; padding-top: 10px;';
        
        // Add messages
        messageStore.slice().reverse().forEach(message => {
            const messageItem = document.createElement('div');
            messageItem.style.cssText = 'margin-bottom: 15px; padding: 10px; border-radius: 4px; background: #2a2a2a;';
            
            const header = document.createElement('div');
            header.style.cssText = 'display: flex; justify-content: space-between; margin-bottom: 5px;';
            
            const role = document.createElement('span');
            role.innerHTML = `<strong>${message.role}</strong>`;
            role.style.color = message.role === 'user' ? '#4dabf7' : '#2ecc71';
            header.appendChild(role);
            
            const time = document.createElement('span');
            time.innerHTML = new Date(message.timestamp).toLocaleString();
            time.style.color = '#888';
            header.appendChild(time);
            
            messageItem.appendChild(header);
            
            const content = document.createElement('div');
            content.innerHTML = message.content;
            content.style.whiteSpace = 'pre-wrap';
            messageItem.appendChild(content);
            
            if (message.tags && message.tags.length > 0) {
                const tags = document.createElement('div');
                tags.style.cssText = 'margin-top: 5px;';
                tags.innerHTML = `Tags: ${message.tags.join(', ')}`;
                tags.style.color = '#888';
                messageItem.appendChild(tags);
            }
            
            messageList.appendChild(messageItem);
        });
        
        content.appendChild(messageList);
        modal.appendChild(content);
        document.body.appendChild(modal);
    }

    // Export messages
    function exportMessages() {
        const data = JSON.stringify({
            messages: messageStore,
            currentTag: currentTag,
            taggingMode: taggingMode,
            exportDate: new Date().toISOString()
        }, null, 2);
        
        const blob = new Blob([data], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        
        const a = document.createElement('a');
        a.href = url;
        a.download = `windsurf-memory-export-${new Date().toISOString().split('T')[0]}.json`;
        a.click();
        
        URL.revokeObjectURL(url);
    }

    // Clear messages
    function clearMessages() {
        messageStore = [];
        saveToLocalStorage();
        updateMessageCount();
    }

    // Analyze and tag messages
    function analyzeAndTagMessages() {
        // Simple keyword-based categorization
        const keywords = {
            technical: ['code', 'function', 'class', 'method', 'variable', 'bug', 'error', 'fix', 'implement', 'feature', 'api', 'database', 'server', 'client', 'test'],
            planning: ['plan', 'schedule', 'timeline', 'milestone', 'goal', 'objective', 'strategy', 'roadmap', 'priority', 'backlog', 'sprint', 'task', 'project'],
            question: ['what', 'how', 'why', 'when', 'where', 'who', 'which', 'can you', 'could you', 'would you', '?'],
            feedback: ['feedback', 'review', 'improve', 'suggestion', 'opinion', 'think', 'feel', 'like', 'dislike', 'prefer']
        };
        
        // Get last few messages
        const recentMessages = messageStore.slice(-config.bufferSize);
        
        // Combine content for analysis
        const combinedContent = recentMessages.map(m => m.content).join(' ').toLowerCase();
        
        // Determine category
        let bestCategory = '';
        let bestScore = 0;
        
        for (const [category, words] of Object.entries(keywords)) {
            const score = words.reduce((count, word) => {
                const regex = new RegExp(`\\b${word}\\b`, 'gi');
                const matches = combinedContent.match(regex);
                return count + (matches ? matches.length : 0);
            }, 0);
            
            if (score > bestScore) {
                bestScore = score;
                bestCategory = category;
            }
        }
        
        // Apply tag if category found
        if (bestCategory && bestScore > 0) {
            // Add category tag to recent messages
            recentMessages.forEach(message => {
                if (!message.tags.includes(bestCategory)) {
                    message.tags.push(bestCategory);
                }
            });
            
            // Save changes
            saveToLocalStorage();
            
            console.log(`Auto-tagged ${recentMessages.length} messages as "${bestCategory}"`);
        }
    }

    // Initialize when the document is ready
    if (document.readyState === 'complete') {
        initialize();
    } else {
        document.addEventListener('DOMContentLoaded', initialize);
    }

    // Expose API for external use
    window.windsurfMemory = {
        storeMessage,
        setTag: (tag) => {
            currentTag = tag;
            updateTagDisplay();
            saveToLocalStorage();
        },
        getTag: () => currentTag,
        setTaggingMode: (mode) => {
            if (mode === 'automatic' || mode === 'manual') {
                taggingMode = mode;
                updateTaggingModeDisplay();
                saveToLocalStorage();
            }
        },
        getTaggingMode: () => taggingMode,
        getMessages: () => [...messageStore],
        clearMessages,
        exportMessages
    };
})();
