const vscode = require('vscode');
const fetch = require('node-fetch');

// Global variables
let currentConversationTag = '';
let messageBuffer = [];
const MESSAGE_BUFFER_SIZE = 5;

/**
 * @param {vscode.ExtensionContext} context
 */
function activate(context) {
    console.log('VS Code Memory Extension is now active');

    // Register commands
    let setTagCommand = vscode.commands.registerCommand('vscode-memory.setTag', async () => {
        const tag = await vscode.window.showInputBox({
            placeHolder: 'Enter conversation tag',
            prompt: 'Set the current conversation tag'
        });

        if (tag) {
            setConversationTag(tag);
        }
    });

    let getTagCommand = vscode.commands.registerCommand('vscode-memory.getTag', async () => {
        const tag = await getCurrentTag();
        if (tag) {
            vscode.window.showInformationMessage(`Current conversation tag: ${tag}`);
        } else {
            vscode.window.showInformationMessage('No conversation tag is currently set');
        }
    });

    let captureMessageCommand = vscode.commands.registerCommand('vscode-memory.captureMessage', async () => {
        const role = await vscode.window.showQuickPick(['user', 'assistant'], {
            placeHolder: 'Select message role'
        });

        if (!role) return;

        const content = await vscode.window.showInputBox({
            placeHolder: 'Enter message content',
            prompt: 'Message to capture'
        });

        if (content) {
            sendMessage(role, content);
        }
    });

    // Register event listeners for chat messages
    if (vscode.chat) {
        // This will work in VS Code with chat features
        const chatListener = vscode.chat.onDidReceiveMessage(async (message) => {
            if (getConfig('autoCapture')) {
                // Determine the role based on the participant (user or assistant)
                const role = message.participant.name === 'user' ? 'user' : 'assistant';
                sendMessage(role, message.text);
            }
        });
        context.subscriptions.push(chatListener);
    }

    // For Windsurf, we need a different approach since it has a different API
    // This is a placeholder for Windsurf-specific integration
    try {
        // Check if we're in Windsurf
        if (typeof windsurf !== 'undefined') {
            // Windsurf-specific code would go here
            console.log('Windsurf detected, setting up conversation capture');
            
            // This is hypothetical - actual implementation would depend on Windsurf's API
            windsurf.chat.onMessage((message) => {
                if (getConfig('autoCapture')) {
                    sendMessage(message.role, message.content);
                }
            });
        }
    } catch (error) {
        // Not in Windsurf, that's fine
    }

    context.subscriptions.push(setTagCommand);
    context.subscriptions.push(getTagCommand);
    context.subscriptions.push(captureMessageCommand);

    // Status bar item to show current tag
    const statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Left, 100);
    statusBarItem.command = 'vscode-memory.setTag';
    updateStatusBar(statusBarItem);
    statusBarItem.show();
    
    // Update status bar periodically
    setInterval(() => updateStatusBar(statusBarItem), 30000);
    
    context.subscriptions.push(statusBarItem);
}

function deactivate() {
    // Clean up resources
}

// Helper functions
function getConfig(key) {
    return vscode.workspace.getConfiguration('vscode-memory').get(key);
}

function getServerUrl() {
    return getConfig('serverUrl') || 'http://localhost:10010';
}

async function setConversationTag(tag) {
    try {
        const response = await fetch(`${getServerUrl()}/api/set-conversation-tag`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ tag })
        });

        if (response.ok) {
            currentConversationTag = tag;
            vscode.window.showInformationMessage(`Conversation tag set to: ${tag}`);
            return true;
        } else {
            const errorText = await response.text();
            vscode.window.showErrorMessage(`Failed to set conversation tag: ${errorText}`);
            return false;
        }
    } catch (error) {
        vscode.window.showErrorMessage(`Error connecting to memory server: ${error.message}`);
        return false;
    }
}

async function getCurrentTag() {
    try {
        const response = await fetch(`${getServerUrl()}/api/get-conversation-tag`);
        
        if (response.ok) {
            const data = await response.json();
            currentConversationTag = data.tag;
            return data.tag;
        } else {
            const errorText = await response.text();
            vscode.window.showErrorMessage(`Failed to get conversation tag: ${errorText}`);
            return null;
        }
    } catch (error) {
        vscode.window.showErrorMessage(`Error connecting to memory server: ${error.message}`);
        return null;
    }
}

async function sendMessage(role, content) {
    try {
        const response = await fetch(`${getServerUrl()}/api/message`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ role, content })
        });

        if (response.ok) {
            // Add to local buffer for UI feedback
            messageBuffer.push({ role, content, timestamp: new Date() });
            if (messageBuffer.length > MESSAGE_BUFFER_SIZE) {
                messageBuffer.shift(); // Remove oldest message
            }
            
            // Show subtle notification
            const truncatedContent = content.length > 30 ? content.substring(0, 30) + '...' : content;
            vscode.window.setStatusBarMessage(`Captured ${role} message: ${truncatedContent}`, 3000);
            return true;
        } else {
            const errorText = await response.text();
            vscode.window.showErrorMessage(`Failed to send message: ${errorText}`);
            return false;
        }
    } catch (error) {
        vscode.window.showErrorMessage(`Error connecting to memory server: ${error.message}`);
        return false;
    }
}

function updateStatusBar(statusBarItem) {
    if (currentConversationTag) {
        statusBarItem.text = `$(tag) ${currentConversationTag}`;
        statusBarItem.tooltip = `Current conversation tag: ${currentConversationTag}`;
    } else {
        statusBarItem.text = '$(tag) Set Tag';
        statusBarItem.tooltip = 'No conversation tag set. Click to set one.';
    }
}

module.exports = {
    activate,
    deactivate
};
