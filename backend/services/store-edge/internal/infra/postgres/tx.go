package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type txKey struct{}

type dbConn interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (s *Store) conn(ctx context.Context) dbConn {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return s.pool
}

func (s *Store) Run(ctx context.Context, fn func(context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return fn(ctx)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}

	txCtx := context.WithValue(ctx, txKey{}, tx)
	defer tx.Rollback(txCtx) //nolint:errcheck

	if err := fn(txCtx); err != nil {
		return err
	}
	return tx.Commit(txCtx)
}
