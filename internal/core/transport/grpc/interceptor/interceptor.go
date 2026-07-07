package grpcintrcpt

import (
	"context"

	"github.com/shitaiv1ck/sso/internal/core/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Logger(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		l := log.With(
			zap.String("method", info.FullMethod),
		)

		ctx = logger.ContextWithLogger(ctx, l)

		response, err := handler(ctx, req)

		return response, err
	}
}

func Trace() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		log := logger.FromContext(ctx)

		log.Debug(
			">>> incoming grpc request",
		)

		response, err := handler(ctx, req)

		respStatus := status.Code(err)

		log.Debug(
			"<<< done grpc request",
			zap.String("status", respStatus.String()),
		)

		return response, err
	}
}

func Panic() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		log := logger.FromContext(ctx)

		defer func() {
			if p := recover(); p != nil {
				log.Error(
					"unexpected panic",
					zap.Any("panic", p),
					zap.String("status", codes.Internal.String()),
				)

				err = status.Error(codes.Internal, "unexpected panic")
			}
		}()

		response, err := handler(ctx, req)

		return response, err
	}
}
