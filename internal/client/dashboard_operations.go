package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/user/memory-client-go/internal/models"
)

// ClearAllMemories clears all memories (messages and project files)
func (c *MemoryClient) ClearAllMemories(ctx context.Context) error {
	if c.verbose {
		fmt.Println("Clearing all memories")
	}
	
	// Recreate collection to clear all data
	return c.recreateCollection(ctx)
}

// ClearMessages clears all messages
func (c *MemoryClient) ClearMessages(ctx context.Context) error {
	if c.verbose {
		fmt.Println("Clearing all messages")
	}
	
	return c.DeleteAllMessages(ctx)
}

// ClearProjectFiles clears all project files
func (c *MemoryClient) ClearProjectFiles(ctx context.Context) error {
	if c.verbose {
		fmt.Println("Clearing all project files")
	}
	
	return c.DeleteAllProjectFiles(ctx)
}

// DeleteProjectFilesByTag deletes project files with a specific tag
func (c *MemoryClient) DeleteProjectFilesByTag(ctx context.Context, tag string) error {
	if c.verbose {
		fmt.Printf("Deleting project files with tag: %s\n", tag)
	}
	
	// Create filter for project files with the specified tag
	url := fmt.Sprintf("%s/collections/%s/points/delete", c.qdrantURL, c.collectionName)
	
	request := map[string]interface{}{
		"filter": map[string]interface{}{
			"must": []map[string]interface{}{
				{
					"key": "type",
					"match": map[string]interface{}{
						"value": "project_file",
					},
				},
				{
					"key": "tag",
					"match": map[string]interface{}{
						"value": tag,
					},
				},
			},
		},
	}
	
	jsonData, err := json.Marshal(request)
	if err != nil {
		return err
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete project files: %s - %s", resp.Status, string(body))
	}
	
	return nil
}

// ListProjectFilesByTag lists project files with a specific tag
func (c *MemoryClient) ListProjectFilesByTag(ctx context.Context, tag string, limit int) ([]models.ProjectFile, error) {
	if c.verbose {
		fmt.Printf("Listing project files with tag: %s\n", tag)
	}
	
	// Create filter for project files with the specified tag
	url := fmt.Sprintf("%s/collections/%s/points/scroll", c.qdrantURL, c.collectionName)
	
	request := map[string]interface{}{
		"filter": map[string]interface{}{
			"must": []map[string]interface{}{
				{
					"key": "type",
					"match": map[string]interface{}{
						"value": "project_file",
					},
				},
				{
					"key": "tag",
					"match": map[string]interface{}{
						"value": tag,
					},
				},
			},
		},
		"limit": limit,
		"with_payload": true,
	}
	
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list project files: %s - %s", resp.Status, string(body))
	}
	
	// Parse response
	var result struct {
		Result struct {
			Points []struct {
				ID      string                 `json:"id"`
				Payload map[string]interface{} `json:"payload"`
			} `json:"points"`
		} `json:"result"`
	}
	
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	
	// Convert to ProjectFile objects
	files := make([]models.ProjectFile, 0, len(result.Result.Points))
	for _, point := range result.Result.Points {
		file := models.ProjectFile{
			ID:        point.ID,
			Path:      point.Payload["path"].(string),
			Content:   point.Payload["content"].(string),
			Language:  point.Payload["language"].(string),
			Tag:       point.Payload["tag"].(string),
			ModTime:   int64(point.Payload["mod_time"].(float64)),
		}
		
		// Parse timestamp if available
		if ts, ok := point.Payload["timestamp"].(string); ok {
			timestamp, err := time.Parse(time.RFC3339, ts)
			if err == nil {
				file.Timestamp = timestamp
			}
		}
		
		files = append(files, file)
	}
	
	return files, nil
}
