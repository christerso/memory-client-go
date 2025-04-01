package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
	
	"github.com/christerso/memory-client-go/internal/models"
)

// MockMemoryClient implements MemoryClientInterface for testing
type MockMemoryClient struct {
	// Control behavior
	returnError bool
	errorMsg    string
	
	// Mock data
	messages     []*models.Message
	projectFiles []*models.ProjectFile
	
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
	listProjectFilesCalled   bool
}

// NewMockClient creates a new mock client with specified behavior
func NewMockClient(returnError bool, errorMsg string) *MockMemoryClient {
	return &MockMemoryClient{
		returnError:  returnError,
		errorMsg:     errorMsg,
		messages:     []*models.Message{},
		projectFiles: []*models.ProjectFile{},
	}
}

// AddMessage implements MemoryClientInterface
func (m *MockMemoryClient) AddMessage(ctx context.Context, message *models.Message) error {
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
func (m *MockMemoryClient) GetConversationHistory(ctx context.Context, limit int, filter *models.HistoryFilter) ([]models.Message, error) {
	m.getConversationCalled = true
	
	if m.returnError {
		return nil, errors.New(m.errorMsg)
	}
	
	if limit <= 0 {
		limit = 10
	}
	
	// Convert []*models.Message to []models.Message
	result := make([]models.Message, 0, len(m.messages))
	for _, msg := range m.messages {
		if msg != nil {
			result = append(result, *msg)
		}
	}
	
	if len(result) > limit {
		return result[:limit], nil
	}
	return result, nil
}

// SearchMessages implements MemoryClientInterface
func (m *MockMemoryClient) SearchMessages(ctx context.Context, query string, limit int) ([]models.Message, error) {
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
	
	// Create mock messages for testing
	result := []models.Message{
		{
			ID:      "msg1",
			Role:    "user",
			Content: "Test message 1",
		},
		{
			ID:      "msg2",
			Role:    "assistant",
			Content: "Test message 2",
		},
	}
	
	if len(result) > limit {
		return result[:limit], nil
	}
	return result, nil
}

// GetMemoryStats implements MemoryClientInterface
func (m *MockMemoryClient) GetMemoryStats(ctx context.Context) (*models.MemoryStats, error) {
	m.getStatsCalled = true
	
	if m.returnError {
		return nil, errors.New(m.errorMsg)
	}
	
	return &models.MemoryStats{
		TotalVectors:      len(m.messages) + len(m.projectFiles),
		MessageCount:      map[string]int{"total": len(m.messages)},
		ProjectFileCount:  len(m.projectFiles),
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
	
	m.messages = []*models.Message{}
	return nil
}

// TagMessages implements MemoryClientInterface
func (m *MockMemoryClient) TagMessages(ctx context.Context, ids []string, tag string) error {
	m.tagMessagesCalled = true
	
	if m.returnError {
		return errors.New(m.errorMsg)
	}
	
	if len(ids) == 0 {
		return errors.New("ids cannot be empty")
	}
	
	if tag == "" {
		return errors.New("tag cannot be empty")
	}
	
	return nil
}

// SummarizeAndTagMessages is no longer part of the interface, but we keep it for testing
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
	
	return 2, nil
}

// GetMessagesByTag implements MemoryClientInterface
func (m *MockMemoryClient) GetMessagesByTag(ctx context.Context, tag string, limit int) ([]models.Message, error) {
	m.getMessagesByTagCalled = true
	
	if m.returnError {
		return nil, errors.New(m.errorMsg)
	}
	
	if tag == "" {
		return nil, errors.New("tag cannot be empty")
	}
	
	// Convert []*models.Message to []models.Message
	result := make([]models.Message, 0, len(m.messages))
	for _, msg := range m.messages {
		if msg != nil {
			result = append(result, *msg)
		}
	}
	
	if len(result) > limit {
		return result[:limit], nil
	}
	return result, nil
}

// IndexProjectFiles implements MemoryClientInterface
func (m *MockMemoryClient) IndexProjectFiles(ctx context.Context, path string, tag string) (int, error) {
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
func (m *MockMemoryClient) SearchProjectFiles(ctx context.Context, query string, limit int) ([]models.ProjectFile, error) {
	m.searchProjectFilesCalled = true
	
	if m.returnError {
		return nil, errors.New(m.errorMsg)
	}
	
	// Convert []*models.ProjectFile to []models.ProjectFile
	result := make([]models.ProjectFile, 0, len(m.projectFiles))
	for _, file := range m.projectFiles {
		if file != nil {
			result = append(result, *file)
		}
	}
	
	// Allow empty query for listing all files
	if query == "" && limit > 0 {
		if len(result) > limit {
			return result[:limit], nil
		}
		return result, nil
	}
	
	if query == "" {
		return nil, errors.New("query cannot be empty")
	}
	
	if limit <= 0 {
		limit = 10
	}
	
	if len(result) > limit {
		return result[:limit], nil
	}
	return result, nil
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
	
	m.projectFiles = []*models.ProjectFile{}
	return nil
}

// ListProjectFiles implements MemoryClientInterface
func (m *MockMemoryClient) ListProjectFiles(ctx context.Context, limit int) ([]models.ProjectFile, error) {
	m.listProjectFilesCalled = true
	
	if m.returnError {
		return nil, errors.New(m.errorMsg)
	}
	
	if limit <= 0 {
		limit = 10
	}
	
	// Create mock project files for testing
	result := []models.ProjectFile{
		{
			ID:        "file1",
			Path:      "/path/to/file1.go",
			Content:   "package main\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}",
			Language:  "go",
			Tag:       "test",
			Timestamp: time.Now(),
		},
		{
			ID:        "file2",
			Path:      "/path/to/file2.go",
			Content:   "package main\n\nfunc init() {\n\tfmt.Println(\"Initializing...\")\n}",
			Language:  "go",
			Tag:       "test",
			Timestamp: time.Now(),
		},
	}
	
	if len(result) > limit {
		return result[:limit], nil
	}
	return result, nil
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
			args:        json.RawMessage(`{"ids":["id1","id2"],"tag":"test-tag"}`),
			wantError:   false,
			mockError:   false,
		},
		{
			name:        "missing ids",
			args:        json.RawMessage(`{"tag":"test-tag"}`),
			wantError:   true,
			mockError:   false,
		},
		{
			name:        "missing tag",
			args:        json.RawMessage(`{"ids":["id1","id2"]}`),
			wantError:   true,
			mockError:   false,
		},
		{
			name:        "client error",
			args:        json.RawMessage(`{"ids":["id1","id2"],"tag":"test-tag"}`),
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
			
			// Check if SearchMessages was called instead of SummarizeAndTagMessages
			if !tt.wantError && !mock.searchMessagesCalled {
				t.Error("SearchMessages was not called")
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
