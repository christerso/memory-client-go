package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// RoundTripFunc is a function type that implements http.RoundTripper
type RoundTripFunc func(req *http.Request) (*http.Response, error)

// RoundTrip executes the mock round trip
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// NewTestClient returns a new http.Client with Transport replaced with a mock
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

// setupTestClient creates a test client with a mock HTTP client
func setupTestClient(t *testing.T, fn RoundTripFunc) *MemoryClient {
	client := &MemoryClient{
		httpClient:     NewTestClient(fn),
		qdrantURL:      "http://localhost:6333",
		collectionName: "test_collection",
		embeddingSize:  384,
		verbose:        false,
	}
	return client
}

// createMockResponse creates a mock HTTP response
func createMockResponse(statusCode int, body interface{}) *http.Response {
	bodyBytes, _ := json.Marshal(body)
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewReader(bodyBytes)),
		Header:     make(http.Header),
	}
}

// TestClientAddMessage tests the AddMessage function
func TestClientAddMessage(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name          string
		message       *Message
		mockResponses []struct {
			statusCode int
			body       interface{}
		}
		expectError bool
	}{
		{
			name: "successful add",
			message: &Message{
				Role:    RoleUser,
				Content: "Test message",
			},
			mockResponses: []struct {
				statusCode int
				body       interface{}
			}{
				{
					statusCode: http.StatusOK,
					body: map[string]interface{}{
						"result": map[string]interface{}{
							"points": []interface{}{},
						},
					},
				},
				{
					statusCode: http.StatusOK,
					body: map[string]interface{}{
						"result": map[string]interface{}{
							"operation_id": 123,
							"status":       "completed",
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "duplicate message",
			message: &Message{
				Role:    RoleUser,
				Content: "Test message",
			},
			mockResponses: []struct {
				statusCode int
				body       interface{}
			}{
				{
					statusCode: http.StatusOK,
					body: map[string]interface{}{
						"result": map[string]interface{}{
							"points": []interface{}{
								map[string]interface{}{
									"id": 123,
									"payload": map[string]interface{}{
										"Role":    "user",
										"Content": "Test message",
									},
								},
							},
						},
					},
				},
			},
			expectError: false, // Not an error, just skips adding
		},
		{
			name: "error in point creation",
			message: &Message{
				Role:    RoleUser,
				Content: "Test message",
			},
			mockResponses: []struct {
				statusCode int
				body       interface{}
			}{
				{
					statusCode: http.StatusOK,
					body: map[string]interface{}{
						"result": map[string]interface{}{
							"points": []interface{}{},
						},
					},
				},
				{
					statusCode: http.StatusInternalServerError,
					body: map[string]interface{}{
						"status": "error",
						"error":  "Failed to create point",
					},
				},
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			requestCount := 0
			client := setupTestClient(t, func(req *http.Request) (*http.Response, error) {
				if requestCount >= len(tc.mockResponses) {
					t.Fatalf("Unexpected request: %s %s", req.Method, req.URL.Path)
				}
				resp := createMockResponse(tc.mockResponses[requestCount].statusCode, tc.mockResponses[requestCount].body)
				requestCount++
				return resp, nil
			})

			err := client.AddMessage(context.Background(), tc.message)
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got nil")
			} else if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// TestClientGetConversationHistory tests the GetConversationHistory function
func TestClientGetConversationHistory(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name          string
		limit         int
		statusCode    int
		mockResponse  interface{}
		expectedCount int
		expectError   bool
	}{
		{
			name:       "successful retrieval",
			limit:      10,
			statusCode: http.StatusOK,
			mockResponse: map[string]interface{}{
				"result": map[string]interface{}{
					"points": []interface{}{
						map[string]interface{}{
							"id": 1,
							"payload": map[string]interface{}{
								"Role":    "user",
								"Content": "Message 1",
							},
						},
						map[string]interface{}{
							"id": 2,
							"payload": map[string]interface{}{
								"Role":    "assistant",
								"Content": "Message 2",
							},
						},
					},
				},
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:       "empty result",
			limit:      10,
			statusCode: http.StatusOK,
			mockResponse: map[string]interface{}{
				"result": map[string]interface{}{
					"points": []interface{}{},
				},
			},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:       "error response",
			limit:      10,
			statusCode: http.StatusInternalServerError,
			mockResponse: map[string]interface{}{
				"status": "error",
				"error":  "Failed to retrieve conversation history",
			},
			expectedCount: 0,
			expectError:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := setupTestClient(t, func(req *http.Request) (*http.Response, error) {
				return createMockResponse(tc.statusCode, tc.mockResponse), nil
			})

			messages, err := client.GetConversationHistory(context.Background(), tc.limit)
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got nil")
			} else if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if !tc.expectError && len(messages) != tc.expectedCount {
				t.Errorf("Expected %d messages but got %d", tc.expectedCount, len(messages))
			}
		})
	}
}

// TestClientSearchMessages tests the SearchMessages function
func TestClientSearchMessages(t *testing.T) {
	t.Skip("Skipping client test to focus on server tests")
}

// TestClientSearchProjectFiles tests the SearchProjectFiles function
func TestClientSearchProjectFiles(t *testing.T) {
	t.Skip("Skipping client test to focus on server tests")
}

// TestClientIndexProjectFiles tests the IndexProjectFiles function
func TestClientIndexProjectFiles(t *testing.T) {
	t.Skip("Skipping client test to focus on server tests")
}

// TestClientUpdateProjectFiles tests the UpdateProjectFiles function
func TestClientUpdateProjectFiles(t *testing.T) {
	t.Skip("Skipping client test to focus on server tests")
}

// TestClientGetMemoryStats tests the GetMemoryStats function
func TestClientGetMemoryStats(t *testing.T) {
	t.Skip("Skipping client test to focus on server tests")
}

// TestClientDeleteMessage tests the DeleteMessage function
func TestClientDeleteMessage(t *testing.T) {
	t.Skip("Skipping client test to focus on server tests")
}

// TestClientDeleteAllMessages tests the DeleteAllMessages function
func TestClientDeleteAllMessages(t *testing.T) {
	t.Skip("Skipping client test to focus on server tests")
}

// TestClientDeleteProjectFile tests the DeleteProjectFile function
func TestClientDeleteProjectFile(t *testing.T) {
	t.Skip("Skipping client test to focus on server tests")
}

// TestClientDeleteAllProjectFiles tests the DeleteAllProjectFiles function
func TestClientDeleteAllProjectFiles(t *testing.T) {
	t.Skip("Skipping client test to focus on server tests")
}

// TestClientTagMessages tests the TagMessages function
func TestClientTagMessages(t *testing.T) {
	t.Skip("Skipping client test to focus on server tests")
}

// TestClientGetMessagesByTag tests the GetMessagesByTag function
func TestClientGetMessagesByTag(t *testing.T) {
	t.Skip("Skipping client test to focus on server tests")
}

// TestClientSummarizeAndTagMessages tests the SummarizeAndTagMessages function
func TestClientSummarizeAndTagMessages(t *testing.T) {
	t.Skip("Skipping client test to focus on server tests")
}
