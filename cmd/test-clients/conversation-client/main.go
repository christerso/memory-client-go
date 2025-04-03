package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Simple client to test sending conversation messages to the MCP server

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: test-conversation-client [role] [content]")
		fmt.Println("Example: test-conversation-client user \"Hello, how are you?\"")
		os.Exit(1)
	}

	role := os.Args[1]
	content := os.Args[2]

	// MCP server URL (from config)
	mcpServerURL := "http://localhost:8080"

	// Create the request payload
	payload := map[string]interface{}{
		"id":   fmt.Sprintf("req-%d", time.Now().UnixNano()),
		"type": "tool_call",
		"data": map[string]interface{}{
			"name": "add_message",
			"arguments": map[string]string{
				"role":    role,
				"content": content,
			},
		},
	}

	// Convert to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	// Send the request to the MCP server
	resp, err := http.Post(mcpServerURL+"/api/mcp", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending request to MCP server: %v\n", err)
		fmt.Println("Make sure the MCP server is running (memory-client mcp)")
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read and print the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error response from server: %s\n", string(body))
		os.Exit(1)
	}

	fmt.Println("Message successfully sent to memory client!")
	fmt.Printf("Role: %s\n", role)
	fmt.Printf("Content: %s\n", content)
}
