package main

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

// MockMemoryClient implements MemoryClientInterface for testing
type MockMemoryClient struct {
	// Control behavior
	returnError bool
	errorMsg    string
	
	// Mock data
	messages     []*Message
	projectFiles []*ProjectFile
	
	// Track calls
	addMessageCalled         bool
	getConversationCalled    bool
	searchMessagesCalled     bool
	getStatsCalled           bool
	deleteMessageCalled      bool
	deleteAllMessagesCalled  bool
	tagMessagesCalled        bool
	summarizeAndTagCalled    bool
	getMessagesByTagCalled   bool
	indexProjectFilesCalled  bool
	updateProjectFilesCalled bool
	searchProjectFilesCalled bool
	deleteProjectFileCalled  bool
	deleteAllFilesCalled     bool
}

// NewMockClient creates a new mock client with specified behavior
func NewMockClient(returnError bool, errorMsg string) *MockMemoryClient {
	return &MockMemoryClient{
		returnError:  returnError,
		errorMsg:     errorMsg,
		messages:     []*Message{},
		projectFiles: []*ProjectFile{},
	}
}

// AddMessage implements MemoryClientInterface
func (m *MockMemoryClient) AddMessage(ctx context.Context, message *Message) error {
	m.addMessageCalled = true
	
	if m.returnError {
		return errors.New(m.errorMsg)
	}
	
	if message == nil || message.Role == "" || message.Content == "" {
		return errors.New("invalid message")
	}
	
	m.messages = append(m.messages, message)
	return nil
}

// GetConversationHistory implements MemoryClientInterface
func (m *MockMemoryClient) GetConversationHistory(ctx context.Context, limit int) ([]*Message, error) {
	m.getConversationCalled = true
	
	if m.returnError {
		return nil, errors.New(m.errorMsg)
	}
	
	if limit <= 0 {
		limit = 10
	}
	
	if len(m.messages) > limit {
		return m.messages[:limit], nil
	}
	return m.messages, nil
}

// SearchMessages implements MemoryClientInterface
func (m *MockMemoryClient) SearchMessages(ctx context.Context, query string, limit int) ([]*Message, error) {
	m.searchMessagesCalled = true
	
	if m.returnError {
		return nil, errors.New(m.errorMsg)
	}
	
	if query == "" {
		return nil, errors.New("query cannot be empty")
	}
	
	if limit <= 0 {
		limit = 10
	}
	
	if len(m.messages) > limit {
		return m.messages[:limit], nil
	}
	return m.messages, nil
}

// GetMemoryStats implements MemoryClientInterface
func (m *MockMemoryClient) GetMemoryStats(ctx context.Context) (map[string]interface{}, error) {
	m.getStatsCalled = true
	
	if m.returnError {
		return nil, errors.New(m.errorMsg)
	}
	
	return map[string]interface{}{
		"message_count":      len(m.messages),
		"project_file_count": len(m.projectFiles),
	}, nil
}

// DeleteMessage implements MemoryClientInterface
func (m *MockMemoryClient) DeleteMessage(ctx context.Context, id string) error {
	m.deleteMessageCalled = true
	
	if m.returnError {
		return errors.New(m.errorMsg)
	}
	
	if id == "" {
		return errors.New("id cannot be empty")
	}
	
	return nil
}

// DeleteAllMessages implements MemoryClientInterface
func (m *MockMemoryClient) DeleteAllMessages(ctx context.Context) error {
	m.deleteAllMessagesCalled = true
	
	if m.returnError {
		return errors.New(m.errorMsg)
	}
	
	m.messages = []*Message{}
	return nil
}

// TagMessages implements MemoryClientInterface
func (m *MockMemoryClient) TagMessages(ctx context.Context, query string, tags []string, limit int) (int, error) {
	m.tagMessagesCalled = true
	
	if m.returnError {
		return 0, errors.New(m.errorMsg)
	}
	
	if query == "" {
		return 0, errors.New("query cannot be empty")
	}
	
	if len(tags) == 0 {
		return 0, errors.New("tags cannot be empty")
	}
	
	return 2, nil
}

