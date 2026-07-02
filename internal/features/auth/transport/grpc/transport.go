package authgrpc

import (
	"context"

	ssov1 "github.com/shitaiv1ck/protos/gen/go/sso"
	"github.com/shitaiv1ck/sso/internal/core/domain"
	errs "github.com/shitaiv1ck/sso/internal/core/errors"
	"github.com/shitaiv1ck/sso/internal/core/logger"
	grpcstatus "github.com/shitaiv1ck/sso/internal/core/transport/grpc/status"
	"github.com/shitaiv1ck/sso/internal/core/validation"
)

type AuthGRPC struct {
	ssov1.UnimplementedAuthServer

	service AuthService
	log     *logger.Logger
}

type AuthService interface {
	Register(ctx context.Context, user domain.User) (int, error)
	Login(ctx context.Context, user domain.User, app domain.App) (string, error)
}

func NewAuthGRPC(service AuthService, log *logger.Logger) *AuthGRPC {
	return &AuthGRPC{
		service: service,
		log:     log,
	}
}

func (t *AuthGRPC) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	t.log.Debug("invoke register user")

	grpcStatus := grpcstatus.NewGRPCStatus(t.log)

	if err := validation.ValidateEmail(req.GetEmail()); err != nil {
		return nil, grpcStatus.Error("failed to validate email", errs.ErrInvalidArg)
	}

	if err := validation.ValidatePassword(req.GetPassword()); err != nil {
		return nil, grpcStatus.Error("failed to validate password", errs.ErrInvalidArg)
	}

	user := domain.NewUnknownUser(req.GetEmail(), req.GetPassword())

	userID, err := t.service.Register(ctx, user)
	if err != nil {
		return nil, grpcStatus.Error("failed to register user", err)
	}

	response := &ssov1.RegisterResponse{UserId: int64(userID)}

	return response, nil
}

func (t *AuthGRPC) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	t.log.Debug("invoke register user")

	grpcStatus := grpcstatus.NewGRPCStatus(t.log)

	if err := validation.ValidateEmail(req.GetEmail()); err != nil {
		return nil, grpcStatus.Error("failed to validate email", err)
	}

	if err := validation.ValidatePassword(req.GetPassword()); err != nil {
		return nil, grpcStatus.Error("failed to validate password", err)
	}

	if err := validation.ValidateID(int(req.GetAppId())); err != nil {
		return nil, grpcStatus.Error("failed to validate app ID", err)
	}

	user := domain.NewUnknownUser(req.GetEmail(), req.GetPassword())

	app := domain.NewUnnamedApp(int(req.GetAppId()))

	token, err := t.service.Login(ctx, user, app)
	if err != nil {
		return nil, grpcStatus.Error("failed to login user", err)
	}

	response := &ssov1.LoginResponse{Token: token}

	return response, nil
}
