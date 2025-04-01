package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/user/memory-client-go/internal/models"
)

// IndexProjectFiles indexes all files in a project directory
func (c *MemoryClient) IndexProjectFiles(ctx context.Context, projectPath, tag string) (int, error) {
	if c.verbose {
		fmt.Printf("Indexing project directory: %s\n", projectPath)
		if tag != "" {
			fmt.Printf("Using tag: %s\n", tag)
		}
	}

	// Get list of files to process
	filesToProcess, err := c.getProjectFiles(projectPath)
	if err != nil {
		return 0, fmt.Errorf("failed to get project files: %w", err)
	}

	if c.verbose {
		fmt.Printf("Found %d files to index\n", len(filesToProcess))
	}

	// Process files
	count := 0
	for i, path := range filesToProcess {
		if c.verbose && len(filesToProcess) > 10 {
			progress := float64(i+1) / float64(len(filesToProcess)) * 100
			fmt.Printf("Progress: %d%% (%d/%d files)\n", int(progress), i+1, len(filesToProcess))
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", path, err)
			continue
		}

		// Skip empty files
		if len(content) == 0 {
			continue
		}

		// Skip binary files
		if isBinary(content) {
			continue
		}

		// Create project file
		relPath, err := filepath.Rel(projectPath, path)
		if err != nil {
			relPath = path
		}

		// Use forward slashes for consistency
		relPath = strings.ReplaceAll(relPath, "\\", "/")

		// Detect language based on file extension
		ext := strings.ToLower(filepath.Ext(path))
		language := "unknown"
		if lang, ok := models.LanguageMap[ext]; ok {
			language = lang
		}

		projectFile := models.ProjectFile{
			ID:        generateID(),
			Path:      relPath,
			Content:   string(content),
			Timestamp: time.Now(),
			Tag:       tag,
			Language:  language,
			ModTime:   time.Now().Unix(),
		}

		// Index file
		err = c.indexProjectFile(ctx, projectFile)
		if err != nil {
			fmt.Printf("Error indexing file %s: %v\n", path, err)
			continue
		}

		count++
	}

	if c.verbose {
		fmt.Printf("Successfully indexed %d files\n", count)
	}

	return count, nil
}

// UpdateProjectFiles updates modified project files
func (c *MemoryClient) UpdateProjectFiles(ctx context.Context, projectPath string) (int, int, error) {
	if c.verbose {
		fmt.Printf("Updating project files in: %s\n", projectPath)
	}

	// Get list of files to process
	filesToProcess, err := c.getProjectFiles(projectPath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get project files: %w", err)
	}

	// Get existing project files
	existingFiles, err := c.getExistingProjectFiles(ctx, projectPath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get existing project files: %w", err)
	}

	// Create map of existing files
	existingFileMap := make(map[string]models.ProjectFile)
	for _, file := range existingFiles {
		existingFileMap[file.Path] = file
	}

	// Process files
	newCount := 0
	updateCount := 0

	for _, path := range filesToProcess {
		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", path, err)
			continue
		}

		// Skip empty files
		if len(content) == 0 {
			continue
		}

		// Skip binary files
		if isBinary(content) {
			continue
		}

		// Create project file
		relPath, err := filepath.Rel(projectPath, path)
		if err != nil {
			relPath = path
		}

		// Use forward slashes for consistency
		relPath = strings.ReplaceAll(relPath, "\\", "/")

		// Check if file exists
		existingFile, exists := existingFileMap[relPath]
		if exists {
			// Check if content has changed
			if existingFile.Content == string(content) {
				continue
			}

			// Update file
			existingFile.Content = string(content)
			existingFile.Timestamp = time.Now()

			err = c.indexProjectFile(ctx, existingFile)
			if err != nil {
				fmt.Printf("Error updating file %s: %v\n", relPath, err)
				continue
			}

			updateCount++
		} else {
			// New file
			ext := strings.ToLower(filepath.Ext(path))
			language := "unknown"
			if lang, ok := models.LanguageMap[ext]; ok {
				language = lang
			}

			projectFile := models.ProjectFile{
				ID:        generateID(),
				Path:      relPath,
				Content:   string(content),
				Timestamp: time.Now(),
				Tag:       "", // No tag for updates
				Language:  language,
				ModTime:   time.Now().Unix(),
			}

			err = c.indexProjectFile(ctx, projectFile)
			if err != nil {
				fmt.Printf("Error indexing file %s: %v\n", relPath, err)
				continue
			}

			newCount++
		}
	}

	if c.verbose {
		fmt.Printf("Successfully added %d new files and updated %d files\n", newCount, updateCount)
	}

	return newCount, updateCount, nil
}

