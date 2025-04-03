package mcp

import (
	"context"
	"encoding/json"
	"testing"
)

// TestAddMessage tests the handleAddMessage function
func TestAddMessage(t *testing.T) {
	tests := []struct {
		name      string
		args      json.RawMessage
		wantError bool
		mockError bool
		errorMsg  string
	}{
		{
			name:      "valid message",
			args:      json.RawMessage(`{"role":"user","content":"test message"}`),
			wantError: false,
			mockError: false,
		},
		{
			name:      "missing role",
			args:      json.RawMessage(`{"content":"test message"}`),
			wantError: true,
			mockError: false,
		},
		{
			name:      "missing content",
			args:      json.RawMessage(`{"role":"user"}`),
			wantError: true,
			mockError: false,
		},
		{
			name:      "client error",
			args:      json.RawMessage(`{"role":"user","content":"test message"}`),
			wantError: true,
			mockError: true,
			errorMsg:  "mock error",
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

			if !tt.wantError && !mock.AddMessageCalled {
				t.Error("AddMessage was not called")
			}
		})
	}
}

// TestGetConversationHistory tests the handleGetConversationHistory function
func TestGetConversationHistory(t *testing.T) {
	tests := []struct {
		name      string
		args      json.RawMessage
		wantError bool
		mockError bool
		errorMsg  string
	}{
		{
			name:      "valid request",
			args:      json.RawMessage(`{"limit":10}`),
			wantError: false,
			mockError: false,
		},
		{
			name:      "default limit",
			args:      json.RawMessage(`{}`),
			wantError: false,
			mockError: false,
		},
		{
			name:      "client error",
			args:      json.RawMessage(`{"limit":10}`),
			wantError: true,
			mockError: true,
			errorMsg:  "mock error",
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

			if !tt.wantError && !mock.GetConversationCalled {
				t.Error("GetConversationHistory was not called")
			}
		})
	}
}

// TestSearchMessages tests the handleSearchSimilarMessages function
func TestSearchMessages(t *testing.T) {
	tests := []struct {
		name      string
		args      json.RawMessage
		wantError bool
		mockError bool
		errorMsg  string
	}{
		{
			name:      "valid search",
			args:      json.RawMessage(`{"query":"test","limit":10}`),
			wantError: false,
			mockError: false,
		},
		{
			name:      "missing query",
			args:      json.RawMessage(`{"limit":10}`),
			wantError: true,
			mockError: false,
		},
		{
			name:      "client error",
			args:      json.RawMessage(`{"query":"test","limit":10}`),
			wantError: true,
			mockError: true,
			errorMsg:  "mock error",
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

			if !tt.wantError && !mock.SearchMessagesCalled {
				t.Error("SearchMessages was not called")
			}
		})
	}
}

// TestGetMemoryStats tests the handleGetMemoryStats function
func TestGetMemoryStats(t *testing.T) {
	tests := []struct {
		name      string
		wantError bool
		mockError bool
		errorMsg  string
	}{
		{
			name:      "valid request",
			wantError: false,
			mockError: false,
		},
		{
			name:      "client error",
			wantError: true,
			mockError: true,
			errorMsg:  "mock error",
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

			if !tt.wantError && !mock.GetStatsCalled {
				t.Error("GetMemoryStats was not called")
			}
		})
	}
}

// TestDeleteMessage tests the handleDeleteMessage function
func TestDeleteMessage(t *testing.T) {
	tests := []struct {
		name      string
		args      json.RawMessage
		wantError bool
		mockError bool
		errorMsg  string
	}{
		{
			name:      "valid delete",
			args:      json.RawMessage(`{"id":"test-id"}`),
			wantError: false,
			mockError: false,
		},
		{
			name:      "missing id",
			args:      json.RawMessage(`{}`),
			wantError: true,
			mockError: false,
		},
		{
			name:      "client error",
			args:      json.RawMessage(`{"id":"test-id"}`),
			wantError: true,
			mockError: true,
			errorMsg:  "mock error",
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

			if !tt.wantError && !mock.DeleteMessageCalled {
				t.Error("DeleteMessage was not called")
			}
		})
	}
}

// TestDeleteAllMessages tests the handleDeleteAllMessages function
func TestDeleteAllMessages(t *testing.T) {
	tests := []struct {
		name      string
		wantError bool
		mockError bool
		errorMsg  string
	}{
		{
			name:      "valid delete all",
			wantError: false,
			mockError: false,
		},
		{
			name:      "client error",
			wantError: true,
			mockError: true,
			errorMsg:  "mock error",
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

			if !tt.wantError && !mock.DeleteAllMessagesCalled {
				t.Error("DeleteAllMessages was not called")
			}
		})
	}
}

