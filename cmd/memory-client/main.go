package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/christerso/memory-client-go/internal/client"
	"github.com/christerso/memory-client-go/internal/config"
	"github.com/christerso/memory-client-go/internal/dashboard"
	"github.com/christerso/memory-client-go/internal/mcp"
	"github.com/christerso/memory-client-go/internal/models"

	"github.com/qdrant/go-client/qdrant"
)

var rootCmd = &cobra.Command{
	Use:   "memory-client",
	Short: "MCP Memory Client for persistent conversation storage",
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a message to memory",
	Run: func(cmd *cobra.Command, args []string) {
		memClient := initClient()

		role, _ := cmd.Flags().GetString("role")
		content, _ := cmd.Flags().GetString("content")

		if content == "" {
			fmt.Println("Error: content is required")
			os.Exit(1)
		}

		ctx := context.Background()
		message := &models.Message{
			Role:      models.Role(role),
			Content:   content,
			Timestamp: time.Now(),
		}

		err := memClient.AddMessage(ctx, message)
		if err != nil {
			fmt.Printf("Error adding message: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Message added successfully")
	},
}

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search conversation memory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		memClient := initClient()

		query := args[0]
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := context.Background()
		results, err := memClient.SearchMessages(ctx, query, limit)
		if err != nil {
			fmt.Printf("Error searching messages: %v\n", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			fmt.Println("No results found")
			return
		}

		fmt.Printf("Found %d results:\n\n", len(results))
		for i, msg := range results {
			fmt.Printf("%d. [%s] %s: %s\n", i+1, msg.Timestamp.Format(time.RFC3339), msg.Role, msg.Content)
		}
	},
}

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear messages from memory",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		memClient := initClient()
		defer memClient.Close()

		timeRange := cmd.Flag("time-range").Value.String()
		switch timeRange {
		case "day":
			count, err := memClient.DeleteMessagesForCurrentDay(ctx)
			if err != nil {
				fmt.Printf("Error clearing messages: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Cleared %d messages from today\n", count)
		case "week":
			count, err := memClient.DeleteMessagesForCurrentWeek(ctx)
			if err != nil {
				fmt.Printf("Error clearing messages: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Cleared %d messages from this week\n", count)
		case "month":
			count, err := memClient.DeleteMessagesForCurrentMonth(ctx)
			if err != nil {
				fmt.Printf("Error clearing messages: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Cleared %d messages from this month\n", count)
		case "range":
			if cmd.Flag("from").Changed && cmd.Flag("to").Changed {
				from, err := time.Parse(time.RFC3339, cmd.Flag("from").Value.String())
				if err != nil {
					fmt.Printf("Error parsing from date: %v\n", err)
					os.Exit(1)
				}

				to, err := time.Parse(time.RFC3339, cmd.Flag("to").Value.String())
				if err != nil {
					fmt.Printf("Error parsing to date: %v\n", err)
					os.Exit(1)
				}

				count, err := memClient.DeleteMessagesByTimeRange(ctx, from, to)
				if err != nil {
					fmt.Printf("Error clearing messages: %v\n", err)
					os.Exit(1)
				}
				fmt.Printf("Cleared %d messages from %s to %s\n", count, from.Format(time.RFC3339), to.Format(time.RFC3339))
			} else {
				fmt.Println("Error: from and to dates are required for range period")
				os.Exit(1)
			}
		default:
			fmt.Println("Error: invalid period. Use day, week, month, or range")
			os.Exit(1)
		}
	},
}

var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Completely purge all data from Qdrant",
	Run: func(cmd *cobra.Command, args []string) {
		memClient := initClient()

		ctx := context.Background()
		err := memClient.ClearAllMemories(ctx)
		if err != nil {
			fmt.Printf("Error purging data: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("All data purged successfully")
	},
}

var indexProjectCmd = &cobra.Command{
	Use:   "index-project [path]",
	Short: "Index project files in a directory",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		memClient := initClient()

		projectPath := "."
		if len(args) > 0 {
			projectPath = args[0]
		}

		tag, _ := cmd.Flags().GetString("tag")

		absPath, err := filepath.Abs(projectPath)
		if err != nil {
			fmt.Printf("Error getting absolute path: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Indexing project files in: %s\n", absPath)
		if tag != "" {
			fmt.Printf("Using tag: %s\n", tag)
		}

		ctx := context.Background()
		count, err := memClient.IndexProjectFiles(ctx, absPath, tag)
		if err != nil {
			fmt.Printf("Error indexing project files: %v\n", err)
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
		ctx := context.Background()
		memClient := initClient()
		defer memClient.Close()

		projectPath := "."
		if len(args) > 0 {
			projectPath = args[0]
		}

		added, updated, err := memClient.UpdateProjectFiles(ctx, projectPath)
		if err != nil {
			fmt.Printf("Error updating project files: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Added %d new files, updated %d existing files\n", added, updated)
	},
}

var watchProjectCmd = &cobra.Command{
	Use:   "watch-project [path]",
	Short: "Watch a project directory for changes",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		memClient := initClient()
		defer memClient.Close()

		projectPath := "."
		if len(args) > 0 {
			projectPath = args[0]
		}

		// Since WatchProjectFiles is not implemented, we'll use a simple polling approach
		fmt.Printf("Watching project directory: %s\n", projectPath)
		fmt.Println("Press Ctrl+C to stop")

		// Set up signal handling for graceful shutdown
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

		// Start a goroutine to handle signals
		go func() {
			<-sigCh
			fmt.Println("\nStopping project watcher...")
			cancel()
		}()

		// Poll for changes every 5 seconds
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				added, updated, err := memClient.UpdateProjectFiles(ctx, projectPath)
				if err != nil {
					fmt.Printf("Error updating project files: %v\n", err)
					continue
				}

				if added > 0 || updated > 0 {
					fmt.Printf("[%s] Added %d new files, updated %d existing files\n",
						time.Now().Format(time.RFC3339), added, updated)
				}
			}
		}
	},
}

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Start the web dashboard for monitoring memory usage",
	Run: func(cmd *cobra.Command, args []string) {
		memClient := initClient()

		port, _ := cmd.Flags().GetInt("port")

		fmt.Printf("Starting memory dashboard on http://localhost:%d\n", port)
		fmt.Println("Press Ctrl+C to stop")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Handle Ctrl+C
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			fmt.Println("\nStopping dashboard server...")
			cancel()
			os.Exit(0)
		}()

		dashboardServer := dashboard.NewDashboardServer(memClient, port)
		err := dashboardServer.Start(ctx)
		if err != nil {
			fmt.Printf("Error starting dashboard server: %v\n", err)
			os.Exit(1)
		}
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check if the MCP server is running",
	Run: func(cmd *cobra.Command, args []string) {
		// First try to connect to the MCP server directly
		mcpStatusURL := "http://127.0.0.1:9580/status"
		mcpAPIURL := "http://127.0.0.1:10010/api/get-tagging-mode"
		dashboardURL := "http://127.0.0.1:9581/"

		client := http.Client{
			Timeout: 2 * time.Second,
		}

		// Check MCP HTTP server
		mcpHTTPRunning := false
		resp, err := client.Get(mcpStatusURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			mcpHTTPRunning = true
			fmt.Println("✅ MCP HTTP server is running at http://127.0.0.1:9580/status")
		} else {
			fmt.Println("❌ MCP HTTP server is not running")
		}

		// Check MCP API server
		mcpAPIRunning := false
		resp, err = client.Get(mcpAPIURL)
		if err == nil && (resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusMethodNotAllowed) {
			mcpAPIRunning = true
			fmt.Println("✅ MCP API server is running at http://127.0.0.1:10010")
		} else {
			fmt.Println("❌ MCP API server is not running")
		}

		// Check Dashboard server
		dashboardRunning := false
		resp, err = client.Get(dashboardURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			dashboardRunning = true
			fmt.Println("✅ Dashboard is running at http://127.0.0.1:9581")
		} else {
			fmt.Println("❌ Dashboard is not running")
		}

		// If direct connection fails, check for the process
		if !mcpHTTPRunning && !mcpAPIRunning {
			processes, err := findProcessByName("memory-client")
			if err != nil {
				fmt.Printf("Error checking MCP server status: %v\n", err)
				os.Exit(1)
			}

			for _, proc := range processes {
				// Check if it's running with the mcp command
				if strings.Contains(strings.ToLower(proc.CommandLine), "memory-client") && strings.Contains(strings.ToLower(proc.CommandLine), "mcp") {
					fmt.Println("\nFound MCP server process:")
					fmt.Printf("  PID: %d\n", proc.PID)
					fmt.Printf("  Command: %s\n", proc.CommandLine)
					fmt.Printf("  Started: %s\n", proc.StartTime.Format(time.RFC3339))
					fmt.Printf("  Uptime: %s\n", time.Since(proc.StartTime).Round(time.Second))
				}
			}
		}

		// Print instructions if services are not running
		if !mcpHTTPRunning || !mcpAPIRunning || !dashboardRunning {
			fmt.Println("\nTo start the missing services:")

			if !mcpHTTPRunning || !mcpAPIRunning {
				fmt.Println("  1. Start MCP server:    memory-client mcp")
			}

			if !dashboardRunning {
				fmt.Println("  2. Start dashboard:     memory-client dashboard")
			}
		}
	},
}

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the MCP server for handling memory operations",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		memClient := initClient()
		defer memClient.Close()

		// Load config to get Qdrant URL
		cfg := config.LoadConfig()

		// Create Qdrant client instance
		qdrantConfig := &qdrant.Config{
			Host: cfg.QdrantURL,
		}
		qdrantClient, err := qdrant.NewClient(qdrantConfig)
		if err != nil {
			fmt.Printf("Error creating Qdrant client: %v\n", err)
			os.Exit(1)
		}

		// Create MCP server with the Qdrant client directly
		server := mcp.NewMCPServer(memClient, qdrantClient)
		if err := server.Start(ctx); err != nil {
			fmt.Printf("MCP server error: %v\n", err)
			os.Exit(1)
		}
	},
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run tests against the vector database",
	Run: func(cmd *cobra.Command, args []string) {
		memClient := initClient()
		ctx := context.Background()

		testType, _ := cmd.Flags().GetString("type")
		count, _ := cmd.Flags().GetInt("count")

		switch testType {
		case "add":
			fmt.Println("Running add message test...")
			runAddMessageTest(ctx, memClient, count)
		case "search":
			fmt.Println("Running search test...")
			runSearchTest(ctx, memClient)
		case "history":
			fmt.Println("Running history test...")
			runHistoryTest(ctx, memClient, count)
		case "all":
			fmt.Println("Running all tests...")
			runAddMessageTest(ctx, memClient, count)
			runSearchTest(ctx, memClient)
			runHistoryTest(ctx, memClient, count)
		default:
			fmt.Println("Unknown test type. Available types: add, search, history, all")
		}
	},
}

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Display conversation history",
	Long:  `Display the conversation history from the memory client database.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		memClient := initClient()
		defer memClient.Close()

		// Get limit flag
		limit, _ := cmd.Flags().GetInt("limit")
		if limit <= 0 {
			limit = 20 // Default limit
		}

		// Get role filter flag
		roleFilter, _ := cmd.Flags().GetString("role")

		fmt.Printf("Retrieving last %d messages", limit)
		if roleFilter != "" {
			fmt.Printf(" with role '%s'", roleFilter)
		}
		fmt.Println()

		// Get conversation history
		var filter *models.HistoryFilter
		if roleFilter != "" {
			filter = &models.HistoryFilter{
				Role: models.Role(roleFilter),
			}
		}

		messages, err := memClient.GetConversationHistory(ctx, limit, filter)
		if err != nil {
			fmt.Printf("Error retrieving conversation history: %v\n", err)
			os.Exit(1)
		}

		if len(messages) == 0 {
			fmt.Println("No messages found in conversation history.")
			return
		}

		// Sort messages by timestamp (newest first)
		sort.Slice(messages, func(i, j int) bool {
			return messages[i].Timestamp.After(messages[j].Timestamp)
		})

		// Print messages
		fmt.Printf("Found %d messages:\n\n", len(messages))
		for i, msg := range messages {
			// Format timestamp
			timestamp := msg.Timestamp.Format(time.RFC3339)

			// Format role with color if supported
			roleDisplay := string(msg.Role)

			// Print message header
			fmt.Printf("[%d] %s | %s\n", i+1, timestamp, roleDisplay)

			// Print message content with indentation
			contentLines := strings.Split(msg.Content, "\n")
			for _, line := range contentLines {
				fmt.Printf("    %s\n", line)
			}

			// Add separator between messages
			fmt.Println("----------------------------------------")
		}
	},
}

func runAddMessageTest(ctx context.Context, memClient *client.MemoryClient, count int) {
	fmt.Printf("Adding %d test messages to the database...\n", count)

	roles := []models.Role{models.RoleUser, models.RoleAssistant}
	topics := []string{"programming", "database", "vector search", "golang", "memory client"}

	successCount := 0
	startTime := time.Now()

	for i := 0; i < count; i++ {
		role := roles[rand.Intn(len(roles))]
		topic := topics[rand.Intn(len(topics))]

		message := &models.Message{
			ID:        uuid.New().String(),
			Role:      role,
			Content:   fmt.Sprintf("This is a test message about %s (test ID: %d)", topic, i+1),
			Timestamp: time.Now(),
			Tags:      []string{"test", topic},
		}

		err := memClient.AddMessage(ctx, message)
		if err != nil {
			fmt.Printf("Error adding message %d: %v\n", i+1, err)
		} else {
			successCount++
		}
	}

	duration := time.Since(startTime)
	fmt.Printf("Test completed: Successfully added %d/%d messages in %v\n", successCount, count, duration)
	fmt.Printf("Average time per message: %v\n", duration/time.Duration(count))
}

func runSearchTest(ctx context.Context, memClient *client.MemoryClient) {
	queries := []string{
		"golang programming",
		"vector database search",
		"memory client functionality",
		"test message",
	}

	fmt.Println("Running search tests with different queries...")

	for i, query := range queries {
		fmt.Printf("\nTest %d: Searching for '%s'\n", i+1, query)
		startTime := time.Now()

		results, err := memClient.SearchMessages(ctx, query, 5)
		if err != nil {
			fmt.Printf("Error searching for '%s': %v\n", query, err)
			continue
		}

		duration := time.Since(startTime)
		fmt.Printf("Found %d results in %v\n", len(results), duration)

		for j, result := range results {
			fmt.Printf("  %d. [%s] %s: %s\n",
				j+1,
				result.Timestamp.Format(time.RFC3339),
				result.Role,
				result.Content,
			)
		}
	}
}

func runHistoryTest(ctx context.Context, memClient *client.MemoryClient, limit int) {
	fmt.Printf("Testing conversation history retrieval (limit: %d)...\n", limit)
	startTime := time.Now()

	// Test with no filter
	messages, err := memClient.GetConversationHistory(ctx, limit, nil)
	if err != nil {
		fmt.Printf("Error retrieving conversation history: %v\n", err)
		return
	}

	duration := time.Since(startTime)
	fmt.Printf("Retrieved %d messages in %v\n", len(messages), duration)

	// Test with tag filter
	fmt.Println("\nTesting history with tag filter...")
	filter := &models.HistoryFilter{
		Tags: []string{"test"},
	}

	startTime = time.Now()
	taggedMessages, err := memClient.GetConversationHistory(ctx, limit, filter)
	if err != nil {
		fmt.Printf("Error retrieving filtered conversation history: %v\n", err)
		return
	}

	duration = time.Since(startTime)
	fmt.Printf("Retrieved %d messages with tag 'test' in %v\n", len(taggedMessages), duration)
}

// ProcessInfo contains information about a running process
type ProcessInfo struct {
	PID         int
	CommandLine string
	StartTime   time.Time
}

// findProcessByName finds processes by name (Windows implementation)
func findProcessByName(name string) ([]ProcessInfo, error) {
	cmd := exec.Command("tasklist", "/fo", "csv", "/nh")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute tasklist: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var processes []ProcessInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse CSV format
		fields := strings.Split(line, ",")
		if len(fields) < 2 {
			continue
		}

		// Remove quotes
		processName := strings.Trim(fields[0], "\"")
		pidStr := strings.Trim(fields[1], "\"")

		if !strings.Contains(strings.ToLower(processName), strings.ToLower(name)) {
			continue
		}

		var pid int
		_, err := fmt.Sscanf(pidStr, "%d", &pid)
		if err != nil {
			continue
		}

		// Get more details using wmic
		cmdDetails := exec.Command("wmic", "process", "where", fmt.Sprintf("ProcessId=%d", pid), "get", "CommandLine,CreationDate", "/format:csv")
		detailsOutput, err := cmdDetails.Output()
		if err != nil {
			continue
		}

		detailsLines := strings.Split(string(detailsOutput), "\n")
		if len(detailsLines) < 2 {
			continue
		}

		// Find the line with the process details
		var commandLine string
		var startTime time.Time

		for _, detailLine := range detailsLines {
			if strings.Contains(detailLine, fmt.Sprintf(",%d,", pid)) {
				detailFields := strings.Split(detailLine, ",")
				if len(detailFields) >= 3 {
					commandLine = detailFields[1]
					// Parse creation date (format: yyyymmddHHMMSS.mmmmmm+zzz)
					if len(detailFields) >= 4 && len(detailFields[3]) > 14 {
						dateStr := detailFields[3][:14]
						startTime, _ = time.Parse("20060102150405", dateStr)
					}
				}
				break
			}
		}

		processes = append(processes, ProcessInfo{
			PID:         pid,
			CommandLine: commandLine,
			StartTime:   startTime,
		})
	}

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
	// Add command flags
	addCmd.Flags().StringP("role", "r", "user", "Message role (user or assistant)")
	addCmd.Flags().StringP("content", "c", "", "Message content")

	searchCmd.Flags().IntP("limit", "l", 10, "Maximum number of results to return")

	clearCmd.Flags().StringP("time-range", "t", "", "Time range to clear (day, week, month, or range)")
	clearCmd.Flags().StringP("from", "f", "", "Start date (YYYY-MM-DDTHH:MM:SSZ) for range period")
	clearCmd.Flags().StringP("to", "e", "", "End date (YYYY-MM-DDTHH:MM:SSZ) for range period")

	indexProjectCmd.Flags().StringP("tag", "t", "", "Tag to associate with indexed files")
	updateProjectCmd.Flags().StringP("tag", "t", "", "Tag to associate with updated files")
	watchProjectCmd.Flags().StringP("tag", "t", "", "Tag to associate with watched files")

	dashboardCmd.Flags().IntP("port", "p", 9581, "Port to run the dashboard server on")

	mcpCmd.Flags().IntP("port", "p", 9580, "Port to run the MCP server on")

	testCmd.Flags().StringP("type", "t", "all", "Test type (add, search, history, all)")
	testCmd.Flags().IntP("count", "c", 10, "Number of test messages to add")

	historyCmd.Flags().IntP("limit", "l", 20, "Maximum number of messages to retrieve")
	historyCmd.Flags().StringP("role", "r", "", "Filter messages by role (user, assistant, system)")

	// Add commands to root command
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(clearCmd)
	rootCmd.AddCommand(purgeCmd)
	rootCmd.AddCommand(indexProjectCmd)
	rootCmd.AddCommand(updateProjectCmd)
	rootCmd.AddCommand(watchProjectCmd)
	rootCmd.AddCommand(dashboardCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(mcpCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(historyCmd)
}

// Execute executes the root command
func Execute() error {
	return rootCmd.Execute()
}

func main() {
	if err := Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initClient() *client.MemoryClient {
	cfg := config.LoadConfig()

	qdrantURL := cfg.QdrantURL
	collectionName := cfg.CollectionName
	embeddingSize := cfg.EmbeddingSize

	memClient, err := client.NewMemoryClient(qdrantURL, collectionName, embeddingSize, false)
	if err != nil {
		fmt.Printf("Error initializing memory client: %v\n", err)
		os.Exit(1)
	}

	// The EnsureCollection method is not exported, so we can't call it directly
	// We'll assume the collection is already created

	return memClient
}

func runBackgroundIndexer(ctx context.Context, memClient *client.MemoryClient) {
	fmt.Println("Background indexer started, but no project path configured")
	return
}
