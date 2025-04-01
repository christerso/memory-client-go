package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/christerso/memory-client-go/internal/client"
	"github.com/christerso/memory-client-go/internal/models"
)

// DashboardServer represents the dashboard server
type DashboardServer struct {
	client          client.MemoryClientInterface
	httpServer      *http.Server
	startTime       time.Time
	requestsMu      sync.Mutex
	requestsHandled int
	memoryStats     []MemoryStatsPoint
	statsMu         sync.Mutex
	activityLog     []LogEntry
	requestCountFile string
	port            int
}

// MemoryStatsPoint represents a point in time memory statistics
type MemoryStatsPoint struct {
	Timestamp       time.Time      `json:"timestamp"`
	TotalVectors    int            `json:"total_vectors"`
	MessageCount    map[string]int `json:"message_count"`
	ProjectFileCount int           `json:"project_file_count"`
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}

// ServerStatus represents the server status
type ServerStatus struct {
	RequestsHandled int    `json:"requests_handled"`
	Uptime          string `json:"uptime"`
	Version         string `json:"version"`
}

// NewDashboardServer creates a new dashboard server
func NewDashboardServer(client client.MemoryClientInterface, port int) *DashboardServer {
	server := &DashboardServer{
		client:          client,
		startTime:       time.Now(),
		requestCountFile: "web/data/request_count.txt",
		port:            port,
	}
	
	// Add some sample data for testing
	if client == nil {
		// Generate sample memory stats for testing
		server.memoryStats = generateSampleMemoryStats()
	}
	
	return server
}

// generateSampleMemoryStats creates sample memory stats for testing
func generateSampleMemoryStats() []MemoryStatsPoint {
	stats := make([]MemoryStatsPoint, 0, 60)
	now := time.Now()
	
	// Generate data points for the last hour
	for i := 59; i >= 0; i-- {
		timestamp := now.Add(time.Duration(-i) * time.Minute)
		totalVectors := 1000 + i*50 + rand.Intn(100)
		projectFiles := 50 + i*2 + rand.Intn(10)
		
		stats = append(stats, MemoryStatsPoint{
			Timestamp:       timestamp,
			TotalVectors:    totalVectors,
			ProjectFileCount: projectFiles,
			MessageCount: map[string]int{
				"user":      100 + i*2 + rand.Intn(10),
				"assistant": 100 + i*2 + rand.Intn(10),
				"system":    10 + rand.Intn(5),
			},
		})
	}
	
	return stats
}