// TestTagMessages tests the handleTagMessages function
func TestTagMessages(t *testing.T) {
	tests := []struct {
		name      string
		args      json.RawMessage
		wantError bool
		mockError bool
		errorMsg  string
	}{
		{
			name:      "valid tag",
			args:      json.RawMessage(`{"ids":["id1","id2"],"tag":"test-tag"}`),
			wantError: false,
			mockError: false,
		},
		{
			name:      "missing ids",
			args:      json.RawMessage(`{"tag":"test-tag"}`),
			wantError: true,
			mockError: false,
		},
		{
			name:      "missing tag",
			args:      json.RawMessage(`{"ids":["id1","id2"]}`),
			wantError: true,
			mockError: false,
		},
		{
			name:      "client error",
			args:      json.RawMessage(`{"ids":["id1","id2"],"tag":"test-tag"}`),
			wantError: true,
			mockError: true,
			errorMsg:  "mock error",
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

			if !tt.wantError && !mock.TagMessagesCalled {
				t.Error("TagMessages was not called")
			}
		})
	}
}

// TestSummarizeAndTagMessages tests the handleSummarizeAndTagMessages function
func TestSummarizeAndTagMessages(t *testing.T) {
	tests := []struct {
		name      string
		args      json.RawMessage
		wantError bool
		mockError bool
		errorMsg  string
	}{
		{
			name:      "valid summarize and tag",
			args:      json.RawMessage(`{"query":"test","summary":"test summary","tags":["tag1","tag2"],"limit":10}`),
			wantError: false,
			mockError: false,
		},
		{
			name:      "missing query",
			args:      json.RawMessage(`{"summary":"test summary","tags":["tag1","tag2"],"limit":10}`),
			wantError: true,
			mockError: false,
		},
		{
			name:      "missing summary",
			args:      json.RawMessage(`{"query":"test","tags":["tag1","tag2"],"limit":10}`),
			wantError: true,
			mockError: false,
		},
		{
			name:      "missing tags",
			args:      json.RawMessage(`{"query":"test","summary":"test summary","limit":10}`),
			wantError: true,
			mockError: false,
		},
		{
			name:      "client error",
			args:      json.RawMessage(`{"query":"test","summary":"test summary","tags":["tag1","tag2"],"limit":10}`),
			wantError: true,
			mockError: true,
			errorMsg:  "mock error",
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
			if !tt.wantError && !mock.SearchMessagesCalled {
				t.Error("SearchMessages was not called")
			}
		})
	}
}

// TestGetMessagesByTag tests the handleGetMessagesByTag function
func TestGetMessagesByTag(t *testing.T) {
	tests := []struct {
		name      string
		args      json.RawMessage
		wantError bool
		mockError bool
		errorMsg  string
	}{
		{
			name:      "valid get by tag",
			args:      json.RawMessage(`{"tag":"test-tag","limit":10}`),
			wantError: false,
			mockError: false,
		},
		{
			name:      "missing tag",
			args:      json.RawMessage(`{"limit":10}`),
			wantError: true,
			mockError: false,
		},
		{
			name:      "client error",
			args:      json.RawMessage(`{"tag":"test-tag","limit":10}`),
			wantError: true,
			mockError: true,
			errorMsg:  "mock error",
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

			if !tt.wantError && !mock.GetMessagesByTagCalled {
				t.Error("GetMessagesByTag was not called")
			}
		})
	}
}

// TestHandleResourceAccess tests the handleResourceAccess function
func TestHandleResourceAccess(t *testing.T) {
	tests := []struct {
		name      string
		uri       string
		wantError bool
		mockError bool
		errorMsg  string
	}{
		{
			name:      "conversation history resource",
			uri:       "memory:///conversation_history",
			wantError: false,
			mockError: false,
		},
		{
			name:      "project files resource",
			uri:       "memory:///project_files",
			wantError: false,
			mockError: false,
		},
		{
			name:      "unknown resource",
			uri:       "memory:///unknown_resource",
			wantError: true,
			mockError: false,
		},
		{
			name:      "client error",
			uri:       "memory:///conversation_history",
			wantError: true,
			mockError: true,
			errorMsg:  "mock error",
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
