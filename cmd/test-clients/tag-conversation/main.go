package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Simple client to test conversation tagging with the MCP server

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  Set conversation tag: test-tag-conversation set-tag [tag]")
		fmt.Println("  Get current tag: test-tag-conversation get-tag")
		fmt.Println("  Send message: test-tag-conversation send [role] [content]")
		fmt.Println("")
		fmt.Println("Examples:")
		fmt.Println("  test-tag-conversation set-tag project-planning")
		fmt.Println("  test-tag-conversation send user \"Let's discuss the database schema\"")
		os.Exit(1)
	}

	// MCP server URL
	mcpServerURL := "http://localhost:10010"

	command := os.Args[1]

	switch command {
	case "set-tag":
		if len(os.Args) < 3 {
			fmt.Println("Error: Missing tag parameter")
			fmt.Println("Usage: test-tag-conversation set-tag [tag]")
			os.Exit(1)
		}
		tag := os.Args[2]
		setConversationTag(mcpServerURL, tag)

	case "get-tag":
		getCurrentTag(mcpServerURL)

	case "send":
		if len(os.Args) < 4 {
			fmt.Println("Error: Missing role or content parameters")
			fmt.Println("Usage: test-tag-conversation send [role] [content]")
			os.Exit(1)
		}
		role := os.Args[2]
		content := os.Args[3]
		sendMessage(mcpServerURL, role, content)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func setConversationTag(serverURL string, tag string) {
	// Create the request payload
	payload := map[string]interface{}{
		"tag": tag,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	// Send the request to the MCP server
	resp, err := http.Post(serverURL+"/api/set-conversation-tag", "application/json", bytes.NewBuffer(jsonData))
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

	fmt.Printf("Successfully set conversation tag to '%s'\n", tag)
}

func getCurrentTag(serverURL string) {
	// Send the request to the MCP server
	resp, err := http.Get(serverURL + "/api/get-conversation-tag")
	if err != nil {
		fmt.Printf("Error sending request to MCP server: %v\n", err)
		fmt.Println("Make sure the MCP server is running (memory-client mcp)")
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read and parse the response
	var result struct {
		Tag string `json:"tag"`
	}
	
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if result.Tag == "" {
		fmt.Println("No conversation tag is currently set")
	} else {
		fmt.Printf("Current conversation tag: %s\n", result.Tag)
	}
}

func sendMessage(serverURL string, role string, content string) {
	// Create the request payload
	payload := map[string]interface{}{
		"role":    role,
		"content": content,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	// Send the request to the MCP server
	resp, err := http.Post(serverURL+"/api/message", "application/json", bytes.NewBuffer(jsonData))
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
