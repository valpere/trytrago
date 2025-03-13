package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/valpere/trytrago/application/service"
	"github.com/valpere/trytrago/domain"
	"github.com/valpere/trytrago/domain/logging"
	"github.com/valpere/trytrago/interface/api/rest"
)

// Server is the main server that handles HTTP traffic
type Server struct {
	cfg          *domain.Config
	logger       logging.Logger
	entryService service.EntryService
	transService service.TranslationService
	userService  service.UserService

	httpServer *http.Server

	shutdownWg sync.WaitGroup
	shutdownCh chan os.Signal
}

// NewServer creates a new server instance
func NewServer(
	cfg *domain.Config,
	logger logging.Logger,
	entryService service.EntryService,
	transService service.TranslationService,
	userService service.UserService,
) *Server {
	return &Server{
		cfg:          cfg,
		logger:       logger.With(logging.String("component", "server")),
		entryService: entryService,
		transService: transService,
		userService:  userService,
		shutdownCh:   make(chan os.Signal, 1),
	}
}

// Start initializes and starts HTTP server
func (s *Server) Start() error {
	// Set up signal handling for graceful shutdown
	signal.Notify(s.shutdownCh, os.Interrupt, syscall.SIGTERM)

	// Set Gin mode based on environment
	if s.cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Start HTTP server in a goroutine
	s.shutdownWg.Add(1)
	go func() {
		defer s.shutdownWg.Done()

		router := rest.SetupRouter(s.logger, s.entryService, s.transService, s.userService)

		s.httpServer = &http.Server{
			Addr:         fmt.Sprintf(":%d", s.cfg.Server.HTTPPort),
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		s.logger.Info("starting HTTP server", logging.Int("port", s.cfg.Server.HTTPPort))

		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("failed to start HTTP server", logging.Error(err))
		}
	}()

	// Wait for shutdown signal
	<-s.shutdownCh
	return s.Shutdown()
}

// Shutdown gracefully stops all servers
func (s *Server) Shutdown() error {
	s.logger.Info("shutting down servers")

	// Shutdown HTTP server
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		s.logger.Info("shutting down HTTP server")
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.Error("HTTP server shutdown error", logging.Error(err))
		}
	}

	// Wait for all goroutines to finish
	s.shutdownWg.Wait()
	s.logger.Info("all servers shutdown complete")

	return nil
}
