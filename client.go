package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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

	// Ensure collection exists
	err := client.ensureCollection(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to ensure collection exists: %w", err)
	}

	return client, nil
}

// Close closes the client
func (c *MemoryClient) Close() {
	// Nothing to close for HTTP client
}

// ensureCollection ensures that the collection exists
func (c *MemoryClient) ensureCollection(ctx context.Context) error {
	// Check if collection exists
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/collections/%s", c.qdrantURL, c.collectionName), nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// If collection doesn't exist (404), create it
	if resp.StatusCode == http.StatusNotFound {
		// Create collection
		createReq := struct {
			Vectors map[string]interface{} `json:"vectors"`
		}{
			Vectors: map[string]interface{}{
				"default": map[string]interface{}{
					"size":     c.embeddingSize,
					"distance": "Cosine",
				},
			},
		}

		createBody, err := json.Marshal(createReq)
		if err != nil {
			return err
		}

		req, err = http.NewRequestWithContext(
			ctx,
			"PUT",
			fmt.Sprintf("%s/collections/%s", c.qdrantURL, c.collectionName),
			bytes.NewBuffer(createBody),
		)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err = c.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("failed to create collection: %s", body)
		}

		// Create index
		indexReq := struct {
			FieldName string `json:"field_name"`
			FieldType string `json:"field_type"`
		}{
			FieldName: "Role",
			FieldType: "keyword",
		}

		indexBody, err := json.Marshal(indexReq)
		if err != nil {
			return err
		}

		req, err = http.NewRequestWithContext(
			ctx,
			"PUT",
			fmt.Sprintf("%s/collections/%s/index", c.qdrantURL, c.collectionName),
			bytes.NewBuffer(indexBody),
		)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err = c.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("failed to create index: %s", body)
		}
	} else if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to check collection: %s", body)
	}

	return nil
}

// AddMessage adds a message to the memory with exact duplication checking
func (c *MemoryClient) AddMessage(ctx context.Context, message *Message) error {
	// Check for exact duplicates before adding
	isDuplicate, err := c.isExactDuplicate(ctx, message)
	if err != nil {
		// If there's an error checking for duplicates, log it but continue
		if c.verbose {
			fmt.Printf("Warning: Failed to check for duplicates: %v\n", err)
		}
	} else if isDuplicate {
		// Skip adding if it's an exact duplicate
		if c.verbose {
			fmt.Println("Exact duplicate message detected, skipping")
		}
		return nil
	}

	// Generate random vector for now
	// In a real implementation, you would use an embedding service
	if message.Vector == nil {
		message.Vector = make([]float32, c.embeddingSize)
		for i := range message.Vector {
			message.Vector[i] = rand.Float32()
		}
	}

	// Create point with simple numeric ID
	pointID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Convert ID to int64
	numericID, _ := strconv.ParseInt(pointID, 10, 64)

	// Prepare request
	req := struct {
		Points []struct {
			ID      int64                  `json:"id"`
			Vector  []float32              `json:"vector"`
			Payload map[string]interface{} `json:"payload"`
		} `json:"points"`
	}{
		Points: []struct {
			ID      int64                  `json:"id"`
			Vector  []float32              `json:"vector"`
			Payload map[string]interface{} `json:"payload"`
		}{
			{
				ID:     numericID,
				Vector: message.Vector,
				Payload: map[string]interface{}{
					"Role":    string(message.Role),
					"Content": message.Content,
				},
			},
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		"PUT",
		fmt.Sprintf("%s/collections/%s/points", c.qdrantURL, c.collectionName),
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add message: %s", respBody)
	}

	return nil
}

// GetConversationHistory gets the conversation history
func (c *MemoryClient) GetConversationHistory(ctx context.Context, limit int) ([]*Message, error) {
	// Prepare request
	req := struct {
		Limit      int  `json:"limit"`
		WithVector bool `json:"with_vector"`
	}{
		Limit:      limit,
		WithVector: true,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/collections/%s/points/scroll", c.qdrantURL, c.collectionName),
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get conversation history: %s", respBody)
	}

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse response using a more flexible approach
	var resultMap map[string]interface{}
	if err := json.Unmarshal(respBody, &resultMap); err != nil {
		return nil, err
	}

	// Only print debug info if verbose flag is set
	if c.verbose {
		fmt.Println("Response received from Qdrant (vector data omitted for clarity)")
	}

	// Extract results - the structure is different than expected
	resultObj, ok := resultMap["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format: missing result object")
	}

	pointsList, ok := resultObj["points"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format: missing points array")
	}

	// Convert to messages
	messages := make([]*Message, 0, len(pointsList))
	for _, item := range pointsList {
		point, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		payload, ok := point["payload"].(map[string]interface{})
		if !ok {
			continue
		}

		role, _ := payload["Role"].(string)
		content, _ := payload["Content"].(string)

		// Extract vector if available
		var vector []float32
		if vectorData, ok := point["vector"].([]interface{}); ok {
			vector = make([]float32, len(vectorData))
			for i, v := range vectorData {
				if f, ok := v.(float64); ok {
					vector[i] = float32(f)
				}
			}
		}

		messages = append(messages, &Message{
			Role:    Role(role),
			Content: content,
			Vector:  vector,
		})
	}

	return messages, nil
}

// SearchMessages searches for messages
func (c *MemoryClient) SearchMessages(ctx context.Context, query string, limit int) ([]*Message, error) {
	// For now, just return the most recent messages
	// In a real implementation, you would:
	// 1. Get embedding for the query
	// 2. Perform vector search
	return c.GetConversationHistory(ctx, limit)
}

