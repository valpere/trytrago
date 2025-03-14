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
    "github.com/valpere/trytrago/interface/api/rest/handler"
    "github.com/valpere/trytrago/interface/api/rest/middleware"
)

// Server is the interface for all server implementations
type Server interface {
    Start() error
    Shutdown(ctx context.Context) error
}

// AppServer is the main server that handles HTTP traffic
type AppServer struct {
    cfg          domain.Config
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
    cfg domain.Config,
    logger logging.Logger,
    entryService service.EntryService,
    transService service.TranslationService,
    userService service.UserService,
) *AppServer {
    return &AppServer{
        cfg:          cfg,
        logger:       logger.With(logging.String("component", "server")),
        entryService: entryService,
        transService: transService,
        userService:  userService,
        shutdownCh:   make(chan os.Signal, 1),
    }
}

// Start initializes and starts HTTP server
func (s *AppServer) Start() error {
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

        // Create router with handlers
        entryHandler := handler.NewEntryHandler(s.entryService, s.logger)
        transHandler := handler.NewTranslationHandler(s.transService, s.logger)
        userHandler := handler.NewUserHandler(s.userService, s.logger)
        authMiddleware := middleware.NewAuthMiddleware(s.logger)

        // Create router
        router := rest.NewRouter(
            s.cfg,
            s.logger,
            entryHandler,
            transHandler,
            userHandler,
            authMiddleware,
        )

        s.httpServer = &http.Server{
            Addr:         fmt.Sprintf(":%d", s.cfg.Server.Port),
            Handler:      router.Handler(),
            ReadTimeout:  s.cfg.Server.ReadTimeout,
            WriteTimeout: s.cfg.Server.WriteTimeout,
            IdleTimeout:  60 * time.Second,
        }

        s.logger.Info("Starting HTTP server", logging.Int("port", s.cfg.Server.Port))

        if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            s.logger.Error("Failed to start HTTP server", logging.Error(err))
        }
    }()

    // Wait for shutdown signal
    <-s.shutdownCh
    return s.Shutdown(context.Background())
}

// Shutdown gracefully stops all servers
func (s *AppServer) Shutdown(ctx context.Context) error {
    s.logger.Info("Shutting down servers")

    // Shutdown HTTP server
    if s.httpServer != nil {
        shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
        defer cancel()

        s.logger.Info("Shutting down HTTP server")
        if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
            s.logger.Error("HTTP server shutdown error", logging.Error(err))
        }
    }

    // Wait for all goroutines to finish
    s.shutdownWg.Wait()
    s.logger.Info("All servers shutdown complete")

    return nil
}
