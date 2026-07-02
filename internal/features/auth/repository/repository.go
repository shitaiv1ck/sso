package authrep

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

type AuthRepository struct {
	store postgres.Postgres
}

func NewAuthRep(store postgres.Postgres) *AuthRepository {
	return &AuthRepository{
		store: store,
	}
}

func (r *AuthRepository) SaveUser(ctx context.Context, user domain.User) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.store.GetTimeout())
	defer cancel()

	query := `
		INSERT INTO sso.users(email, pass_hash)
		VALUES($1, $2)
		RETURNING id;
	`

	var userID int
	if err := r.store.QueryRow(
		ctx,
		query,
		user.Email,
		user.PasswordHash,
	).Scan(&userID); err != nil {
		var pgxErr *pgconn.PgError
		if errors.As(err, &pgxErr) {
			if pgxErr.Code == "23505" {
				return -1, fmt.Errorf("user with email=%v alredy exist: %w", user.Email, errs.ErrAlreadyExist)
			}
		}

		return -1, err
	}

	return userID, nil
}

func (r *AuthRepository) FindUser(ctx context.Context, user domain.User) (domain.User, error) {
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

func (r *AuthRepository) FindApp(ctx context.Context, app domain.App) (domain.App, error) {
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
