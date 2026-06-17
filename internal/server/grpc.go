// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/internal/server/grpc.go
// Role: Frameworks & drivers — gRPC server bootstrap
// Description: Builds the gRPC server, installs the auth interceptor, registers the auth handler
// and serves with graceful shutdown driven by the caller's context. Takes explicit dependencies
// (not the DI container) so the transport layer stays decoupled from the composition root.

package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"

	"neptune-pamm/github.com/ratheeshkumar25/internal/handler"
	authv1 "neptune-pamm/github.com/ratheeshkumar25/internal/proto/auth/v1"
	"neptune-pamm/github.com/ratheeshkumar25/internal/services"
	"neptune-pamm/github.com/ratheeshkumar25/pkg/utilis"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server wraps a configured gRPC server and its listen port.
type Server struct {
	grpc *grpc.Server
	port int
	log  *slog.Logger
}

// New builds the gRPC server with the auth interceptor and registers the auth service.
func New(port int, auth services.AuthService, jwt *utilis.JWTManager, cache services.SessionCache, log *slog.Logger) *Server {
	interceptor := handler.NewAuthInterceptor(jwt, cache, log)

	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptor.Unary()),
	)
	authv1.RegisterAuthServiceServer(grpcSrv, handler.NewAuthHandler(auth, log))

	// Reflection lets grpcurl / IDE clients introspect the API while there is no UI yet.
	reflection.Register(grpcSrv)

	return &Server{grpc: grpcSrv, port: port, log: log}
}

// Run serves until ctx is cancelled, then stops gracefully (drains in-flight RPCs).
func (s *Server) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("listen on :%d: %w", s.port, err)
	}

	go func() {
		<-ctx.Done()
		s.log.Info("stopping grpc server")
		s.grpc.GracefulStop()
	}()

	s.log.Info("grpc server listening", "port", s.port)
	if err := s.grpc.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return fmt.Errorf("serve grpc: %w", err)
	}
	return nil
}