// IndexProjectFiles indexes files in a project directory with progress reporting
func (c *MemoryClient) IndexProjectFiles(ctx context.Context, projectDir string) (int, error) {
	// Get absolute path
	absPath, err := filepath.Abs(projectDir)
	if err != nil {
		return 0, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if directory exists
	info, err := os.Stat(absPath)
	if err != nil {
		return 0, fmt.Errorf("failed to stat directory: %w", err)
	}
	if !info.IsDir() {
		return 0, fmt.Errorf("not a directory: %s", absPath)
	}

	fmt.Printf("Indexing project directory: %s\n", absPath)

	// Create project collection if it doesn't exist
	projectCollection := c.collectionName + "_project"
	err = c.ensureProjectCollection(ctx, projectCollection)
	if err != nil {
		return 0, fmt.Errorf("failed to ensure project collection: %w", err)
	}

	// First, collect all eligible files to process
	var filesToProcess []string
	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get file extension
		ext := strings.ToLower(filepath.Ext(path))

		// Skip media and binary files
		if MediaExtensions[ext] || BinaryExtensions[ext] {
			return nil
		}

		// Skip files larger than 1MB
		if info.Size() > 1024*1024 {
			return nil
		}

		// Skip hidden files and directories
		if strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		filesToProcess = append(filesToProcess, path)
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("error walking directory: %w", err)
	}

	totalFiles := len(filesToProcess)
	fmt.Printf("Found %d files to index\n", totalFiles)

	// Process files in batches
	const batchSize = 100
	var count int
	var lastPercent int

	for i := 0; i < totalFiles; i += batchSize {
		end := i + batchSize
		if end > totalFiles {
			end = totalFiles
		}

		batch := filesToProcess[i:end]
		for _, path := range batch {
			// Create a context with timeout for each file
			fileCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			
			// Read file content
			info, err := os.Stat(path)
			if err != nil {
				cancel()
				continue // Skip files we can't stat
			}
			
			content, err := os.ReadFile(path)
			if err != nil {
				cancel()
				continue // Skip files we can't read
			}

			// Get relative path
			relPath, err := filepath.Rel(absPath, path)
			if err != nil {
				relPath = path // Use absolute path if relative path fails
			}

			// Determine language
			ext := strings.ToLower(filepath.Ext(path))
			language := "Unknown"
			if lang, ok := LanguageMap[ext]; ok {
				language = lang
			}

			// Create project file
			projectFile := &ProjectFile{
				Path:     relPath,
				Content:  string(content),
				Language: language,
				Vector:   make([]float32, c.embeddingSize),
				ModTime:  info.ModTime().Unix(), // Store modification time
			}

			// Generate random vector for now
			// In a real implementation, you would use an embedding service
			for i := range projectFile.Vector {
				projectFile.Vector[i] = rand.Float32()
			}

			// Add to Qdrant with retry logic
			var addErr error
			for retries := 0; retries < 3; retries++ {
				addErr = c.addProjectFile(fileCtx, projectCollection, projectFile)
				if addErr == nil {
					break
				}
				
				// If context deadline exceeded, wait a bit and retry
				if strings.Contains(addErr.Error(), "context deadline exceeded") || 
				   strings.Contains(addErr.Error(), "connection refused") {
					time.Sleep(time.Duration(retries+1) * 500 * time.Millisecond)
					continue
				}
				
				// For other errors, don't retry
				break
			}
			
			cancel() // Release the file context
			
			if addErr != nil {
				// Log the error but continue with other files
				fmt.Printf("Error indexing file %s: %v\n", relPath, addErr)
				continue
			}

			count++

			// Report progress
			percent := (count * 100) / totalFiles
			if percent > lastPercent {
				fmt.Printf("Indexing progress: %d%% (%d/%d files)\n", percent, count, totalFiles)
				lastPercent = percent
			}
			
			// Check if the main context is done
			select {
			case <-ctx.Done():
				return count, fmt.Errorf("indexing interrupted: %w", ctx.Err())
			default:
				// Continue processing
			}
		}
	}

	if count == 0 && totalFiles > 0 {
		return 0, fmt.Errorf("failed to index any files")
	}

	return count, nil
}

// UpdateProjectFiles updates only modified files in a project directory
func (c *MemoryClient) UpdateProjectFiles(ctx context.Context, projectDir string) (int, int, error) {
	// Get absolute path
	absPath, err := filepath.Abs(projectDir)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if directory exists
	info, err := os.Stat(absPath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to stat directory: %w", err)
	}
	if !info.IsDir() {
		return 0, 0, fmt.Errorf("not a directory: %s", absPath)
	}

	fmt.Printf("Checking for updated files in: %s\n", absPath)

	// Create project collection if it doesn't exist
	projectCollection := c.collectionName + "_project"
	err = c.ensureProjectCollection(ctx, projectCollection)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to ensure project collection: %w", err)
	}

	// Get existing files from the collection
	existingFiles, err := c.SearchProjectFiles(ctx, "", 10000) // Get all files
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get existing files: %w", err)
	}

	// Create a map of existing files by path
	existingFileMap := make(map[string]*ProjectFile)
	for _, file := range existingFiles {
		existingFileMap[file.Path] = file
	}

	// Track new and updated files
	var newCount, updateCount int

	// Walk directory and check for new or modified files
	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get file extension
		ext := strings.ToLower(filepath.Ext(path))

		// Skip media and binary files
		if MediaExtensions[ext] || BinaryExtensions[ext] {
			return nil
		}

		// Skip files larger than 1MB
		if info.Size() > 1024*1024 {
			return nil
		}

		// Skip hidden files and directories
		if strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(absPath, path)
		if err != nil {
			relPath = path // Use absolute path if relative path fails
		}

		// Check if file exists in collection and if it's been modified
		existingFile, exists := existingFileMap[relPath]
		modTime := info.ModTime().Unix()

		if !exists || (exists && existingFile.ModTime < modTime) {
			// File is new or has been modified
			// Read file content
			content, err := os.ReadFile(path)
			if err != nil {
				return nil // Skip files we can't read
			}

			// Determine language
			language := "Unknown"
			if lang, ok := LanguageMap[ext]; ok {
				language = lang
			}

			// Create project file
			projectFile := &ProjectFile{
				Path:     relPath,
				Content:  string(content),
				Language: language,
				Vector:   make([]float32, c.embeddingSize),
				ModTime:  modTime,
			}

			// Generate random vector for now
			// In a real implementation, you would use an embedding service
			for i := range projectFile.Vector {
				projectFile.Vector[i] = rand.Float32()
			}

			// Add to Qdrant
			err = c.addProjectFile(ctx, projectCollection, projectFile)
			if err != nil {
				return err
			}

			if exists {
				updateCount++
				fmt.Printf("Updated file: %s\n", relPath)
			} else {
				newCount++
				fmt.Printf("Added new file: %s\n", relPath)
			}
		}

		// Remove from map to track what's been processed
		delete(existingFileMap, relPath)
		return nil
	})

	if err != nil {
		return newCount, updateCount, fmt.Errorf("error walking directory: %w", err)
	}

	// TODO: Handle deleted files (remaining in existingFileMap)

	fmt.Printf("Update complete: %d new files, %d updated files\n", newCount, updateCount)
	return newCount, updateCount, nil
}

