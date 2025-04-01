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

// AddMessage adds a message to memory
func (c *MemoryClient) AddMessage(ctx context.Context, message *models.Message) error {
	// Generate embedding for message
	embedding, err := c.generateEmbedding(ctx, message.Content)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Create point
	point := map[string]interface{}{
		"id": message.ID,
		"vector": embedding,
		"payload": map[string]interface{}{
			"role":       message.Role,
			"content":    message.Content,
			"timestamp":  message.Timestamp.Format(time.RFC3339),
			"metadata":   message.Metadata,
			"tags":       message.Tags,
		},
	}

	// Add point to collection
	url := fmt.Sprintf("%s/collections/%s/points", c.qdrantURL, c.collectionName)
	
	request := map[string]interface{}{
		"points": []interface{}{point},
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
		return fmt.Errorf("failed to add point: %s - %s", resp.Status, string(body))
	}

	return nil
}

// GetConversationHistory retrieves conversation history
func (c *MemoryClient) GetConversationHistory(ctx context.Context, limit int, filter *models.HistoryFilter) ([]models.Message, error) {
	url := fmt.Sprintf("%s/collections/%s/points/scroll", c.qdrantURL, c.collectionName)

	// Build filter
	filterObj := map[string]interface{}{}
	if filter != nil {
		if !filter.StartTime.IsZero() || !filter.EndTime.IsZero() {
			dateFilter := map[string]interface{}{}
			
			if !filter.StartTime.IsZero() {
				dateFilter["gte"] = filter.StartTime.Format(time.RFC3339)
			}
			
			if !filter.EndTime.IsZero() {
				dateFilter["lte"] = filter.EndTime.Format(time.RFC3339)
			}
			
			filterObj["timestamp"] = dateFilter
		}
	}

	// Build request
	request := map[string]interface{}{
		"limit": limit,
		"with_payload": true,
		"with_vector": false,
	}

	if len(filterObj) > 0 {
		request["filter"] = map[string]interface{}{
			"must": []map[string]interface{}{
				{
					"payload": filterObj,
				},
			},
		}
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
		return nil, fmt.Errorf("failed to get conversation history: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Result struct {
			Points []struct {
				ID      string `json:"id"`
				Payload struct {
					Role      string                 `json:"role"`
					Content   string                 `json:"content"`
					Timestamp string                 `json:"timestamp"`
					Metadata  map[string]interface{} `json:"metadata"`
					Tags      []string               `json:"tags"`
				} `json:"payload"`
			} `json:"points"`
		} `json:"result"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	messages := make([]models.Message, 0, len(result.Result.Points))
	for _, point := range result.Result.Points {
		timestamp, err := time.Parse(time.RFC3339, point.Payload.Timestamp)
		if err != nil {
			timestamp = time.Now() // Fallback to current time if parsing fails
		}

		// Convert map[string]interface{} to map[string]string
		metadata := make(map[string]string)
		for k, v := range point.Payload.Metadata {
			if str, ok := v.(string); ok {
				metadata[k] = str
			} else {
				// Convert non-string values to string
				metadata[k] = fmt.Sprintf("%v", v)
			}
		}

		message := models.Message{
			ID:        point.ID,
			Role:      models.Role(point.Payload.Role),
			Content:   point.Payload.Content,
			Timestamp: timestamp,
			Metadata:  metadata,
			Tags:      point.Payload.Tags,
		}
		messages = append(messages, message)
	}

	return messages, nil
}

// SearchSimilarMessages searches for similar messages
func (c *MemoryClient) SearchSimilarMessages(ctx context.Context, query string, limit int) ([]models.Message, error) {
	// Generate embedding for query
	embedding, err := c.generateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Search for similar messages
	url := fmt.Sprintf("%s/collections/%s/points/search", c.qdrantURL, c.collectionName)

	request := map[string]interface{}{
		"vector":       embedding,
		"limit":        limit,
		"with_payload": true,
		"with_vector":  false,
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
		return nil, fmt.Errorf("failed to search similar messages: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Result []struct {
			ID      string  `json:"id"`
			Score   float64 `json:"score"`
			Payload struct {
				Role      string                 `json:"role"`
				Content   string                 `json:"content"`
				Timestamp string                 `json:"timestamp"`
				Metadata  map[string]interface{} `json:"metadata"`
				Tags      []string               `json:"tags"`
			} `json:"payload"`
		} `json:"result"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	messages := make([]models.Message, 0, len(result.Result))
	for _, item := range result.Result {
		timestamp, err := time.Parse(time.RFC3339, item.Payload.Timestamp)
		if err != nil {
			timestamp = time.Now() // Fallback to current time if parsing fails
		}

		// Convert map[string]interface{} to map[string]string
		metadata := make(map[string]string)
		for k, v := range item.Payload.Metadata {
			if str, ok := v.(string); ok {
				metadata[k] = str
			} else {
				// Convert non-string values to string
				metadata[k] = fmt.Sprintf("%v", v)
			}
		}

		message := models.Message{
			ID:        item.ID,
			Role:      models.Role(item.Payload.Role),
			Content:   item.Payload.Content,
			Timestamp: timestamp,
			Metadata:  metadata,
			Tags:      item.Payload.Tags,
			Score:     item.Score,
		}
		messages = append(messages, message)
	}

	return messages, nil
}

// DeleteMessage deletes a message by ID
func (c *MemoryClient) DeleteMessage(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/collections/%s/points/%s", c.qdrantURL, c.collectionName, id)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete message: %s - %s", resp.Status, string(body))
	}

	return nil
}

