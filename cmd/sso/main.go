package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/shitaiv1ck/sso/internal/core/logger"
	"github.com/shitaiv1ck/sso/internal/core/repository/postgres"
	grpcserver "github.com/shitaiv1ck/sso/internal/core/transport/grpc/server"
	authrep "github.com/shitaiv1ck/sso/internal/features/auth/repository"
	authsrvc "github.com/shitaiv1ck/sso/internal/features/auth/service"
	authgrpc "github.com/shitaiv1ck/sso/internal/features/auth/transport/grpc"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	log, err := logger.NewLogger(logger.NewConfigMust())
	if err != nil {
		panic(err)
	}

	log.Debug("init connection pool...")
	connPool, err := postgres.NewConnPool(ctx, postgres.NewConfigMust())
	if err != nil {
		log.Error("failed to init connection pool", zap.Error(err))

		panic(err)
	}
	defer connPool.Close()

	log.Debug("init feature: auth...")
	authRep := authrep.NewAuthRep(connPool)
	authService := authsrvc.NewAuthService(authRep)
	authGRPC := authgrpc.NewAuthGRPC(authService, log)

	server := grpcserver.NewGRPCServer(log, grpcserver.NewConfigMust())
	server.RegisterServices(authGRPC)

	if err := server.Run(ctx); err != nil {
		panic(err)
	}
}
