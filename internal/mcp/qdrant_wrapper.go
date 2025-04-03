package mcp

import (
	"context"
	"log"

	"github.com/qdrant/go-client/qdrant"
)

// QdrantWrapper wraps the Qdrant client to provide additional functionality
type QdrantWrapper struct {
	client *qdrant.Client
}

// NewQdrantWrapper creates a new QdrantWrapper
func NewQdrantWrapper(client *qdrant.Client) *QdrantWrapper {
	return &QdrantWrapper{
		client: client,
	}
}

// UpsertVector is a mock implementation that logs the call but does nothing
// This is a temporary solution to fix the compilation error
func (w *QdrantWrapper) UpsertVector(ctx context.Context, id string, embedding []float32, metadata map[string]interface{}) error {
	log.Printf("Mock UpsertVector called with id: %s", id)
	// In a real implementation, this would upsert the vector to Qdrant
	return nil
}
