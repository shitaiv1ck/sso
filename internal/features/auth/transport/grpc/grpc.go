package authgrpc

import (
	"context"

	ssov1 "github.com/shitaiv1ck/protos/gen/go/sso"
	errs "github.com/shitaiv1ck/sso/internal/core/errors"
	"github.com/shitaiv1ck/sso/internal/core/logger"
	grpcstatus "github.com/shitaiv1ck/sso/internal/core/transport/grpc/status"
	"github.com/shitaiv1ck/sso/internal/core/validation"
)

//go:generate mockgen -source=grpc.go -destination=mocks/mock.go

type AuthGRPC struct {
	ssov1.UnimplementedAuthServer

	service AuthService
}

type AuthService interface {
	Register(ctx context.Context, email string, password string) (int, error)
	Login(ctx context.Context, email string, password string, appID int) (string, string, error)
	Refresh(ctx context.Context, refreshToken string, appID int) (string, string, error)
	Logout(ctx context.Context, refreshToken string, accessToken string) error
}

func NewAuthGRPC(service AuthService) *AuthGRPC {
	return &AuthGRPC{
		service: service,
	}
}

func (t *AuthGRPC) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	log := logger.FromContext(ctx)
	grpcStatus := grpcstatus.NewGRPCStatus(log)

	log.Debug("invoke register user")

	if err := validation.ValidateEmail(req.GetEmail()); err != nil {
		return nil, grpcStatus.Error("failed to validate email", errs.ErrInvalidArg)
	}

	if err := validation.ValidatePassword(req.GetPassword()); err != nil {
		return nil, grpcStatus.Error("failed to validate password", errs.ErrInvalidArg)
	}

	userID, err := t.service.Register(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, grpcStatus.Error("failed to register user", err)
	}

	response := &ssov1.RegisterResponse{UserId: int64(userID)}

	return response, nil
}

func (t *AuthGRPC) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	log := logger.FromContext(ctx)
	grpcStatus := grpcstatus.NewGRPCStatus(log)

	log.Debug("invoke login user")

	if err := validation.ValidateEmail(req.GetEmail()); err != nil {
		return nil, grpcStatus.Error("failed to validate email", err)
	}

	if err := validation.ValidatePassword(req.GetPassword()); err != nil {
		return nil, grpcStatus.Error("failed to validate password", err)
	}

	if err := validation.ValidateID(int(req.GetAppId())); err != nil {
		return nil, grpcStatus.Error("failed to validate app ID", err)
	}

	accessToken, refreshToken, err := t.service.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		return nil, grpcStatus.Error("failed to login user", err)
	}

	response := &ssov1.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return response, nil
}

func (t *AuthGRPC) Refresh(ctx context.Context, req *ssov1.RefreshRequest) (*ssov1.RefreshResponse, error) {
	log := logger.FromContext(ctx)
	grpcStatus := grpcstatus.NewGRPCStatus(log)

	log.Debug("invoke refresh session")

	if err := validation.ValidateRefreshToken(req.GetRefreshToken()); err != nil {
		return nil, grpcStatus.Error("failed to validate refresh token", err)
	}

	if err := validation.ValidateID(int(req.GetAppId())); err != nil {
		return nil, grpcStatus.Error("failed to validate app ID", err)
	}

	accessToken, refreshToken, err := t.service.Refresh(ctx, req.GetRefreshToken(), int(req.GetAppId()))
	if err != nil {
		return nil, grpcStatus.Error("failed to refresh session", err)
	}

	response := &ssov1.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return response, nil
}

func (t *AuthGRPC) Logout(ctx context.Context, req *ssov1.LogoutRequest) (*ssov1.Empty, error) {
	log := logger.FromContext(ctx)
	grpcStatus := grpcstatus.NewGRPCStatus(log)

	log.Debug("invoke logout user")

	if err := validation.ValidateRefreshToken(req.GetRefreshToken()); err != nil {
		return nil, grpcStatus.Error("failed to validate refresh token", err)
	}

	if err := t.service.Logout(ctx, req.GetRefreshToken(), req.GetAccessToken()); err != nil {
		return nil, grpcStatus.Error("failed to logout user", err)
	}

	return &ssov1.Empty{}, nil
}
