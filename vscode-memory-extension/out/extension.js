"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || function (mod) {
    if (mod && mod.__esModule) return mod;
    var result = {};
    if (mod != null) for (var k in mod) if (k !== "default" && Object.prototype.hasOwnProperty.call(mod, k)) __createBinding(result, mod, k);
    __setModuleDefault(result, mod);
    return result;
};
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.deactivate = exports.activate = void 0;
const vscode = __importStar(require("vscode"));
const websocket_client_1 = require("./src/websocket-client");
let wsClient;
let currentTag;
function activate(context) {
    return __awaiter(this, void 0, void 0, function* () {
        const config = vscode.workspace.getConfiguration('vscode-memory');
        wsClient = new websocket_client_1.WebSocketClient(config.get('serverUrl') || 'ws://localhost:10010');
        // Connect with retry logic
        const connectWithRetry = (attempt = 0) => __awaiter(this, void 0, void 0, function* () {
            try {
                yield wsClient.connect();
                vscode.window.showInformationMessage('Connected to memory server');
                // Auto-index project when workspace opens
                if (vscode.workspace.workspaceFolders) {
                    for (const folder of vscode.workspace.workspaceFolders) {
                        yield wsClient.indexProject(folder.uri.fsPath, currentTag);
                    }
                }
            }
            catch (error) {
                if (attempt < 5) {
                    setTimeout(() => connectWithRetry(attempt + 1), 2000 * attempt);
                }
                else {
                    vscode.window.showErrorMessage(`Failed to connect to memory server: ${error}`);
                }
            }
        });
        // Setup listeners
        context.subscriptions.push(vscode.workspace.onDidOpenTextDocument((doc) => __awaiter(this, void 0, void 0, function* () {
            if (config.get('autoCapture')) {
                yield wsClient.addMessage('user', `Opened file: ${doc.fileName}`);
            }
        })), vscode.workspace.onDidChangeTextDocument(e => {
            if (config.get('autoCapture') && e.contentChanges.length > 0) {
                const change = e.contentChanges[0];
                wsClient.addMessage('user', `Code change at ${e.document.fileName}: ${change.text}`);
            }
        }), vscode.window.onDidChangeTextEditorSelection(e => {
            if (config.get('autoCapture') && e.selections.length > 0) {
                const selection = e.selections[0];
                const text = e.textEditor.document.getText(selection);
                if (text.trim().length > 0) {
                    wsClient.addMessage('user', `Selected code in ${e.textEditor.document.fileName}: ${text}`);
                }
            }
        }), vscode.commands.registerCommand('vscode-memory.setTag', () => __awaiter(this, void 0, void 0, function* () {
            currentTag = yield vscode.window.showInputBox({
                prompt: 'Enter conversation tag',
                validateInput: value => (value === null || value === void 0 ? void 0 : value.includes(' ')) ? 'No spaces allowed' : null
            });
        })), vscode.commands.registerCommand('vscode-memory.getTag', () => {
            vscode.window.showInformationMessage(currentTag || 'No active tag');
        }));
        // Start connection
        connectWithRetry();
    });
}
exports.activate = activate;
function deactivate() {
    if (wsClient) {
        wsClient.disconnect();
    }
}
exports.deactivate = deactivate;
//# sourceMappingURL=extension.js.map