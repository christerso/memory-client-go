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

	"github.com/user/memory-client-go/internal/client"
	"github.com/user/memory-client-go/internal/models"
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
	// Ensure web directories exist
	if err := s.ensureWebDirs(); err != nil {
		return fmt.Errorf("failed to ensure web directories: %w", err)
	}

	// Add initial log entries for startup
	s.addLogEntry(ctx, "Dashboard server started")
	s.addLogEntry(ctx, fmt.Sprintf("Loaded %d memory stats points", len(s.memoryStats)))

	// Add log entry for indexed project files
	files, err := s.client.ListProjectFiles(ctx, 100)
	if err == nil {
		s.addLogEntry(ctx, fmt.Sprintf("Found %d indexed project files", len(files)))
		for _, file := range files {
			s.addLogEntry(ctx, fmt.Sprintf("Project file: %s (Tag: %s)", file.Path, file.Tag))
		}
	}

	// Start stats collection in background
	go s.collectStats(ctx)

	// Start HTTP server
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

	mux.HandleFunc("/api/memory/clear", func(w http.ResponseWriter, r *http.Request) {
		clearType := r.URL.Query().Get("type")
		tag := r.URL.Query().Get("tag")
		
		var err error
		var message string
		
		switch clearType {
		case "all":
			err = s.client.ClearAllMemories(ctx)
			message = "Cleared all memories"
		case "messages":
			err = s.client.ClearMessages(ctx)
			message = "Cleared all messages"
		case "project_files":
			if tag != "" {
				err = s.client.DeleteProjectFilesByTag(ctx, tag)
				message = fmt.Sprintf("Cleared project files with tag: %s", tag)
			} else {
				err = s.client.ClearProjectFiles(ctx)
				message = "Cleared all project files"
			}
		default:
			http.Error(w, "Invalid clear type", http.StatusBadRequest)
			return
		}
		
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		s.addLogEntry(ctx, message)
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": message})
	})

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
	s.statsMu.Lock()
	defer s.statsMu.Unlock()
	
	// If we're in test mode (no client), generate sample data
	if s.client == nil {
		// Add a new data point with random variations
		if len(s.memoryStats) > 0 {
			lastStat := s.memoryStats[len(s.memoryStats)-1]
			newStat := MemoryStatsPoint{
				Timestamp:       time.Now(),
				TotalVectors:    lastStat.TotalVectors + rand.Intn(100) - 50,
				ProjectFileCount: lastStat.ProjectFileCount + rand.Intn(10) - 5,
				MessageCount: map[string]int{
					"user":      lastStat.MessageCount["user"] + rand.Intn(5) - 2,
					"assistant": lastStat.MessageCount["assistant"] + rand.Intn(5) - 2,
					"system":    lastStat.MessageCount["system"] + rand.Intn(2) - 1,
				},
			}
			
			// Ensure counts don't go below zero
			if newStat.TotalVectors < 0 {
				newStat.TotalVectors = 0
			}
			if newStat.ProjectFileCount < 0 {
				newStat.ProjectFileCount = 0
			}
			for role, count := range newStat.MessageCount {
				if count < 0 {
					newStat.MessageCount[role] = 0
				}
			}
			
			s.memoryStats = append(s.memoryStats, newStat)
		}
		
		// Keep only the last 60 data points
		if len(s.memoryStats) > 60 {
			s.memoryStats = s.memoryStats[len(s.memoryStats)-60:]
		}
		return
	}
	
	// If we have a real client, get actual stats
	if s.client != nil {
		// Get memory stats from client
		// This is the actual implementation that would be used with a real client
		// For now, we're using the test implementation above
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
	s.statsMu.Lock()
	defer s.statsMu.Unlock()

	if len(s.memoryStats) == 0 {
		return MemoryStatsPoint{}, nil
	}

	return s.memoryStats[len(s.memoryStats)-1], nil
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