// SummarizeAndTagMessages implements MemoryClientInterface
func (m *MockMemoryClient) SummarizeAndTagMessages(ctx context.Context, query string, summary string, tags []string, limit int) (int, error) {
	m.summarizeAndTagCalled = true
	
	if m.returnError {
		return 0, errors.New(m.errorMsg)
	}
	
	if query == "" {
		return 0, errors.New("query cannot be empty")
	}
	
	if summary == "" {
		return 0, errors.New("summary cannot be empty")
	}
	
	if len(tags) == 0 {
		return 0, errors.New("tags cannot be empty")
	}
	
	return 1, nil
}

// GetMessagesByTag implements MemoryClientInterface
func (m *MockMemoryClient) GetMessagesByTag(ctx context.Context, tag string, limit int) ([]*Message, error) {
	m.getMessagesByTagCalled = true
	
	if m.returnError {
		return nil, errors.New(m.errorMsg)
	}
	
	if tag == "" {
		return nil, errors.New("tag cannot be empty")
	}
	
	if len(m.messages) > limit {
		return m.messages[:limit], nil
	}
	return m.messages, nil
}

// IndexProjectFiles implements MemoryClientInterface
func (m *MockMemoryClient) IndexProjectFiles(ctx context.Context, path string) (int, error) {
	m.indexProjectFilesCalled = true
	
	if m.returnError {
		return 0, errors.New(m.errorMsg)
	}
	
	if path == "" {
		return 0, errors.New("path cannot be empty")
	}
	
	return 5, nil
}

// UpdateProjectFiles implements MemoryClientInterface
func (m *MockMemoryClient) UpdateProjectFiles(ctx context.Context, path string) (int, int, error) {
	m.updateProjectFilesCalled = true
	
	if m.returnError {
		return 0, 0, errors.New(m.errorMsg)
	}
	
	if path == "" {
		return 0, 0, errors.New("path cannot be empty")
	}
	
	return 3, 1, nil
}

// SearchProjectFiles implements MemoryClientInterface
func (m *MockMemoryClient) SearchProjectFiles(ctx context.Context, query string, limit int) ([]*ProjectFile, error) {
	m.searchProjectFilesCalled = true
	
	if m.returnError {
		return nil, errors.New(m.errorMsg)
	}
	
	// Allow empty query for listing all files
	if query == "" && limit > 0 {
		if len(m.projectFiles) > limit {
			return m.projectFiles[:limit], nil
		}
		return m.projectFiles, nil
	}
	
	if query == "" {
		return nil, errors.New("query cannot be empty")
	}
	
	if limit <= 0 {
		limit = 10
	}
	
	if len(m.projectFiles) > limit {
		return m.projectFiles[:limit], nil
	}
	return m.projectFiles, nil
}

// DeleteProjectFile implements MemoryClientInterface
func (m *MockMemoryClient) DeleteProjectFile(ctx context.Context, path string) error {
	m.deleteProjectFileCalled = true
	
	if m.returnError {
		return errors.New(m.errorMsg)
	}
	
	if path == "" {
		return errors.New("path cannot be empty")
	}
	
	return nil
}

// DeleteAllProjectFiles implements MemoryClientInterface
func (m *MockMemoryClient) DeleteAllProjectFiles(ctx context.Context) error {
	m.deleteAllFilesCalled = true
	
	if m.returnError {
		return errors.New(m.errorMsg)
	}
	
	m.projectFiles = []*ProjectFile{}
	return nil
}

// TestAddMessage tests the handleAddMessage function
func TestAddMessage(t *testing.T) {
	tests := []struct {
		name        string
		args        json.RawMessage
		wantError   bool
		mockError   bool
		errorMsg    string
	}{
		{
			name:        "valid message",
			args:        json.RawMessage(`{"role":"user","content":"test message"}`),
			wantError:   false,
			mockError:   false,
		},
		{
			name:        "missing role",
			args:        json.RawMessage(`{"content":"test message"}`),
			wantError:   true,
			mockError:   false,
		},
		{
			name:        "missing content",
			args:        json.RawMessage(`{"role":"user"}`),
			wantError:   true,
			mockError:   false,
		},
		{
			name:        "client error",
			args:        json.RawMessage(`{"role":"user","content":"test message"}`),
			wantError:   true,
			mockError:   true,
			errorMsg:    "mock error",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient(tt.mockError, tt.errorMsg)
			server := &MCPServer{client: mock}
			
			resp, err := server.handleAddMessage(context.Background(), "test-id", tt.args)
			
			if (err != nil) != tt.wantError {
				t.Errorf("handleAddMessage() error = %v, wantError %v", err, tt.wantError)
				return
			}
			
			if err == nil && !resp.Success {
				t.Errorf("handleAddMessage() success = %v, want true", resp.Success)
			}
			
			if !tt.wantError && !mock.addMessageCalled {
				t.Error("AddMessage was not called")
			}
		})
	}
}

