package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	
	"github.com/user/memory-client-go/internal/client"
	"github.com/user/memory-client-go/internal/config"
	"github.com/user/memory-client-go/internal/dashboard"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	
	// Create memory client
	memClient, err := client.NewMemoryClient(cfg.QdrantURL, cfg.CollectionName, cfg.EmbeddingSize, true)
	if err != nil {
		fmt.Printf("Error creating memory client: %v\n", err)
		os.Exit(1)
	}
	
	// Start dashboard server
	port := 9094 // Using a less common port to avoid conflicts
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
	err = dashboardServer.Start(ctx)
	if err != nil {
		fmt.Printf("Error starting dashboard server: %v\n", err)
		os.Exit(1)
	}
}
