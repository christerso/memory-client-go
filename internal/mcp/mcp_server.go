package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/christerso/memory-client-go/internal/models"
	"github.com/fasthttp/websocket"
	"github.com/qdrant/go-client/qdrant"
)

// MemoryClientInterface defines the interface for memory client operations
type MemoryClientInterface interface {
	AddMessage(ctx context.Context, message *models.Message) error
	GetConversationHistory(ctx context.Context, limit int, filter *models.HistoryFilter) ([]models.Message, error)
	SearchMessages(ctx context.Context, query string, limit int) ([]models.Message, error)
	GetMemoryStats(ctx context.Context) (*models.MemoryStats, error)
	DeleteMessage(ctx context.Context, id string) error
	DeleteAllMessages(ctx context.Context) error
	TagMessages(ctx context.Context, ids []string, tag string) error
	GetMessagesByTag(ctx context.Context, tag string, limit int) ([]models.Message, error)
	IndexProjectFiles(ctx context.Context, path string, tag string) (int, error)
	UpdateProjectFiles(ctx context.Context, path string) (int, int, error)
	SearchProjectFiles(ctx context.Context, query string, limit int) ([]models.ProjectFile, error)
	DeleteProjectFile(ctx context.Context, path string) error
	DeleteAllProjectFiles(ctx context.Context) error
	ListProjectFiles(ctx context.Context, limit int) ([]models.ProjectFile, error)
}

// MCPServer represents the MCP server implementation
type MCPServer struct {
	client          MemoryClientInterface
	qdrantClient    *QdrantWrapper
	stdin           *os.File
	stdout          *os.File
	httpServer      *http.Server
	apiServer       *http.Server
	startTime       time.Time
	requestsMu      sync.Mutex
	requestsHandled int
	recentOps       []OperationLog
	recentOpsMu     sync.Mutex
	maxRecentOps    int

	// VS Code extension state
	contexts   map[string]CodeContext // sessionID -> context
	contextsMu sync.Mutex
	threads    map[string]Thread // threadID -> thread
	threadsMu  sync.Mutex
}

// OperationLog represents a log of a recent operation
type OperationLog struct {
	Timestamp time.Time
	Operation string
	Details   string
	Success   bool
}

// NewMCPServer creates a new MCP server
func NewMCPServer(client MemoryClientInterface, qdrantClient *qdrant.Client) *MCPServer {
	// Wrap the qdrantClient with our QdrantWrapper
	var qdrantWrapper *QdrantWrapper
	if qdrantClient != nil {
		qdrantWrapper = NewQdrantWrapper(qdrantClient)
	}

	return &MCPServer{
		client:       client,
		qdrantClient: qdrantWrapper,
		stdin:        os.Stdin,
		stdout:       os.Stdout,
		startTime:    time.Now(),
		maxRecentOps: 50, // Keep track of last 50 operations
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

	// Start HTTP server for status checks and API access
	go s.startHTTPServer(ctx)

	// Start API server for external clients
	go s.startAPIServer(ctx)

	// Log server start
	s.logOperation("Server Start", "MCP server started", true)

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
			s.logOperation("Server Shutdown", "MCP server shutting down", true)
			return ctx.Err()
		default:
			var request MCPRequest
			err := decoder.Decode(&request)
			if err != nil {
				log.Printf("Error decoding request: %v", err)
				s.logOperation("Request Decode", fmt.Sprintf("Failed to decode request: %v", err), false)
				continue
			}

			// Log the incoming request
			s.logOperation("Request Received", fmt.Sprintf("Type: %s", request.Type), true)

			response, err := s.handleRequest(ctx, &request)
			if err != nil {
				log.Printf("Error handling request: %v", err)
				s.logOperation("Request Handling", fmt.Sprintf("Failed to handle request of type %s: %v", request.Type, err), false)
				s.sendErrorResponse(request.ID, err)
				continue
			}

			// Log the successful response
			s.logOperation("Response Sent", fmt.Sprintf("Type: %s, Success: true", request.Type), true)

			err = s.sendResponse(response)
			if err != nil {
				log.Printf("Error sending response: %v", err)
				s.logOperation("Response Sending", fmt.Sprintf("Failed to send response: %v", err), false)
			}

			// Increment request counter
			s.requestsMu.Lock()
			s.requestsHandled++
			s.requestsMu.Unlock()
		}
	}
}