// TestGetConversationHistory tests the handleGetConversationHistory function
func TestGetConversationHistory(t *testing.T) {
	tests := []struct {
		name        string
		args        json.RawMessage
		wantError   bool
		mockError   bool
		errorMsg    string
	}{
		{
			name:        "valid request",
			args:        json.RawMessage(`{"limit":10}`),
			wantError:   false,
			mockError:   false,
		},
		{
			name:        "default limit",
			args:        json.RawMessage(`{}`),
			wantError:   false,
			mockError:   false,
		},
		{
			name:        "client error",
			args:        json.RawMessage(`{"limit":10}`),
			wantError:   true,
			mockError:   true,
			errorMsg:    "mock error",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient(tt.mockError, tt.errorMsg)
			server := &MCPServer{client: mock}
			
			resp, err := server.handleGetConversationHistory(context.Background(), "test-id", tt.args)
			
			if (err != nil) != tt.wantError {
				t.Errorf("handleGetConversationHistory() error = %v, wantError %v", err, tt.wantError)
				return
			}
			
			if err == nil && !resp.Success {
				t.Errorf("handleGetConversationHistory() success = %v, want true", resp.Success)
			}
			
			if !tt.wantError && !mock.getConversationCalled {
				t.Error("GetConversationHistory was not called")
			}
		})
	}
}

// TestSearchMessages tests the handleSearchSimilarMessages function
func TestSearchMessages(t *testing.T) {
	tests := []struct {
		name        string
		args        json.RawMessage
		wantError   bool
		mockError   bool
		errorMsg    string
	}{
		{
			name:        "valid search",
			args:        json.RawMessage(`{"query":"test","limit":10}`),
			wantError:   false,
			mockError:   false,
		},
		{
			name:        "missing query",
			args:        json.RawMessage(`{"limit":10}`),
			wantError:   true,
			mockError:   false,
		},
		{
			name:        "client error",
			args:        json.RawMessage(`{"query":"test","limit":10}`),
			wantError:   true,
			mockError:   true,
			errorMsg:    "mock error",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient(tt.mockError, tt.errorMsg)
			server := &MCPServer{client: mock}
			
			resp, err := server.handleSearchSimilarMessages(context.Background(), "test-id", tt.args)
			
			if (err != nil) != tt.wantError {
				t.Errorf("handleSearchSimilarMessages() error = %v, wantError %v", err, tt.wantError)
				return
			}
			
			if err == nil && !resp.Success {
				t.Errorf("handleSearchSimilarMessages() success = %v, want true", resp.Success)
			}
			
			if !tt.wantError && !mock.searchMessagesCalled {
				t.Error("SearchMessages was not called")
			}
		})
	}
}

// TestGetMemoryStats tests the handleGetMemoryStats function
func TestGetMemoryStats(t *testing.T) {
	tests := []struct {
		name        string
		wantError   bool
		mockError   bool
		errorMsg    string
	}{
		{
			name:        "valid request",
			wantError:   false,
			mockError:   false,
		},
		{
			name:        "client error",
			wantError:   true,
			mockError:   true,
			errorMsg:    "mock error",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient(tt.mockError, tt.errorMsg)
			server := &MCPServer{client: mock}
			
			resp, err := server.handleGetMemoryStats(context.Background(), "test-id", json.RawMessage(`{}`))
			
			if (err != nil) != tt.wantError {
				t.Errorf("handleGetMemoryStats() error = %v, wantError %v", err, tt.wantError)
				return
			}
			
			if err == nil && !resp.Success {
				t.Errorf("handleGetMemoryStats() success = %v, want true", resp.Success)
			}
			
			if !tt.wantError && !mock.getStatsCalled {
				t.Error("GetMemoryStats was not called")
			}
		})
	}
}