// ensureProjectCollection ensures the project collection exists
func (c *MemoryClient) ensureProjectCollection(ctx context.Context, collectionName string) error {
	// Check if collection exists
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/collections/%s", c.qdrantURL, collectionName), nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// If collection doesn't exist (404), create it
	if resp.StatusCode == http.StatusNotFound {
		// Create collection
		createReq := struct {
			Vectors map[string]interface{} `json:"vectors"`
		}{
			Vectors: map[string]interface{}{
				"default": map[string]interface{}{
					"size":     c.embeddingSize,
					"distance": "Cosine",
				},
			},
		}

		createBody, err := json.Marshal(createReq)
		if err != nil {
			return err
		}

		req, err = http.NewRequestWithContext(
			ctx,
			"PUT",
			fmt.Sprintf("%s/collections/%s", c.qdrantURL, collectionName),
			bytes.NewBuffer(createBody),
		)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err = c.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("failed to create collection: %s", body)
		}

		// Create index for path and language
		for _, field := range []string{"Path", "Language"} {
			indexReq := struct {
				FieldName string `json:"field_name"`
				FieldType string `json:"field_type"`
			}{
				FieldName: field,
				FieldType: "keyword",
			}

			indexBody, err := json.Marshal(indexReq)
			if err != nil {
				return err
			}

			req, err = http.NewRequestWithContext(
				ctx,
				"PUT",
				fmt.Sprintf("%s/collections/%s/index", c.qdrantURL, collectionName),
				bytes.NewBuffer(indexBody),
			)
			if err != nil {
				return err
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err = c.httpClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("failed to create index: %s", body)
			}
		}
	} else if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to check collection: %s", body)
	}

	return nil
}

// addProjectFile adds a project file to the collection
func (c *MemoryClient) addProjectFile(ctx context.Context, collectionName string, file *ProjectFile) error {
	// Create point ID
	pointID := fmt.Sprintf("%d", time.Now().UnixNano())
	numericID, _ := strconv.ParseInt(pointID, 10, 64)

	// Prepare request
	req := struct {
		Points []struct {
			ID      int64                  `json:"id"`
			Vector  []float32              `json:"vector"`
			Payload map[string]interface{} `json:"payload"`
		} `json:"points"`
	}{
		Points: []struct {
			ID      int64                  `json:"id"`
			Vector  []float32              `json:"vector"`
			Payload map[string]interface{} `json:"payload"`
		}{
			{
				ID:     numericID,
				Vector: file.Vector,
				Payload: map[string]interface{}{
					"Path":     file.Path,
					"Content":  file.Content,
					"Language": file.Language,
					"ModTime":  file.ModTime,
				},
			},
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		"PUT",
		fmt.Sprintf("%s/collections/%s/points", c.qdrantURL, collectionName),
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add project file: %s", respBody)
	}

	return nil
}

// SearchProjectFiles searches for project files
func (c *MemoryClient) SearchProjectFiles(ctx context.Context, query string, limit int) ([]*ProjectFile, error) {
	projectCollection := c.collectionName + "_project"

	// For now, just return the most recent files
	// In a real implementation, you would:
	// 1. Get embedding for the query
	// 2. Perform vector search

	// Prepare request
	req := struct {
		Limit      int  `json:"limit"`
		WithVector bool `json:"with_vector"`
	}{
		Limit:      limit,
		WithVector: true,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/collections/%s/points/scroll", c.qdrantURL, projectCollection),
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get project files: %s", respBody)
	}

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse response using a more flexible approach
	var resultMap map[string]interface{}
	if err := json.Unmarshal(respBody, &resultMap); err != nil {
		return nil, err
	}

	// Extract results
	resultObj, ok := resultMap["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format: missing result object")
	}

	pointsList, ok := resultObj["points"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format: missing points array")
	}

	// Convert to project files
	files := make([]*ProjectFile, 0, len(pointsList))
	for _, item := range pointsList {
		point, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		payload, ok := point["payload"].(map[string]interface{})
		if !ok {
			continue
		}

		path, _ := payload["Path"].(string)
		content, _ := payload["Content"].(string)
		language, _ := payload["Language"].(string)

		// Extract ModTime
		var modTime int64
		if mt, ok := payload["ModTime"].(float64); ok {
			modTime = int64(mt)
		}

		// Extract vector if available
		var vector []float32
		if vectorData, ok := point["vector"].([]interface{}); ok {
			vector = make([]float32, len(vectorData))
			for i, v := range vectorData {
				if f, ok := v.(float64); ok {
					vector[i] = float32(f)
				}
			}
		}

		files = append(files, &ProjectFile{
			Path:     path,
			Content:  content,
			Language: language,
			Vector:   vector,
			ModTime:  modTime,
		})
	}

	return files, nil
}

