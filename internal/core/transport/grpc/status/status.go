package grpcstatus

import (
	"errors"
	"fmt"

	errs "github.com/shitaiv1ck/sso/internal/core/errors"
	"github.com/shitaiv1ck/sso/internal/core/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Status struct {
	statusCode codes.Code

	log *logger.Logger
}

func NewGRPCStatus(log *logger.Logger) *Status {
	return &Status{
		log: log,
	}
}

func (s *Status) Error(msg string, err error) error {
	s.SetStatusCode(err)

	if s.statusCode == codes.Internal {
		s.log.Error(msg, zap.Error(err), zap.String("status", s.statusCode.String()))
	} else {
		s.log.Warn(msg, zap.Error(err), zap.String("status", s.statusCode.String()))
	}

	return status.Error(s.statusCode, fmt.Sprintf("%s: %s", msg, err))
}

func (s *Status) SetStatusCode(err error) {
	switch {
	case errors.Is(err, errs.ErrInvalidArg):
		s.statusCode = codes.InvalidArgument
	case errors.Is(err, errs.ErrAlreadyExist):
		s.statusCode = codes.AlreadyExists
	case errors.Is(err, errs.ErrInvalidCredentials):
		s.statusCode = codes.Unauthenticated
	case errors.Is(err, errs.ErrNotFound):
		s.statusCode = codes.NotFound
	case errors.Is(err, errs.ErrRefSession):
		s.statusCode = codes.Unauthenticated
	default:
		s.statusCode = codes.Internal
	}
}
