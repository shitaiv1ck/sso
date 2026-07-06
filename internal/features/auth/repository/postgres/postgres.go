package authpg

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shitaiv1ck/sso/internal/core/domain"
	errs "github.com/shitaiv1ck/sso/internal/core/errors"
	"github.com/shitaiv1ck/sso/internal/core/repository/postgres"
)

type AuthPostgres struct {
	store postgres.Postgres
}

func NewAuthPG(store postgres.Postgres) *AuthPostgres {
	return &AuthPostgres{
		store: store,
	}
}

func (r *AuthPostgres) SaveUser(ctx context.Context, user domain.User) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.store.GetTimeout())
	defer cancel()

	query := `
		INSERT INTO sso.users(email, pass_hash)
		VALUES($1, $2)
		RETURNING id, email;
	`

	var savedUser domain.User
	if err := r.store.QueryRow(
		ctx,
		query,
		user.Email,
		user.PasswordHash,
	).Scan(
		&savedUser.ID,
		&savedUser.Email,
	); err != nil {
		var pgxErr *pgconn.PgError
		if errors.As(err, &pgxErr) {
			if pgxErr.Code == "23505" {
				return domain.User{}, fmt.Errorf("user with email=%v alredy exist: %w", user.Email, errs.ErrAlreadyExist)
			}
		}

		return domain.User{}, err
	}

	return savedUser, nil
}

func (r *AuthPostgres) FindUser(ctx context.Context, user domain.User) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.store.GetTimeout())
	defer cancel()

	query := `
		SELECT id, email, pass_hash FROM sso.users
		WHERE email = $1;
	`

	var foundUser domain.User
	if err := r.store.QueryRow(
		ctx,
		query,
		user.Email,
	).Scan(
		&foundUser.ID,
		&foundUser.Email,
		&foundUser.PasswordHash,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, errs.ErrNotFound
		}

		return domain.User{}, err
	}

	return foundUser, nil
}

func (r *AuthPostgres) FindApp(ctx context.Context, app domain.App) (domain.App, error) {
	ctx, cancel := context.WithTimeout(ctx, r.store.GetTimeout())
	defer cancel()

	query := `
		SELECT id, name FROM sso.apps
		WHERE id = $1;
	`

	var foundApp domain.App
	if err := r.store.QueryRow(
		ctx,
		query,
		app.ID,
	).Scan(
		&foundApp.ID,
		&foundApp.Name,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.App{}, errs.ErrNotFound
		}

		return domain.App{}, err
	}

	return foundApp, nil
}

func (r *AuthPostgres) SaveSession(ctx context.Context, session domain.Session) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.store.GetTimeout())
	defer cancel()

	query := `
		INSERT INTO sso.sessions(refresh_token, user_id, created_at, expires_at)
		VALUES($1, $2, $3, $4)
		RETURNING refresh_token;
	`

	var refreshToken string
	if err := r.store.QueryRow(
		ctx,
		query,
		session.RefreshToken,
		session.UserID,
		session.CreatedAt,
		session.ExpiresAt,
	).Scan(
		&refreshToken,
	); err != nil {
		return "", err
	}

	return refreshToken, nil
}

func (r *AuthPostgres) SaveTXSession(ctx context.Context, tx postgres.SQLExecuter, session domain.Session) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.store.GetTimeout())
	defer cancel()

	query := `
		INSERT INTO sso.sessions(refresh_token, user_id, created_at, expires_at)
		VALUES($1, $2, $3, $4)
		RETURNING refresh_token;
	`

	var refreshToken string
	if err := tx.QueryRow(
		ctx,
		query,
		session.RefreshToken,
		session.UserID,
		session.CreatedAt,
		session.ExpiresAt,
	).Scan(
		&refreshToken,
	); err != nil {
		return "", err
	}

	return refreshToken, nil
}

func (r *AuthPostgres) DeleteSession(ctx context.Context, refreshToken string) error {
	ctx, cancel := context.WithTimeout(ctx, r.store.GetTimeout())
	defer cancel()

	query := `
		DELETE FROM sso.sessions
		WHERE refresh_token = $1;
	`

	result, err := r.store.Exec(ctx, query, refreshToken)
	if err != nil {
		return err
	}

	if result.RowsAffected() != 1 {
		return errs.ErrNotFound
	}

	return nil
}

func (r *AuthPostgres) DeleteTXSession(
	ctx context.Context,
	tx postgres.SQLExecuter,
	refreshToken string,
) (domain.Session, domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.store.GetTimeout())
	defer cancel()

	query := `
		DELETE FROM sso.sessions AS s
		USING sso.users AS u
		WHERE s.user_id = u.id AND s.refresh_token = $1
		RETURNING s.expires_at, u.id, u.email;
	`

	var session domain.Session
	var user domain.User
	if err := tx.QueryRow(
		ctx,
		query,
		refreshToken,
	).Scan(
		&session.ExpiresAt,
		&user.ID,
		&user.Email,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Session{}, domain.User{}, errs.ErrNotFound
		}

		return domain.Session{}, domain.User{}, err
	}

	return session, user, nil
}
