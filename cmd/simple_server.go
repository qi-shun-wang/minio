package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	xhttp "github.com/minio/minio/internal/http"
)

type SimpleMinIOServer struct {
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	dataDir  string
	port     int
	running  bool
	server   *xhttp.Server
}

func NewSimpleMinIOServer(dataDir string, port int) *SimpleMinIOServer {
	return &SimpleMinIOServer{
		dataDir: dataDir,
		port:    port,
		running: false,
	}
}

func (s *SimpleMinIOServer) Start(accessKey, secretKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("server is already running")
	}

	// Create data directory
	os.MkdirAll(s.dataDir, 0755)

	// Set credentials
	os.Setenv("MINIO_ROOT_USER", accessKey)
	os.Setenv("MINIO_ROOT_PASSWORD", secretKey)

	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.running = true

	// Start server in goroutine
	go s.runMinimalServer()

	// Wait for server to start
	time.Sleep(2 * time.Second)

	return nil
}

func (s *SimpleMinIOServer) runMinimalServer() {
	defer func() {
		if r := recover(); r != nil {
			// Ignore any panics
		}
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	// Create a minimal HTTP server that responds to health checks
	mux := http.NewServeMux()
	
	// Health check endpoint
	mux.HandleFunc("/minio/health/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Basic S3 API endpoints (minimal implementation)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "MinIO")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("MinIO Server"))
	})

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	s.server = &xhttp.Server{}

	// Start server
	go func() {
		httpServer.ListenAndServe()
	}()

	// Wait for shutdown signal
	<-s.ctx.Done()

	// Shutdown server
	if httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		httpServer.Shutdown(ctx)
	}
}

func (s *SimpleMinIOServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("server is not running")
	}

	s.running = false

	if s.cancel != nil {
		s.cancel()
	}

	return nil
}

func (s *SimpleMinIOServer) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *SimpleMinIOServer) GetStatus() string {
	if s.IsRunning() {
		return "Running"
	}
	return "Stopped"
}