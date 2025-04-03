"use strict";
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.WebSocketClient = void 0;
const events_1 = require("events");
const ws_1 = __importDefault(require("ws"));
class WebSocketClient {
    constructor(url) {
        this.url = url;
        this.ws = null;
        this.eventEmitter = new events_1.EventEmitter();
        this.retryCount = 0;
        this.maxRetries = 3;
        this.retryDelay = 1000;
    }
    connect() {
        return __awaiter(this, void 0, void 0, function* () {
            return new Promise((resolve, reject) => {
                if (this.ws) {
                    resolve();
                    return;
                }
                this.ws = new ws_1.default(this.url);
                this.ws.on('open', () => {
                    this.retryCount = 0;
                    this.eventEmitter.emit('connected');
                    resolve();
                });
                this.ws.on('error', (error) => {
                    this.eventEmitter.emit('error', error);
                    if (this.retryCount < this.maxRetries) {
                        this.retryCount++;
                        setTimeout(() => this.connect(), this.retryDelay);
                    }
                    else {
                        reject(error);
                    }
                });
                this.ws.on('close', () => {
                    this.eventEmitter.emit('disconnected');
                    this.ws = null;
                });
            });
        });
    }
    disconnect() {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
    }
    addMessage(role, content) {
        return __awaiter(this, void 0, void 0, function* () {
            if (!this.ws) {
                throw new Error('Not connected to MCP server');
            }
            return new Promise((resolve, reject) => {
                const message = JSON.stringify({
                    tool: 'add_message',
                    arguments: {
                        role,
                        content
                    }
                });
                this.ws.send(message, (error) => {
                    error ? reject(error) : resolve();
                });
            });
        });
    }
    indexProject(path, tag) {
        return __awaiter(this, void 0, void 0, function* () {
            if (!this.ws) {
                throw new Error('Not connected to MCP server');
            }
            return new Promise((resolve, reject) => {
                const message = JSON.stringify({
                    tool: 'index_project',
                    arguments: Object.assign({ path }, (tag && { tag }))
                });
                this.ws.send(message, (error) => {
                    error ? reject(error) : resolve();
                });
            });
        });
    }
    retrieve(key) {
        return __awaiter(this, void 0, void 0, function* () {
            if (!this.ws) {
                throw new Error('Not connected to MCP server');
            }
            return new Promise((resolve, reject) => {
                const message = JSON.stringify({
                    type: 'retrieve',
                    key
                });
                const handler = (data) => {
                    try {
                        const response = JSON.parse(data.toString());
                        if (response.key === key) {
                            this.ws.off('message', handler);
                            resolve(response.value);
                        }
                    }
                    catch (error) {
                        reject(error);
                    }
                };
                this.ws.on('message', handler);
                this.ws.send(message);
            });
        });
    }
    on(event, listener) {
        this.eventEmitter.on(event, listener);
        return this;
    }
}
exports.WebSocketClient = WebSocketClient;
//# sourceMappingURL=websocket-client.js.map