package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
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

// MCPResponse represents an MCP response
type MCPResponse struct {
	ID     string          `json:"id"`
	Status string          `json:"status"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}

func main() {
	// Create a unique test message
	testMessage := fmt.Sprintf("Test message from MCP protocol test at %s", time.Now().Format(time.RFC3339))
	
	// Test adding a message
	fmt.Println("Testing add_message with:", testMessage)
	addMessageArgs, _ := json.Marshal(map[string]interface{}{
		"role":    "user",
		"content": testMessage,
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

	// Convert request to JSON
	addRequestJSON, _ := json.Marshal(addRequest)
	
	// Create a temporary file to store the request
	tmpFile, err := os.CreateTemp("", "mcp-request-*.json")
	if err != nil {
		fmt.Printf("Error creating temp file: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(tmpFile.Name())
	
	// Write the request to the temp file
	if _, err := tmpFile.Write(addRequestJSON); err != nil {
		fmt.Printf("Error writing to temp file: %v\n", err)
		os.Exit(1)
	}
	if err := tmpFile.Close(); err != nil {
		fmt.Printf("Error closing temp file: %v\n", err)
		os.Exit(1)
	}
	
	// Execute the MCP command with the request as input
	cmd := exec.Command("go", "run", "main.go", "mcp")
	cmd.Stdin, err = os.Open(tmpFile.Name())
	if err != nil {
		fmt.Printf("Error opening temp file for stdin: %v\n", err)
		os.Exit(1)
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error executing MCP command: %v\n", err)
		fmt.Println("Output:", string(output))
		os.Exit(1)
	}
	
	// Parse the response
	outputStr := string(output)
	fmt.Println("Raw output:", outputStr)
	
	// Find the JSON response in the output
	jsonStart := strings.Index(outputStr, "{")
	if jsonStart == -1 {
		fmt.Println("No JSON response found in output")
		os.Exit(1)
	}
	
	jsonStr := outputStr[jsonStart:]
	var response MCPResponse
	if err := json.Unmarshal([]byte(jsonStr), &response); err != nil {
		fmt.Printf("Error parsing JSON response: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Response: %+v\n", response)
	
	// Now search for the message we just added
	fmt.Println("\nTesting search_similar_messages for:", testMessage)
	
	// Extract a unique search term from the test message
	searchTerm := fmt.Sprintf("MCP protocol test at %s", time.Now().Format("2006-01-02"))
	
	searchArgs, _ := json.Marshal(map[string]interface{}{
		"query": searchTerm,
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
	
	// Convert search request to JSON
	searchRequestJSON, _ := json.Marshal(searchRequest)
	
	// Create a temporary file for the search request
	searchTmpFile, err := os.CreateTemp("", "mcp-search-*.json")
	if err != nil {
		fmt.Printf("Error creating temp file for search: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(searchTmpFile.Name())
	
	// Write the search request to the temp file
	if _, err := searchTmpFile.Write(searchRequestJSON); err != nil {
		fmt.Printf("Error writing to search temp file: %v\n", err)
		os.Exit(1)
	}
	if err := searchTmpFile.Close(); err != nil {
		fmt.Printf("Error closing search temp file: %v\n", err)
		os.Exit(1)
	}
	
	// Execute the MCP command with the search request as input
	searchCmd := exec.Command("go", "run", "main.go", "mcp")
	searchCmd.Stdin, err = os.Open(searchTmpFile.Name())
	if err != nil {
		fmt.Printf("Error opening search temp file for stdin: %v\n", err)
		os.Exit(1)
	}
	
	searchOutput, err := searchCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error executing MCP search command: %v\n", err)
		fmt.Println("Search Output:", string(searchOutput))
		os.Exit(1)
	}
	
	// Parse the search response
	searchOutputStr := string(searchOutput)
	fmt.Println("Raw search output:", searchOutputStr)
	
	// Find the JSON response in the search output
	searchJsonStart := strings.Index(searchOutputStr, "{")
	if searchJsonStart == -1 {
		fmt.Println("No JSON response found in search output")
		os.Exit(1)
	}
	
	searchJsonStr := searchOutputStr[searchJsonStart:]
	var searchResponse MCPResponse
	if err := json.Unmarshal([]byte(searchJsonStr), &searchResponse); err != nil {
		fmt.Printf("Error parsing JSON search response: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Search Response: %+v\n", searchResponse)
	
	// Try to parse the search results
	if searchResponse.Result != nil {
		var searchResults map[string]interface{}
		if err := json.Unmarshal(searchResponse.Result, &searchResults); err != nil {
			fmt.Printf("Error parsing search results: %v\n", err)
		} else {
			fmt.Printf("Search Results: %+v\n", searchResults)
		}
	}
	
	fmt.Println("\nMCP protocol test completed")
}
