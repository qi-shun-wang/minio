package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/minio/minio/cmd"
)

func main() {
	// Example 1: Basic usage
	basicExample()
	
	// Example 2: With signal handling
	signalExample()
	
	// Example 3: Using convenience function
	convenienceExample()
}

func basicExample() {
	fmt.Println("=== Basic Example ===")
	
	// Create embedded server with storage path
	server := cmd.NewEmbeddedServer([]string{"/tmp/minio-data"})

	// Start the server
	fmt.Println("Starting MinIO server...")
	if err := server.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}

	// Wait for server to start
	server.WaitForStart()
	fmt.Println("MinIO server started successfully!")
	fmt.Println("Server accessible at: http://localhost:9000")
	fmt.Println("Console accessible at: http://localhost:9001")

	// Check if server is running
	if server.IsRunning() {
		fmt.Println("Server status: RUNNING")
	}

	// Let it run for 10 seconds
	fmt.Println("Server will run for 10 seconds...")
	time.Sleep(10 * time.Second)

	// Stop the server
	fmt.Println("Stopping MinIO server...")
	if err := server.Stop(); err != nil {
		log.Fatal("Failed to stop server:", err)
	}

	// Wait for server to stop
	server.WaitForStop()
	fmt.Println("MinIO server stopped successfully!")
	fmt.Println()
}

func signalExample() {
	fmt.Println("=== Signal Handling Example ===")
	fmt.Println("Press Ctrl+C to stop the server")
	
	// Create server with custom configuration
	server := cmd.NewEmbeddedServer([]string{
		"--address", ":9010",
		"--console-address", ":9011",
		"/tmp/minio-data-2",
	})

	// Start the server
	if err := server.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}

	// Wait for server to start
	server.WaitForStart()
	fmt.Println("MinIO server started on port 9010!")

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for signal or server to stop
	go func() {
		<-sigChan
		fmt.Println("\nReceived shutdown signal...")
		if err := server.Stop(); err != nil {
			log.Printf("Error stopping server: %v", err)
		}
	}()

	// Wait for server to stop
	server.WaitForStop()
	fmt.Println("Server stopped gracefully!")
	fmt.Println()
}

func convenienceExample() {
	fmt.Println("=== Convenience Function Example ===")
	
	// Use convenience function
	server, err := cmd.StartEmbeddedServer([]string{"/tmp/minio-data-3"})
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}

	server.WaitForStart()
	fmt.Println("Server started using convenience function!")

	// Run for 5 seconds
	time.Sleep(5 * time.Second)

	// Stop
	server.Stop()
	server.WaitForStop()
	fmt.Println("Convenience example completed!")
}