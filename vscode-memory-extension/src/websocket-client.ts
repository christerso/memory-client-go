import { EventEmitter } from 'events';
import WebSocket from 'ws';

type WebSocketEvent = 'connected' | 'disconnected' | 'error';

export class WebSocketClient implements McpClient {
    private ws: WebSocket | null = null;
    private eventEmitter = new EventEmitter();
    private retryCount = 0;
    private maxRetries = 3;
    private retryDelay = 1000;

    constructor(private url: string) {}

    async connect(): Promise<void> {
        return new Promise((resolve, reject) => {
            if (this.ws) {
                resolve();
                return;
            }

            this.ws = new WebSocket(this.url);

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
                } else {
                    reject(error);
                }
            });

            this.ws.on('close', () => {
                this.eventEmitter.emit('disconnected');
                this.ws = null;
            });
        });
    }

    disconnect(): void {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
    }

    async addMessage(role: 'user'|'assistant'|'system', content: string): Promise<void> {
        if (!this.ws) {
            throw new Error('Not connected to MCP server');
        }

        return new Promise((resolve, reject) => {
            const message = JSON.stringify({
                type: "tool_call",
                data: JSON.stringify({
                    name: "add_message",
                    arguments: {
                        role,
                        content,
                        embedding: await this.generateEmbedding(content)
                    }
                })
            });

            this.ws!.send(message, (error) => {
                error ? reject(error) : resolve();
            });
        });
    }

    async indexProject(path: string, tag?: string): Promise<void> {
        if (!this.ws) {
            throw new Error('Not connected to MCP server');
        }

        return new Promise((resolve, reject) => {
            const message = JSON.stringify({
                type: "tool_call",
                data: JSON.stringify({
                    name: "index_project",
                    arguments: {
                        path,
                        ...(tag && { tag })
                    }
                })
            });

            this.ws!.send(message, (error) => {
                error ? reject(error) : resolve();
            });
        });
    }

    async retrieve(key: string): Promise<string | null> {
        if (!this.ws) {
            throw new Error('Not connected to MCP server');
        }

        return new Promise((resolve, reject) => {
            const message = JSON.stringify({
                type: "tool_call",
                data: JSON.stringify({
                    name: "retrieve",
                    arguments: { key }
                })
            });

            const handler = (data: WebSocket.Data) => {
                try {
                    const response = JSON.parse(data.toString());
                    if (response.key === key) {
                        this.ws!.off('message', handler);
                        resolve(response.value);
                    }
                } catch (error) {
                    reject(error);
                }
            };

            this.ws!.on('message', handler);
            this.ws!.send(message);
        });
    }

    on(event: 'connected', listener: () => void): this;
    on(event: 'disconnected', listener: () => void): this;
    on(event: 'error', listener: (error: Error) => void): this;
    on(event: WebSocketEvent, listener: (...args: any[]) => void): this {
        this.eventEmitter.on(event, listener);
        return this;
    }

    private async generateEmbedding(text: string): Promise<number[]> {
        // Temporary embedding implementation for testing
        const words = text.toLowerCase().split(/\W+/).filter(w => w.length > 2);
        const embedding = new Array(384).fill(0);
        words.forEach(word => {
            const hash = this.hashString(word);
            embedding[hash % embedding.length] += 1;
        });
        return embedding;
    }

    private hashString(str: string): number {
        let hash = 0;
        for (let i = 0; i < str.length; i++) {
            hash = (hash << 5) - hash + str.charCodeAt(i);
            hash |= 0; // Convert to 32-bit integer
        }
        return Math.abs(hash);
    }
}