// SearchProjectFiles searches for content in project files
func (c *MemoryClient) SearchProjectFiles(ctx context.Context, query string, limit int) ([]models.ProjectFile, error) {
	// Generate embedding for query
	embedding, err := c.generateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Search for similar project files
	url := fmt.Sprintf("%s/collections/%s/points/search", c.qdrantURL, c.collectionName)

	request := map[string]interface{}{
		"vector":       embedding,
		"limit":        limit,
		"with_payload": true,
		"with_vector":  false,
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
		return nil, fmt.Errorf("failed to search project files: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Result []struct {
			ID      string `json:"id"`
			Score   float64 `json:"score"`
			Payload struct {
				Path      string    `json:"path"`
				Content   string    `json:"content"`
				Timestamp string    `json:"timestamp"`
				Type      string    `json:"type"`
				Tag       string    `json:"tag"`
			} `json:"payload"`
		} `json:"result"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	files := make([]models.ProjectFile, 0, len(result.Result))
	for _, item := range result.Result {
		timestamp, err := time.Parse(time.RFC3339, item.Payload.Timestamp)
		if err != nil {
			timestamp = time.Now() // Fallback to current time if parsing fails
		}

		file := models.ProjectFile{
			ID:        item.ID,
			Path:      item.Payload.Path,
			Content:   item.Payload.Content,
			Timestamp: timestamp,
			Score:     item.Score,
			Tag:       item.Payload.Tag,
		}
		files = append(files, file)
	}

	return files, nil
}

// DeleteProjectFile deletes a project file by ID
func (c *MemoryClient) DeleteProjectFile(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/collections/%s/points/delete", c.qdrantURL, c.collectionName)

	request := map[string]interface{}{
		"points": []string{id},
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
		return fmt.Errorf("failed to delete project file: %s - %s", resp.Status, string(body))
	}

	return nil
}

// DeleteAllProjectFiles deletes all project files
func (c *MemoryClient) DeleteAllProjectFiles(ctx context.Context) error {
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
		return fmt.Errorf("failed to delete all project files: %s - %s", resp.Status, string(body))
	}

	return nil
}

// ListProjectFiles retrieves a list of project files with a specified limit
func (c *MemoryClient) ListProjectFiles(ctx context.Context, limit int) ([]models.ProjectFile, error) {
	url := fmt.Sprintf("%s/collections/%s/points/scroll", c.qdrantURL, c.collectionName)

	request := map[string]interface{}{
		"limit":        limit,
		"with_payload": true,
		"with_vector":  false,
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

	var result struct {
		Result struct {
			Points []struct {
				ID      string `json:"id"`
				Payload struct {
					Path      string `json:"path"`
					Content   string `json:"content"`
					Timestamp string `json:"timestamp"`
					Type      string `json:"type"`
					Tag       string `json:"tag"`
					Language  string `json:"language"`
					ModTime   int64  `json:"mod_time"`
				} `json:"payload"`
			} `json:"points"`
		} `json:"result"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	files := make([]models.ProjectFile, 0, len(result.Result.Points))
	for _, point := range result.Result.Points {
		timestamp, err := time.Parse(time.RFC3339, point.Payload.Timestamp)
		if err != nil {
			timestamp = time.Now() // Fallback to current time if parsing fails
		}

		file := models.ProjectFile{
			ID:        point.ID,
			Path:      point.Payload.Path,
			Content:   point.Payload.Content,
			Timestamp: timestamp,
			Tag:       point.Payload.Tag,
			Language:  point.Payload.Language,
			ModTime:   point.Payload.ModTime,
		}
		files = append(files, file)
	}

	return files, nil
}

// Helper functions

// getProjectFiles gets all files in a project directory
func (c *MemoryClient) getProjectFiles(projectPath string) ([]string, error) {
	var filesToProcess []string

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// Skip hidden directories
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		// Skip binary files and non-text files
		ext := strings.ToLower(filepath.Ext(path))
		if isIgnoredExtension(ext) {
			return nil
		}

		filesToProcess = append(filesToProcess, path)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return filesToProcess, nil
}

// isIgnoredExtension checks if a file extension should be ignored
func isIgnoredExtension(ext string) bool {
	ignoredExtensions := map[string]bool{
		".exe":  true,
		".dll":  true,
		".so":   true,
		".dylib": true,
		".a":    true,
		".o":    true,
		".obj":  true,
		".bin":  true,
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".ico":  true,
		".mp3":  true,
		".mp4":  true,
		".wav":  true,
		".avi":  true,
		".mov":  true,
		".zip":  true,
		".tar":  true,
		".gz":   true,
		".7z":   true,
		".rar":  true,
		".pdf":  true,
		".doc":  true,
		".docx": true,
		".xls":  true,
		".xlsx": true,
		".ppt":  true,
		".pptx": true,
	}

	return ignoredExtensions[ext]
}

// isBinary checks if content is binary
func isBinary(content []byte) bool {
	// Check for null bytes which are common in binary files
	for _, b := range content {
		if b == 0 {
			return true
		}
	}

	// Check if the content contains too many non-printable characters
	nonPrintable := 0
	for _, b := range content {
		if b < 32 && b != 9 && b != 10 && b != 13 { // Tab, LF, CR are allowed
			nonPrintable++
		}
	}

	// If more than 10% of the content is non-printable, consider it binary
	return float64(nonPrintable)/float64(len(content)) > 0.1
}

// getExistingProjectFiles gets existing project files from the database
func (c *MemoryClient) getExistingProjectFiles(ctx context.Context, projectPath string) ([]models.ProjectFile, error) {
	url := fmt.Sprintf("%s/collections/%s/points/scroll", c.qdrantURL, c.collectionName)

	request := map[string]interface{}{
		"limit":        1000, // Adjust as needed
		"with_payload": true,
		"with_vector":  false,
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
		return nil, fmt.Errorf("failed to get existing project files: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Result struct {
			Points []struct {
				ID      string `json:"id"`
				Payload struct {
					Path      string `json:"path"`
					Content   string `json:"content"`
					Timestamp string `json:"timestamp"`
					Type      string `json:"type"`
					Tag       string `json:"tag"`
					Language  string `json:"language"`
					ModTime   int64  `json:"mod_time"`
				} `json:"payload"`
			} `json:"points"`
		} `json:"result"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	files := make([]models.ProjectFile, 0, len(result.Result.Points))
	for _, point := range result.Result.Points {
		timestamp, err := time.Parse(time.RFC3339, point.Payload.Timestamp)
		if err != nil {
			timestamp = time.Now() // Fallback to current time if parsing fails
		}

		file := models.ProjectFile{
			ID:        point.ID,
			Path:      point.Payload.Path,
			Content:   point.Payload.Content,
			Timestamp: timestamp,
			Tag:       point.Payload.Tag,
			Language:  point.Payload.Language,
			ModTime:   point.Payload.ModTime,
		}
		files = append(files, file)
	}

	return files, nil
}

// indexProjectFile indexes a project file
func (c *MemoryClient) indexProjectFile(ctx context.Context, file models.ProjectFile) error {
	// Generate embedding for file content
	embedding, err := c.generateEmbedding(ctx, file.Content)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Detect language from file extension if not already set
	if file.Language == "" {
		ext := strings.ToLower(filepath.Ext(file.Path))
		if lang, ok := models.LanguageMap[ext]; ok {
			file.Language = lang
		} else {
			file.Language = "Text"
		}
	}

	// Set mod time if not already set
	if file.ModTime == 0 {
		file.ModTime = time.Now().Unix()
	}

	// Create point
	url := fmt.Sprintf("%s/collections/%s/points", c.qdrantURL, c.collectionName)
	
	point := map[string]interface{}{
		"id": file.ID,
		"vector": embedding,
		"payload": map[string]interface{}{
			"path":      file.Path,
			"content":   file.Content,
			"timestamp": file.Timestamp.Format(time.RFC3339),
			"type":      "project_file",
			"tag":       file.Tag,
			"language":  file.Language,
			"mod_time":  file.ModTime,
		},
	}

	// Add point to collection
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
		return fmt.Errorf("failed to index project file: %s - %s", resp.Status, string(body))
	}

	return nil
}
