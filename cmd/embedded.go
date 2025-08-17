package cmd

import (
	"context"
	"errors"
	"sync"

	"github.com/minio/cli"
	xhttp "github.com/minio/minio/internal/http"
)

type EmbeddedServer struct {
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	args       []string
	running    bool
	started    chan struct{}
	stopped    chan struct{}
	httpServer *xhttp.Server
}

func NewEmbeddedServer(args []string) *EmbeddedServer {
	return &EmbeddedServer{
		args:    args,
		started: make(chan struct{}),
		stopped: make(chan struct{}),
	}
}

func (s *EmbeddedServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.running {
		return errors.New("server is already running")
	}
	
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.running = true
	
	go func() {
		defer func() {
			s.mu.Lock()
			s.running = false
			s.mu.Unlock()
			close(s.stopped)
		}()
		
		close(s.started)
		s.runEmbeddedServer()
	}()
	
	return nil
}

func (s *EmbeddedServer) runEmbeddedServer() {
	// Reset global state for port reuse
	if globalHTTPServer != nil {
		globalHTTPServer.Shutdown()
		globalHTTPServer = nil
	}
	
	// Create proper CLI context with flagset
	app := cli.NewApp()
	app.Name = "minio"
	app.Commands = []cli.Command{serverCmd}
	cliArgs := append([]string{"minio", "server"}, s.args...)
	
	// Run through CLI app to get proper context
	app.Action = func(c *cli.Context) error {
		// Minimal server initialization from serverMain
		loadEnvVarsFromFiles()
		serverHandleEarlyEnvVars()
		
		err := buildServerCtxt(c, &globalServerCtxt)
		if err != nil {
			return err
		}
		
		serverHandleCmdArgs(globalServerCtxt)
		runDNSCache(c)
		serverHandleEnvVars()
		
		// Initialize subsystems
		initAllSubsystems(s.ctx)
		
		// Configure and start HTTP server
		handler, err := configureServerHandler(globalEndpoints)
		if err != nil {
			return err
		}
		
		s.httpServer = xhttp.NewServer(getServerListenAddrs()).
			UseHandler(setCriticalErrorHandler(corsHandler(handler))).
			UseBaseContext(s.ctx)
		
		globalHTTPServer = s.httpServer
		
		// Start server
		serveFn, err := s.httpServer.Init(s.ctx, func(string, error) {})
		if err != nil {
			return err
		}
		
		// Wait for shutdown signal
		go func() {
			<-s.ctx.Done()
			s.httpServer.Shutdown()
		}()
		
		serveFn()
		return nil
	}
	
	app.Run(cliArgs)
}

func (s *EmbeddedServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !s.running {
		return errors.New("server is not running")
	}
	
	// Mark as not running immediately
	s.running = false
	
	// Shutdown HTTP server first
	if s.httpServer != nil {
		s.httpServer.Shutdown()
		s.httpServer = nil
	}
	
	// Reset global HTTP server
	if globalHTTPServer != nil {
		globalHTTPServer.Shutdown()
		globalHTTPServer = nil
	}
	
	// Cancel context
	if s.cancel != nil {
		s.cancel()
	}
	
	return nil
}

func (s *EmbeddedServer) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *EmbeddedServer) WaitForStart() {
	<-s.started
}

func (s *EmbeddedServer) WaitForStop() {
	<-s.stopped
}

func StartEmbeddedServer(args []string) (*EmbeddedServer, error) {
	server := NewEmbeddedServer(args)
	if err := server.Start(); err != nil {
		return nil, err
	}
	return server, nil
}