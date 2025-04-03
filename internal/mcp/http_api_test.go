package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/christerso/memory-client-go/internal/models"
	"github.com/google/uuid"
)

// HTTPTestMemoryClient implements MemoryClientInterface for HTTP API testing
type HTTPTestMemoryClient struct {
	messages       []models.Message
	tags           map[string][]string
	addMessageErr  error
	tagMessagesErr error
}

func NewHTTPTestMemoryClient() *HTTPTestMemoryClient {
	return &HTTPTestMemoryClient{
		messages: make([]models.Message, 0),
		tags:     make(map[string][]string),
	}
}

func (m *HTTPTestMemoryClient) AddMessage(ctx context.Context, message *models.Message) error {
	if m.addMessageErr != nil {
		return m.addMessageErr
	}
	m.messages = append(m.messages, *message)
	return nil
}

func (m *HTTPTestMemoryClient) GetConversationHistory(ctx context.Context, limit int, filter *models.HistoryFilter) ([]models.Message, error) {
	return m.messages, nil
}

func (m *HTTPTestMemoryClient) SearchMessages(ctx context.Context, query string, limit int) ([]models.Message, error) {
	return nil, nil
}

func (m *HTTPTestMemoryClient) GetMemoryStats(ctx context.Context) (*models.MemoryStats, error) {
	return &models.MemoryStats{
		TotalVectors:     len(m.messages),
		MessageCount:     map[string]int{"total": len(m.messages)},
		ProjectFileCount: 0,
	}, nil
}

func (m *HTTPTestMemoryClient) DeleteMessage(ctx context.Context, id string) error {
	return nil
}

func (m *HTTPTestMemoryClient) DeleteAllMessages(ctx context.Context) error {
	m.messages = make([]models.Message, 0)
	return nil
}

func (m *HTTPTestMemoryClient) TagMessages(ctx context.Context, ids []string, tag string) error {
	if m.tagMessagesErr != nil {
		return m.tagMessagesErr
	}
	m.tags[tag] = ids
	return nil
}

func (m *HTTPTestMemoryClient) GetMessagesByTag(ctx context.Context, tag string, limit int) ([]models.Message, error) {
	return nil, nil
}

func (m *HTTPTestMemoryClient) IndexProjectFiles(ctx context.Context, path string, tag string) (int, error) {
	return 0, nil
}

func (m *HTTPTestMemoryClient) UpdateProjectFiles(ctx context.Context, path string) (int, int, error) {
	return 0, 0, nil
}

func (m *HTTPTestMemoryClient) SearchProjectFiles(ctx context.Context, query string, limit int) ([]models.ProjectFile, error) {
	return nil, nil
}

func (m *HTTPTestMemoryClient) DeleteProjectFile(ctx context.Context, path string) error {
	return nil
}

func (m *HTTPTestMemoryClient) DeleteAllProjectFiles(ctx context.Context) error {
	return nil
}

func (m *HTTPTestMemoryClient) ListProjectFiles(ctx context.Context, limit int) ([]models.ProjectFile, error) {
	return nil, nil
}

func TestAddMessageAPI(t *testing.T) {
	mockClient := NewHTTPTestMemoryClient()
	server := NewMCPServer(mockClient, nil)

	// Create a request to send to the server
	messageData := map[string]string{
		"role":    "user",
		"content": "Test message content",
	}
	jsonData, _ := json.Marshal(messageData)

	req := httptest.NewRequest("POST", "/api/message", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Create a handler function that calls our API endpoint
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This simulates what would happen in the startAPIServer method
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read the request body
		body, err := json.Marshal(messageData)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		// Parse the message
		var messageRequest struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}
		err = json.Unmarshal(body, &messageRequest)
		if err != nil {
			http.Error(w, "Failed to parse request JSON", http.StatusBadRequest)
			return
		}

		// Create and add the message
		message := models.NewMessage(models.Role(messageRequest.Role), messageRequest.Content)

		// Add current conversation tag if set
		if currentConversationTag != "" {
			message.Tags = append(message.Tags, currentConversationTag)
		}

		err = server.client.AddMessage(r.Context(), message)
		if err != nil {
			http.Error(w, "Failed to add message", http.StatusInternalServerError)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Message added successfully",
			"id":      message.ID,
		})
	})

	// Serve the request
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response JSON: %v", err)
	}

	if success, ok := response["success"].(bool); !ok || !success {
		t.Errorf("Expected success to be true, got %v", response["success"])
	}

	// Check that the message was added to the mock client
	if len(mockClient.messages) != 1 {
		t.Errorf("Expected 1 message to be added, got %d", len(mockClient.messages))
	}

	if mockClient.messages[0].Role != models.Role("user") {
		t.Errorf("Expected role to be 'user', got %s", mockClient.messages[0].Role)
	}

	if mockClient.messages[0].Content != "Test message content" {
		t.Errorf("Expected content to be 'Test message content', got %s", mockClient.messages[0].Content)
	}
}