// TestDeleteMessage tests the handleDeleteMessage function
func TestDeleteMessage(t *testing.T) {
	tests := []struct {
		name        string
		args        json.RawMessage
		wantError   bool
		mockError   bool
		errorMsg    string
	}{
		{
			name:        "valid delete",
			args:        json.RawMessage(`{"id":"test-id"}`),
			wantError:   false,
			mockError:   false,
		},
		{
			name:        "missing id",
			args:        json.RawMessage(`{}`),
			wantError:   true,
			mockError:   false,
		},
		{
			name:        "client error",
			args:        json.RawMessage(`{"id":"test-id"}`),
			wantError:   true,
			mockError:   true,
			errorMsg:    "mock error",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient(tt.mockError, tt.errorMsg)
			server := &MCPServer{client: mock}
			
			resp, err := server.handleDeleteMessage(context.Background(), "test-id", tt.args)
			
			if (err != nil) != tt.wantError {
				t.Errorf("handleDeleteMessage() error = %v, wantError %v", err, tt.wantError)
				return
			}
			
			if err == nil && !resp.Success {
				t.Errorf("handleDeleteMessage() success = %v, want true", resp.Success)
			}
			
			if !tt.wantError && !mock.deleteMessageCalled {
				t.Error("DeleteMessage was not called")
			}
		})
	}
}

// TestDeleteAllMessages tests the handleDeleteAllMessages function
func TestDeleteAllMessages(t *testing.T) {
	tests := []struct {
		name        string
		wantError   bool
		mockError   bool
		errorMsg    string
	}{
		{
			name:        "valid delete all",
			wantError:   false,
			mockError:   false,
		},
		{
			name:        "client error",
			wantError:   true,
			mockError:   true,
			errorMsg:    "mock error",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient(tt.mockError, tt.errorMsg)
			server := &MCPServer{client: mock}
			
			resp, err := server.handleDeleteAllMessages(context.Background(), "test-id", json.RawMessage(`{}`))
			
			if (err != nil) != tt.wantError {
				t.Errorf("handleDeleteAllMessages() error = %v, wantError %v", err, tt.wantError)
				return
			}
			
			if err == nil && !resp.Success {
				t.Errorf("handleDeleteAllMessages() success = %v, want true", resp.Success)
			}
			
			if !tt.wantError && !mock.deleteAllMessagesCalled {
				t.Error("DeleteAllMessages was not called")
			}
		})
	}
}

// TestTagMessages tests the handleTagMessages function
func TestTagMessages(t *testing.T) {
	tests := []struct {
		name        string
		args        json.RawMessage
		wantError   bool
		mockError   bool
		errorMsg    string
	}{
		{
			name:        "valid tag",
			args:        json.RawMessage(`{"query":"test","tags":["tag1","tag2"],"limit":10}`),
			wantError:   false,
			mockError:   false,
		},
		{
			name:        "missing query",
			args:        json.RawMessage(`{"tags":["tag1","tag2"],"limit":10}`),
			wantError:   true,
			mockError:   false,
		},
		{
			name:        "missing tags",
			args:        json.RawMessage(`{"query":"test","limit":10}`),
			wantError:   true,
			mockError:   false,
		},
		{
			name:        "client error",
			args:        json.RawMessage(`{"query":"test","tags":["tag1","tag2"],"limit":10}`),
			wantError:   true,
			mockError:   true,
			errorMsg:    "mock error",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient(tt.mockError, tt.errorMsg)
			server := &MCPServer{client: mock}
			
			resp, err := server.handleTagMessages(context.Background(), "test-id", tt.args)
			
			if (err != nil) != tt.wantError {
				t.Errorf("handleTagMessages() error = %v, wantError %v", err, tt.wantError)
				return
			}
			
			if err == nil && !resp.Success {
				t.Errorf("handleTagMessages() success = %v, want true", resp.Success)
			}
			
			if !tt.wantError && !mock.tagMessagesCalled {
				t.Error("TagMessages was not called")
			}
		})
	}
}

