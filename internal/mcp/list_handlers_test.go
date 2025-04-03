package mcp

import (
	"context"
	"encoding/json"
	"testing"
)

// TestListToolsRequest tests the handleListToolsRequest function
func TestListToolsRequest(t *testing.T) {
	// Create a new server with a mock client
	mock := NewMockClient(false, "")
	server := &MCPServer{client: mock}

	// Call the handler
	resp, err := server.handleListToolsRequest(context.Background(), "test-id")

	// Check for errors
	if err != nil {
		t.Errorf("handleListToolsRequest() error = %v, want nil", err)
		return
	}

	// Check response properties
	if !resp.Success {
		t.Errorf("handleListToolsRequest() success = %v, want true", resp.Success)
	}

	if resp.Type != "tools_list" {
		t.Errorf("handleListToolsRequest() type = %v, want tools_list", resp.Type)
	}

	if resp.ID != "test-id" {
		t.Errorf("handleListToolsRequest() id = %v, want test-id", resp.ID)
	}

	// Unmarshal the tools from the response data
	var tools []MCPTool
	err = json.Unmarshal(resp.Data, &tools)
	if err != nil {
		t.Errorf("Failed to unmarshal tools: %v", err)
		return
	}

	// Check that we have the expected number of tools
	expectedTools := 4 // add_message, get_conversation_history, search_similar_messages, get_milestones
	if len(tools) != expectedTools {
		t.Errorf("Expected %d tools, got %d", expectedTools, len(tools))
	}

	// Check that each tool has the required fields
	for _, tool := range tools {
		if tool.Name == "" {
			t.Errorf("Tool name is empty")
		}
		if tool.Description == "" {
			t.Errorf("Tool description is empty")
		}
		if len(tool.InputSchema) == 0 {
			t.Errorf("Tool input schema is empty")
		}
	}
}

// TestListResourcesRequest tests the handleListResourcesRequest function
func TestListResourcesRequest(t *testing.T) {
	// Create a new server with a mock client
	mock := NewMockClient(false, "")
	server := &MCPServer{client: mock}

	// Call the handler
	resp, err := server.handleListResourcesRequest(context.Background(), "test-id")

	// Check for errors
	if err != nil {
		t.Errorf("handleListResourcesRequest() error = %v, want nil", err)
		return
	}

	// Check response properties
	if !resp.Success {
		t.Errorf("handleListResourcesRequest() success = %v, want true", resp.Success)
	}

	if resp.Type != "resources_list" {
		t.Errorf("handleListResourcesRequest() type = %v, want resources_list", resp.Type)
	}

	if resp.ID != "test-id" {
		t.Errorf("handleListResourcesRequest() id = %v, want test-id", resp.ID)
	}

	// Unmarshal the resources from the response data
	var resources []MCPResource
	err = json.Unmarshal(resp.Data, &resources)
	if err != nil {
		t.Errorf("Failed to unmarshal resources: %v", err)
		return
	}

	// Check that we have the expected number of resources
	expectedResources := 2 // conversation_history, milestones
	if len(resources) != expectedResources {
		t.Errorf("Expected %d resources, got %d", expectedResources, len(resources))
	}

	// Check that each resource has the required fields
	for _, resource := range resources {
		if resource.URI == "" {
			t.Errorf("Resource URI is empty")
		}
		if resource.Name == "" {
			t.Errorf("Resource name is empty")
		}
		if resource.Description == "" {
			t.Errorf("Resource description is empty")
		}
	}
}
