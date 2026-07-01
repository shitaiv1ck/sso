package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/shitaiv1ck/sso/internal/core/logger"
	"github.com/shitaiv1ck/sso/internal/core/repository/postgres"
	grpcserver "github.com/shitaiv1ck/sso/internal/core/transport/grpc/server"
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

	server := grpcserver.NewGRPCServer(log, grpcserver.NewConfigMust())
	if err := server.Run(ctx); err != nil {
		panic(err)
	}
}