// logOperation logs an operation to the recent operations list
func (s *MCPServer) logOperation(operation, details string, success bool) {
	s.recentOpsMu.Lock()
	defer s.recentOpsMu.Unlock()

	// Add new operation log
	s.recentOps = append(s.recentOps, OperationLog{
		Timestamp: time.Now(),
		Operation: operation,
		Details:   details,
		Success:   success,
	})

	// Trim if exceeding max size
	if len(s.recentOps) > s.maxRecentOps {
		s.recentOps = s.recentOps[len(s.recentOps)-s.maxRecentOps:]
	}
}

// getRecentOperations returns the recent operations
func (s *MCPServer) getRecentOperations() []OperationLog {
	s.recentOpsMu.Lock()
	defer s.recentOpsMu.Unlock()

	// Return a copy to avoid race conditions
	result := make([]OperationLog, len(s.recentOps))
	copy(result, s.recentOps)

	// Reverse the order so newest are first
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// startHTTPServer starts the HTTP server for status checks and API access
func (s *MCPServer) startHTTPServer(ctx context.Context) {
	mux := http.NewServeMux()

	// Add WebSocket endpoint for VS Code extension
	mux.HandleFunc("/api/vscode/ws", func(w http.ResponseWriter, r *http.Request) {
		s.handleVSCodeWebSocket(w, r)
	})

	// Add status endpoint for JSON API
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		s.requestsMu.Lock()
		requestCount := s.requestsHandled
		s.requestsMu.Unlock()

		uptime := time.Since(s.startTime).Round(time.Second)

		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		status := map[string]interface{}{
			"status":            "running",
			"uptime":            uptime.String(),
			"start_time":        s.startTime.Format(time.RFC3339),
			"requests_handled":  requestCount,
			"memory_usage_mb":   float64(memStats.Alloc) / 1024 / 1024,
			"goroutines":        runtime.NumGoroutine(),
			"recent_operations": s.getRecentOperations(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	// Add health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Add web UI status page
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		s.serveStatusPageMCP(w, r)
	})

	// Redirect root to status page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/status", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})

	s.httpServer = &http.Server{
		Addr:    ":9580",
		Handler: mux,
	}

	log.Printf("Starting HTTP server on :9580")
	s.logOperation("HTTP Server", "Started HTTP server on port 9580", true)

	// Start the server in a goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
			s.logOperation("HTTP Server", fmt.Sprintf("HTTP server error: %v", err), false)
		}
	}()

	// Shutdown the server when context is done
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		log.Printf("Shutting down HTTP server")
		s.logOperation("HTTP Server", "Shutting down HTTP server", true)
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
			s.logOperation("HTTP Server", fmt.Sprintf("HTTP server shutdown error: %v", err), false)
		}
	}()
}

