package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
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
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

// Helper function for min of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func init() {
	// Add global flags
	rootCmd.PersistentFlags().StringVar(&qdrantURL, "url", "", "Qdrant server URL")
	rootCmd.PersistentFlags().StringVar(&collectionName, "collection", "", "Collection name")
	rootCmd.PersistentFlags().IntVar(&embeddingSize, "embedding-size", 0, "Embedding vector size")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVar(&projectPath, "project", "", "Project directory path")

	// Add commands
	rootCmd.AddCommand(serveCmd, mcpCmd, addCmd, searchCmd, historyCmd, indexProjectCmd, searchProjectCmd, updateProjectCmd, watchProjectCmd)
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
