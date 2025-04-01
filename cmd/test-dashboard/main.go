package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	
	"github.com/christerso/memory-client-go/internal/dashboard"
)

func main() {
	// Create a context that will be canceled on Ctrl+C
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Create and start the dashboard server with nil client to use sample data
	port := 9095 // Using a different port to avoid conflicts
	fmt.Printf("Starting memory dashboard on http://localhost:%d with sample data\n", port)
	fmt.Println("Press Ctrl+C to stop")

	server := dashboard.NewDashboardServer(nil, port)
	go func() {
		if err := server.Start(ctx); err != nil {
			log.Fatalf("Error starting dashboard server: %v", err)
		}
	}()

	// Wait for termination signal
	<-sigCh
	fmt.Println("\nShutting down...")
}
