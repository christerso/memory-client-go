package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// MemoryClientInterface defines the interface for memory client operations
type MemoryClientInterface interface {
	AddMessage(ctx context.Context, message *Message) error
	GetConversationHistory(ctx context.Context, limit int) ([]*Message, error)
	SearchMessages(ctx context.Context, query string, limit int) ([]*Message, error)
	GetMemoryStats(ctx context.Context) (map[string]interface{}, error)
	DeleteMessage(ctx context.Context, id string) error
	DeleteAllMessages(ctx context.Context) error
	TagMessages(ctx context.Context, query string, tags []string, limit int) (int, error)
	SummarizeAndTagMessages(ctx context.Context, query string, summary string, tags []string, limit int) (int, error)
	GetMessagesByTag(ctx context.Context, tag string, limit int) ([]*Message, error)
	IndexProjectFiles(ctx context.Context, path string) (int, error)
	UpdateProjectFiles(ctx context.Context, path string) (int, int, error)
	SearchProjectFiles(ctx context.Context, query string, limit int) ([]*ProjectFile, error)
	DeleteProjectFile(ctx context.Context, path string) error
	DeleteAllProjectFiles(ctx context.Context) error
}

// MCPServer represents the MCP server implementation
type MCPServer struct {
	client MemoryClientInterface
	stdin  *os.File
	stdout *os.File
}

// NewMCPServer creates a new MCP server
func NewMCPServer(client MemoryClientInterface) *MCPServer {
	return &MCPServer{
		client: client,
		stdin:  os.Stdin,
		stdout: os.Stdout,
	}
}

// Start starts the MCP server
func (s *MCPServer) Start(ctx context.Context) error {
	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Set up signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	// Send server info
	err := s.sendServerInfo()
	if err != nil {
		return err
	}

	// Process incoming requests
	decoder := json.NewDecoder(s.stdin)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			var request MCPRequest
			err := decoder.Decode(&request)
			if err != nil {
				log.Printf("Error decoding request: %v", err)
				continue
			}

			response, err := s.handleRequest(ctx, &request)
			if err != nil {
				log.Printf("Error handling request: %v", err)
				s.sendErrorResponse(request.ID, err)
				continue
			}

			err = s.sendResponse(response)
			if err != nil {
				log.Printf("Error sending response: %v", err)
			}
		}
	}
}