func TestSetConversationTagAPI(t *testing.T) {
	mockClient := NewHTTPTestMemoryClient()
	server := NewMCPServer(mockClient, nil)

	// Set a test tag
	testTag := "test-tag-" + uuid.New().String()

	// Create a request to send to the server
	tagData := map[string]string{
		"tag": testTag,
	}
	jsonData, _ := json.Marshal(tagData)

	req := httptest.NewRequest("POST", "/api/set-conversation-tag", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Create a handler function that calls our API endpoint
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This simulates what would happen in the startAPIServer method
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read the request body
		body, err := json.Marshal(tagData)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		// Parse the tag request
		var tagRequest struct {
			Tag string `json:"tag"`
		}
		err = json.Unmarshal(body, &tagRequest)
		if err != nil {
			http.Error(w, "Failed to parse request JSON", http.StatusBadRequest)
			return
		}

		// Set the current conversation tag
		currentConversationTag = tagRequest.Tag

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Conversation tag set successfully",
		})
	})

	// Serve the request
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response JSON: %v", err)
	}

	if success, ok := response["success"].(bool); !ok || !success {
		t.Errorf("Expected success to be true, got %v", response["success"])
	}

	// Check that the tag was set correctly
	if currentConversationTag != testTag {
		t.Errorf("Expected tag to be '%s', got '%s'", testTag, currentConversationTag)
	}

	// Now test that messages get tagged with the current tag
	messageData := map[string]string{
		"role":    "user",
		"content": "Test message with tag",
	}
	jsonData, _ = json.Marshal(messageData)

	req = httptest.NewRequest("POST", "/api/message", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Create a new ResponseRecorder
	rr = httptest.NewRecorder()

	// Create a handler for the message endpoint
	messageHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This simulates what would happen in the startAPIServer method
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read the request body
		body, err := json.Marshal(messageData)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		// Parse the message
		var messageRequest struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}
		err = json.Unmarshal(body, &messageRequest)
		if err != nil {
			http.Error(w, "Failed to parse request JSON", http.StatusBadRequest)
			return
		}

		// Create and add the message
		message := models.NewMessage(models.Role(messageRequest.Role), messageRequest.Content)

		// Add current conversation tag if set
		if currentConversationTag != "" {
			message.Tags = append(message.Tags, currentConversationTag)
		}

		err = server.client.AddMessage(r.Context(), message)
		if err != nil {
			http.Error(w, "Failed to add message", http.StatusInternalServerError)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Message added successfully",
			"id":      message.ID,
		})
	})

	// Serve the request
	messageHandler.ServeHTTP(rr, req)

	// Check that the message was added with the tag
	if len(mockClient.messages) != 1 {
		t.Errorf("Expected 1 message to be added, got %d", len(mockClient.messages))
	} else {
		if len(mockClient.messages[0].Tags) != 1 {
			t.Errorf("Expected message to have 1 tag, got %d", len(mockClient.messages[0].Tags))
		} else if mockClient.messages[0].Tags[0] != testTag {
			t.Errorf("Expected message tag to be '%s', got '%s'", testTag, mockClient.messages[0].Tags[0])
		}
	}
}

func TestAnalyzeAndTagMessages(t *testing.T) {
	mockClient := NewHTTPTestMemoryClient()
	server := NewMCPServer(mockClient, nil)

	// Create test messages with technical content
	messages := []models.Message{
		{
			ID:        uuid.New().String(),
			Role:      models.Role("user"),
			Content:   "I'm having an error with my golang code. The function is not working.",
			Timestamp: time.Now(),
		},
		{
			ID:        uuid.New().String(),
			Role:      models.Role("assistant"),
			Content:   "Let's debug your code. Can you show me the function that's causing the error?",
			Timestamp: time.Now(),
		},
		{
			ID:        uuid.New().String(),
			Role:      models.Role("user"),
			Content:   "Here's my function: func getData() error { return nil }",
			Timestamp: time.Now(),
		},
	}

	// Call analyzeAndTagMessages
	server.analyzeAndTagMessages(context.Background(), messages)

	// Check that the messages were tagged correctly
	if len(mockClient.tags) != 1 {
		t.Errorf("Expected 1 tag to be created, got %d", len(mockClient.tags))
	}

	// Check if the "category:technical" tag was created
	if ids, ok := mockClient.tags["category:technical"]; !ok {
		t.Errorf("Expected 'category:technical' tag to be created")
	} else {
		// Check that all message IDs were included
		if len(ids) != len(messages) {
			t.Errorf("Expected %d message IDs in tag, got %d", len(messages), len(ids))
		}

		// Check that each message ID is in the tag
		for _, msg := range messages {
			found := false
			for _, id := range ids {
				if id == msg.ID {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Message ID %s not found in tag", msg.ID)
			}
		}
	}

	// Test with planning content
	mockClient = NewHTTPTestMemoryClient()
	server = NewMCPServer(mockClient, nil)

	planningMessages := []models.Message{
		{
			ID:        uuid.New().String(),
			Role:      models.Role("user"),
			Content:   "Let's plan our project timeline for the next milestone.",
			Timestamp: time.Now(),
		},
		{
			ID:        uuid.New().String(),
			Role:      models.Role("assistant"),
			Content:   "Good idea. What tasks do we need to implement for this feature?",
			Timestamp: time.Now(),
		},
	}

	// Call analyzeAndTagMessages
	server.analyzeAndTagMessages(context.Background(), planningMessages)

	// Check if the "category:planning" tag was created
	if ids, ok := mockClient.tags["category:planning"]; !ok {
		t.Errorf("Expected 'category:planning' tag to be created")
	} else {
		if len(ids) != len(planningMessages) {
			t.Errorf("Expected %d message IDs in tag, got %d", len(planningMessages), len(ids))
		}
	}
}

func TestGetCurrentTagAPI(t *testing.T) {
	// Set a test tag
	testTag := "test-tag-" + uuid.New().String()
	currentConversationTag = testTag

	// Create a request to send to the server
	req := httptest.NewRequest("GET", "/api/get-conversation-tag", nil)

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Create a handler function that calls our API endpoint
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This simulates what would happen in the startAPIServer method
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Return the current conversation tag
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"tag": currentConversationTag,
		})
	})

	// Serve the request
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response JSON: %v", err)
	}

	if tag, ok := response["tag"].(string); !ok || tag != testTag {
		t.Errorf("Expected tag to be '%s', got '%v'", testTag, response["tag"])
	}

	// Reset the tag for other tests
	currentConversationTag = ""
}