// TestSummarizeAndTagMessages tests the handleSummarizeAndTagMessages function
func TestSummarizeAndTagMessages(t *testing.T) {
	tests := []struct {
		name        string
		args        json.RawMessage
		wantError   bool
		mockError   bool
		errorMsg    string
	}{
		{
			name:        "valid summarize and tag",
			args:        json.RawMessage(`{"query":"test","summary":"test summary","tags":["tag1","tag2"],"limit":10}`),
			wantError:   false,
			mockError:   false,
		},
		{
			name:        "missing query",
			args:        json.RawMessage(`{"summary":"test summary","tags":["tag1","tag2"],"limit":10}`),
			wantError:   true,
			mockError:   false,
		},
		{
			name:        "missing summary",
			args:        json.RawMessage(`{"query":"test","tags":["tag1","tag2"],"limit":10}`),
			wantError:   true,
			mockError:   false,
		},
		{
			name:        "missing tags",
			args:        json.RawMessage(`{"query":"test","summary":"test summary","limit":10}`),
			wantError:   true,
			mockError:   false,
		},
		{
			name:        "client error",
			args:        json.RawMessage(`{"query":"test","summary":"test summary","tags":["tag1","tag2"],"limit":10}`),
			wantError:   true,
			mockError:   true,
			errorMsg:    "mock error",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient(tt.mockError, tt.errorMsg)
			server := &MCPServer{client: mock}
			
			resp, err := server.handleSummarizeAndTagMessages(context.Background(), "test-id", tt.args)
			
			if (err != nil) != tt.wantError {
				t.Errorf("handleSummarizeAndTagMessages() error = %v, wantError %v", err, tt.wantError)
				return
			}
			
			if err == nil && !resp.Success {
				t.Errorf("handleSummarizeAndTagMessages() success = %v, want true", resp.Success)
			}
			
			if !tt.wantError && !mock.summarizeAndTagCalled {
				t.Error("SummarizeAndTagMessages was not called")
			}
		})
	}
}

// TestGetMessagesByTag tests the handleGetMessagesByTag function
func TestGetMessagesByTag(t *testing.T) {
	tests := []struct {
		name        string
		args        json.RawMessage
		wantError   bool
		mockError   bool
		errorMsg    string
	}{
		{
			name:        "valid get by tag",
			args:        json.RawMessage(`{"tag":"test-tag","limit":10}`),
			wantError:   false,
			mockError:   false,
		},
		{
			name:        "missing tag",
			args:        json.RawMessage(`{"limit":10}`),
			wantError:   true,
			mockError:   false,
		},
		{
			name:        "client error",
			args:        json.RawMessage(`{"tag":"test-tag","limit":10}`),
			wantError:   true,
			mockError:   true,
			errorMsg:    "mock error",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient(tt.mockError, tt.errorMsg)
			server := &MCPServer{client: mock}
			
			resp, err := server.handleGetMessagesByTag(context.Background(), "test-id", tt.args)
			
			if (err != nil) != tt.wantError {
				t.Errorf("handleGetMessagesByTag() error = %v, wantError %v", err, tt.wantError)
				return
			}
			
			if err == nil && !resp.Success {
				t.Errorf("handleGetMessagesByTag() success = %v, want true", resp.Success)
			}
			
			if !tt.wantError && !mock.getMessagesByTagCalled {
				t.Error("GetMessagesByTag was not called")
			}
		})
	}
}

// TestHandleResourceAccess tests the handleResourceAccess function
func TestHandleResourceAccess(t *testing.T) {
	tests := []struct {
		name        string
		uri         string
		wantError   bool
		mockError   bool
		errorMsg    string
	}{
		{
			name:        "conversation history resource",
			uri:         "memory:///conversation_history",
			wantError:   false,
			mockError:   false,
		},
		{
			name:        "project files resource",
			uri:         "memory:///project_files",
			wantError:   false,
			mockError:   false,
		},
		{
			name:        "unknown resource",
			uri:         "memory:///unknown_resource",
			wantError:   true,
			mockError:   false,
		},
		{
			name:        "client error",
			uri:         "memory:///conversation_history",
			wantError:   true,
			mockError:   true,
			errorMsg:    "mock error",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient(tt.mockError, tt.errorMsg)
			server := &MCPServer{client: mock}
			
			data, _ := json.Marshal(map[string]string{"uri": tt.uri})
			request := &MCPRequest{
				ID:   "test-id",
				Type: "resource_access",
				Data: data,
			}
			
			resp, err := server.handleResourceAccess(context.Background(), request)
			
			if (err != nil) != tt.wantError {
				t.Errorf("handleResourceAccess() error = %v, wantError %v", err, tt.wantError)
				return
			}
			
			if err == nil && !resp.Success {
				t.Errorf("handleResourceAccess() success = %v, want true", resp.Success)
			}
		})
	}
}