// sendServerInfo sends the server info to the client
func (s *MCPServer) sendServerInfo() error {
	serverInfo := MCPServerInfo{
		Name:        "memory-server",
		Version:     "1.0.0",
		Description: "Memory server for conversation history",
		Tools: []MCPTool{
			{
				Name:        "add_message",
				Description: "Add a message to the conversation history",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"role": {
							"type": "string",
							"enum": ["user", "assistant", "system"],
							"description": "Role of the message sender"
						},
						"content": {
							"type": "string",
							"description": "Content of the message"
						}
					},
					"required": ["role", "content"]
				}`),
			},
			{
				Name:        "get_conversation_history",
				Description: "Retrieve the conversation history",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"limit": {
							"type": "number",
							"description": "Maximum number of messages to retrieve"
						}
					}
				}`),
			},
			{
				Name:        "search_similar_messages",
				Description: "Search for messages similar to a query",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"query": {
							"type": "string",
							"description": "Query text to search for similar messages"
						},
						"limit": {
							"type": "number",
							"description": "Maximum number of similar messages to retrieve"
						}
					},
					"required": ["query"]
				}`),
			},
			{
				Name:        "index_project",
				Description: "Index files in a project directory",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"path": {
							"type": "string",
							"description": "Path to the project directory"
						},
						"verbose": {
							"type": "boolean",
							"description": "Show detailed progress information"
						}
					},
					"required": ["path"]
				}`),
			},
			{
				Name:        "update_project",
				Description: "Update modified files in a project directory",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"path": {
							"type": "string",
							"description": "Path to the project directory"
						},
						"verbose": {
							"type": "boolean",
							"description": "Show detailed progress information"
						}
					},
					"required": ["path"]
				}`),
			},
			{
				Name:        "search_project_files",
				Description: "Search for files in the project",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"query": {
							"type": "string",
							"description": "Query text to search for in project files"
						},
						"limit": {
							"type": "number",
							"description": "Maximum number of files to retrieve"
						}
					},
					"required": ["query"]
				}`),
			},
			{
				Name:        "get_memory_stats",
				Description: "Get statistics about memory usage",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {}
				}`),
			},
			{
				Name:        "delete_message",
				Description: "Delete a message from the conversation history by ID",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"id": {
							"type": "string",
							"description": "ID of the message to delete"
						}
					},
					"required": ["id"]
				}`),
			},
			{
				Name:        "delete_all_messages",
				Description: "Delete all messages from the conversation history",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {}
				}`),
			},
			{
				Name:        "delete_project_file",
				Description: "Delete a project file by path",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"path": {
							"type": "string",
							"description": "Path of the file to delete"
						}
					},
					"required": ["path"]
				}`),
			},
			{
				Name:        "delete_all_project_files",
				Description: "Delete all project files",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {}
				}`),
			},
			{
				Name:        "tag_messages",
				Description: "Add tags to messages matching a query",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"query": {
							"type": "string",
							"description": "Query text to search for messages to tag"
						},
						"tags": {
							"type": "array",
							"items": {
								"type": "string"
							},
							"description": "Tags to add to the matching messages"
						},
						"limit": {
							"type": "number",
							"description": "Maximum number of messages to tag"
						}
					},
					"required": ["query", "tags"]
				}`),
			},
			{
				Name:        "summarize_and_tag_messages",
				Description: "Summarize and tag messages matching a query",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"query": {
							"type": "string",
							"description": "Query text to search for messages to summarize and tag"
						},
						"summary": {
							"type": "string",
							"description": "Summary to add to the matching messages"
						},
						"tags": {
							"type": "array",
							"items": {
								"type": "string"
							},
							"description": "Tags to add to the matching messages"
						},
						"limit": {
							"type": "number",
							"description": "Maximum number of messages to summarize and tag"
						}
					},
					"required": ["query", "summary", "tags"]
				}`),
			},
			{
				Name:        "get_messages_by_tag",
				Description: "Retrieve messages with a specific tag",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"tag": {
							"type": "string",
							"description": "Tag to search for"
						},
						"limit": {
							"type": "number",
							"description": "Maximum number of messages to retrieve"
						}
					},
					"required": ["tag"]
				}`),
			},
		},
		Resources: []MCPResource{
			{
				URI:         "memory:///conversation_history",
				Name:        "Conversation History",
				Description: "Complete history of the conversation",
			},
			{
				URI:         "memory:///project_files",
				Name:        "Project Files",
				Description: "Source code and other files from the current project",
			},
		},
	}

	return json.NewEncoder(s.stdout).Encode(serverInfo)
}

// handleRequest handles an MCP request
func (s *MCPServer) handleRequest(ctx context.Context, request *MCPRequest) (*MCPResponse, error) {
	switch request.Type {
	case "tool_call":
		return s.handleToolCall(ctx, request)
	case "resource_access":
		return s.handleResourceAccess(ctx, request)
	default:
		return nil, fmt.Errorf("unsupported request type: %s", request.Type)
	}
}