// serveStatusPageMCP serves the HTML status page for MCP
func (s *MCPServer) serveStatusPageMCP(w http.ResponseWriter, r *http.Request) {
	s.requestsMu.Lock()
	requestCount := s.requestsHandled
	s.requestsMu.Unlock()

	uptime := time.Since(s.startTime).Round(time.Second)

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Get memory stats
	memoryUsageMB := float64(memStats.Alloc) / 1024 / 1024

	// Get recent operations
	recentOps := s.getRecentOperations()

	// Create data for template
	data := map[string]interface{}{
		"Status":          "Running",
		"Uptime":          uptime.String(),
		"StartTime":       s.startTime.Format(time.RFC3339),
		"RequestsHandled": requestCount,
		"MemoryUsageMB":   fmt.Sprintf("%.2f", memoryUsageMB),
		"Goroutines":      runtime.NumGoroutine(),
		"RecentOps":       recentOps,
		"RefreshTime":     time.Now().Format(time.RFC3339),
	}

	// HTML template for the status page
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>MCP Server Status</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        :root {
            color-scheme: dark;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            line-height: 1.6;
            color: #e0e0e0;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background-color: #121212 !important;
            background-image: linear-gradient(rgba(255, 193, 7, 0.05) 1px, transparent 1px), 
                             linear-gradient(90deg, rgba(255, 193, 7, 0.05) 1px, transparent 1px);
            background-size: 20px 20px;
        }
        h1, h2 {
            color: #4dabf7;
        }
        .card {
            background: #1e1e1e;
            border-radius: 5px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.3);
            padding: 20px;
            margin-bottom: 20px;
            border: 1px solid #333;
        }
        .status-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 20px;
        }
        .status-item {
            background: #2d2d2d;
            padding: 15px;
            border-radius: 5px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.3);
            border: 1px solid #333;
        }
        .status-label {
            font-weight: bold;
            color: #adb5bd;
            font-size: 0.9em;
            margin-bottom: 5px;
        }
        .status-value {
            font-size: 1.4em;
            color: #4dabf7;
        }
        table {
            width: 100%;
            border-collapse: collapse;
        }
        th, td {
            padding: 12px 15px;
            text-align: left;
            border-bottom: 1px solid #333;
        }
        th {
            background-color: #2d2d2d;
            color: #e0e0e0;
        }
        tr:hover {
            background-color: #2d2d2d;
        }
        .success {
            color: #28a745;
        }
        .failure {
            color: #dc3545;
        }
        .refresh-bar {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 20px;
        }
        .refresh-button {
            background-color: #4dabf7;
            color: white;
            border: none;
            padding: 8px 16px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
        }
        .refresh-button:hover {
            background-color: #3c8dbc;
        }
        .last-refresh {
            color: #adb5bd;
            font-size: 0.9em;
        }
        @media (max-width: 768px) {
            .status-grid {
                grid-template-columns: 1fr;
            }
        }
        
        /* Debug indicator */
        .debug-indicator {
            position: fixed;
            top: 10px;
            right: 10px;
            background-color: #dc3545;
            color: white;
            padding: 5px 10px;
            border-radius: 4px;
            z-index: 1000;
            display: none;
        }
    </style>
    <script>
        function setupAutoRefresh() {
            // Auto refresh every 30 seconds
            setTimeout(function() {
                window.location.reload();
            }, 30000);
        }
        
        window.onload = function() {
            setupAutoRefresh();
            
            // Debug logging
            console.log('Dark mode active');
            
            // Force dark mode
            document.documentElement.style.colorScheme = 'dark';
            document.body.style.backgroundColor = '#121212';
            document.body.style.color = '#e0e0e0';
        };
    </script>
</head>
<body>
    <div class="debug-indicator">DARK MODE ACTIVE</div>
    <div class="card">
        <h1>MCP Server Status</h1>
        
        <div class="refresh-bar">
            <button class="refresh-button" onclick="window.location.reload();">Refresh Now</button>
            <span class="last-refresh">Last refreshed: {{.RefreshTime}}</span>
        </div>
        
        <div class="status-grid">
            <div class="status-item">
                <div class="status-label">Status</div>
                <div class="status-value">{{.Status}}</div>
            </div>
            <div class="status-item">
                <div class="status-label">Uptime</div>
                <div class="status-value">{{.Uptime}}</div>
            </div>
            <div class="status-item">
                <div class="status-label">Start Time</div>
                <div class="status-value">{{.StartTime}}</div>
            </div>
            <div class="status-item">
                <div class="status-label">Requests Handled</div>
                <div class="status-value">{{.RequestsHandled}}</div>
            </div>
            <div class="status-item">
                <div class="status-label">Memory Usage</div>
                <div class="status-value">{{.MemoryUsageMB}} MB</div>
            </div>
            <div class="status-item">
                <div class="status-label">Goroutines</div>
                <div class="status-value">{{.Goroutines}}</div>
            </div>
        </div>
    </div>
    
    <div class="card">
        <h2>Recent Operations</h2>
        <table>
            <thead>
                <tr>
                    <th>Time</th>
                    <th>Operation</th>
                    <th>Details</th>
                    <th>Status</th>
                </tr>
            </thead>
            <tbody>
                {{range .RecentOps}}
                <tr>
                    <td>{{.Timestamp.Format "15:04:05"}}</td>
                    <td>{{.Operation}}</td>
                    <td>{{.Details}}</td>
                    <td class="{{if .Success}}success{{else}}failure{{end}}">
                        {{if .Success}}Success{{else}}Failed{{end}}
                    </td>
                </tr>
                {{else}}
                <tr>
                    <td colspan="4">No operations recorded yet</td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>
