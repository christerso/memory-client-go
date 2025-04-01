package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

const (
	// Version information
	Version   = "1.2.0"
	BuildDate = "2025-04-01"
	Author    = "Christer Söderlund"
)

var (
	// Command line flags
	qdrantURL      string
	collectionName string
	embeddingSize  int
	verbose        bool
)

var (
	projectPath string
)

var rootCmd = &cobra.Command{
	Use:   "memory-client",
	Short: "MCP Memory Client for persistent conversation storage",
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start memory server daemon",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client := initClient()
		defer client.Close()

		fmt.Println("Starting memory server daemon...")
		runBackgroundIndexer(ctx, client)
		select {} // Block forever
	},
}

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start as MCP server",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		client := initClient()
		defer client.Close()

		if verbose {
			fmt.Println("Starting MCP memory server...")
		}

		server := NewMCPServer(client)
		if err := server.Start(ctx); err != nil {
			if verbose {
				fmt.Printf("MCP server error: %v\n", err)
			}
			os.Exit(1)
		}
	},
}

var addCmd = &cobra.Command{
	Use:   "add [role] [content]",
	Short: "Add a message to memory",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		client := initClient()
		defer client.Close()

		msg := NewMessage(Role(args[0]), args[1])
		if err := client.AddMessage(ctx, msg); err != nil {
			fmt.Printf("Error adding message: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Message added successfully")
	},
}

var indexProjectCmd = &cobra.Command{
	Use:   "index-project [path]",
	Short: "Index project files in a directory",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Increase timeout to 10 minutes for large projects
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		client := initClient()
		defer client.Close()

		// Use provided path or current directory
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		// Use projectPath flag if provided
		if projectPath != "" {
			path = projectPath
		}

		fmt.Printf("Indexing project files in: %s\n", path)
		count, err := client.IndexProjectFiles(ctx, path)
		if err != nil {
			fmt.Printf("Error indexing project: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully indexed %d project files\n", count)
	},
}

var updateProjectCmd = &cobra.Command{
	Use:   "update-project [path]",
	Short: "Update modified project files in a directory",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Increase timeout to 10 minutes for large projects
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		client := initClient()
		defer client.Close()

		// Use provided path or current directory
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		// Use projectPath flag if provided
		if projectPath != "" {
			path = projectPath
		}

		fmt.Printf("Updating project files in: %s\n", path)
		newCount, updateCount, err := client.UpdateProjectFiles(ctx, path)
		if err != nil {
			fmt.Printf("Error updating project: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Project update complete: %d new files, %d updated files\n", newCount, updateCount)
	},
}

var watchProjectCmd = &cobra.Command{
	Use:   "watch-project [path]",
	Short: "Watch a project directory for changes and update automatically",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		client := initClient()
		defer client.Close()

		// Use provided path or current directory
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		// Use projectPath flag if provided
		if projectPath != "" {
			path = projectPath
		}

		// Get absolute path
		absPath, err := filepath.Abs(path)
		if err != nil {
			fmt.Printf("Error getting absolute path: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Watching project directory: %s\n", absPath)
		fmt.Println("Press Ctrl+C to stop")

		// First, index the project
		_, err = client.IndexProjectFiles(ctx, absPath)
		if err != nil {
			fmt.Printf("Error indexing project: %v\n", err)
			os.Exit(1)
		}

		// Set up a ticker to check for changes every 5 seconds
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		// Handle Ctrl+C
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

		// Watch for changes
		for {
			select {
			case <-ticker.C:
				newCount, updateCount, err := client.UpdateProjectFiles(ctx, absPath)
				if err != nil {
					fmt.Printf("Error updating project: %v\n", err)
					continue
				}

				if newCount > 0 || updateCount > 0 {
					fmt.Printf("Project updated: %d new files, %d modified files\n", newCount, updateCount)
				}
			case <-sigCh:
				fmt.Println("Stopping project watcher")
				return
			}
		}
	},
}

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search conversation memory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		client := initClient()
		defer client.Close()

		results, err := client.SearchMessages(ctx, args[0], 10)
		if err != nil {
			fmt.Printf("Search failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Found %d results:\n", len(results))
		for i, msg := range results {
			fmt.Printf("%d. [%s] %s\n", i+1, msg.Role, msg.Content)
		}
	},
}

