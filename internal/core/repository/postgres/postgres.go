package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)

	GetTimeout() time.Duration
}

type ConnPool struct {
	*pgxpool.Pool

	timeout time.Duration
}

func NewConnPool(ctx context.Context, config Config) (*ConnPool, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DB,
	)

	poolcfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolcfg)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &ConnPool{
		Pool:    pool,
		timeout: config.Timeout,
	}, nil
}

func (c *ConnPool) GetTimeout() time.Duration {
	return c.timeout
}