// IndexMessages indexes messages
func (c *MemoryClient) IndexMessages(ctx context.Context) error {
	// This would be used to batch process messages and update vectors
	// For now, it's a no-op since we're adding vectors at insertion time
	return nil
}

// isExactDuplicate checks if a message with exactly the same content and role already exists
func (c *MemoryClient) isExactDuplicate(ctx context.Context, message *Message) (bool, error) {
	// Prepare request to search for messages with the same role and content
	req := struct {
		Filter     map[string]interface{} `json:"filter"`
		Limit      int                    `json:"limit"`
		WithVector bool                   `json:"with_vector"`
	}{
		Filter: map[string]interface{}{
			"must": []map[string]interface{}{
				{
					"key":   "Role",
					"match": map[string]interface{}{
						"value": string(message.Role),
					},
				},
				{
					"key":   "Content",
					"match": map[string]interface{}{
						"value": message.Content,
					},
				},
			},
		},
		Limit:      1, // We only need to know if at least one exists
		WithVector: false, // No need for vectors
	}

	body, err := json.Marshal(req)
	if err != nil {
		return false, err
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/collections/%s/points/scroll", c.qdrantURL, c.collectionName),
		bytes.NewBuffer(body),
	)
	if err != nil {
		return false, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("failed to check for duplicates: %s", respBody)
	}

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	// Parse response
	var resultMap map[string]interface{}
	if err := json.Unmarshal(respBody, &resultMap); err != nil {
		return false, err
	}

	// Extract results
	resultObj, ok := resultMap["result"].(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("unexpected response format: missing result object")
	}

	pointsList, ok := resultObj["points"].([]interface{})
	if !ok {
		return false, fmt.Errorf("unexpected response format: missing points array")
	}

	// If we found any points, it's a duplicate
	return len(pointsList) > 0, nil
}

