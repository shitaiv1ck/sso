package accsrvc

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/shitaiv1ck/sso/internal/core/domain"
	errs "github.com/shitaiv1ck/sso/internal/core/errors"
	"github.com/shitaiv1ck/sso/internal/core/repository/postgres"
	"github.com/shitaiv1ck/sso/internal/core/validation"
)

type AccountService struct {
	pg    AccountPostgres
	tx    postgres.TxManager
	kafka AccountKafka
}

type AccountPostgres interface {
	FindUserByID(ctx context.Context, userID int) (domain.User, error)
	UpdateTXUser(ctx context.Context, tx postgres.SQLExecuter, user domain.User) error
	DeleteTXSessionsByUserID(ctx context.Context, tx postgres.SQLExecuter, userID int) error
}

type AccountKafka interface {
	EventUserUpdated(ctx context.Context, user domain.User) error
}

func NewAccountService(pg AccountPostgres, tx postgres.TxManager, kafka AccountKafka) *AccountService {
	return &AccountService{
		pg:    pg,
		tx:    tx,
		kafka: kafka,
	}
}

func (s *AccountService) ChangePassword(ctx context.Context, userID int, oldPassword string, newPassword string) error {
	if err := validation.ValidateID(userID); err != nil {
		return err
	}

	if err := validation.ValidatePassword(oldPassword); err != nil {
		return err
	}

	if err := validation.ValidatePassword(newPassword); err != nil {
		return err
	}

	if oldPassword == newPassword {
		return errs.ErrInvalidArg
	}

	user, err := s.pg.FindUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := user.ChangePassword(oldPassword, newPassword); err != nil {
		return err
	}

	tx, err := s.tx.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := s.pg.UpdateTXUser(ctx, tx, user); err != nil {
		return err
	}

	if err := s.pg.DeleteTXSessionsByUserID(ctx, tx, user.ID); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (s *AccountService) ChangeEmail(ctx context.Context, userID int, password string, newEmail string) error {
	if err := validation.ValidateID(userID); err != nil {
		return err
	}

	if err := validation.ValidatePassword(password); err != nil {
		return err
	}

	if err := validation.ValidateEmail(newEmail); err != nil {
		return err
	}

	user, err := s.pg.FindUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if !user.ComparePassword(password) {
		return fmt.Errorf("invalid password: %w", errs.ErrInvalidArg)
	}

	if user.Email == newEmail {
		return fmt.Errorf("new email can't be equal current email: %w", errs.ErrInvalidArg)
	}

	if err := user.ChangeEmail(newEmail); err != nil {
		return err
	}

	tx, err := s.tx.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := s.pg.UpdateTXUser(ctx, tx, user); err != nil {
		return err
	}

	if err := s.pg.DeleteTXSessionsByUserID(ctx, tx, user.ID); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	go func() {
		bgCtx := context.Background()
		s.kafka.EventUserUpdated(bgCtx, user)
	}()

	return nil
}
