package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// MCPRequest represents an MCP request
type MCPRequest struct {
	ID   string          `json:"id"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// MCPToolCall represents a tool call in an MCP request
type MCPToolCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

func main() {
	// Check if the MCP server is running
	resp, err := http.Get("http://localhost:8080/status")
	if err != nil {
		fmt.Printf("Error checking MCP server status: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("MCP server not running (status code: %d)\n", resp.StatusCode)
		os.Exit(1)
	}

	fmt.Println("MCP server is running!")

	// Test adding a message
	fmt.Println("\nTesting add_message:")
	addMessageArgs, _ := json.Marshal(map[string]interface{}{
		"role":    "user",
		"content": "This is a test message from the test client",
	})

	toolCall, _ := json.Marshal(MCPToolCall{
		Name:      "add_message",
		Arguments: addMessageArgs,
	})

	addRequest := MCPRequest{
		ID:   "test-add-message",
		Type: "tool_call",
		Data: toolCall,
	}

	// Send the request to stdin of the MCP server
	addRequestJSON, _ := json.Marshal(addRequest)
	fmt.Printf("Sending request: %s\n", string(addRequestJSON))

	// Since we can't directly write to the MCP server's stdin in this test,
	// we'll print the JSON that would be sent for manual testing
	fmt.Println("\nTo test manually, pipe this JSON to the MCP server:")
	fmt.Printf("echo '%s' | go run main.go mcp\n", string(addRequestJSON))

	// Test getting conversation history
	fmt.Println("\nTesting get_conversation_history:")
	historyArgs, _ := json.Marshal(map[string]interface{}{
		"limit": 10,
	})

	historyToolCall, _ := json.Marshal(MCPToolCall{
		Name:      "get_conversation_history",
		Arguments: historyArgs,
	})

	historyRequest := MCPRequest{
		ID:   "test-get-history",
		Type: "tool_call",
		Data: historyToolCall,
	}

	historyRequestJSON, _ := json.Marshal(historyRequest)
	fmt.Printf("Sending request: %s\n", string(historyRequestJSON))
	fmt.Println("\nTo test manually, pipe this JSON to the MCP server:")
	fmt.Printf("echo '%s' | go run main.go mcp\n", string(historyRequestJSON))

	// Test searching messages
	fmt.Println("\nTesting search_similar_messages:")
	searchArgs, _ := json.Marshal(map[string]interface{}{
		"query": "test",
		"limit": 5,
	})

	searchToolCall, _ := json.Marshal(MCPToolCall{
		Name:      "search_similar_messages",
		Arguments: searchArgs,
	})

	searchRequest := MCPRequest{
		ID:   "test-search",
		Type: "tool_call",
		Data: searchToolCall,
	}

	searchRequestJSON, _ := json.Marshal(searchRequest)
	fmt.Printf("Sending request: %s\n", string(searchRequestJSON))
	fmt.Println("\nTo test manually, pipe this JSON to the MCP server:")
	fmt.Printf("echo '%s' | go run main.go mcp\n", string(searchRequestJSON))

	fmt.Println("\nTest client completed")
}
