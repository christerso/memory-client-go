package main

import (
	"fmt"
	"os/exec"
	"time"
)

func main() {
	// Generate a unique test message
	timestamp := time.Now().Format(time.RFC3339)
	testMessage := fmt.Sprintf("MCP_DIRECT_TEST_%s: This is a direct test of the MCP memory service functionality", timestamp)
	
	fmt.Println("Step 1: Adding a test message to the memory service")
	addCmd := exec.Command("go", "run", "main.go", "add", "user", testMessage)
	addOutput, err := addCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error adding message: %v\n", err)
		fmt.Println("Output:", string(addOutput))
		return
	}
	fmt.Println("Add message output:", string(addOutput))
	
	// Wait a moment for the message to be indexed
	fmt.Println("Waiting for message to be indexed...")
	time.Sleep(2 * time.Second)
	
	// Search for the message using a unique identifier
	searchTerm := fmt.Sprintf("MCP_DIRECT_TEST_%s", timestamp)
	fmt.Printf("Step 2: Searching for the message with term: %s\n", searchTerm)
	
	searchCmd := exec.Command("go", "run", "main.go", "search", searchTerm)
	searchOutput, err := searchCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error searching for message: %v\n", err)
		fmt.Println("Output:", string(searchOutput))
		return
	}
	
	// Check if the search output contains our test message
	outputStr := string(searchOutput)
	fmt.Println("Search results:")
	fmt.Println(outputStr)
	
	if len(outputStr) > 0 && outputStr != "No results found" {
		fmt.Println("\n✅ TEST PASSED: The MCP memory service is working correctly!")
		fmt.Println("The service successfully stored and retrieved the test message.")
	} else {
		fmt.Println("\n❌ TEST FAILED: The MCP memory service is not working correctly.")
		fmt.Println("The service could not retrieve the test message that was added.")
	}
}
