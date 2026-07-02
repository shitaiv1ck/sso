package grpcserver

import (
	"context"
	"errors"
	"net"

	ssov1 "github.com/shitaiv1ck/protos/gen/go/sso"
	"github.com/shitaiv1ck/sso/internal/core/logger"
	authgrpc "github.com/shitaiv1ck/sso/internal/features/auth/transport/grpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type GRPCServer struct {
	server *grpc.Server
	log    *logger.Logger
	config Config
}

func NewGRPCServer(log *logger.Logger, config Config) *GRPCServer {
	return &GRPCServer{
		server: grpc.NewServer(),
		log:    log,
		config: config,
	}
}

func (s *GRPCServer) RegisterServices(auth *authgrpc.AuthGRPC) {
	ssov1.RegisterAuthServer(s.server, auth)
}

func (s GRPCServer) Run(ctx context.Context) error {
	s.log.Debug("run gRPC server", zap.String("addres", s.config.Addr))

	errChan := make(chan error)

	go func() {
		defer close(errChan)

		l, err := net.Listen("tcp", s.config.Addr)
		if err != nil {
			errChan <- err
		}

		err = s.server.Serve(l)
		if !errors.Is(err, grpc.ErrServerStopped) {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		s.server.GracefulStop()

		s.log.Debug("gRPC server stopped gracefully")
	case err := <-errChan:
		s.log.Error("failed to run gRPC server", zap.Error(err))

		return err
	}

	return nil
}