var searchProjectCmd = &cobra.Command{
	Use:   "search-project [query]",
	Short: "Search project files",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		client := initClient()
		defer client.Close()

		results, err := client.SearchProjectFiles(ctx, args[0], 10)
		if err != nil {
			fmt.Printf("Project search failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Found %d project files:\n", len(results))
		for i, file := range results {
			fmt.Printf("%d. [%s] %s\n", i+1, file.Path, file.Content[:min(100, len(file.Content))]+"...")
		}
	},
}

var historyCmd = &cobra.Command{
	Use:   "history [limit]",
	Short: "Show conversation history",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		client := initClient()
		defer client.Close()

		limit := 10
		if len(args) > 0 {
			fmt.Sscanf(args[0], "%d", &limit)
		}

		messages, err := client.GetConversationHistory(ctx, limit)
		if err != nil {
			fmt.Printf("Error retrieving history: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Conversation History (last %d messages):\n", limit)
		for i, msg := range messages {
			fmt.Printf("%d. [%s] %s\n", i+1, msg.Role, msg.Content)
		}
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check if the MCP server is running",
	Run: func(cmd *cobra.Command, args []string) {
		// Check if the process is running
		processes, err := findProcessByName("memory-client")
		if err != nil {
			fmt.Printf("Error checking process status: %v\n", err)
			return
		}

		mcpServerRunning := false
		for _, proc := range processes {
			// Check if it's running with mcp-server command
			if strings.Contains(proc.CommandLine, "mcp-server") {
				mcpServerRunning = true
				fmt.Printf("MCP server is running (PID: %d)\n", proc.PID)
				fmt.Printf("Command: %s\n", proc.CommandLine)
				fmt.Printf("Started at: %s\n", proc.StartTime.Format(time.RFC3339))
				fmt.Printf("Running for: %s\n", time.Since(proc.StartTime).Round(time.Second))
				break
			}
		}

		if !mcpServerRunning {
			fmt.Println("MCP server is not running")
			fmt.Println("To start the server, run: memory-client mcp-server")
		}
	},
}

var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Completely purge all data from Qdrant",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		client := initClient()
		defer client.Close()

		fmt.Println("WARNING: This will delete ALL data from Qdrant!")
		fmt.Println("This action cannot be undone.")
		fmt.Print("Are you sure you want to continue? (y/N): ")
		
		var response string
		fmt.Scanln(&response)
		
		if strings.ToLower(response) != "y" {
			fmt.Println("Operation cancelled.")
			return
		}

		fmt.Println("Purging all data from Qdrant...")
		err := client.PurgeQdrant(ctx)
		if err != nil {
			fmt.Printf("Error purging data: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("All data has been purged successfully.")
	},
}

var clearCmd = &cobra.Command{
	Use:   "clear [period]",
	Short: "Clear messages for a specific time period",
	Long: `Clear messages for a specific time period.
Available periods:
  day     - Clear messages from the current day
  week    - Clear messages from the current week
  month   - Clear messages from the current month
  range   - Clear messages from a specific date range (requires --from and --to flags)`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		client := initClient()
		defer client.Close()

		period := args[0]
		
		var count int
		var err error
		
		switch period {
		case "day":
			fmt.Println("Clearing messages from the current day...")
			count, err = client.DeleteMessagesForCurrentDay(ctx)
		case "week":
			fmt.Println("Clearing messages from the current week...")
			count, err = client.DeleteMessagesForCurrentWeek(ctx)
		case "month":
			fmt.Println("Clearing messages from the current month...")
			count, err = client.DeleteMessagesForCurrentMonth(ctx)
		case "range":
			fromStr, _ := cmd.Flags().GetString("from")
			toStr, _ := cmd.Flags().GetString("to")
			
			if fromStr == "" || toStr == "" {
				fmt.Println("Error: --from and --to flags are required for range period")
				os.Exit(1)
			}
			
			from, err := time.Parse("2006-01-02", fromStr)
			if err != nil {
				fmt.Printf("Error parsing from date: %v\n", err)
				fmt.Println("Date format should be YYYY-MM-DD")
				os.Exit(1)
			}
			
			to, err := time.Parse("2006-01-02", toStr)
			if err != nil {
				fmt.Printf("Error parsing to date: %v\n", err)
				fmt.Println("Date format should be YYYY-MM-DD")
				os.Exit(1)
			}
			
			// Set to end of day for the to date
			to = time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, to.Location())
			
			fmt.Printf("Clearing messages from %s to %s...\n", from.Format("2006-01-02"), to.Format("2006-01-02"))
			count, err = client.DeleteMessagesByTimeRange(ctx, from, to)
		default:
			fmt.Printf("Unknown period: %s\n", period)
			fmt.Println("Available periods: day, week, month, range")
			os.Exit(1)
		}
		
		if err != nil {
			fmt.Printf("Error clearing messages: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("Successfully deleted %d messages.\n", count)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Memory Client for MCP v%s\n", Version)
		fmt.Printf("Build date: %s\n", BuildDate)
		fmt.Printf("Author: %s\n", Author)
		fmt.Println("Copyright (c) 2025")
		fmt.Println("License: MIT")
	},
}