</body>
</html>
`

	// Parse and execute the template
	t, err := template.New("status").Parse(tmpl)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing template: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
		return
	}
}

// handleRequest handles an MCP request
func (s *MCPServer) handleRequest(ctx context.Context, request *MCPRequest) (*MCPResponse, error) {
	switch request.Type {
	case "tool_call":
		return s.handleToolCall(ctx, request)
	case "resource_access":
		return s.handleResourceAccess(ctx, request)
	case "vscode_extension":
		return s.handleVSCodeRequest(ctx, request)
	case "list_tools_request":
		return s.handleListToolsRequest(ctx, request.ID)
	case "list_resources_request":
		return s.handleListResourcesRequest(ctx, request.ID)
	default:
		return nil, fmt.Errorf("unsupported request type: %s", request.Type)
	}
}

// handleVSCodeWebSocket handles WebSocket connections from VS Code
func (s *MCPServer) handleVSCodeWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	for {
		var msg VSCodeMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		response, err := s.processVSCodeMessage(r.Context(), msg)
		if err != nil {
			log.Printf("Error processing message: %v", err)
			conn.WriteJSON(VSCodeErrorResponse{
				Error: err.Error(),
			})
			continue
		}

		if err := conn.WriteJSON(response); err != nil {
			log.Printf("WebSocket write error: %v", err)
			break
		}
	}
}

// processVSCodeMessage handles messages from VS Code extension
func (s *MCPServer) processVSCodeMessage(ctx context.Context, msg VSCodeMessage) (interface{}, error) {
	switch msg.Type {
	case "store_context":
		return s.handleStoreContext(ctx, msg)
	case "get_context":
		return s.handleGetContext(ctx, msg)
	case "create_thread":
		return s.handleCreateThread(ctx, msg)
	case "get_threads":
		return s.handleGetThreads(ctx, msg)
	default:
		return nil, fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// handleStoreContext stores code context from VS Code
func (s *MCPServer) handleStoreContext(ctx context.Context, msg VSCodeMessage) (interface{}, error) {
	if msg.Context == nil {
		return nil, fmt.Errorf("missing context in store_context message")
	}

	s.contextsMu.Lock()
	defer s.contextsMu.Unlock()

	if s.contexts == nil {
		s.contexts = make(map[string]CodeContext)
	}
	s.contexts[msg.Context.SessionID] = *msg.Context

	return map[string]interface{}{
		"status":    "ok",
		"sessionId": msg.Context.SessionID,
		"storedAt":  time.Now().Format(time.RFC3339),
	}, nil
}

// handleGetContext retrieves stored code context
func (s *MCPServer) handleGetContext(ctx context.Context, msg VSCodeMessage) (interface{}, error) {
	if msg.Context == nil || msg.Context.SessionID == "" {
		return nil, fmt.Errorf("missing sessionID in get_context message")
	}

	s.contextsMu.Lock()
	defer s.contextsMu.Unlock()

	context, exists := s.contexts[msg.Context.SessionID]
	if !exists {
		return nil, fmt.Errorf("context not found for sessionID: %s", msg.Context.SessionID)
	}

	return map[string]interface{}{
		"status":  "ok",
		"context": context,
	}, nil
}

// handleCreateThread creates a new conversation thread
func (s *MCPServer) handleCreateThread(ctx context.Context, msg VSCodeMessage) (interface{}, error) {
	if msg.Context == nil {
		return nil, fmt.Errorf("missing context in create_thread message")
	}

	threadID := fmt.Sprintf("thread_%d", time.Now().UnixNano())
	newThread := Thread{
		ID:        threadID,
		Title:     fmt.Sprintf("Discussion about %s", msg.Context.File),
		CreatedAt: time.Now(),
		Context:   *msg.Context,
	}

	s.threadsMu.Lock()
	defer s.threadsMu.Unlock()

	if s.threads == nil {
		s.threads = make(map[string]Thread)
	}
	s.threads[threadID] = newThread

	return map[string]interface{}{
		"status": "ok",
		"thread": newThread,
	}, nil
}

// handleGetThreads lists all conversation threads
func (s *MCPServer) handleGetThreads(ctx context.Context, msg VSCodeMessage) (interface{}, error) {
	s.threadsMu.Lock()
	defer s.threadsMu.Unlock()

	threads := make([]Thread, 0, len(s.threads))
	for _, t := range s.threads {
		threads = append(threads, t)
	}

	// Sort threads by creation time (newest first)
	sort.Slice(threads, func(i, j int) bool {
		return threads[i].CreatedAt.After(threads[j].CreatedAt)
	})

	return map[string]interface{}{
		"status":  "ok",
		"threads": threads,
	}, nil
}

// handleVSCodeRequest handles MCP requests from VS Code extension
func (s *MCPServer) handleVSCodeRequest(ctx context.Context, request *MCPRequest) (*MCPResponse, error) {
	var vscodeReq VSCodeMessage
	err := json.Unmarshal(request.Data, &vscodeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal vscode request: %w", err)
	}

	result, err := s.processVSCodeMessage(ctx, vscodeReq)
	if err != nil {
		return nil, err
	}

	responseData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return &MCPResponse{
		ID:      request.ID,
		Type:    "vscode_response",
		Success: true,
		Data:    responseData,
	}, nil
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
		Role      string    `json:"role"`
		Content   string    `json:"content"`
		Embedding []float32 `json:"embedding"`
	}
	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, err
	}

	// Create message with embedding
	message := models.NewMessage(models.Role(params.Role), params.Content)
	message.Embedding = params.Embedding

	// Store in both memory client and Qdrant
	err = s.client.AddMessage(ctx, message)
	if err != nil {
		return nil, err
	}

	// Upsert to Qdrant vector storage if qdrantClient is available
	if s.qdrantClient != nil {
		err = s.qdrantClient.UpsertVector(ctx, message.ID, params.Embedding, map[string]interface{}{
			"role":    params.Role,
			"content": params.Content,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to upsert to Qdrant: %w", err)
		}
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

	messages, err := s.client.GetConversationHistory(ctx, params.Limit, nil)
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
	messages, err := s.client.GetConversationHistory(ctx, 100, nil) // Get last 100 messages
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
		Tag     string `json:"tag"`
		Verbose bool   `json:"verbose"`
	}
	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, err
	}

	// Index project files
	count, err := s.client.IndexProjectFiles(ctx, params.Path, params.Tag)
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

	if params.ID == "" {
		return nil, fmt.Errorf("missing required parameter 'id'")
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
		IDs []string `json:"ids"`
		Tag string   `json:"tag"`
	}
	err := json.Unmarshal(args, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Validate parameters
	if len(params.IDs) == 0 {
		return nil, fmt.Errorf("ids cannot be empty")
	}

	if params.Tag == "" {
		return nil, fmt.Errorf("tag cannot be empty")
	}

	// Tag messages
	err = s.client.TagMessages(ctx, params.IDs, params.Tag)
	if err != nil {
		return nil, fmt.Errorf("failed to tag messages: %w", err)
	}

	// Prepare response data
	responseData, err := json.Marshal(map[string]interface{}{
		"tagged_count": len(params.IDs),
		"tag":          params.Tag,
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

	// Validate parameters
	if params.Query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	if params.Summary == "" {
		return nil, fmt.Errorf("summary cannot be empty")
	}

	if len(params.Tags) == 0 {
		return nil, fmt.Errorf("tags cannot be empty")
	}

	// Set default limit if not provided
	if params.Limit <= 0 {
		params.Limit = 10
	}

	// First, search for messages matching the query
	messages, err := s.client.SearchMessages(ctx, params.Query, params.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search messages: %w", err)
	}

	// Extract message IDs
	var messageIDs []string
	for _, msg := range messages {
		messageIDs = append(messageIDs, msg.ID)
	}

	if len(messageIDs) == 0 {
		return nil, fmt.Errorf("no messages found matching the query")
	}

	// Tag each message with all the provided tags
	var taggedCount int
	for _, tag := range params.Tags {
		err = s.client.TagMessages(ctx, messageIDs, tag)
		if err != nil {
			return nil, fmt.Errorf("failed to tag messages with tag %s: %w", tag, err)
		}
		taggedCount += len(messageIDs)
	}

	// Prepare response data
	responseData, err := json.Marshal(map[string]interface{}{
		"summarized_and_tagged_count": taggedCount,
		"summary":                     params.Summary,
		"tags":                        params.Tags,
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

	if params.Tag == "" {
		return nil, fmt.Errorf("missing required parameter 'tag'")
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
						"tag": {
							"type": "string",
							"description": "Tag to apply to the indexed files"
						},
						"verbose": {
							"type": "boolean",
							"description": "Show detailed progress information"
						}
					},
					"required": ["path", "tag"]
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
						"ids": {
							"type": "array",
							"items": {
								"type": "string"
							},
							"description": "IDs of the messages to tag"
						},
						"tag": {
							"type": "string",
							"description": "Tag to add to the matching messages"
						}
					},
					"required": ["ids", "tag"]
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

// VS Code extension protocol types
type VSCodeMessage struct {
	Type    string          `json:"type"`
	Data    json.RawMessage `json:"data"`
	Context *CodeContext    `json:"context,omitempty"`
}

type CodeContext struct {
	File       string   `json:"file"`
	Lines      []int    `json:"lines"`
	Symbols    []string `json:"symbols"`
	LanguageID string   `json:"languageId"`
	SessionID  string   `json:"sessionId"`
}

type Thread struct {
	ID        string      `json:"id"`
	Title     string      `json:"title"`
	CreatedAt time.Time   `json:"createdAt"`
	Context   CodeContext `json:"context"`
}

type VSCodeErrorResponse struct {
	Error string `json:"error"`
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

// handleListToolsRequest handles a request to list available tools
func (s *MCPServer) handleListToolsRequest(ctx context.Context, requestID string) (*MCPResponse, error) {
	// Log the operation
	s.logOperation("List Tools Request", "Handling request to list available tools", true)

	// Get the tools from the server info
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
				Name:        "get_milestones",
				Description: "Retrieve milestones from the conversation",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"type": {
							"type": "string",
							"enum": ["personal_info", "preference", "action", "decision", "goal"],
							"description": "Type of milestones to retrieve (optional)"
						},
						"limit": {
							"type": "number",
							"description": "Maximum number of milestones to retrieve"
						}
					}
				}`),
			},
		},
	}

	// Marshal the tools to JSON
	responseData, err := json.Marshal(serverInfo.Tools)
	if err != nil {
		s.logOperation("List Tools Request", fmt.Sprintf("Failed to marshal tools: %v", err), false)
		return nil, err
	}

	// Return the response
	return &MCPResponse{
		ID:      requestID,
		Type:    "tools_list",
		Success: true,
		Data:    responseData,
	}, nil
}

// handleListResourcesRequest handles a request to list available resources
func (s *MCPServer) handleListResourcesRequest(ctx context.Context, requestID string) (*MCPResponse, error) {
	// Log the operation
	s.logOperation("List Resources Request", "Handling request to list available resources", true)

	// Get the resources from the server info
	resources := []MCPResource{
		{
			URI:         "memory:///conversation_history",
			Name:        "Conversation History",
			Description: "Complete history of the conversation",
		},
		{
			URI:         "memory:///milestones",
			Name:        "Milestones",
			Description: "Detected milestones from the conversation",
		},
	}

	// Marshal the resources to JSON
	responseData, err := json.Marshal(resources)
	if err != nil {
		s.logOperation("List Resources Request", fmt.Sprintf("Failed to marshal resources: %v", err), false)
		return nil, err
	}

	// Return the response
	return &MCPResponse{
		ID:      requestID,
		Type:    "resources_list",
		Success: true,
		Data:    responseData,
	}, nil
}