// GetMemoryStats retrieves statistics about the memory storage
func (c *MemoryClient) GetMemoryStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get collection info
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("%s/collections/%s", c.qdrantURL, c.collectionName),
		nil,
	)
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
		return nil, fmt.Errorf("failed to get collection info: %s", body)
	}

	var collectionInfo struct {
		Result struct {
			Status string `json:"status"`
			Config struct {
				Params struct {
					VectorsCount  int `json:"vectors_count"`
					SegmentsCount int `json:"segments_count"`
				} `json:"params"`
			} `json:"config"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&collectionInfo); err != nil {
		return nil, fmt.Errorf("failed to decode collection info: %w", err)
	}

	// Get message count
	messageCount, err := c.getMessageCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get message count: %w", err)
	}

	// Get project files count
	projectFilesCount, err := c.getProjectFilesCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get project files count: %w", err)
	}

	// Assemble stats
	stats["message_count"] = messageCount
	stats["project_files_count"] = projectFilesCount
	stats["collection_status"] = collectionInfo.Result.Status
	stats["vectors_count"] = collectionInfo.Result.Config.Params.VectorsCount
	stats["segments_count"] = collectionInfo.Result.Config.Params.SegmentsCount
	stats["timestamp"] = time.Now().Unix()

	return stats, nil
}

// getMessageCount gets the count of messages in the collection
func (c *MemoryClient) getMessageCount(ctx context.Context) (int, error) {
	// Prepare search request to count messages
	searchReq := struct {
		Filter struct {
			MustNot struct {
				Field struct {
					Key   string `json:"key"`
					Match struct {
						Value string `json:"value"`
					} `json:"match"`
				} `json:"field"`
			} `json:"must_not"`
		} `json:"filter"`
		Limit int `json:"limit"`
	}{
		Filter: struct {
			MustNot struct {
				Field struct {
					Key   string `json:"key"`
					Match struct {
						Value string `json:"value"`
					} `json:"match"`
				} `json:"field"`
			} `json:"must_not"`
		}{
			MustNot: struct {
				Field struct {
					Key   string `json:"key"`
					Match struct {
						Value string `json:"value"`
					} `json:"match"`
				} `json:"field"`
			}{
				Field: struct {
					Key   string `json:"key"`
					Match struct {
						Value string `json:"value"`
					} `json:"match"`
				}{
					Key: "Role",
					Match: struct {
						Value string `json:"value"`
					}{
						Value: string(RoleProject),
					},
				},
			},
		},
		Limit: 0, // We only need the count
	}

	searchBody, err := json.Marshal(searchReq)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/collections/%s/points/scroll", c.qdrantURL, c.collectionName),
		bytes.NewBuffer(searchBody),
	)
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
		return 0, fmt.Errorf("failed to count messages: %s", body)
	}

	var result struct {
		Result struct {
			Points []interface{} `json:"points"`
			NextPageOffset string `json:"next_page_offset"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	return len(result.Result.Points), nil
}

// getProjectFilesCount gets the count of project files in the collection
func (c *MemoryClient) getProjectFilesCount(ctx context.Context) (int, error) {
	// Create project collection name
	projectCollectionName := c.collectionName + "_project"

	// Check if collection exists
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("%s/collections/%s", c.qdrantURL, projectCollectionName),
		nil,
	)
	if err != nil {
		return 0, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// If collection doesn't exist, return 0
	if resp.StatusCode == http.StatusNotFound {
		return 0, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("failed to check project collection: %s", body)
	}

	// Prepare search request to count project files
	searchReq := struct {
		Limit int `json:"limit"`
	}{
		Limit: 0, // We only need the count
	}

	searchBody, err := json.Marshal(searchReq)
	if err != nil {
		return 0, err
	}

	req, err = http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/collections/%s/points/scroll", c.qdrantURL, projectCollectionName),
		bytes.NewBuffer(searchBody),
	)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("failed to count project files: %s", body)
	}

	var result struct {
		Result struct {
			Points []interface{} `json:"points"`
			NextPageOffset string `json:"next_page_offset"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	return len(result.Result.Points), nil
}

// DeleteMessage deletes a message by ID from the memory
func (c *MemoryClient) DeleteMessage(ctx context.Context, id string) error {
	// Convert ID to int64
	numericID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	// Prepare request
	req, err := http.NewRequestWithContext(
		ctx,
		"DELETE",
		fmt.Sprintf("%s/collections/%s/points/%d", c.qdrantURL, c.collectionName, numericID),
		nil,
	)
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
		return fmt.Errorf("failed to delete message: %s", body)
	}

	return nil
}

// DeleteAllMessages deletes all messages from the memory
func (c *MemoryClient) DeleteAllMessages(ctx context.Context) error {
	// Prepare filter to exclude project files
	filter := struct {
		Filter struct {
			MustNot struct {
				Field struct {
					Key   string `json:"key"`
					Match struct {
						Value string `json:"value"`
					} `json:"match"`
				} `json:"field"`
			} `json:"must_not"`
		} `json:"filter"`
	}{
		Filter: struct {
			MustNot struct {
				Field struct {
					Key   string `json:"key"`
					Match struct {
						Value string `json:"value"`
					} `json:"match"`
				} `json:"field"`
			} `json:"must_not"`
		}{
			MustNot: struct {
				Field struct {
					Key   string `json:"key"`
					Match struct {
						Value string `json:"value"`
					} `json:"match"`
				} `json:"field"`
			}{
				Field: struct {
					Key   string `json:"key"`
					Match struct {
						Value string `json:"value"`
					} `json:"match"`
				}{
					Key: "Role",
					Match: struct {
						Value string `json:"value"`
					}{
						Value: string(RoleProject),
					},
				},
			},
		},
	}

	filterBody, err := json.Marshal(filter)
	if err != nil {
		return err
	}

	// Prepare request
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/collections/%s/points/delete", c.qdrantURL, c.collectionName),
		bytes.NewBuffer(filterBody),
	)
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
		return fmt.Errorf("failed to delete all messages: %s", body)
	}

	return nil
}

// DeleteProjectFile deletes a project file by path from the memory
func (c *MemoryClient) DeleteProjectFile(ctx context.Context, path string) error {
	// Create project collection name
	projectCollectionName := c.collectionName + "_project"

	// Check if collection exists
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("%s/collections/%s", c.qdrantURL, projectCollectionName),
		nil,
	)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// If collection doesn't exist, return error
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("project collection does not exist")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to check project collection: %s", body)
	}

	// Prepare filter to find the file by path
	filter := struct {
		Filter struct {
			Must struct {
				Field struct {
					Key   string `json:"key"`
					Match struct {
						Value string `json:"value"`
					} `json:"match"`
				} `json:"field"`
			} `json:"must"`
		} `json:"filter"`
	}{
		Filter: struct {
			Must struct {
				Field struct {
					Key   string `json:"key"`
					Match struct {
						Value string `json:"value"`
					} `json:"match"`
				} `json:"field"`
			} `json:"must"`
		}{
			Must: struct {
				Field struct {
					Key   string `json:"key"`
					Match struct {
						Value string `json:"value"`
					} `json:"match"`
				} `json:"field"`
			}{
				Field: struct {
					Key   string `json:"key"`
					Match struct {
						Value string `json:"value"`
					} `json:"match"`
				}{
					Key: "Path",
					Match: struct {
						Value string `json:"value"`
					}{
						Value: path,
					},
				},
			},
		},
	}

	filterBody, err := json.Marshal(filter)
	if err != nil {
		return err
	}

	// Prepare request
	req, err = http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/collections/%s/points/delete", c.qdrantURL, projectCollectionName),
		bytes.NewBuffer(filterBody),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete project file: %s", body)
	}

	return nil
}

// DeleteAllProjectFiles deletes all project files from the memory
func (c *MemoryClient) DeleteAllProjectFiles(ctx context.Context) error {
	// Create project collection name
	projectCollectionName := c.collectionName + "_project"

	// Check if collection exists
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("%s/collections/%s", c.qdrantURL, projectCollectionName),
		nil,
	)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// If collection doesn't exist, return success (nothing to delete)
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to check project collection: %s", body)
	}

	// Delete all points (empty filter means delete all)
	emptyFilter := struct{}{}
	filterBody, err := json.Marshal(emptyFilter)
	if err != nil {
		return err
	}

	// Prepare request
	req, err = http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/collections/%s/points/delete", c.qdrantURL, projectCollectionName),
		bytes.NewBuffer(filterBody),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete all project files: %s", body)
	}

	return nil
}

// TagMessages adds tags to messages matching a query
func (c *MemoryClient) TagMessages(ctx context.Context, query string, tags []string, limit int) (int, error) {
	// Find messages matching the query
	messages, err := c.SearchMessages(ctx, query, limit)
	if err != nil {
		return 0, fmt.Errorf("failed to search messages: %w", err)
	}

	if len(messages) == 0 {
		return 0, nil // No messages found
	}

	// Update each message with the new tags
	updatedCount := 0
	for _, message := range messages {
		// Add new tags if they don't already exist
		existingTags := make(map[string]bool)
		for _, tag := range message.Tags {
			existingTags[tag] = true
		}

		// Add new tags that don't already exist
		tagsAdded := false
		for _, tag := range tags {
			if !existingTags[tag] {
				message.Tags = append(message.Tags, tag)
				tagsAdded = true
			}
		}

		// If no new tags were added, skip updating this message
		if !tagsAdded {
			continue
		}

		// Update the message in the database
		err := c.updateMessageTags(ctx, message)
		if err != nil {
			return updatedCount, fmt.Errorf("failed to update message tags: %w", err)
		}

		updatedCount++
	}

	return updatedCount, nil
}

// updateMessageTags updates the tags for a message in the database
func (c *MemoryClient) updateMessageTags(ctx context.Context, message *Message) error {
	// Find the message ID first
	searchReq := struct {
		Filter struct {
			Must []struct {
				Field struct {
					Key   string `json:"key"`
					Match struct {
						Value string `json:"value"`
					} `json:"match"`
				} `json:"field"`
			} `json:"must"`
		} `json:"filter"`
		Limit int `json:"limit"`
	}{
		Filter: struct {
			Must []struct {
				Field struct {
					Key   string `json:"key"`
					Match struct {
						Value string `json:"value"`
					} `json:"match"`
				} `json:"field"`
			} `json:"must"`
		}{
			Must: []struct {
				Field struct {
					Key   string `json:"key"`
					Match struct {
						Value string `json:"value"`
					} `json:"match"`
				} `json:"field"`
			}{
				{
					Field: struct {
						Key   string `json:"key"`
						Match struct {
							Value string `json:"value"`
						} `json:"match"`
					}{
						Key: "Role",
						Match: struct {
							Value string `json:"value"`
						}{
							Value: string(message.Role),
						},
					},
				},
				{
					Field: struct {
						Key   string `json:"key"`
						Match struct {
							Value string `json:"value"`
						} `json:"match"`
					}{
						Key: "Content",
						Match: struct {
							Value string `json:"value"`
						}{
							Value: message.Content,
						},
					},
				},
			},
		},
		Limit: 1,
	}

	searchBody, err := json.Marshal(searchReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/collections/%s/points/scroll", c.qdrantURL, c.collectionName),
		bytes.NewBuffer(searchBody),
	)
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
		return fmt.Errorf("failed to find message: %s", body)
	}

	var result struct {
		Result struct {
			Points []struct {
				ID      int64                  `json:"id"`
				Payload map[string]interface{} `json:"payload"`
			} `json:"points"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if len(result.Result.Points) == 0 {
		return fmt.Errorf("message not found")
	}

	// Get the ID of the first matching message
	pointID := result.Result.Points[0].ID

	// Update the message with new tags
	updateReq := struct {
		Payload map[string]interface{} `json:"payload"`
	}{
		Payload: map[string]interface{}{
			"Tags": message.Tags,
		},
	}

	updateBody, err := json.Marshal(updateReq)
	if err != nil {
		return err
	}

	req, err = http.NewRequestWithContext(
		ctx,
		"PATCH",
		fmt.Sprintf("%s/collections/%s/points/%d", c.qdrantURL, c.collectionName, pointID),
		bytes.NewBuffer(updateBody),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update message: %s", body)
	}

	return nil
}

// SummarizeAndTagMessages summarizes messages matching a query and adds tags to them
func (c *MemoryClient) SummarizeAndTagMessages(ctx context.Context, query string, summary string, tags []string, limit int) (int, error) {
	// Find messages matching the query
	messages, err := c.SearchMessages(ctx, query, limit)
	if err != nil {
		return 0, fmt.Errorf("failed to search messages: %w", err)
	}

	if len(messages) == 0 {
		return 0, nil // No messages found
	}

	// Update each message with the summary and tags
	updatedCount := 0
	for _, message := range messages {
		// Set the summary
		message.Summary = summary

		// Add new tags if they don't already exist
		existingTags := make(map[string]bool)
		for _, tag := range message.Tags {
			existingTags[tag] = true
		}

		// Add new tags that don't already exist
		for _, tag := range tags {
			if !existingTags[tag] {
				message.Tags = append(message.Tags, tag)
			}
		}

		// Update the message in the database
		err := c.updateMessageSummaryAndTags(ctx, message)
		if err != nil {
			return updatedCount, fmt.Errorf("failed to update message: %w", err)
		}

		updatedCount++
	}

	return updatedCount, nil
}

// updateMessageSummaryAndTags updates the summary and tags for a message in the database
func (c *MemoryClient) updateMessageSummaryAndTags(ctx context.Context, message *Message) error {
	// Find the message ID first (same as in updateMessageTags)
	searchReq := struct {
		Filter struct {
			Must []struct {
				Field struct {
					Key   string `json:"key"`
					Match struct {
						Value string `json:"value"`
					} `json:"match"`
				} `json:"field"`
			} `json:"must"`
		} `json:"filter"`
		Limit int `json:"limit"`
	}{
		Filter: struct {
			Must []struct {
				Field struct {
					Key   string `json:"key"`
					Match struct {
						Value string `json:"value"`
					} `json:"match"`
				} `json:"field"`
			} `json:"must"`
		}{
			Must: []struct {
				Field struct {
					Key   string `json:"key"`
					Match struct {
						Value string `json:"value"`
					} `json:"match"`
				} `json:"field"`
			}{
				{
					Field: struct {
						Key   string `json:"key"`
						Match struct {
							Value string `json:"value"`
						} `json:"match"`
					}{
						Key: "Role",
						Match: struct {
							Value string `json:"value"`
						}{
							Value: string(message.Role),
						},
					},
				},
				{
					Field: struct {
						Key   string `json:"key"`
						Match struct {
							Value string `json:"value"`
						} `json:"match"`
					}{
						Key: "Content",
						Match: struct {
							Value string `json:"value"`
						}{
							Value: message.Content,
						},
					},
				},
			},
		},
		Limit: 1,
	}

	searchBody, err := json.Marshal(searchReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/collections/%s/points/scroll", c.qdrantURL, c.collectionName),
		bytes.NewBuffer(searchBody),
	)
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
		return fmt.Errorf("failed to find message: %s", body)
	}

	var result struct {
		Result struct {
			Points []struct {
				ID      int64                  `json:"id"`
				Payload map[string]interface{} `json:"payload"`
			} `json:"points"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if len(result.Result.Points) == 0 {
		return fmt.Errorf("message not found")
	}

	// Get the ID of the first matching message
	pointID := result.Result.Points[0].ID

	// Update the message with new summary and tags
	updateReq := struct {
		Payload map[string]interface{} `json:"payload"`
	}{
		Payload: map[string]interface{}{
			"Tags":    message.Tags,
			"Summary": message.Summary,
		},
	}

	updateBody, err := json.Marshal(updateReq)
	if err != nil {
		return err
	}

	req, err = http.NewRequestWithContext(
		ctx,
		"PATCH",
		fmt.Sprintf("%s/collections/%s/points/%d", c.qdrantURL, c.collectionName, pointID),
		bytes.NewBuffer(updateBody),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update message: %s", body)
	}

	return nil
}

// GetMessagesByTag retrieves messages with a specific tag
func (c *MemoryClient) GetMessagesByTag(ctx context.Context, tag string, limit int) ([]*Message, error) {
	// Prepare search request
	searchReq := struct {
		Filter struct {
			Must struct {
				Field struct {
					Key   string `json:"key"`
					Match struct {
						Value string `json:"value"`
					} `json:"match"`
				} `json:"field"`
			} `json:"must"`
		} `json:"filter"`
		Limit int `json:"limit"`
	}{
		Filter: struct {
			Must struct {
				Field struct {
					Key   string `json:"key"`
					Match struct {
						Value string `json:"value"`
					} `json:"match"`
				} `json:"field"`
			} `json:"must"`
		}{
			Must: struct {
				Field struct {
					Key   string `json:"key"`
					Match struct {
						Value string `json:"value"`
					} `json:"match"`
				} `json:"field"`
			}{
				Field: struct {
					Key   string `json:"key"`
					Match struct {
						Value string `json:"value"`
					} `json:"match"`
				}{
					Key: "Tags",
					Match: struct {
						Value string `json:"value"`
					}{
						Value: tag,
					},
				},
			},
		},
		Limit: limit,
	}

	searchBody, err := json.Marshal(searchReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/collections/%s/points/scroll", c.qdrantURL, c.collectionName),
		bytes.NewBuffer(searchBody),
	)
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
		return nil, fmt.Errorf("failed to search messages by tag: %s", body)
	}

	var result struct {
		Result struct {
			Points []struct {
				ID      int64                  `json:"id"`
				Payload map[string]interface{} `json:"payload"`
			} `json:"points"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Convert to messages
	messages := make([]*Message, 0, len(result.Result.Points))
	for _, point := range result.Result.Points {
		payload := point.Payload

		role, _ := payload["Role"].(string)
		content, _ := payload["Content"].(string)

		// Extract tags
		var tags []string
		if tagsData, ok := payload["Tags"].([]interface{}); ok {
			tags = make([]string, len(tagsData))
			for i, t := range tagsData {
				if tag, ok := t.(string); ok {
					tags[i] = tag
				}
			}
		}

		// Extract summary
		summary, _ := payload["Summary"].(string)

		messages = append(messages, &Message{
			Role:    Role(role),
			Content: content,
			Tags:    tags,
			Summary: summary,
		})
	}

	return messages, nil
}

// CheckServerStatus checks if the MCP server is running and returns its status
func (c *MemoryClient) CheckServerStatus(ctx context.Context) (bool, string, error) {
	// Try to connect to the MCP server
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/status", nil)
	if err != nil {
		return false, "", fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{
		Timeout: 2 * time.Second, // Short timeout for quick status check
	}

	resp, err := client.Do(req)
	if err != nil {
		// Check if it's a connection error
		if strings.Contains(err.Error(), "connection refused") || 
		   strings.Contains(err.Error(), "dial tcp") ||
		   strings.Contains(err.Error(), "context deadline exceeded") {
			return false, "MCP server is not running", nil
		}
		return false, "", fmt.Errorf("error connecting to server: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", fmt.Errorf("error reading response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Sprintf("Server returned non-OK status: %d %s", resp.StatusCode, resp.Status), nil
	}

	return true, fmt.Sprintf("MCP server is running: %s", string(body)), nil
}

// PurgeQdrant completely purges all data from Qdrant
func (c *MemoryClient) PurgeQdrant(ctx context.Context) error {
	// Delete the collection
	req, err := http.NewRequestWithContext(
		ctx,
		"DELETE",
		fmt.Sprintf("%s/collections/%s", c.qdrantURL, c.collectionName),
		nil,
	)
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
		return fmt.Errorf("failed to delete collection: %s", body)
	}

	// Recreate the collection
	err = c.ensureCollection(ctx)
	if err != nil {
		return fmt.Errorf("failed to recreate collection: %w", err)
	}

	// Also delete and recreate the project collection
	projectCollectionName := c.collectionName + "_projects"
	req, err = http.NewRequestWithContext(
		ctx,
		"DELETE",
		fmt.Sprintf("%s/collections/%s", c.qdrantURL, projectCollectionName),
		nil,
	)
	if err != nil {
		return err
	}

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Recreate the project collection
	err = c.ensureProjectCollection(ctx, projectCollectionName)
	if err != nil {
		return fmt.Errorf("failed to recreate project collection: %w", err)
	}

	return nil
}

// DeleteMessagesByTimeRange deletes messages within a specific time range
func (c *MemoryClient) DeleteMessagesByTimeRange(ctx context.Context, startTime, endTime time.Time) (int, error) {
	// Prepare filter for time range
	filter := struct {
		Filter struct {
			Must []struct {
				Range struct {
					Key     string `json:"key"`
					GTE     *int64 `json:"gte,omitempty"`
					LTE     *int64 `json:"lte,omitempty"`
					GT      *int64 `json:"gt,omitempty"`
					LT      *int64 `json:"lt,omitempty"`
				} `json:"range"`
			} `json:"must"`
			MustNot struct {
				Field struct {
					Key   string `json:"key"`
					Match struct {
						Value string `json:"value"`
					} `json:"match"`
				} `json:"field"`
			} `json:"must_not"`
		} `json:"filter"`
	}{}

	// Set up the time range filter
	filter.Filter.Must = make([]struct {
		Range struct {
			Key     string `json:"key"`
			GTE     *int64 `json:"gte,omitempty"`
			LTE     *int64 `json:"lte,omitempty"`
			GT      *int64 `json:"gt,omitempty"`
			LT      *int64 `json:"lt,omitempty"`
		} `json:"range"`
	}, 0)

	// Add start time if provided
	if !startTime.IsZero() {
		startUnix := startTime.Unix()
		rangeFilter := struct {
			Range struct {
				Key     string `json:"key"`
				GTE     *int64 `json:"gte,omitempty"`
				LTE     *int64 `json:"lte,omitempty"`
				GT      *int64 `json:"gt,omitempty"`
				LT      *int64 `json:"lt,omitempty"`
			} `json:"range"`
		}{
			Range: struct {
				Key     string `json:"key"`
				GTE     *int64 `json:"gte,omitempty"`
				LTE     *int64 `json:"lte,omitempty"`
				GT      *int64 `json:"gt,omitempty"`
				LT      *int64 `json:"lt,omitempty"`
			}{
				Key: "Timestamp",
				GTE: &startUnix,
			},
		}
		filter.Filter.Must = append(filter.Filter.Must, rangeFilter)
	}

	// Add end time if provided
	if !endTime.IsZero() {
		endUnix := endTime.Unix()
		rangeFilter := struct {
			Range struct {
				Key     string `json:"key"`
				GTE     *int64 `json:"gte,omitempty"`
				LTE     *int64 `json:"lte,omitempty"`
				GT      *int64 `json:"gt,omitempty"`
				LT      *int64 `json:"lt,omitempty"`
			} `json:"range"`
		}{
			Range: struct {
				Key     string `json:"key"`
				GTE     *int64 `json:"gte,omitempty"`
				LTE     *int64 `json:"lte,omitempty"`
				GT      *int64 `json:"gt,omitempty"`
				LT      *int64 `json:"lt,omitempty"`
			}{
				Key: "Timestamp",
				LTE: &endUnix,
			},
		}
		filter.Filter.Must = append(filter.Filter.Must, rangeFilter)
	}

	// Exclude project files
	filter.Filter.MustNot = struct {
		Field struct {
			Key   string `json:"key"`
			Match struct {
				Value string `json:"value"`
			} `json:"match"`
		} `json:"field"`
	}{
		Field: struct {
			Key   string `json:"key"`
			Match struct {
				Value string `json:"value"`
			} `json:"match"`
		}{
			Key: "Role",
			Match: struct {
				Value string `json:"value"`
			}{
				Value: string(RoleProject),
			},
		},
	}

	// First, get the count of messages to be deleted
	countFilter := struct {
		Filter struct {
			Must []struct {
				Range struct {
					Key     string `json:"key"`
					GTE     *int64 `json:"gte,omitempty"`
					LTE     *int64 `json:"lte,omitempty"`
					GT      *int64 `json:"gt,omitempty"`
					LT      *int64 `json:"lt,omitempty"`
				} `json:"range"`
			} `json:"must"`
			MustNot struct {
				Field struct {
					Key   string `json:"key"`
					Match struct {
						Value string `json:"value"`
					} `json:"match"`
				} `json:"field"`
			} `json:"must_not"`
		} `json:"filter"`
	}{}
	countFilter.Filter = filter.Filter

	countBody, err := json.Marshal(countFilter)
	if err != nil {
		return 0, err
	}

	countReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/collections/%s/points/count", c.qdrantURL, c.collectionName),
		bytes.NewBuffer(countBody),
	)
	if err != nil {
		return 0, err
	}
	countReq.Header.Set("Content-Type", "application/json")

	countResp, err := c.httpClient.Do(countReq)
	if err != nil {
		return 0, err
	}
	defer countResp.Body.Close()

	if countResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(countResp.Body)
		return 0, fmt.Errorf("failed to count messages: %s", body)
	}

	var countResult struct {
		Result struct {
			Count int `json:"count"`
		} `json:"result"`
	}
	if err := json.NewDecoder(countResp.Body).Decode(&countResult); err != nil {
		return 0, err
	}

	// If no messages to delete, return early
	if countResult.Result.Count == 0 {
		return 0, nil
	}

	// Now delete the messages
	filterBody, err := json.Marshal(filter)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/collections/%s/points/delete", c.qdrantURL, c.collectionName),
		bytes.NewBuffer(filterBody),
	)
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
		return 0, fmt.Errorf("failed to delete messages: %s", body)
	}

	return countResult.Result.Count, nil
}

// DeleteMessagesForCurrentDay deletes all messages from the current day
func (c *MemoryClient) DeleteMessagesForCurrentDay(ctx context.Context) (int, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return c.DeleteMessagesByTimeRange(ctx, startOfDay, now)
}

// DeleteMessagesForCurrentWeek deletes all messages from the current week
func (c *MemoryClient) DeleteMessagesForCurrentWeek(ctx context.Context) (int, error) {
	now := time.Now()
	// Calculate the start of the week (Monday)
	daysToMonday := (int(now.Weekday()) - 1) % 7
	if daysToMonday < 0 {
		daysToMonday += 7
	}
	startOfWeek := time.Date(now.Year(), now.Month(), now.Day()-daysToMonday, 0, 0, 0, 0, now.Location())
	return c.DeleteMessagesByTimeRange(ctx, startOfWeek, now)
}

// DeleteMessagesForCurrentMonth deletes all messages from the current month
func (c *MemoryClient) DeleteMessagesForCurrentMonth(ctx context.Context) (int, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	return c.DeleteMessagesByTimeRange(ctx, startOfMonth, now)
}
