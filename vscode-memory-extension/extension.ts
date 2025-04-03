import * as vscode from 'vscode';
import { WebSocketClient } from './src/websocket-client';

let wsClient: WebSocketClient;
let currentTag: string | undefined;

export async function activate(context: vscode.ExtensionContext) {
    const config = vscode.workspace.getConfiguration('vscode-memory');
    wsClient = new WebSocketClient(config.get('serverUrl') || 'ws://localhost:10010');
    
    // Connect with retry logic
    const connectWithRetry = async (attempt = 0) => {
        try {
            await wsClient.connect();
            vscode.window.showInformationMessage('Connected to memory server');
            
            // Auto-index project when workspace opens
            if (vscode.workspace.workspaceFolders) {
                for (const folder of vscode.workspace.workspaceFolders) {
                    await wsClient.indexProject(folder.uri.fsPath, currentTag);
                }
            }
        } catch (error) {
            if (attempt < 5) {
                setTimeout(() => connectWithRetry(attempt + 1), 2000 * attempt);
            } else {
                vscode.window.showErrorMessage(`Failed to connect to memory server: ${error}`);
            }
        }
    };

    // Setup listeners
    context.subscriptions.push(
        vscode.workspace.onDidOpenTextDocument(async doc => {
            if (config.get('autoCapture')) {
                await wsClient.addMessage('user', `Opened file: ${doc.fileName}`);
            }
        }),
        
        vscode.workspace.onDidChangeTextDocument(e => {
            if (config.get('autoCapture') && e.contentChanges.length > 0) {
                const change = e.contentChanges[0];
                wsClient.addMessage('user', `Code change at ${e.document.fileName}: ${change.text}`);
            }
        }),
        
        vscode.window.onDidChangeTextEditorSelection(e => {
            if (config.get('autoCapture') && e.selections.length > 0) {
                const selection = e.selections[0];
                const text = e.textEditor.document.getText(selection);
                if (text.trim().length > 0) {
                    wsClient.addMessage('user', `Selected code in ${e.textEditor.document.fileName}: ${text}`);
                }
            }
        }),
        
        vscode.commands.registerCommand('vscode-memory.setTag', async () => {
            currentTag = await vscode.window.showInputBox({
                prompt: 'Enter conversation tag',
                validateInput: value => value?.includes(' ') ? 'No spaces allowed' : null
            });
        }),
        
        vscode.commands.registerCommand('vscode-memory.getTag', () => {
            vscode.window.showInformationMessage(currentTag || 'No active tag');
        })
    );

    // Start connection
    connectWithRetry();
}

export function deactivate() {
    if (wsClient) {
        wsClient.disconnect();
    }
}