// handleToolCall handles a tool call request
func (s *MCPServer) handleToolCall(ctx context.Context, request *MCPRequest) (*MCPResponse, error) {
	var toolCall MCPToolCall
	err := json.Unmarshal(request.Data, &toolCall)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool call: %w", err)
	}

	switch toolCall.Name {
	case "add_message":
		return s.handleAddMessage(ctx, request.ID, toolCall.Arguments)
	case "get_conversation_history":
		return s.handleGetConversationHistory(ctx, request.ID, toolCall.Arguments)
	case "search_similar_messages":
		return s.handleSearchSimilarMessages(ctx, request.ID, toolCall.Arguments)
	case "index_project":
		return s.handleIndexProject(ctx, request.ID, toolCall.Arguments)
	case "update_project":
		return s.handleUpdateProject(ctx, request.ID, toolCall.Arguments)
	case "search_project_files":
		return s.handleSearchProjectFiles(ctx, request.ID, toolCall.Arguments)
	case "get_memory_stats":
		return s.handleGetMemoryStats(ctx, request.ID, toolCall.Arguments)
	case "delete_message":
		return s.handleDeleteMessage(ctx, request.ID, toolCall.Arguments)
	case "delete_all_messages":
		return s.handleDeleteAllMessages(ctx, request.ID, toolCall.Arguments)
	case "delete_project_file":
		return s.handleDeleteProjectFile(ctx, request.ID, toolCall.Arguments)
	case "delete_all_project_files":
		return s.handleDeleteAllProjectFiles(ctx, request.ID, toolCall.Arguments)
	case "tag_messages":
		return s.handleTagMessages(ctx, request.ID, toolCall.Arguments)
	case "summarize_and_tag_messages":
		return s.handleSummarizeAndTagMessages(ctx, request.ID, toolCall.Arguments)
	case "get_messages_by_tag":
		return s.handleGetMessagesByTag(ctx, request.ID, toolCall.Arguments)
	default:
		return nil, fmt.Errorf("unsupported tool: %s", toolCall.Name)
	}
}

// handleAddMessage handles the add_message tool call
func (s *MCPServer) handleAddMessage(ctx context.Context, requestID string, args json.RawMessage) (*MCPResponse, error) {
	var params struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, err
	}

	message := NewMessage(Role(params.Role), params.Content)
	err = s.client.AddMessage(ctx, message)
	if err != nil {
		return nil, err
	}

	return &MCPResponse{
		ID:      requestID,
		Type:    "tool_call_result",
		Success: true,
		Data:    json.RawMessage(`{"success": true}`),
	}, nil
}

