package domain

import (
	"fmt"
	"time"

	errs "github.com/shitaiv1ck/sso/internal/core/errors"
	"github.com/shitaiv1ck/sso/internal/core/validation"
)

type Session struct {
	RefreshToken string
	UserID       int
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

func NewSession(
	refreshToken string,
	userID int,
	expiresIn time.Duration,
) (Session, error) {
	createdAt := time.Now()

	s := Session{
		RefreshToken: refreshToken,
		UserID:       userID,
		CreatedAt:    createdAt,
		ExpiresAt:    createdAt.Add(expiresIn),
	}

	if err := s.Validate(); err != nil {
		return Session{}, err
	}

	return s, nil
}

func (s *Session) Validate() error {
	if err := validation.ValidateRefreshToken(s.RefreshToken); err != nil {
		return err
	}

	if err := validation.ValidateID(s.UserID); err != nil {
		return err
	}

	if !s.CreatedAt.Before(s.ExpiresAt) {
		return fmt.Errorf("create time must be before expired time: %w", errs.ErrInvalidArg)
	}

	return nil
}

func (s *Session) IsExpired() bool {
	return s.ExpiresAt.Before(time.Now())
}

type SessionShort struct {
	RefreshToken string
	AccessToken  string
	TTL          time.Duration
}
