package mcp

import (
	"context"
	"errors"

	"github.com/christerso/memory-client-go/internal/models"
)

// MockMemoryClient implements MemoryClientInterface for testing
type MockMemoryClient struct {
	// Control behavior
	ReturnError bool
	ErrorMsg    string

	// Mock data
	Messages     []*models.Message
	ProjectFiles []*models.ProjectFile

	// Track calls
	AddMessageCalled         bool
	GetConversationCalled    bool
	SearchMessagesCalled     bool
	GetStatsCalled           bool
	DeleteMessageCalled      bool
	DeleteAllMessagesCalled  bool
	TagMessagesCalled        bool
	SummarizeAndTagCalled    bool
	GetMessagesByTagCalled   bool
	IndexProjectFilesCalled  bool
	UpdateProjectFilesCalled bool
	SearchProjectFilesCalled bool
	DeleteProjectFileCalled  bool
	DeleteAllFilesCalled     bool
	ListProjectFilesCalled   bool
}

// NewMockClient creates a new mock client with specified behavior
func NewMockClient(returnError bool, errorMsg string) *MockMemoryClient {
	return &MockMemoryClient{
		ReturnError:  returnError,
		ErrorMsg:     errorMsg,
		Messages:     []*models.Message{},
		ProjectFiles: []*models.ProjectFile{},
	}
}

// AddMessage implements MemoryClientInterface
func (m *MockMemoryClient) AddMessage(ctx context.Context, message *models.Message) error {
	m.AddMessageCalled = true
	if m.ReturnError {
		return errors.New(m.ErrorMsg)
	}
	if message == nil || message.Role == "" || message.Content == "" {
		return errors.New("invalid message")
	}
	m.Messages = append(m.Messages, message)
	return nil
}

// GetConversationHistory implements MemoryClientInterface
func (m *MockMemoryClient) GetConversationHistory(ctx context.Context, limit int, filter *models.HistoryFilter) ([]models.Message, error) {
	m.GetConversationCalled = true
	if m.ReturnError {
		return nil, errors.New(m.ErrorMsg)
	}
	if limit <= 0 {
		limit = 10
	}
	result := make([]models.Message, 0, len(m.Messages))
	for _, msg := range m.Messages {
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
	m.SearchMessagesCalled = true
	if m.ReturnError {
		return nil, errors.New(m.ErrorMsg)
	}
	if query == "" {
		return nil, errors.New("query cannot be empty")
	}
	return []models.Message{
		{ID: "msg1", Role: "user", Content: "Test message 1"},
		{ID: "msg2", Role: "assistant", Content: "Test message 2"},
	}, nil
}

// GetMemoryStats implements MemoryClientInterface
func (m *MockMemoryClient) GetMemoryStats(ctx context.Context) (*models.MemoryStats, error) {
	m.GetStatsCalled = true
	if m.ReturnError {
		return nil, errors.New(m.ErrorMsg)
	}
	return &models.MemoryStats{
		TotalVectors:     len(m.Messages) + len(m.ProjectFiles),
		MessageCount:     map[string]int{"total": len(m.Messages)},
		ProjectFileCount: len(m.ProjectFiles),
	}, nil
}

// DeleteMessage implements MemoryClientInterface
func (m *MockMemoryClient) DeleteMessage(ctx context.Context, id string) error {
	m.DeleteMessageCalled = true
	if m.ReturnError {
		return errors.New(m.ErrorMsg)
	}
	return nil
}

// DeleteAllMessages implements MemoryClientInterface
func (m *MockMemoryClient) DeleteAllMessages(ctx context.Context) error {
	m.DeleteAllMessagesCalled = true
	if m.ReturnError {
		return errors.New(m.ErrorMsg)
	}
	m.Messages = []*models.Message{}
	return nil
}

// TagMessages implements MemoryClientInterface
func (m *MockMemoryClient) TagMessages(ctx context.Context, ids []string, tag string) error {
	m.TagMessagesCalled = true
	if m.ReturnError {
		return errors.New(m.ErrorMsg)
	}
	return nil
}

// GetMessagesByTag implements MemoryClientInterface
func (m *MockMemoryClient) GetMessagesByTag(ctx context.Context, tag string, limit int) ([]models.Message, error) {
	m.GetMessagesByTagCalled = true
	if m.ReturnError {
		return nil, errors.New(m.ErrorMsg)
	}
	result := make([]models.Message, 0, len(m.Messages))
	for _, msg := range m.Messages {
		if msg != nil {
			result = append(result, *msg)
		}
	}
	return result, nil
}

// IndexProjectFiles implements MemoryClientInterface
func (m *MockMemoryClient) IndexProjectFiles(ctx context.Context, path string, tag string) (int, error) {
	m.IndexProjectFilesCalled = true
	if m.ReturnError {
		return 0, errors.New(m.ErrorMsg)
	}
	return 5, nil
}

// UpdateProjectFiles implements MemoryClientInterface
func (m *MockMemoryClient) UpdateProjectFiles(ctx context.Context, path string) (int, int, error) {
	m.UpdateProjectFilesCalled = true
	if m.ReturnError {
		return 0, 0, errors.New(m.ErrorMsg)
	}
	return 3, 1, nil
}

// SearchProjectFiles implements MemoryClientInterface
func (m *MockMemoryClient) SearchProjectFiles(ctx context.Context, query string, limit int) ([]models.ProjectFile, error) {
	m.SearchProjectFilesCalled = true
	if m.ReturnError {
		return nil, errors.New(m.ErrorMsg)
	}
	return []models.ProjectFile{}, nil
}

// DeleteProjectFile implements MemoryClientInterface
func (m *MockMemoryClient) DeleteProjectFile(ctx context.Context, path string) error {
	m.DeleteProjectFileCalled = true
	if m.ReturnError {
		return errors.New(m.ErrorMsg)
	}
	return nil
}

// DeleteAllProjectFiles implements MemoryClientInterface
func (m *MockMemoryClient) DeleteAllProjectFiles(ctx context.Context) error {
	m.DeleteAllFilesCalled = true
	if m.ReturnError {
		return errors.New(m.ErrorMsg)
	}
	return nil
}

// ListProjectFiles implements MemoryClientInterface
func (m *MockMemoryClient) ListProjectFiles(ctx context.Context, limit int) ([]models.ProjectFile, error) {
	m.ListProjectFilesCalled = true
	if m.ReturnError {
		return nil, errors.New(m.ErrorMsg)
	}
	return []models.ProjectFile{}, nil
}