// handleGetConversationHistory handles the get_conversation_history tool call
func (s *MCPServer) handleGetConversationHistory(ctx context.Context, requestID string, args json.RawMessage) (*MCPResponse, error) {
	var params struct {
		Limit int `json:"limit"`
	}
	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, err
	}

	if params.Limit <= 0 {
		params.Limit = 10 // Default limit
	}

	messages, err := s.client.GetConversationHistory(ctx, params.Limit)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	type messageResponse struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	response := make([]messageResponse, 0, len(messages))
	for _, msg := range messages {
		response = append(response, messageResponse{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	return &MCPResponse{
		ID:      requestID,
		Type:    "tool_call_result",
		Success: true,
		Data:    responseData,
	}, nil
}

// handleSearchSimilarMessages handles the search_similar_messages tool call
func (s *MCPServer) handleSearchSimilarMessages(ctx context.Context, requestID string, args json.RawMessage) (*MCPResponse, error) {
	var params struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, err
	}

	if params.Limit <= 0 {
		params.Limit = 5 // Default limit
	}

	messages, err := s.client.SearchMessages(ctx, params.Query, params.Limit)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	type messageResponse struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	response := make([]messageResponse, 0, len(messages))
	for _, msg := range messages {
		response = append(response, messageResponse{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	return &MCPResponse{
		ID:      requestID,
		Type:    "tool_call_result",
		Success: true,
		Data:    responseData,
	}, nil
}

// handleResourceAccess handles a resource access request
func (s *MCPServer) handleResourceAccess(ctx context.Context, request *MCPRequest) (*MCPResponse, error) {
	var resourceAccess MCPResourceAccess
	err := json.Unmarshal(request.Data, &resourceAccess)
	if err != nil {
		return nil, err
	}

	switch resourceAccess.URI {
	case "memory:///conversation_history":
		return s.handleConversationHistoryResource(ctx, request.ID)
	case "memory:///project_files":
		return s.handleProjectFilesResource(ctx, request.ID)
	default:
		return nil, fmt.Errorf("unsupported resource URI: %s", resourceAccess.URI)
	}
}

// handleConversationHistoryResource handles the conversation_history resource access
func (s *MCPServer) handleConversationHistoryResource(ctx context.Context, requestID string) (*MCPResponse, error) {
	messages, err := s.client.GetConversationHistory(ctx, 100) // Get last 100 messages
	if err != nil {
		return nil, err
	}

	// Convert to response format
	type messageResponse struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	response := make([]messageResponse, 0, len(messages))
	for _, msg := range messages {
		response = append(response, messageResponse{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	return &MCPResponse{
		ID:      requestID,
		Type:    "resource_content",
		Success: true,
		Data:    responseData,
	}, nil
}

// handleProjectFilesResource handles the project_files resource access
func (s *MCPServer) handleProjectFilesResource(ctx context.Context, requestID string) (*MCPResponse, error) {
	// Get project files from the project collection
	files, err := s.client.SearchProjectFiles(ctx, "", 100) // Get up to 100 files
	if err != nil {
		return nil, err
	}

	// Convert to response format
	type fileResponse struct {
		Path     string `json:"path"`
		Language string `json:"language"`
		Content  string `json:"content"`
	}
	response := make([]fileResponse, 0, len(files))
	for _, file := range files {
		response = append(response, fileResponse{
			Path:     file.Path,
			Language: file.Language,
			Content:  file.Content,
		})
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	return &MCPResponse{
		ID:      requestID,
		Type:    "resource_content",
		Success: true,
		Data:    responseData,
	}, nil
}

// handleIndexProject handles the index_project tool call
func (s *MCPServer) handleIndexProject(ctx context.Context, requestID string, args json.RawMessage) (*MCPResponse, error) {
	var params struct {
		Path    string `json:"path"`
		Verbose bool   `json:"verbose"`
	}
	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, err
	}

	// Index project files
	count, err := s.client.IndexProjectFiles(ctx, params.Path)
	if err != nil {
		return nil, err
	}

	// Prepare response
	responseData, err := json.Marshal(map[string]interface{}{
		"count": count,
		"path":  params.Path,
	})
	if err != nil {
		return nil, err
	}

	return &MCPResponse{
		ID:      requestID,
		Type:    "tool_call_result",
		Success: true,
		Data:    responseData,
	}, nil
}

// handleUpdateProject handles the update_project tool call
func (s *MCPServer) handleUpdateProject(ctx context.Context, requestID string, args json.RawMessage) (*MCPResponse, error) {
	var params struct {
		Path    string `json:"path"`
		Verbose bool   `json:"verbose"`
	}
	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, err
	}

	// Update project files
	newCount, updateCount, err := s.client.UpdateProjectFiles(ctx, params.Path)
	if err != nil {
		return nil, err
	}

	// Prepare response
	responseData, err := json.Marshal(map[string]interface{}{
		"new_files":     newCount,
		"updated_files": updateCount,
		"path":          params.Path,
	})
	if err != nil {
		return nil, err
	}

	return &MCPResponse{
		ID:      requestID,
		Type:    "tool_call_result",
		Success: true,
		Data:    responseData,
	}, nil
}

// handleSearchProjectFiles handles the search_project_files tool call
func (s *MCPServer) handleSearchProjectFiles(ctx context.Context, requestID string, args json.RawMessage) (*MCPResponse, error) {
	var params struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, err
	}

	if params.Limit <= 0 {
		params.Limit = 10 // Default limit
	}

	// Search project files
	files, err := s.client.SearchProjectFiles(ctx, params.Query, params.Limit)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	type fileResponse struct {
		Path     string `json:"path"`
		Language string `json:"language"`
		Content  string `json:"content"`
		Excerpt  string `json:"excerpt"`
	}
	response := make([]fileResponse, 0, len(files))
	for _, file := range files {
		// Create a short excerpt
		excerpt := file.Content
		if len(excerpt) > 200 {
			excerpt = excerpt[:200] + "..."
		}

		response = append(response, fileResponse{
			Path:     file.Path,
			Language: file.Language,
			Content:  file.Content,
			Excerpt:  excerpt,
		})
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	return &MCPResponse{
		ID:      requestID,
		Type:    "tool_call_result",
		Success: true,
		Data:    responseData,
	}, nil
}

// handleGetMemoryStats handles the get_memory_stats tool call
func (s *MCPServer) handleGetMemoryStats(ctx context.Context, requestID string, args json.RawMessage) (*MCPResponse, error) {
	// Get memory stats
	stats, err := s.client.GetMemoryStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory stats: %w", err)
	}

	// Prepare response data
	responseData, err := json.Marshal(stats)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response data: %w", err)
	}

	// Return response
	return &MCPResponse{
		ID:      requestID,
		Type:    "tool_call_result",
		Success: true,
		Data:    responseData,
	}, nil
}

// handleDeleteMessage handles the delete_message tool call
func (s *MCPServer) handleDeleteMessage(ctx context.Context, requestID string, args json.RawMessage) (*MCPResponse, error) {
	// Parse arguments
	var params struct {
		ID string `json:"id"`
	}
	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Delete message
	err = s.client.DeleteMessage(ctx, params.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete message: %w", err)
	}

	// Return success response
	return &MCPResponse{
		ID:      requestID,
		Type:    "tool_call_result",
		Success: true,
		Data:    json.RawMessage(`{"deleted": true}`),
	}, nil
}

// handleDeleteAllMessages handles the delete_all_messages tool call
func (s *MCPServer) handleDeleteAllMessages(ctx context.Context, requestID string, args json.RawMessage) (*MCPResponse, error) {
	// Delete all messages
	err := s.client.DeleteAllMessages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to delete all messages: %w", err)
	}

	// Return success response
	return &MCPResponse{
		ID:      requestID,
		Type:    "tool_call_result",
		Success: true,
		Data:    json.RawMessage(`{"deleted": true}`),
	}, nil
}

// handleDeleteProjectFile handles the delete_project_file tool call
func (s *MCPServer) handleDeleteProjectFile(ctx context.Context, requestID string, args json.RawMessage) (*MCPResponse, error) {
	// Parse arguments
	var params struct {
		Path string `json:"path"`
	}
	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Delete project file
	err = s.client.DeleteProjectFile(ctx, params.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to delete project file: %w", err)
	}

	// Return success response
	return &MCPResponse{
		ID:      requestID,
		Type:    "tool_call_result",
		Success: true,
		Data:    json.RawMessage(`{"deleted": true}`),
	}, nil
}

// handleDeleteAllProjectFiles handles the delete_all_project_files tool call
func (s *MCPServer) handleDeleteAllProjectFiles(ctx context.Context, requestID string, args json.RawMessage) (*MCPResponse, error) {
	// Delete all project files
	err := s.client.DeleteAllProjectFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to delete all project files: %w", err)
	}

	// Return success response
	return &MCPResponse{
		ID:      requestID,
		Type:    "tool_call_result",
		Success: true,
		Data:    json.RawMessage(`{"deleted": true}`),
	}, nil
}

// handleTagMessages handles the tag_messages tool call
func (s *MCPServer) handleTagMessages(ctx context.Context, requestID string, args json.RawMessage) (*MCPResponse, error) {
	// Parse arguments
	var params struct {
		Query string   `json:"query"`
		Tags  []string `json:"tags"`
		Limit int      `json:"limit"`
	}
	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Set default limit if not provided
	if params.Limit <= 0 {
		params.Limit = 10
	}

	// Tag messages
	count, err := s.client.TagMessages(ctx, params.Query, params.Tags, params.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to tag messages: %w", err)
	}

	// Prepare response data
	responseData, err := json.Marshal(map[string]interface{}{
		"tagged_count": count,
		"tags":         params.Tags,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response data: %w", err)
	}

	// Return response
	return &MCPResponse{
		ID:      requestID,
		Type:    "tool_call_result",
		Success: true,
		Data:    responseData,
	}, nil
}

// handleSummarizeAndTagMessages handles the summarize_and_tag_messages tool call
func (s *MCPServer) handleSummarizeAndTagMessages(ctx context.Context, requestID string, args json.RawMessage) (*MCPResponse, error) {
	// Parse arguments
	var params struct {
		Query   string   `json:"query"`
		Summary string   `json:"summary"`
		Tags    []string `json:"tags"`
		Limit   int      `json:"limit"`
	}
	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Set default limit if not provided
	if params.Limit <= 0 {
		params.Limit = 10
	}

	// Summarize and tag messages
	count, err := s.client.SummarizeAndTagMessages(ctx, params.Query, params.Summary, params.Tags, params.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to summarize and tag messages: %w", err)
	}

	// Prepare response data
	responseData, err := json.Marshal(map[string]interface{}{
		"updated_count": count,
		"summary":       params.Summary,
		"tags":          params.Tags,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response data: %w", err)
	}

	// Return response
	return &MCPResponse{
		ID:      requestID,
		Type:    "tool_call_result",
		Success: true,
		Data:    responseData,
	}, nil
}

// handleGetMessagesByTag handles the get_messages_by_tag tool call
func (s *MCPServer) handleGetMessagesByTag(ctx context.Context, requestID string, args json.RawMessage) (*MCPResponse, error) {
	// Parse arguments
	var params struct {
		Tag   string `json:"tag"`
		Limit int    `json:"limit"`
	}
	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Set default limit if not provided
	if params.Limit <= 0 {
		params.Limit = 10
	}

	// Get messages by tag
	messages, err := s.client.GetMessagesByTag(ctx, params.Tag, params.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by tag: %w", err)
	}

	// Convert messages to response format
	responseMessages := make([]map[string]interface{}, len(messages))
	for i, msg := range messages {
		responseMessages[i] = map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
			"tags":    msg.Tags,
			"summary": msg.Summary,
		}
	}

	// Prepare response data
	responseData, err := json.Marshal(map[string]interface{}{
		"messages": responseMessages,
		"tag":      params.Tag,
		"count":    len(messages),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response data: %w", err)
	}

	// Return response
	return &MCPResponse{
		ID:      requestID,
		Type:    "tool_call_result",
		Success: true,
		Data:    responseData,
	}, nil
}

// sendErrorResponse sends an error response
func (s *MCPServer) sendErrorResponse(requestID string, err error) error {
	response := MCPResponse{
		ID:      requestID,
		Type:    "error",
		Success: false,
		Error:   err.Error(),
	}
	return json.NewEncoder(s.stdout).Encode(response)
}

// sendResponse sends a response
func (s *MCPServer) sendResponse(response *MCPResponse) error {
	return json.NewEncoder(s.stdout).Encode(response)
}

// MCP protocol types
type MCPServerInfo struct {
	Name        string        `json:"name"`
	Version     string        `json:"version"`
	Description string        `json:"description"`
	Tools       []MCPTool     `json:"tools"`
	Resources   []MCPResource `json:"resources"`
}

type MCPTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

type MCPResource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type MCPRequest struct {
	ID   string          `json:"id"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type MCPToolCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type MCPResourceAccess struct {
	URI string `json:"uri"`
}

type MCPResponse struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}
