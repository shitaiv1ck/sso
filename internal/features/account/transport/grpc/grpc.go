package accgrpc

import (
	"context"

	ssov1 "github.com/shitaiv1ck/protos/gen/go/sso"
	errs "github.com/shitaiv1ck/sso/internal/core/errors"
	"github.com/shitaiv1ck/sso/internal/core/logger"
	grpcstatus "github.com/shitaiv1ck/sso/internal/core/transport/grpc/status"
	"github.com/shitaiv1ck/sso/internal/core/validation"
)

type AccountGRPC struct {
	ssov1.UnimplementedAccountServer

	service AccountService
}

type AccountService interface {
	ChangePassword(ctx context.Context, userID int, oldPassword string, newPassword string) error
	ChangeEmail(ctx context.Context, userID int, password string, newEmail string) error
}

func NewAccountGRPC(service AccountService) *AccountGRPC {
	return &AccountGRPC{
		service: service,
	}
}

func (t *AccountGRPC) ChangePassword(ctx context.Context, req *ssov1.PasswordRequest) (*ssov1.Empty, error) {
	log := logger.FromContext(ctx)
	grpcStatus := grpcstatus.NewGRPCStatus(log)

	log.Debug("invoke change user password")

	if err := validation.ValidateID(int(req.GetUserId())); err != nil {
		return nil, grpcStatus.Error("failed to validate user ID", err)
	}

	if err := validation.ValidatePassword(req.GetOldPassword()); err != nil {
		return nil, grpcStatus.Error("failed to validate password", err)
	}

	if err := validation.ValidatePassword(req.GetNewPassword()); err != nil {
		return nil, grpcStatus.Error("failed to validate password", err)
	}

	if req.GetOldPassword() == req.GetNewPassword() {
		return nil, grpcStatus.Error("new password can't be equal old password", errs.ErrInvalidArg)
	}

	if err := t.service.ChangePassword(ctx, int(req.GetUserId()), req.GetOldPassword(), req.GetNewPassword()); err != nil {
		return nil, grpcStatus.Error("failed to change password", err)
	}

	return &ssov1.Empty{}, nil
}

func (t *AccountGRPC) ChangeEmail(ctx context.Context, req *ssov1.EmailRequest) (*ssov1.Empty, error) {
	log := logger.FromContext(ctx)
	grpcStatus := grpcstatus.NewGRPCStatus(log)

	log.Debug("invoke change user emaiil")

	if err := validation.ValidateID(int(req.GetUserId())); err != nil {
		return nil, grpcStatus.Error("failed to validate user ID", err)
	}

	if err := validation.ValidatePassword(req.GetPassword()); err != nil {
		return nil, grpcStatus.Error("failed to validate password", err)
	}

	if err := validation.ValidateEmail(req.GetNewEmail()); err != nil {
		return nil, grpcStatus.Error("failed to validate email", err)
	}

	if err := t.service.ChangeEmail(ctx, int(req.GetUserId()), req.GetPassword(), req.GetNewEmail()); err != nil {
		return nil, grpcStatus.Error("failed to change email", err)
	}

	return &ssov1.Empty{}, nil
}
