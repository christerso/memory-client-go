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
)

// startStatusHTTPServer starts the HTTP server for status page
func (s *MCPServer) startStatusHTTPServer(ctx context.Context) {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		s.requestsMu.Lock()
		requestCount := s.requestsHandled
		s.requestsMu.Unlock()

		uptime := time.Since(s.startTime)

		runtimeMemStats := runtime.MemStats{}
		runtime.ReadMemStats(&runtimeMemStats)

		// Get memory client stats
		clientStats, err := s.client.GetMemoryStats(r.Context())
		var clientStatsMap map[string]interface{}
		
		if err != nil {
			// Create default stats if there's an error
			clientStatsMap = map[string]interface{}{
				"total_vectors":      0,
				"message_count":      map[string]int{"total": 0},
				"project_file_count": 0,
			}
		} else {
			// Convert MemoryStats to map[string]interface{}
			clientStatsMap = map[string]interface{}{
				"total_vectors":      clientStats.TotalVectors,
				"message_count":      clientStats.MessageCount,
				"project_file_count": clientStats.ProjectFileCount,
			}
		}

		status := map[string]interface{}{
			"status":          "running",
			"version":         "1.3.0",
			"uptime":          uptime.String(),
			"requests":        requestCount,
			"memory_usage_mb": int64(runtimeMemStats.Alloc / 1024 / 1024),
			"goroutines":      runtime.NumGoroutine(),
			"client_stats":    clientStatsMap,
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
		Addr:    ":9580",
		Handler: mux,
	}

	// Start HTTP server
	log.Println("Starting HTTP server on :9580")
	s.logOperation("HTTP", "Starting HTTP server on :9580", true)

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

	s.requestsMu.Lock()
	requestCount := s.requestsHandled
	s.requestsMu.Unlock()

	uptime := time.Since(s.startTime)

	runtimeMemStats := runtime.MemStats{}
	runtime.ReadMemStats(&runtimeMemStats)

	// Get memory client stats
	clientStats, err := s.client.GetMemoryStats(r.Context())
	var clientStatsMap map[string]interface{}
	
	if err != nil {
		// Create default stats if there's an error
		clientStatsMap = map[string]interface{}{
			"total_vectors":      0,
			"message_count":      map[string]int{"total": 0},
			"project_file_count": 0,
		}
	} else {
		// Convert MemoryStats to map[string]interface{}
		clientStatsMap = map[string]interface{}{
			"total_vectors":      clientStats.TotalVectors,
			"message_count":      clientStats.MessageCount,
			"project_file_count": clientStats.ProjectFileCount,
		}
	}

	// Get recent operations
	operations := s.getRecentOperations()

	// Create data for template
	data := map[string]interface{}{
		"ServerName":      "Memory Client MCP Server",
		"Version":         "1.3.0",
		"Uptime":          uptime.String(),
		"StartTime":       s.startTime.Format(time.RFC1123),
		"RequestsHandled": requestCount,
		"MemoryUsageMB":   int64(runtimeMemStats.Alloc / 1024 / 1024),
		"Goroutines":      runtime.NumGoroutine(),
		"ClientStats":     clientStatsMap,
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
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            margin: 0;
            padding: 20px;
            color: #e0e0e0;
            background-color: #121212;
            background-image: linear-gradient(rgba(255, 193, 7, 0.05) 1px, transparent 1px), 
                             linear-gradient(90deg, rgba(255, 193, 7, 0.05) 1px, transparent 1px);
            background-size: 20px 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background-color: #1e1e1e;
            padding: 20px;
            border-radius: 6px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.3);
            border: 1px solid #333;
        }
        h1 {
            color: #e0e0e0;
            margin-top: 0;
        }
        h2 {
            color: #4dabf7;
            margin-top: 20px;
        }
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
            gap: 20px;
            margin-bottom: 20px;
        }
        .stat-card {
            background-color: #2d2d2d;
            padding: 15px;
            border-radius: 5px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.3);
            border: 1px solid #333;
        }
        .stat-title {
            font-weight: bold;
            margin-bottom: 5px;
            color: #adb5bd;
        }
        .stat-value {
            font-size: 24px;
            color: #4dabf7;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 10px;
        }
        th, td {
            padding: 8px 12px;
            text-align: left;
            border-bottom: 1px solid #333;
        }
        th {
            background-color: #2d2d2d;
            color: #e0e0e0;
        }
        tr:hover {
            background-color: #2d2d2d;
        }
        .success {
            color: #28a745;
        }
        .error {
            color: #dc3545;
        }
        .refresh-btn {
            background-color: #4dabf7;
            color: white;
            border: none;
            padding: 10px 15px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
            margin-bottom: 20px;
        }
        .refresh-btn:hover {
            background-color: #3c8dbc;
        }
        
        /* Light/Dark mode toggle */
        .theme-toggle {
            position: absolute;
            top: 20px;
            right: 20px;
            display: flex;
            align-items: center;
        }
        .toggle-label {
            margin-right: 8px;
            color: #e0e0e0;
        }
        .toggle-switch {
            position: relative;
            display: inline-block;
            width: 50px;
            height: 24px;
        }
        .toggle-switch input {
            opacity: 0;
            width: 0;
            height: 0;
        }
        .slider {
            position: absolute;
            cursor: pointer;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background-color: #2d2d2d;
            transition: .4s;
            border-radius: 24px;
            border: 1px solid #444;
        }
        .slider:before {
            position: absolute;
            content: "";
            height: 16px;
            width: 16px;
            left: 4px;
            bottom: 3px;
            background-color: #e0e0e0;
            transition: .4s;
            border-radius: 50%;
        }
        input:checked + .slider {
            background-color: #4dabf7;
        }
        input:checked + .slider:before {
            transform: translateX(26px);
        }
    </style>
    <script>
        function toggleTheme() {
            const body = document.body;
            const isDark = body.classList.contains('dark-theme');
            
            if (isDark) {
                body.classList.remove('dark-theme');
                body.style.backgroundColor = '#f5f5f5';
                body.style.color = '#333';
                body.style.backgroundImage = 'none';
                document.querySelectorAll('.container').forEach(el => {
                    el.style.backgroundColor = 'white';
                    el.style.boxShadow = '0 2px 5px rgba(0,0,0,0.1)';
                    el.style.border = 'none';
                });
                document.querySelectorAll('h1').forEach(el => el.style.color = '#2c3e50');
                document.querySelectorAll('h2').forEach(el => el.style.color = '#3498db');
                document.querySelectorAll('.stat-card').forEach(el => {
                    el.style.backgroundColor = '#f9f9f9';
                    el.style.boxShadow = '0 1px 3px rgba(0,0,0,0.1)';
                    el.style.border = 'none';
                });
                document.querySelectorAll('.stat-title').forEach(el => el.style.color = '#7f8c8d');
                document.querySelectorAll('.stat-value').forEach(el => el.style.color = '#2c3e50');
                document.querySelectorAll('th').forEach(el => {
                    el.style.backgroundColor = '#f2f2f2';
                    el.style.color = '#333';
                });
                document.querySelectorAll('th, td').forEach(el => el.style.borderColor = '#ddd');
                document.querySelectorAll('tr:hover').forEach(el => el.style.backgroundColor = '#f5f5f5');
            } else {
                body.classList.add('dark-theme');
                body.style.backgroundColor = '#121212';
                body.style.color = '#e0e0e0';
                body.style.backgroundImage = 'linear-gradient(rgba(255, 193, 7, 0.05) 1px, transparent 1px), linear-gradient(90deg, rgba(255, 193, 7, 0.05) 1px, transparent 1px)';
                document.querySelectorAll('.container').forEach(el => {
                    el.style.backgroundColor = '#1e1e1e';
                    el.style.boxShadow = '0 2px 5px rgba(0,0,0,0.3)';
                    el.style.border = '1px solid #333';
                });
                document.querySelectorAll('h1').forEach(el => el.style.color = '#e0e0e0');
                document.querySelectorAll('h2').forEach(el => el.style.color = '#4dabf7');
                document.querySelectorAll('.stat-card').forEach(el => {
                    el.style.backgroundColor = '#2d2d2d';
                    el.style.boxShadow = '0 1px 3px rgba(0,0,0,0.3)';
                    el.style.border = '1px solid #333';
                });
                document.querySelectorAll('.stat-title').forEach(el => el.style.color = '#adb5bd');
                document.querySelectorAll('.stat-value').forEach(el => el.style.color = '#4dabf7');
                document.querySelectorAll('th').forEach(el => {
                    el.style.backgroundColor = '#2d2d2d';
                    el.style.color = '#e0e0e0';
                });
                document.querySelectorAll('th, td').forEach(el => el.style.borderColor = '#333');
                document.querySelectorAll('tr:hover').forEach(el => el.style.backgroundColor = '#2d2d2d');
            }
        }
    </script>
</head>
<body class="dark-theme">
    <div class="theme-toggle">
        <span class="toggle-label">Theme</span>
        <label class="toggle-switch">
            <input type="checkbox" checked onclick="toggleTheme()">
            <span class="slider"></span>
        </label>
    </div>
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
                <div class="stat-value">{{.ClientStats.total_vectors}}</div>
            </div>
            <div class="stat-card">
                <div class="stat-title">Messages</div>
                <div class="stat-value">{{index .ClientStats.message_count "total"}}</div>
            </div>
            <div class="stat-card">
                <div class="stat-title">Project Files</div>
                <div class="stat-value">{{.ClientStats.project_file_count}}</div>
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
