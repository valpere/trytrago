package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	"github.com/valpere/trytrago/application/service"
	"github.com/valpere/trytrago/domain/logging"
	"github.com/valpere/trytrago/interface/api/grpc/proto"
	grpcService "github.com/valpere/trytrago/interface/api/grpc/service"
)

// GRPCServer represents a gRPC server
type GRPCServer struct {
	addr              string
	server            *grpc.Server
	listener          net.Listener
	logger            logging.Logger
	entryService      service.EntryService
	translationService service.TranslationService
	userService       service.UserService
}

// NewGRPCServer creates a new gRPC server instance
func NewGRPCServer(
	port int,
	logger logging.Logger,
	entryService service.EntryService,
	translationService service.TranslationService,
	userService service.UserService,
) (*GRPCServer, error) {
	addr := fmt.Sprintf(":%d", port)
	
	// Create listener
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	// Create gRPC server with interceptors
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor(logger)),
		// Add more interceptors here as needed (authentication, error handling, etc.)
	)

	return &GRPCServer{
		addr:              addr,
		server:            grpcServer,
		listener:          listener,
		logger:            logger.With(logging.String("component", "grpc_server")),
		entryService:      entryService,
		translationService: translationService,
		userService:       userService,
	}, nil
}

// NewGRPCServerTLS creates a new gRPC server instance with TLS
func NewGRPCServerTLS(
	port int,
	certFile string,
	keyFile string,
	logger logging.Logger,
	entryService service.EntryService,
	translationService service.TranslationService,
	userService service.UserService,
) (*GRPCServer, error) {
	addr := fmt.Sprintf(":%d", port)
	
	// Create listener
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	// Load TLS credentials
	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS credentials: %w", err)
	}

	// Create gRPC server with TLS and interceptors
	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(loggingInterceptor(logger)),
		// Add more interceptors here as needed (authentication, error handling, etc.)
	)

	return &GRPCServer{
		addr:              addr,
		server:            grpcServer,
		listener:          listener,
		logger:            logger.With(logging.String("component", "grpc_server")),
		entryService:      entryService,
		translationService: translationService,
		userService:       userService,
	}, nil
}

// Start starts the gRPC server
func (s *GRPCServer) Start() error {
	// Register service implementations
	dictionaryService := grpcService.NewDictionaryService(s.entryService, s.translationService, s.logger)
	proto.RegisterDictionaryServiceServer(s.server, dictionaryService)

	userService := grpcService.NewUserService(s.userService, s.logger)
	proto.RegisterUserServiceServer(s.server, userService)

	// Register reflection service for development tools
	reflection.Register(s.server)

	s.logger.Info("starting gRPC server", logging.String("address", s.addr))
	
	// Start the server
	return s.server.Serve(s.listener)
}

// Stop gracefully stops the gRPC server
func (s *GRPCServer) Stop() {
	s.logger.Info("stopping gRPC server")
	s.server.GracefulStop()
	s.logger.Info("gRPC server stopped")
}

// loggingInterceptor creates a gRPC interceptor for logging
func loggingInterceptor(logger logging.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		method := info.FullMethod
		
		// Log incoming request
		logger.Debug("gRPC request received", 
			logging.String("method", method),
		)
		
		// Handle the request
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)
		
		// Log the response
		if err != nil {
			logger.Error("gRPC request failed",
				logging.String("method", method),
				logging.Error(err),
				logging.Duration("duration", duration),
			)
		} else {
			logger.Debug("gRPC request completed",
				logging.String("method", method),
				logging.Duration("duration", duration),
			)
		}
		
		return resp, err
	}
}