// DeleteAllMessages deletes all messages
func (c *MemoryClient) DeleteAllMessages(ctx context.Context) error {
	url := fmt.Sprintf("%s/collections/%s/points/delete", c.qdrantURL, c.collectionName)

	request := map[string]interface{}{
		"filter": map[string]interface{}{
			"must_not": []map[string]interface{}{
				{
					"payload": map[string]interface{}{
						"type": "project_file",
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
		return fmt.Errorf("failed to delete all messages: %s - %s", resp.Status, string(body))
	}

	return nil
}

// TagMessages tags messages with the given tag
func (c *MemoryClient) TagMessages(ctx context.Context, messageIDs []string, tag string) error {
	for _, id := range messageIDs {
		// Get message
		message, err := c.getMessage(ctx, id)
		if err != nil {
			return err
		}

		// Add tag if not already present
		hasTag := false
		for _, t := range message.Tags {
			if t == tag {
				hasTag = true
				break
			}
		}

		if !hasTag {
			message.Tags = append(message.Tags, tag)
			err = c.updateMessage(ctx, message)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetMessagesByTag gets messages with the given tag
func (c *MemoryClient) GetMessagesByTag(ctx context.Context, tag string, limit int) ([]models.Message, error) {
	url := fmt.Sprintf("%s/collections/%s/points/scroll", c.qdrantURL, c.collectionName)

	request := map[string]interface{}{
		"limit":        limit,
		"with_payload": true,
		"with_vector":  false,
		"filter": map[string]interface{}{
			"must": []map[string]interface{}{
				{
					"payload": map[string]interface{}{
						"tags": map[string]interface{}{
							"contains": tag,
						},
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
		return nil, fmt.Errorf("failed to get messages by tag: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Result struct {
			Points []struct {
				ID      string `json:"id"`
				Payload struct {
					Role      string                 `json:"role"`
					Content   string                 `json:"content"`
					Timestamp string                 `json:"timestamp"`
					Metadata  map[string]interface{} `json:"metadata"`
					Tags      []string               `json:"tags"`
				} `json:"payload"`
			} `json:"points"`
		} `json:"result"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	messages := make([]models.Message, 0, len(result.Result.Points))
	for _, point := range result.Result.Points {
		timestamp, err := time.Parse(time.RFC3339, point.Payload.Timestamp)
		if err != nil {
			timestamp = time.Now() // Fallback to current time if parsing fails
		}

		// Convert map[string]interface{} to map[string]string
		metadata := make(map[string]string)
		for k, v := range point.Payload.Metadata {
			if str, ok := v.(string); ok {
				metadata[k] = str
			} else {
				// Convert non-string values to string
				metadata[k] = fmt.Sprintf("%v", v)
			}
		}

		message := models.Message{
			ID:        point.ID,
			Role:      models.Role(point.Payload.Role),
			Content:   point.Payload.Content,
			Timestamp: timestamp,
			Metadata:  metadata,
			Tags:      point.Payload.Tags,
		}
		messages = append(messages, message)
	}

	return messages, nil
}

// Helper functions

// getMessage gets a message by ID
func (c *MemoryClient) getMessage(ctx context.Context, id string) (models.Message, error) {
	url := fmt.Sprintf("%s/collections/%s/points/%s", c.qdrantURL, c.collectionName, id)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return models.Message{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return models.Message{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return models.Message{}, fmt.Errorf("failed to get message: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Result struct {
			Payload struct {
				Role      string                 `json:"role"`
				Content   string                 `json:"content"`
				Timestamp string                 `json:"timestamp"`
				Metadata  map[string]interface{} `json:"metadata"`
				Tags      []string               `json:"tags"`
			} `json:"payload"`
		} `json:"result"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return models.Message{}, err
	}

	timestamp, err := time.Parse(time.RFC3339, result.Result.Payload.Timestamp)
	if err != nil {
		timestamp = time.Now() // Fallback to current time if parsing fails
	}

	// Convert map[string]interface{} to map[string]string
	metadata := make(map[string]string)
	for k, v := range result.Result.Payload.Metadata {
		if str, ok := v.(string); ok {
			metadata[k] = str
		} else {
			// Convert non-string values to string
			metadata[k] = fmt.Sprintf("%v", v)
		}
	}

	return models.Message{
		ID:        id,
		Role:      models.Role(result.Result.Payload.Role),
		Content:   result.Result.Payload.Content,
		Timestamp: timestamp,
		Metadata:  metadata,
		Tags:      result.Result.Payload.Tags,
	}, nil
}

// updateMessage updates a message
func (c *MemoryClient) updateMessage(ctx context.Context, message models.Message) error {
	url := fmt.Sprintf("%s/collections/%s/points", c.qdrantURL, c.collectionName)

	point := map[string]interface{}{
		"id": message.ID,
		"payload": map[string]interface{}{
			"role":       message.Role,
			"content":    message.Content,
			"timestamp":  message.Timestamp.Format(time.RFC3339),
			"metadata":   message.Metadata,
			"tags":       message.Tags,
		},
	}

	request := map[string]interface{}{
		"points": []interface{}{point},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonData))
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
		return fmt.Errorf("failed to update message: %s - %s", resp.Status, string(body))
	}

	return nil
}