// ProcessInfo contains information about a running process
type ProcessInfo struct {
	PID         int
	CommandLine string
	StartTime   time.Time
}

// findProcessByName finds processes by name (Windows implementation)
func findProcessByName(name string) ([]ProcessInfo, error) {
	cmd := exec.Command("powershell", "-Command", 
		fmt.Sprintf("Get-Process | Where-Object {$_.Name -like '*%s*'} | Select-Object Id,StartTime,CommandLine | ConvertTo-Json", name))
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	if len(output) == 0 {
		return []ProcessInfo{}, nil
	}

	// Check if output is an array or a single object
	var processes []ProcessInfo
	
	// Try to parse as array first
	var processArray []struct {
		Id          int       `json:"Id"`
		StartTime   time.Time `json:"StartTime"`
		CommandLine string    `json:"CommandLine"`
	}
	
	err = json.Unmarshal(output, &processArray)
	if err == nil {
		// Successfully parsed as array
		for _, p := range processArray {
			processes = append(processes, ProcessInfo{
				PID:         p.Id,
				CommandLine: p.CommandLine,
				StartTime:   p.StartTime,
			})
		}
		return processes, nil
	}
	
	// Try to parse as single object
	var singleProcess struct {
		Id          int       `json:"Id"`
		StartTime   time.Time `json:"StartTime"`
		CommandLine string    `json:"CommandLine"`
	}
	
	err = json.Unmarshal(output, &singleProcess)
	if err != nil {
		return nil, fmt.Errorf("failed to parse process info: %w", err)
	}
	
	processes = append(processes, ProcessInfo{
		PID:         singleProcess.Id,
		CommandLine: singleProcess.CommandLine,
		StartTime:   singleProcess.StartTime,
	})
	
	return processes, nil
}

// Helper function for min of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func init() {
	// Add commands
	rootCmd.AddCommand(serveCmd, mcpCmd, addCmd, searchCmd, historyCmd, indexProjectCmd, 
		searchProjectCmd, updateProjectCmd, watchProjectCmd, statusCmd, purgeCmd, clearCmd, versionCmd)

	// Add flags
	rootCmd.PersistentFlags().StringVar(&qdrantURL, "qdrant-url", "http://localhost:6333", "URL for Qdrant server")
	rootCmd.PersistentFlags().StringVar(&collectionName, "collection", "memory", "Collection name in Qdrant")
	rootCmd.PersistentFlags().IntVar(&embeddingSize, "embedding-size", 1536, "Size of embeddings")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVar(&projectPath, "project", "", "Project directory path")

	// Project path flag for index-project and update-project commands
	indexProjectCmd.Flags().StringVarP(&projectPath, "path", "p", "", "Path to project directory")
	updateProjectCmd.Flags().StringVarP(&projectPath, "path", "p", "", "Path to project directory")
	watchProjectCmd.Flags().StringVarP(&projectPath, "path", "p", "", "Path to project directory")
	
	// Date range flags for clear command
	clearCmd.Flags().String("from", "", "Start date for range period (format: YYYY-MM-DD)")
	clearCmd.Flags().String("to", "", "End date for range period (format: YYYY-MM-DD)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initClient() *MemoryClient {
	cfg := LoadConfig()

	// Override with command line flags if provided
	if qdrantURL != "" {
		cfg.QdrantURL = qdrantURL
	}
	if collectionName != "" {
		cfg.CollectionName = collectionName
	}
	if embeddingSize > 0 {
		cfg.EmbeddingSize = embeddingSize
	}

	if verbose {
		fmt.Printf("Connecting to Qdrant at %s, collection: %s\n",
			cfg.QdrantURL, cfg.CollectionName)
	}

	client, err := NewMemoryClient(cfg.QdrantURL, cfg.CollectionName, cfg.EmbeddingSize, verbose)
	if err != nil {
		fmt.Printf("Failed to initialize client: %v\n", err)
		os.Exit(1)
	}
	return client
}

func runBackgroundIndexer(ctx context.Context, client *MemoryClient) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := client.IndexMessages(ctx); err != nil {
					fmt.Printf("Background indexing error: %v\n", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
