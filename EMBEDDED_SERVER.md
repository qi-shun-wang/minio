# MinIO Embedded Server

This document describes how to use the MinIO embedded server functionality that allows you to start and stop MinIO programmatically from within your Go applications.

## Overview

The embedded server provides a way to:
- Start MinIO server programmatically
- Stop the server gracefully
- Monitor server status
- Handle server lifecycle from external programs

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/minio/minio/cmd"
)

func main() {
    // Create embedded server with storage path
    server := cmd.NewEmbeddedServer([]string{"/tmp/minio-data"})

    // Start the server
    if err := server.Start(); err != nil {
        log.Fatal("Failed to start server:", err)
    }

    // Wait for server to start
    server.WaitForStart()
    fmt.Println("MinIO server started!")

    // Check if server is running
    if server.IsRunning() {
        fmt.Println("Server is running")
    }

    // Let it run for some time
    time.Sleep(30 * time.Second)

    // Stop the server
    if err := server.Stop(); err != nil {
        log.Fatal("Failed to stop server:", err)
    }

    // Wait for server to stop
    server.WaitForStop()
    fmt.Println("Server stopped!")
}
```

### Advanced Configuration

```go
// Create server with custom configuration
server := cmd.NewEmbeddedServer([]string{
    "--address", ":9010",
    "--console-address", ":9011", 
    "/path/to/data",
})
```

### Convenience Function

```go
// Start server with one function call
server, err := cmd.StartEmbeddedServer([]string{"/tmp/minio-data"})
if err != nil {
    log.Fatal(err)
}
defer server.Stop()
```

## API Reference

### Types

#### EmbeddedServer

The main struct that represents an embedded MinIO server instance.

### Functions

#### NewEmbeddedServer(args []string) *EmbeddedServer

Creates a new embedded server instance with the given arguments.

**Parameters:**
- `args`: Command line arguments for the MinIO server (same as you would pass to `minio server`)

**Returns:**
- `*EmbeddedServer`: A new server instance

#### StartEmbeddedServer(args []string) (*EmbeddedServer, error)

Convenience function that creates and starts a server in one call.

**Parameters:**
- `args`: Command line arguments for the MinIO server

**Returns:**
- `*EmbeddedServer`: The started server instance
- `error`: Any error that occurred during startup

### Methods

#### Start() error

Starts the MinIO server in a background goroutine.

**Returns:**
- `error`: Error if the server is already running or fails to start

#### Stop() error

Stops the running MinIO server.

**Returns:**
- `error`: Error if the server is not running

#### IsRunning() bool

Returns whether the server is currently running.

**Returns:**
- `bool`: True if the server is running, false otherwise

#### WaitForStart()

Blocks until the server has started.

#### WaitForStop()

Blocks until the server has stopped.

#### Wait() error

Blocks until the server starts and then stops.

**Returns:**
- `error`: Any error that occurred

## Examples

### Signal Handling

```go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/minio/minio/cmd"
)

func main() {
    server := cmd.NewEmbeddedServer([]string{"/tmp/minio-data"})
    
    if err := server.Start(); err != nil {
        panic(err)
    }
    
    server.WaitForStart()
    fmt.Println("Server started. Press Ctrl+C to stop.")
    
    // Handle shutdown signals
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        fmt.Println("Shutting down...")
        server.Stop()
    }()
    
    server.WaitForStop()
    fmt.Println("Server stopped.")
}
```

### Multiple Servers

```go
package main

import (
    "fmt"
    "time"
    
    "github.com/minio/minio/cmd"
)

func main() {
    // Start multiple servers on different ports
    server1 := cmd.NewEmbeddedServer([]string{
        "--address", ":9000",
        "/tmp/minio-data-1",
    })
    
    server2 := cmd.NewEmbeddedServer([]string{
        "--address", ":9001", 
        "/tmp/minio-data-2",
    })
    
    // Start both servers
    server1.Start()
    server2.Start()
    
    // Wait for both to start
    server1.WaitForStart()
    server2.WaitForStart()
    
    fmt.Println("Both servers started!")
    
    // Run for some time
    time.Sleep(30 * time.Second)
    
    // Stop both servers
    server1.Stop()
    server2.Stop()
    
    // Wait for both to stop
    server1.WaitForStop()
    server2.WaitForStop()
    
    fmt.Println("Both servers stopped!")
}
```

## Notes

- The embedded server uses the same configuration and arguments as the regular MinIO server
- Default ports are 9000 for API and 9001 for console
- Make sure the data directory exists and is writable
- The server runs with default credentials `minioadmin:minioadmin` unless configured otherwise
- For production use, always configure proper credentials using environment variables

## Troubleshooting

### Server Won't Start

- Check if the data directory exists and is writable
- Ensure the ports are not already in use
- Verify that all required arguments are provided

### Server Won't Stop

- The `Stop()` method sends a shutdown signal but may not immediately terminate the server
- Use `WaitForStop()` to wait for graceful shutdown
- In some cases, you may need to handle cleanup manually

### Port Conflicts

- Use custom `--address` and `--console-address` arguments to avoid port conflicts
- Check that no other MinIO instances are running on the same ports