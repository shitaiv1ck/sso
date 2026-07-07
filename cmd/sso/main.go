package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/shitaiv1ck/sso/internal/core/client/kafka"
	"github.com/shitaiv1ck/sso/internal/core/logger"
	"github.com/shitaiv1ck/sso/internal/core/repository/postgres"
	"github.com/shitaiv1ck/sso/internal/core/repository/redis"
	grpcserver "github.com/shitaiv1ck/sso/internal/core/transport/grpc/server"
	acckafka "github.com/shitaiv1ck/sso/internal/features/account/client/kafka"
	accpg "github.com/shitaiv1ck/sso/internal/features/account/repository/postgres"
	accsrvc "github.com/shitaiv1ck/sso/internal/features/account/service"
	accgrpc "github.com/shitaiv1ck/sso/internal/features/account/transport/grpc"
	authkafka "github.com/shitaiv1ck/sso/internal/features/auth/client/kafka"
	authpg "github.com/shitaiv1ck/sso/internal/features/auth/repository/postgres"
	authredis "github.com/shitaiv1ck/sso/internal/features/auth/repository/redis"
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

	log.Debug("init redis connection...")
	redisConn, err := redis.NewRedis(ctx, redis.NewConfigMust())
	if err != nil {
		log.Error("failed to init redis connection", zap.Error(err))

		panic(err)
	}
	defer redisConn.Close()

	log.Debug("init kafka connection...")
	kafkaConn, err := kafka.NewKafkaConn(ctx, kafka.NewConfigMust())
	if err != nil {
		log.Error("failed to init kafka connection", zap.Error(err))

		panic(err)
	}
	defer kafkaConn.Close()

	log.Debug("init feature: auth...")
	authPG := authpg.NewAuthPG(connPool)
	authRedis := authredis.NewAuthRedis(redisConn)
	authKafka := authkafka.NewAuthKafka(kafkaConn)
	authService := authsrvc.NewAuthService(authsrvc.NewConfigMust(), authPG, connPool, authRedis, authKafka)
	authGRPC := authgrpc.NewAuthGRPC(authService)

	log.Debug("init feature: account...")
	accPG := accpg.NewAccountPG(connPool)
	accKafka := acckafka.NewAccountKafka(kafkaConn)
	accService := accsrvc.NewAccountService(accPG, connPool, accKafka)
	accGRPC := accgrpc.NewAccountGRPC(accService)

	server := grpcserver.NewGRPCServer(log, grpcserver.NewConfigMust())
	server.RegisterServices(authGRPC, accGRPC)

	if err := server.Run(ctx); err != nil {
		panic(err)
	}
}
