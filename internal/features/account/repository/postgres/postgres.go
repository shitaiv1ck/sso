package accpg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shitaiv1ck/sso/internal/core/domain"
	errs "github.com/shitaiv1ck/sso/internal/core/errors"
	"github.com/shitaiv1ck/sso/internal/core/repository/postgres"
)

type AccountPostgres struct {
	store postgres.Postgres
}

func NewAccountPG(store postgres.Postgres) *AccountPostgres {
	return &AccountPostgres{
		store: store,
	}
}

func (r *AccountPostgres) FindUserByID(ctx context.Context, userID int) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.store.GetTimeout())
	defer cancel()

	query := `
		SELECT id, email, pass_hash
		FROM sso.users
		WHERE id = $1;
	`

	var user domain.User
	if err := r.store.QueryRow(
		ctx,
		query,
		userID,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, errs.ErrNotFound
		}

		return domain.User{}, nil
	}

	return user, nil
}

func (r *AccountPostgres) UpdateTXUser(ctx context.Context, tx postgres.SQLExecuter, user domain.User) error {
	ctx, cancel := context.WithTimeout(ctx, r.store.GetTimeout())
	defer cancel()

	query := `
		UPDATE sso.users
		SET email = $1, pass_hash = $2
		WHERE id = $3;
	`

	result, err := tx.Exec(ctx, query, user.Email, user.PasswordHash, user.ID)
	if err != nil {
		var pgxErr *pgconn.PgError
		if errors.As(err, &pgxErr) {
			if pgxErr.Code == "23505" {
				return errs.ErrAlreadyExists
			}
		}

		return err
	}

	if result.RowsAffected() == 0 {
		return errs.ErrNotFound
	}

	return nil
}

func (r *AccountPostgres) DeleteTXSessionsByUserID(ctx context.Context, tx postgres.SQLExecuter, userID int) error {
	ctx, cancel := context.WithTimeout(ctx, r.store.GetTimeout())
	defer cancel()

	query := `
		DELETE FROM sso.sessions
		WHERE user_id = $1;
	`

	_, err := tx.Exec(ctx, query, userID)
	if err != nil {
		return err
	}

	return nil
}
