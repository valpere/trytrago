package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/valpere/trytrago/application/service"
	"github.com/valpere/trytrago/domain/logging"
	"github.com/valpere/trytrago/interface/api/rest"
)

// HTTPServer represents an HTTP server
type HTTPServer struct {
	addr      string
	router    *gin.Engine
	server    *http.Server
	logger    logging.Logger
}

// NewHTTPServer creates a new HTTP server instance
func NewHTTPServer(
	port int,
	logger logging.Logger,
	entryService service.EntryService,
	translationService service.TranslationService,
	userService service.UserService,
) *HTTPServer {
	// Create router
	router := rest.SetupRouter(logger, entryService, translationService, userService)

	// Configure HTTP server
	addr := fmt.Sprintf(":%d", port)
	server := &http.Server{
		Addr:           addr,
		Handler:        router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	return &HTTPServer{
		addr:      addr,
		router:    router,
		server:    server,
		logger:    logger.With(logging.String("component", "http_server")),
	}
}

// Start starts the HTTP server
func (s *HTTPServer) Start() error {
	s.logger.Info("starting HTTP server", logging.String("address", s.addr))

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Error("failed to start HTTP server", logging.Error(err))
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

// StartTLS starts the HTTP server with TLS
func (s *HTTPServer) StartTLS(certFile, keyFile string) error {
	s.logger.Info("starting HTTPS server", logging.String("address", s.addr))

	if err := s.server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
		s.logger.Error("failed to start HTTPS server", logging.Error(err))
		return fmt.Errorf("failed to start HTTPS server: %w", err)
	}

	return nil
}

// Stop gracefully stops the HTTP server
func (s *HTTPServer) Stop(ctx context.Context) error {
	s.logger.Info("stopping HTTP server")

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("failed to shut down HTTP server", logging.Error(err))
		return fmt.Errorf("failed to shut down HTTP server: %w", err)
	}

	s.logger.Info("HTTP server stopped")
	return nil
}

// GetRouter returns the router
func (s *HTTPServer) GetRouter() *gin.Engine {
	return s.router
}
