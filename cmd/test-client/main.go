package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/christerso/memory-client-go/internal/client"
	"github.com/christerso/memory-client-go/internal/models"
)

// Simple test server to verify the Qdrant integration
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Configuration
	qdrantURL := "http://localhost:6333"
	collectionName := "conversation_memory"
	embeddingSize := 384
	port := 10012

	log.Printf("Starting test server on port %d", port)
	log.Printf("Connecting to Qdrant at %s, collection: %s", qdrantURL, collectionName)

	// Create a memory client
	memClient, err := client.NewMemoryClient(qdrantURL, collectionName, embeddingSize, true)
	if err != nil {
		log.Fatalf("Failed to create memory client: %v", err)
	}

	// Set up HTTP server
	mux := http.NewServeMux()

	// API endpoint for adding a message
	mux.HandleFunc("/api/message", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse the message
		var messageRequest struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}
		err := json.NewDecoder(r.Body).Decode(&messageRequest)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse request JSON: %v", err), http.StatusBadRequest)
			return
		}

		// Create a new message
		message := models.NewMessage(models.Role(messageRequest.Role), messageRequest.Content)

		// Add the message to the vector database with the modified AddMessage function
		err = addMessageWithIds(ctx, memClient, message)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to add message: %v", err), http.StatusInternalServerError)
			return
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Message added successfully",
			"id":      message.ID,
		})
	})

	// Start the server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		server.Shutdown(ctx)
	}()

	// Start the server
	log.Printf("Server listening on port %d", port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

// Modified version of AddMessage that includes the ids field
func addMessageWithIds(ctx context.Context, c *client.MemoryClient, message *models.Message) error {
	// Generate embedding for message
	embedding, err := c.GenerateEmbedding(ctx, message.Content)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Create point
	point := map[string]interface{}{
		"id":     message.ID,
		"vector": embedding,
		"payload": map[string]interface{}{
			"role":      message.Role,
			"content":   message.Content,
			"timestamp": message.Timestamp.Format(time.RFC3339),
			"metadata":  message.Metadata,
			"tags":      message.Tags,
		},
	}

	// Add point to collection
	url := fmt.Sprintf("%s/collections/%s/points", c.GetQdrantURL(), c.GetCollectionName())

	// Include the ids field as required by Qdrant
	request := map[string]interface{}{
		"points": []interface{}{point},
		"ids":    []string{message.ID},
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

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add point: %d %s - %s", resp.StatusCode, resp.Status, string(body))
	}

	return nil
}
