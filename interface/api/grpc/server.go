package grpc

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/valpere/trytrago/application/service"
	"github.com/valpere/trytrago/domain/logging"
	"github.com/valpere/trytrago/interface/api/grpc/proto"
	"github.com/valpere/trytrago/interface/api/grpc/service"
)

// Server represents a gRPC server
type Server struct {
	port              int
	server            *grpc.Server
	logger            logging.Logger
	entryService      service.EntryService
	translationService service.TranslationService
	userService       service.UserService
}

// NewServer creates a new gRPC server instance
func NewServer(
	port int,
	logger logging.Logger,
	entryService service.EntryService,
	translationService service.TranslationService,
	userService service.UserService,
) *Server {
	// Create gRPC server with interceptors
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor(logger)),
		// Add more interceptors here as needed (authentication, error handling, etc.)
	)

	return &Server{
		port:              port,
		server:            grpcServer,
		logger:            logger.With(logging.String("component", "grpc_server")),
		entryService:      entryService,
		translationService: translationService,
		userService:       userService,
	}
}

// Start starts the gRPC server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	// Register service implementations
	dictionaryService := service.NewDictionaryService(s.entryService, s.translationService, s.logger)
	proto.RegisterDictionaryServiceServer(s.server, dictionaryService)

	userService := service.NewUserService(s.userService, s.logger)
	proto.RegisterUserServiceServer(s.server, userService)

	// Register reflection service for development tools
	reflection.Register(s.server)

	s.logger.Info("starting gRPC server", logging.String("address", addr))
	if err := s.server.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve gRPC: %w", err)
	}

	return nil
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
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
		resp, err := handler(ctx, req)
		
		// Log the response
		if err != nil {
			logger.Error("gRPC request failed",
				logging.String("method", method),
				logging.Error(err),
			)
		} else {
			logger.Debug("gRPC request completed",
				logging.String("method", method),
			)
		}
		
		return resp, err
	}
}
