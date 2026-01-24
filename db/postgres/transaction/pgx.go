package transaction

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Anton9372/go-lib/db/transaction"
)

type txKey struct{}

type PgxTransactionRunner struct {
	pool *pgxpool.Pool
}

func NewPgxTransactionRunner(pool *pgxpool.Pool) *PgxTransactionRunner {
	return &PgxTransactionRunner{pool: pool}
}

func (r *PgxTransactionRunner) Run(ctx context.Context, fn func(ctx context.Context) error) (txErr error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if txErr != nil {
			rollbackErr := tx.Rollback(ctx)
			if rollbackErr != nil {
				txErr = errors.Join(txErr, fmt.Errorf("rollback transaction: %w", rollbackErr))
			}
		}
	}()

	defer func() {
		if rec := recover(); rec != nil {
			if e, ok := rec.(error); ok {
				txErr = fmt.Errorf("%w: %w\nstack:\n%s", transaction.ErrTransactionPanic, e, string(debug.Stack()))
			} else {
				txErr = fmt.Errorf("%w: panic value: %v\nstack:\n%s", transaction.ErrTransactionPanic, rec, string(debug.Stack()))
			}
		}
	}()

	ctx = context.WithValue(ctx, txKey{}, tx)

	err = fn(ctx)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

//nolint:ireturn // pgx.Tx designed as an interface, it is returned from e.g. pool.Begin
func FromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)

	return tx, ok
}
