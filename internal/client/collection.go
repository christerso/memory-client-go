package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ensureCollection ensures that the collection exists
func (c *MemoryClient) ensureCollection(ctx context.Context) error {
	// Check if collection exists
	exists, err := c.collectionExists(ctx)
	if err != nil {
		return err
	}

	// If collection exists, return
	if exists {
		return nil
	}

	// Create collection
	return c.createCollection(ctx)
}

// collectionExists checks if the collection exists
func (c *MemoryClient) collectionExists(ctx context.Context) (bool, error) {
	url := fmt.Sprintf("%s/collections/%s", c.qdrantURL, c.collectionName)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	return false, nil
}

// createCollection creates a new collection
func (c *MemoryClient) createCollection(ctx context.Context) error {
	url := fmt.Sprintf("%s/collections/%s", c.qdrantURL, c.collectionName)

	// Collection configuration
	config := map[string]interface{}{
		"vectors": map[string]interface{}{
			"size":     c.embeddingSize,
			"distance": "Cosine",
		},
	}

	jsonData, err := json.Marshal(config)
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
		return fmt.Errorf("failed to create collection: %s", resp.Status)
	}

	return nil
}

// recreateCollection deletes and recreates the collection
func (c *MemoryClient) recreateCollection(ctx context.Context) error {
	// Delete collection if it exists
	exists, err := c.collectionExists(ctx)
	if err != nil {
		return err
	}

	if exists {
		err = c.deleteCollection(ctx)
		if err != nil {
			return err
		}
	}

	// Create collection
	return c.createCollection(ctx)
}

// deleteCollection deletes the collection
func (c *MemoryClient) deleteCollection(ctx context.Context) error {
	url := fmt.Sprintf("%s/collections/%s", c.qdrantURL, c.collectionName)
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
		return fmt.Errorf("failed to delete collection: %s", resp.Status)
	}

	return nil
}
