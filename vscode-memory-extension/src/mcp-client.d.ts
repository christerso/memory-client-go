interface McpClient {
    connect(): Promise<void>;
    disconnect(): void;
    addMessage(role: 'user'|'assistant'|'system', content: string): Promise<void>;
    indexProject(path: string, tag?: string): Promise<void>;
    on(event: 'connected', listener: () => void): this;
    on(event: 'disconnected', listener: () => void): this;
    on(event: 'error', listener: (error: Error) => void): this;
}
