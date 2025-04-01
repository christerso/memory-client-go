package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/christerso/memory-client-go/internal/models"
)

// generateID generates a unique ID
func generateID() string {
	return uuid.New().String()
}

// generateEmbedding generates an embedding for text
func (c *MemoryClient) generateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// For now, we'll use a simple random embedding
	// In a real implementation, this would call an embedding API
	embedding := make([]float32, c.embeddingSize)
	for i := range embedding {
		embedding[i] = rand.Float32()*2 - 1 // Random value between -1 and 1
	}
	return embedding, nil
}

// SummarizeAndTagMessages summarizes messages in a time range and tags them
func (c *MemoryClient) SummarizeAndTagMessages(ctx context.Context, timeRange models.TimeRange, tag string) (string, error) {
	// Get messages in time range
	filter := &models.HistoryFilter{
		StartTime: timeRange.StartTime,
		EndTime:   timeRange.EndTime,
	}

	messages, err := c.GetConversationHistory(ctx, 1000, filter)
	if err != nil {
		return "", err
	}

	if len(messages) == 0 {
		return "No messages found in the specified time range", nil
	}

	// Extract message IDs
	messageIDs := make([]string, len(messages))
	for i, msg := range messages {
		messageIDs[i] = msg.ID
	}

	// Tag messages
	err = c.TagMessages(ctx, messageIDs, tag)
	if err != nil {
		return "", err
	}

	// Generate a simple summary
	summary := fmt.Sprintf("Tagged %d messages from %s to %s with tag '%s'",
		len(messages),
		timeRange.StartTime.Format(time.RFC3339),
		timeRange.EndTime.Format(time.RFC3339),
		tag)

	return summary, nil
}

// GetMemoryStats gets memory usage statistics
func (c *MemoryClient) GetMemoryStats(ctx context.Context) (*models.MemoryStats, error) {
	// Get collection info
	url := fmt.Sprintf("%s/collections/%s", c.qdrantURL, c.collectionName)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get collection info: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Result struct {
			VectorsCount int `json:"vectors_count"`
		} `json:"result"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	// Count message types
	messageCount, err := c.countMessagesByType(ctx)
	if err != nil {
		return nil, err
	}

	// Count project files
	projectFileCount, err := c.countProjectFiles(ctx)
	if err != nil {
		return nil, err
	}

	stats := &models.MemoryStats{
		TotalVectors:    result.Result.VectorsCount,
		MessageCount:    messageCount,
		ProjectFileCount: projectFileCount,
	}

	return stats, nil
}

// countMessagesByType counts messages by type
func (c *MemoryClient) countMessagesByType(ctx context.Context) (map[string]int, error) {
	url := fmt.Sprintf("%s/collections/%s/points/count", c.qdrantURL, c.collectionName)

	request := map[string]interface{}{
		"filter": map[string]interface{}{
			"must_not": []map[string]interface{}{
				{
					"key": "type",
					"match": map[string]interface{}{
						"value": "project_file",
					},
				},
			},
		},
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
		return nil, fmt.Errorf("failed to count messages: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Result struct {
			Count int `json:"count"`
		} `json:"result"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	// For now, we just return the total count
	// In a real implementation, we would count by role
	return map[string]int{
		"total": result.Result.Count,
	}, nil
}

// countProjectFiles counts project files
func (c *MemoryClient) countProjectFiles(ctx context.Context) (int, error) {
	url := fmt.Sprintf("%s/collections/%s/points/count", c.qdrantURL, c.collectionName)

	request := map[string]interface{}{
		"filter": map[string]interface{}{
			"must": []map[string]interface{}{
				{
					"key": "type",
					"match": map[string]interface{}{
						"value": "project_file",
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("failed to count project files: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Result struct {
			Count int `json:"count"`
		} `json:"result"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return 0, err
	}

	return result.Result.Count, nil
}
