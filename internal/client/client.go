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

	"github.com/user/memory-client-go/internal/models"
)

// MemoryClient represents a client for the Qdrant vector database
type MemoryClient struct {
	httpClient     *http.Client
	qdrantURL      string
	collectionName string
	embeddingSize  int
	verbose        bool
}

// NewMemoryClient creates a new memory client
func NewMemoryClient(qdrantURL, collectionName string, embeddingSize int, verbose bool) (*MemoryClient, error) {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Ensure URL has proper format
	if qdrantURL[len(qdrantURL)-1] == '/' {
		qdrantURL = qdrantURL[:len(qdrantURL)-1]
	}

	client := &MemoryClient{
		httpClient:     &http.Client{Timeout: 10 * time.Second},
		qdrantURL:      qdrantURL,
		collectionName: collectionName,
		embeddingSize:  embeddingSize,
		verbose:        verbose,
	}

	return client, nil
}

// Close closes the client
func (c *MemoryClient) Close() error {
	// Nothing to close for HTTP client
	return nil
}

// PurgeQdrant completely purges all data from Qdrant
func (c *MemoryClient) PurgeQdrant(ctx context.Context) error {
	if c.verbose {
		fmt.Println("Purging all data from Qdrant")
	}

	// Recreate collection
	return c.recreateCollection(ctx)
}

// DeleteMessagesByTimeRange deletes messages in a specific time range
func (c *MemoryClient) DeleteMessagesByTimeRange(ctx context.Context, from, to time.Time) (int, error) {
	if c.verbose {
		fmt.Printf("Deleting messages from %s to %s\n", from.Format(time.RFC3339), to.Format(time.RFC3339))
	}

	// Format time range for Qdrant
	fromStr := from.Format(time.RFC3339)
	toStr := to.Format(time.RFC3339)

	// Create filter for time range
	url := fmt.Sprintf("%s/collections/%s/points/delete", c.qdrantURL, c.collectionName)
	
	request := map[string]interface{}{
		"filter": map[string]interface{}{
			"must": []map[string]interface{}{
				{
					"must_not": []map[string]interface{}{
						{
							"payload": map[string]interface{}{
								"type": "project_file",
							},
						},
					},
				},
				{
					"range": map[string]interface{}{
						"timestamp": map[string]interface{}{
							"gte": fromStr,
							"lte": toStr,
						},
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
		return 0, fmt.Errorf("failed to delete messages: %s - %s", resp.Status, string(body))
	}

	// Parse response to get count of deleted messages
	var result struct {
		Result struct {
			Deleted int `json:"deleted"`
		} `json:"result"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return 0, err
	}

	if c.verbose {
		fmt.Printf("Deleted %d messages\n", result.Result.Deleted)
	}

	return result.Result.Deleted, nil
}

// DeleteMessagesForCurrentDay deletes all messages from the current day
func (c *MemoryClient) DeleteMessagesForCurrentDay(ctx context.Context) (int, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	
	return c.DeleteMessagesByTimeRange(ctx, startOfDay, endOfDay)
}

// DeleteMessagesForCurrentWeek deletes all messages from the current week
func (c *MemoryClient) DeleteMessagesForCurrentWeek(ctx context.Context) (int, error) {
	now := time.Now()
	
	// Calculate the start of the week (Sunday)
	daysToSunday := int(now.Weekday())
	startOfWeek := time.Date(now.Year(), now.Month(), now.Day()-daysToSunday, 0, 0, 0, 0, now.Location())
	
	// Calculate the end of the week (Saturday)
	endOfWeek := time.Date(now.Year(), now.Month(), now.Day()+(6-daysToSunday), 23, 59, 59, 999999999, now.Location())
	
	return c.DeleteMessagesByTimeRange(ctx, startOfWeek, endOfWeek)
}

// DeleteMessagesForCurrentMonth deletes all messages from the current month
func (c *MemoryClient) DeleteMessagesForCurrentMonth(ctx context.Context) (int, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 999999999, now.Location())
	
	return c.DeleteMessagesByTimeRange(ctx, startOfMonth, endOfMonth)
}

// SearchMessages is an alias for SearchSimilarMessages to match the interface
func (c *MemoryClient) SearchMessages(ctx context.Context, query string, limit int) ([]models.Message, error) {
	return c.SearchSimilarMessages(ctx, query, limit)
}

// IndexMessages indexes all messages
func (c *MemoryClient) IndexMessages(ctx context.Context) error {
	if c.verbose {
		fmt.Println("Indexing messages")
	}
	
	// This is a no-op as messages are indexed when added
	return nil
}