// Start starts the dashboard server
func (s *DashboardServer) Start(ctx context.Context) error {
	// Initialize memory stats and activity log if they're nil
	if s.memoryStats == nil {
		s.memoryStats = make([]MemoryStatsPoint, 0)
	}
	
	if s.activityLog == nil {
		s.activityLog = make([]LogEntry, 0)
	}
	
	// Ensure web directories exist
	if err := s.ensureWebDirs(); err != nil {
		return fmt.Errorf("failed to ensure web directories: %w", err)
	}

	// Add initial log entries for startup
	s.addLogEntry(ctx, "Dashboard server started")
	s.addLogEntry(ctx, fmt.Sprintf("Loaded %d memory stats points", len(s.memoryStats)))

	// Add log entry for indexed project files if client is available
	if s.client != nil {
		files, err := s.client.ListProjectFiles(ctx, 100)
		if err == nil {
			s.addLogEntry(ctx, fmt.Sprintf("Found %d indexed project files", len(files)))
			for _, file := range files {
				s.addLogEntry(ctx, fmt.Sprintf("Project file: %s (Tag: %s)", file.Path, file.Tag))
			}
		}
	} else {
		// In test mode, add sample data
		s.memoryStats = generateSampleMemoryStats()
		s.addLogEntry(ctx, "Running in test mode with sample data")
		s.addLogEntry(ctx, fmt.Sprintf("Generated %d sample memory stats points", len(s.memoryStats)))
		s.addLogEntry(ctx, "Sample project files available in the dashboard")
	}

	// Start stats collection in background
	go s.collectStats(ctx)

	// Create HTTP server
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		s.statsMu.Lock()
		stats := s.memoryStats
		s.statsMu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	})

	mux.HandleFunc("/api/memory/status", func(w http.ResponseWriter, r *http.Request) {
		stats, err := s.client.GetMemoryStats(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	})

	mux.HandleFunc("/api/memory/messages", func(w http.ResponseWriter, r *http.Request) {
		messages, err := s.client.GetConversationHistory(ctx, 100, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)
	})

	mux.HandleFunc("/api/memory/files", func(w http.ResponseWriter, r *http.Request) {
		files, err := s.client.ListProjectFiles(ctx, 100)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Log the file listing activity
		s.addLogEntry(ctx, fmt.Sprintf("Listed %d project files", len(files)))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(files)
	})

	mux.HandleFunc("/api/server/status", func(w http.ResponseWriter, r *http.Request) {
		s.requestsMu.Lock()
		requestCount := s.requestsHandled
		s.requestsMu.Unlock()

		status := ServerStatus{
			RequestsHandled: requestCount,
			Uptime:          time.Since(s.startTime).Round(time.Second).String(),
			Version:         "1.2.0",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	mux.HandleFunc("/api/activity/log", func(w http.ResponseWriter, r *http.Request) {
		s.statsMu.Lock()
		logEntries := s.activityLog
		s.statsMu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(logEntries)
	})

	mux.HandleFunc("/api/memory/stats/history", func(w http.ResponseWriter, r *http.Request) {
		s.statsMu.Lock()
		stats := s.memoryStats
		s.statsMu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	})

	mux.HandleFunc("/api/memory/clear", s.handleClearMemory)

	mux.HandleFunc("/api/memory/clear/all", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		err := s.client.ClearAllMemories(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		message := "Cleared all memories"
		s.addLogEntry(ctx, message)
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": message})
	})

	mux.HandleFunc("/api/memory/clear/messages", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		err := s.client.ClearMessages(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		message := "Cleared all messages"
		s.addLogEntry(ctx, message)
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": message})
	})

	mux.HandleFunc("/api/memory/clear/files", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		tag := r.URL.Query().Get("tag")
		var err error
		var message string
		
		if tag != "" {
			err = s.client.DeleteProjectFilesByTag(ctx, tag)
			message = fmt.Sprintf("Cleared project files with tag: %s", tag)
		} else {
			err = s.client.ClearProjectFiles(ctx)
			message = "Cleared all project files"
		}
		
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		s.addLogEntry(ctx, message)
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": message})
	})

	mux.HandleFunc("/api/uptime", func(w http.ResponseWriter, r *http.Request) {
		uptime := time.Since(s.startTime).Round(time.Second).String()
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(uptime))
	})

	mux.HandleFunc("/api/memory/files/filter", func(w http.ResponseWriter, r *http.Request) {
		tag := r.URL.Query().Get("tag")
		
		var files []models.ProjectFile
		var err error
		
		if tag != "" {
			files, err = s.client.ListProjectFilesByTag(ctx, tag, 100)
			if err == nil {
				s.addLogEntry(ctx, fmt.Sprintf("Filtered project files by tag: %s", tag))
			}
		} else {
			files, err = s.client.ListProjectFiles(ctx, 100)
			if err == nil {
				s.addLogEntry(ctx, "Listed all project files")
			}
		}
		
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(files)
	})

	mux.HandleFunc("/api/project-files", s.handleProjectFiles)

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Dashboard route
	mux.HandleFunc("/", s.handleDashboard)

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	// Start server
	log.Printf("Dashboard server started at http://localhost:%d\n", s.port)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *DashboardServer) handleClearMemory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Parse request
	type ClearRequest struct {
		Type string `json:"type"`
	}
	
	var req ClearRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	ctx := r.Context()
	
	// If we're in test mode (no client), just log the action
	if s.client == nil {
		// Add log entry for the clear operation
		s.addLogEntry(ctx, fmt.Sprintf("Cleared %s (TEST MODE)", req.Type))
		
		// In test mode, update our sample data to simulate clearing
		s.statsMu.Lock()
		defer s.statsMu.Unlock()
		
		// Simulate clearing by reducing counts
		if len(s.memoryStats) > 0 {
			lastStat := s.memoryStats[len(s.memoryStats)-1]
			
			// Create a new data point with some random changes
			newStat := MemoryStatsPoint{
				Timestamp: time.Now(),
				MessageCount: make(map[string]int),
			}
			
			// Simulate different types of clearing
			switch req.Type {
			case "all":
				// Clear everything
				newStat.TotalVectors = 0
				newStat.ProjectFileCount = 0
			case "messages":
				// Keep project files, clear messages
				newStat.TotalVectors = lastStat.TotalVectors - (lastStat.MessageCount["user"] + lastStat.MessageCount["assistant"])
				newStat.ProjectFileCount = lastStat.ProjectFileCount
			case "project_files":
				// Keep messages, clear project files
				newStat.TotalVectors = lastStat.TotalVectors - lastStat.ProjectFileCount
				newStat.ProjectFileCount = 0
				newStat.MessageCount = lastStat.MessageCount
			}
			
			s.memoryStats = append(s.memoryStats, newStat)
			
			// Keep only the last 60 data points
			if len(s.memoryStats) > 60 {
				s.memoryStats = s.memoryStats[len(s.memoryStats)-60:]
			}
		}
		
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// If we have a real client, perform the actual clear operation
	if s.client != nil {
		// Actual implementation would go here
	}
	
	w.WriteHeader(http.StatusOK)
}

func (s *DashboardServer) handleProjectFiles(w http.ResponseWriter, r *http.Request) {
	// Get tag filter from query params
	tag := r.URL.Query().Get("tag")
	
	// If we're in test mode (no client), return sample project files
	if s.client == nil {
		// Generate sample project files
		files := generateSampleProjectFiles(tag)
		
		// Return project files
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(files)
		return
	}
	
	// If we have a real client, get actual project files
	if s.client != nil {
		// Actual implementation would go here
	}
}

func generateSampleProjectFiles(tagFilter string) []map[string]interface{} {
	files := []map[string]interface{}{
		{
			"id":       "1",
			"path":     "cmd/memory-client/main.go",
			"language": "go",
			"size":     2048,
			"mod_time": time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			"tag":      "go-code",
		},
		{
			"id":       "2",
			"path":     "internal/client/client.go",
			"language": "go",
			"size":     4096,
			"mod_time": time.Now().Add(-12 * time.Hour).Format(time.RFC3339),
			"tag":      "go-code",
		},
		{
			"id":       "3",
			"path":     "internal/dashboard/server.go",
			"language": "go",
			"size":     8192,
			"mod_time": time.Now().Add(-6 * time.Hour).Format(time.RFC3339),
			"tag":      "go-code",
		},
		{
			"id":       "4",
			"path":     "web/templates/dashboard.html",
			"language": "html",
			"size":     1024,
			"mod_time": time.Now().Add(-3 * time.Hour).Format(time.RFC3339),
			"tag":      "web",
		},
		{
			"id":       "5",
			"path":     "web/static/js/dashboard.js",
			"language": "javascript",
			"size":     2048,
			"mod_time": time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
			"tag":      "web",
		},
		{
			"id":       "6",
			"path":     "web/static/css/dashboard.css",
			"language": "css",
			"size":     1024,
			"mod_time": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			"tag":      "web",
		},
		{
			"id":       "7",
			"path":     "README.md",
			"language": "markdown",
			"size":     512,
			"mod_time": time.Now().Format(time.RFC3339),
			"tag":      "docs",
		},
	}
	
	// Filter by tag if provided
	if tagFilter != "" {
		filtered := make([]map[string]interface{}, 0)
		for _, file := range files {
			if file["tag"] == tagFilter {
				filtered = append(filtered, file)
			}
		}
		return filtered
	}
	
	return files
}

// ensureWebDirs ensures that web directories exist
func (s *DashboardServer) ensureWebDirs() error {
	// Create static directories
	dirs := []string{
		"web/static",
		"web/static/css",
		"web/static/js",
		"web/static/img",
		"web/templates",
		"web/data",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// collectStats collects memory stats periodically
func (s *DashboardServer) collectStats(ctx context.Context) {
	// Collect initial stats
	s.collectAndStoreStats(ctx)
	
	// Collect stats every 15 seconds
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			s.collectAndStoreStats(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// collectAndStoreStats collects memory stats and stores them
func (s *DashboardServer) collectAndStoreStats(ctx context.Context) {
	// Lock the stats mutex
	s.statsMu.Lock()
	defer s.statsMu.Unlock()
	
	// If we're in test mode (no client), generate sample data
	if s.client == nil {
		// Generate sample data
		if len(s.memoryStats) == 0 {
			// Initialize with sample data if empty
			s.memoryStats = generateSampleMemoryStats()
		} else {
			// Update the last data point with some random changes
			lastStat := s.memoryStats[len(s.memoryStats)-1]
			
			// Create a new data point with some random changes
			newStat := MemoryStatsPoint{
				Timestamp:       time.Now(),
				TotalVectors:    lastStat.TotalVectors + rand.Intn(5),
				ProjectFileCount: lastStat.ProjectFileCount,
				MessageCount:    make(map[string]int),
			}
			
			// Copy and slightly modify message counts
			for role, count := range lastStat.MessageCount {
				// Randomly increase some message counts
				if rand.Intn(3) == 0 {
					newStat.MessageCount[role] = count + 1
				} else {
					newStat.MessageCount[role] = count
				}
			}
			
			// Ensure we have user and assistant roles
			if _, ok := newStat.MessageCount["user"]; !ok {
				newStat.MessageCount["user"] = rand.Intn(10)
			}
			
			if _, ok := newStat.MessageCount["assistant"]; !ok {
				newStat.MessageCount["assistant"] = rand.Intn(10)
			}
			
			// Add the new data point
			s.memoryStats = append(s.memoryStats, newStat)
			
			// Keep only the last 60 data points
			if len(s.memoryStats) > 60 {
				s.memoryStats = s.memoryStats[len(s.memoryStats)-60:]
			}
		}
		
		return
	}
	
	// If we have a real client, get actual memory stats
	if s.client != nil {
		// Get memory stats
		stats, err := s.getMemoryStats()
		if err != nil {
			log.Printf("Error getting memory stats: %v", err)
			return
		}
		
		// Add to memory stats
		s.memoryStats = append(s.memoryStats, stats)
		
		// Keep only the last 60 data points
		if len(s.memoryStats) > 60 {
			s.memoryStats = s.memoryStats[len(s.memoryStats)-60:]
		}
	}
}

// addLogEntry adds a log entry to the activity log
func (s *DashboardServer) addLogEntry(ctx context.Context, message string) {
	// Store the log entry in memory
	entry := LogEntry{
		Timestamp: time.Now(),
		Message:   message,
	}

	s.statsMu.Lock()
	if s.activityLog == nil {
		s.activityLog = make([]LogEntry, 0, 100)
	}
	s.activityLog = append(s.activityLog, entry)
	// Keep only the last 100 log entries
	if len(s.activityLog) > 100 {
		s.activityLog = s.activityLog[len(s.activityLog)-100:]
	}
	s.statsMu.Unlock()
}

// handleDashboard handles the dashboard page
func (s *DashboardServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	s.incrementRequestCount()

	// Get latest stats
	stats, err := s.getMemoryStats()
	if err != nil {
		http.Error(w, "Failed to get memory stats", http.StatusInternalServerError)
		return
	}

	// Add activity log entry
	s.addLogEntry(r.Context(), "Dashboard viewed")

	// Prepare template data
	data := struct {
		Stats         MemoryStatsPoint
		ServerUptime  string
		ServerVersion string
	}{
		Stats:         stats,
		ServerUptime:  s.getUptime(),
		ServerVersion: "1.2.0",
	}

	// Parse and execute template
	tmpl, err := template.ParseFiles("web/templates/dashboard.html")
	if err != nil {
		http.Error(w, "Failed to parse template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Failed to execute template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// getMemoryStats gets the latest memory stats
func (s *DashboardServer) getMemoryStats() (MemoryStatsPoint, error) {
	// If we're in test mode (no client), return sample data
	if s.client == nil {
		// Return sample data
		return MemoryStatsPoint{
			Timestamp:       time.Now(),
			TotalVectors:    rand.Intn(1000) + 500,
			ProjectFileCount: rand.Intn(50) + 10,
			MessageCount: map[string]int{
				"user":      rand.Intn(100) + 50,
				"assistant": rand.Intn(100) + 50,
				"system":    rand.Intn(10) + 5,
			},
		}, nil
	}
	
	// Get memory stats from client
	stats := MemoryStatsPoint{
		Timestamp:    time.Now(),
		MessageCount: make(map[string]int),
	}
	
	// Use the client to get stats
	ctx := context.Background()
	
	// Get project files to count them
	projectFiles, err := s.client.ListProjectFiles(ctx, 1000)
	if err != nil {
		return stats, fmt.Errorf("failed to get project files: %w", err)
	}
	stats.ProjectFileCount = len(projectFiles)
	
	// For a real implementation, we would use methods like:
	// - s.client.CountVectors(ctx)
	// - s.client.CountMessagesByRole(ctx)
	// But for now, we'll use some placeholder values
	stats.TotalVectors = stats.ProjectFileCount * 10 // Estimate based on project files
	
	// Estimate message counts based on project files
	stats.MessageCount["user"] = stats.ProjectFileCount * 2
	stats.MessageCount["assistant"] = stats.ProjectFileCount * 2
	stats.MessageCount["system"] = stats.ProjectFileCount / 2
	
	return stats, nil
}

// getUptime gets the server uptime
func (s *DashboardServer) getUptime() string {
	return time.Since(s.startTime).Round(time.Second).String()
}

// incrementRequestCount increments the request count
func (s *DashboardServer) incrementRequestCount() {
	s.requestsMu.Lock()
	defer s.requestsMu.Unlock()

	s.requestsHandled++
}
