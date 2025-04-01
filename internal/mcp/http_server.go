package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/user/memory-client-go/internal/models"
)

// startHTTPServer starts the HTTP server for status page
func (s *MCPServer) startHTTPServer(ctx context.Context) {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		s.requestsMu.Lock()
		s.requestsHandled++
		requestCount := s.requestsHandled
		s.requestsMu.Unlock()

		uptime := time.Since(s.startTime).Round(time.Second)

		// Get memory stats
		memStats := runtime.MemStats{}
		runtime.ReadMemStats(&memStats)

		// Get memory client stats
		clientStats, err := s.client.GetMemoryStats(r.Context())
		if err != nil {
			clientStats = &models.MemoryStats{
				TotalVectors:     0,
				MessageCount:     map[string]int{"total": 0},
				ProjectFileCount: 0,
			}
		}

		status := map[string]interface{}{
			"status":          "running",
			"uptime":          uptime.String(),
			"requests":        requestCount,
			"memory_usage_mb": memStats.Alloc / 1024 / 1024,
			"goroutines":      runtime.NumGoroutine(),
			"client_stats":    clientStats,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	mux.HandleFunc("/api/operations", func(w http.ResponseWriter, r *http.Request) {
		operations := s.getRecentOperations()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(operations)
	})

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	// Status page
	mux.HandleFunc("/", s.serveStatusPage)

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start HTTP server
	log.Println("Starting HTTP server on :8080")
	s.logOperation("HTTP", "Starting HTTP server on :8080", true)

	if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("HTTP server error: %v", err)
		s.logOperation("HTTP", fmt.Sprintf("HTTP server error: %v", err), false)
	}
}

// serveStatusPage serves the status page
func (s *MCPServer) serveStatusPage(w http.ResponseWriter, r *http.Request) {
	// Create template
	tmpl, err := template.New("status").Parse(statusPageTemplate)
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}

	// Get uptime
	uptime := time.Since(s.startTime).Round(time.Second)

	// Get request count
	s.requestsMu.Lock()
	requestCount := s.requestsHandled
	s.requestsMu.Unlock()

	// Get memory stats
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)

	// Get memory client stats
	clientStats, err := s.client.GetMemoryStats(r.Context())
	if err != nil {
		clientStats = &models.MemoryStats{
			TotalVectors:     0,
			MessageCount:     map[string]int{"total": 0},
			ProjectFileCount: 0,
		}
	}

	// Get recent operations
	operations := s.getRecentOperations()

	// Create data for template
	data := map[string]interface{}{
		"ServerName":      "Memory Client MCP Server",
		"Version":         "1.2.0",
		"Uptime":          uptime.String(),
		"StartTime":       s.startTime.Format(time.RFC1123),
		"RequestsHandled": requestCount,
		"MemoryUsageMB":   memStats.Alloc / 1024 / 1024,
		"Goroutines":      runtime.NumGoroutine(),
		"ClientStats":     clientStats,
		"Operations":      operations,
	}

	// Execute template
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		return
	}
}

// statusPageTemplate is the HTML template for the status page
const statusPageTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.ServerName}} - Status</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 0;
            padding: 20px;
            color: #333;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background-color: white;
            padding: 20px;
            border-radius: 5px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.1);
        }
        h1 {
            color: #2c3e50;
            margin-top: 0;
        }
        h2 {
            color: #3498db;
            margin-top: 20px;
        }
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
            gap: 20px;
            margin-bottom: 20px;
        }
        .stat-card {
            background-color: #f9f9f9;
            padding: 15px;
            border-radius: 5px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }
        .stat-title {
            font-weight: bold;
            margin-bottom: 5px;
            color: #7f8c8d;
        }
        .stat-value {
            font-size: 24px;
            color: #2c3e50;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 10px;
        }
        th, td {
            padding: 8px 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        th {
            background-color: #f2f2f2;
        }
        tr:hover {
            background-color: #f5f5f5;
        }
        .success {
            color: #27ae60;
        }
        .error {
            color: #e74c3c;
        }
        .refresh-btn {
            background-color: #3498db;
            color: white;
            border: none;
            padding: 10px 15px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
            margin-bottom: 20px;
        }
        .refresh-btn:hover {
            background-color: #2980b9;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>{{.ServerName}} Status</h1>
        <p>Version: {{.Version}}</p>
        
        <button class="refresh-btn" onclick="location.reload()">Refresh Status</button>
        
        <h2>Server Statistics</h2>
        <div class="stats">
            <div class="stat-card">
                <div class="stat-title">Uptime</div>
                <div class="stat-value">{{.Uptime}}</div>
                <div>Since {{.StartTime}}</div>
            </div>
            <div class="stat-card">
                <div class="stat-title">Requests Handled</div>
                <div class="stat-value">{{.RequestsHandled}}</div>
            </div>
            <div class="stat-card">
                <div class="stat-title">Memory Usage</div>
                <div class="stat-value">{{.MemoryUsageMB}} MB</div>
            </div>
            <div class="stat-card">
                <div class="stat-title">Goroutines</div>
                <div class="stat-value">{{.Goroutines}}</div>
            </div>
        </div>
        
        <h2>Memory Client Statistics</h2>
        <div class="stats">
            <div class="stat-card">
                <div class="stat-title">Total Vectors</div>
                <div class="stat-value">{{.ClientStats.TotalVectors}}</div>
            </div>
            <div class="stat-card">
                <div class="stat-title">Messages</div>
                <div class="stat-value">{{index .ClientStats.MessageCount "total"}}</div>
            </div>
            <div class="stat-card">
                <div class="stat-title">Project Files</div>
                <div class="stat-value">{{.ClientStats.ProjectFileCount}}</div>
            </div>
        </div>
        
        <h2>Recent Operations</h2>
        <table>
            <thead>
                <tr>
                    <th>Timestamp</th>
                    <th>Operation</th>
                    <th>Details</th>
                    <th>Status</th>
                </tr>
            </thead>
            <tbody>
                {{range .Operations}}
                <tr>
                    <td>{{.Timestamp.Format "2006-01-02 15:04:05"}}</td>
                    <td>{{.Operation}}</td>
                    <td>{{.Details}}</td>
                    <td class="{{if .Success}}success{{else}}error{{end}}">
                        {{if .Success}}Success{{else}}Failed{{end}}
                    </td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>
</body>
</html>
`
