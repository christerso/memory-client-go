package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/christerso/memory-client-go/internal/models"
)

// Current conversation context
var (
	currentConversationTag string
	lastMessageTimestamp   time.Time
	messageBuffer          []models.Message
	messageBufferSize      = 5 // Number of messages to buffer before analysis
	taggingMode            string = "automatic" // Can be "automatic" or "manual"
)

// startAPIServer starts an HTTP API server for the MCP
func (s *MCPServer) startAPIServer(ctx context.Context) {
	mux := http.NewServeMux()

	// API endpoint for receiving messages
	mux.HandleFunc("/api/message", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read the request body
		body, err := io.ReadAll(r.Body)
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
		
		err = s.client.AddMessage(ctx, message)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to add message: %v", err), http.StatusInternalServerError)
			return
		}

		// Add to message buffer for analysis
		messageBuffer = append(messageBuffer, *message)
		
		// Check if we should analyze and tag messages
		if len(messageBuffer) >= messageBufferSize && taggingMode == "automatic" {
			// Copy the buffer to avoid race conditions
			bufferCopy := make([]models.Message, len(messageBuffer))
			copy(bufferCopy, messageBuffer)
			
			// Clear the buffer
			messageBuffer = []models.Message{} // Clear buffer after analysis
			
			// Analyze and tag the messages in a separate goroutine
			go s.analyzeAndTagMessages(ctx, bufferCopy)
		}

		// Log the operation
		s.logOperation("API Message Added", fmt.Sprintf("Role: %s, Content: %s (truncated)", messageRequest.Role, truncateString(messageRequest.Content, 50)), true)

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Message added successfully",
			"id":      message.ID,
		})
	})

	// API endpoint for setting conversation tag
	mux.HandleFunc("/api/set-conversation-tag", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read the request body
		body, err := io.ReadAll(r.Body)
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
		
		// Log the operation
		s.logOperation("Conversation Tag Set", fmt.Sprintf("Tag: %s", currentConversationTag), true)

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": fmt.Sprintf("Conversation tag set to '%s'", currentConversationTag),
		})
	})

	// API endpoint for getting current conversation tag
	mux.HandleFunc("/api/get-conversation-tag", func(w http.ResponseWriter, r *http.Request) {
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

	// API endpoint for MCP protocol requests
	mux.HandleFunc("/api/mcp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		// Parse the MCP request
		var mcpRequest MCPRequest
		err = json.Unmarshal(body, &mcpRequest)
		if err != nil {
			http.Error(w, "Failed to parse request JSON", http.StatusBadRequest)
			return
		}

		// Log the incoming request
		s.logOperation("API Request Received", fmt.Sprintf("Type: %s", mcpRequest.Type), true)

		// Handle the request
		response, err := s.handleRequest(r.Context(), &mcpRequest)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to handle request: %v", err), http.StatusInternalServerError)
			s.logOperation("API Request Handling", fmt.Sprintf("Failed to handle request of type %s: %v", mcpRequest.Type, err), false)
			return
		}

		// Log the successful response
		s.logOperation("API Response Sent", fmt.Sprintf("Type: %s, Success: true", mcpRequest.Type), true)

		// Return the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

	// API endpoint for setting tagging mode
	mux.HandleFunc("/api/set-tagging-mode", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		// Parse the mode request
		var modeRequest struct {
			Mode string `json:"mode"`
		}
		err = json.Unmarshal(body, &modeRequest)
		if err != nil {
			http.Error(w, "Failed to parse request JSON", http.StatusBadRequest)
			return
		}

		// Validate the mode
		if modeRequest.Mode != "automatic" && modeRequest.Mode != "manual" {
			http.Error(w, "Invalid mode. Must be 'automatic' or 'manual'", http.StatusBadRequest)
			return
		}

		// Set the tagging mode
		taggingMode = modeRequest.Mode

		// If switching to automatic and we have buffered messages, process them
		if taggingMode == "automatic" && len(messageBuffer) > 0 {
			// Copy the buffer to avoid race conditions
			bufferCopy := make([]models.Message, len(messageBuffer))
			copy(bufferCopy, messageBuffer)
			
			// Clear the buffer
			messageBuffer = []models.Message{} // Clear buffer after analysis
			
			// Analyze and tag the messages in a separate goroutine
			go s.analyzeAndTagMessages(ctx, bufferCopy)
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Tagging mode set to " + modeRequest.Mode,
		})
	})

	// API endpoint for getting tagging mode
	mux.HandleFunc("/api/get-tagging-mode", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Return the current tagging mode
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"mode": taggingMode,
		})
	})

	// Create HTTP server
	apiServer := &http.Server{
		Addr:    ":10010",
		Handler: mux,
	}

	// Store the server in the MCPServer struct
	s.apiServer = apiServer

	// Start HTTP server
	log.Println("Starting API server on :10010")
	s.logOperation("API Server", "Starting API server on :10010", true)

	go func() {
		if err := apiServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("API server error: %v", err)
			s.logOperation("API Server", fmt.Sprintf("API server error: %v", err), false)
		}
	}()

	// Shutdown the server when context is done
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		apiServer.Shutdown(shutdownCtx)
	}()
}

// analyzeAndTagMessages analyzes a batch of messages and tags them appropriately
func (s *MCPServer) analyzeAndTagMessages(ctx context.Context, messages []models.Message) {
	if len(messages) == 0 {
		return
	}

	// Extract message IDs and content for analysis
	messageIDs := make([]string, len(messages))
	combinedContent := ""
	
	for i, msg := range messages {
		messageIDs[i] = msg.ID
		combinedContent += string(msg.Role) + ": " + msg.Content + "\n"
	}

	// Simple keyword-based categorization
	// In a real implementation, this could use more sophisticated NLP techniques
	categories := map[string][]string{
		"technical": {"code", "programming", "bug", "error", "function", "class", "method", "variable", "golang", "c++", "flatbuffers", "gorm", "postgres", "uuid"},
		"planning":  {"plan", "schedule", "timeline", "milestone", "project", "task", "todo", "implement", "feature", "design"},
		"question":  {"how", "what", "why", "when", "where", "who", "which", "?", "explain", "help", "understand"},
		"feedback":  {"review", "feedback", "improve", "suggestion", "opinion", "think", "feel"},
	}

	// Count category matches
	categoryScores := make(map[string]int)
	for category, keywords := range categories {
		for _, keyword := range keywords {
			count := strings.Count(strings.ToLower(combinedContent), strings.ToLower(keyword))
			categoryScores[category] += count
		}
	}

	// Find the highest scoring category
	var bestCategory string
	var highestScore int
	
	for category, score := range categoryScores {
		if score > highestScore {
			highestScore = score
			bestCategory = category
		}
	}

	// Only tag if we have a clear category and it has a minimum score
	if bestCategory != "" && highestScore >= 2 {
		// Add the category tag to all messages
		tag := "category:" + bestCategory
		err := s.client.TagMessages(ctx, messageIDs, tag)
		if err != nil {
			log.Printf("Error tagging messages: %v", err)
			s.logOperation("Message Tagging", fmt.Sprintf("Failed to tag messages with '%s': %v", tag, err), false)
			return
		}
		
		s.logOperation("Message Tagging", fmt.Sprintf("Tagged %d messages with '%s'", len(messageIDs), tag), true)
	}
}

// Helper function to truncate strings for logging
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
