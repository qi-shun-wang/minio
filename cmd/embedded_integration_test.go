package cmd

import (
	"testing"
	"time"
)

func TestEmbeddedServerRapidStartStop(t *testing.T) {
	// Test rapid start/stop cycles - focus on state management
	for i := 0; i < 5; i++ {
		t.Logf("Starting rapid cycle %d", i)
		
		server := NewEmbeddedServer([]string{
			"--address", ":19030",
			"--console-address", ":19031", 
			"/tmp/minio-test-rapid",
		})
		
		// Test initial state
		if server.IsRunning() {
			t.Fatalf("Server should not be running initially on cycle %d", i)
		}
		
		// Start server
		if err := server.Start(); err != nil {
			t.Fatalf("Failed to start server on rapid cycle %d: %v", i, err)
		}
		
		// Wait for start
		server.WaitForStart()
		
		// Verify running state
		if !server.IsRunning() {
			t.Fatalf("Server should be running on cycle %d", i)
		}
		
		// Stop server
		if err := server.Stop(); err != nil {
			t.Fatalf("Failed to stop server on rapid cycle %d: %v", i, err)
		}
		
		// For rapid testing, we only verify the state change
		// The actual server shutdown may take longer due to background processes
		time.Sleep(10 * time.Millisecond)
		
		// Verify the server is marked as not running
		if server.IsRunning() {
			t.Fatalf("Server should be marked as stopped on cycle %d", i)
		}
		
		t.Logf("Completed rapid cycle %d", i)
	}
}