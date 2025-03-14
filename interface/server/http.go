package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/valpere/trytrago/domain"
	"github.com/valpere/trytrago/domain/logging"
)

// HTTPServer represents the HTTP server implementation
type HTTPServer struct {
	server *http.Server
	logger logging.Logger
	router *gin.Engine
	config domain.Config
}

// NewHTTPServer creates a new HTTP server instance
func NewHTTPServer(config domain.Config, logger logging.Logger, router *gin.Engine) Server {
	return &HTTPServer{
		logger: logger.With(logging.String("component", "http_server")),
		router: router,
		config: config,
	}
}

// Start starts the HTTP server
func (s *HTTPServer) Start() error {
	serverConfig := s.config.Server
	
	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", serverConfig.Port),
		Handler:      s.router,
		ReadTimeout:  serverConfig.ReadTimeout,
		WriteTimeout: serverConfig.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	s.logger.Info("Starting HTTP server", logging.Int("port", serverConfig.Port))
	
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server error: %w", err)
	}
	
	return nil
}

// Shutdown gracefully stops the HTTP server
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server")
	
	if s.server == nil {
		return nil
	}
	
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	
	if err := s.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("HTTP server shutdown error: %w", err)
	}
	
	s.logger.Info("HTTP server stopped")
	return nil
}
