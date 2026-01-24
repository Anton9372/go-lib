package transaction_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Anton9372/go-lib/db/postgres/transaction"
)

func setupPgxPool(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "pass",
			"POSTGRES_USER":     "user",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		},
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		if err = container.Terminate(ctx); err != nil {
			t.Fatal("failed to terminate container", err)
		}
	})

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf(
		"postgres://user:pass@%s/testdb?sslmode=disable",
		net.JoinHostPort(host, port.Port()),
	)

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)

	return pool
}

func TestRun_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := setupPgxPool(t)

	runner := transaction.NewPgxTransactionRunner(pool)

	_ = runner.Run(ctx, func(ctx context.Context) error {
		tx, ok := transaction.FromContext(ctx)
		require.True(t, ok)

		_, err := tx.Exec(ctx, `
			CREATE TABLE IF NOT EXISTS test (
			    id INTEGER PRIMARY KEY
			)
		`)
		require.NoError(t, err)

		_, err = tx.Exec(ctx, `
			INSERT INTO test 
			(id) 
			VALUES ($1)
		`, 1)
		require.NoError(t, err)

		return nil
	})

	var count int
	err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM test`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestRun_Error_Rollback_OneOperation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := setupPgxPool(t)

	runner := transaction.NewPgxTransactionRunner(pool)

	err := runner.Run(ctx, func(ctx context.Context) error {
		tx, ok := transaction.FromContext(ctx)
		require.True(t, ok)

		_, err := tx.Exec(ctx, `
			CREATE TABLE IF NOT EXISTS test (
			    id INTEGER PRIMARY KEY
			)
		`)
		require.NoError(t, err)

		return errors.New("boom")
	})

	require.Error(t, err)

	var exists bool
	row := pool.QueryRow(ctx, `
		SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_name = 'test'
			)
	`)

	require.NoError(t, row.Scan(&exists))
	require.False(t, exists)
}

func TestRun_Error_Rollback_MultipleOperations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := setupPgxPool(t)

	runner := transaction.NewPgxTransactionRunner(pool)

	_, err := pool.Exec(ctx, `
			CREATE TABLE IF NOT EXISTS test (
			    id INTEGER PRIMARY KEY
			)
		`)
	require.NoError(t, err)

	err = runner.Run(ctx, func(ctx context.Context) error {
		tx, ok := transaction.FromContext(ctx)
		require.True(t, ok)

		_, err = tx.Exec(ctx, `
			INSERT INTO test 
			(id) 
			VALUES ($1)
		`, 1)
		require.NoError(t, err)

		_, err = tx.Exec(ctx, `
			INSERT INTO test 
			(id) 
			VALUES ($1)
		`, 2)
		require.NoError(t, err)

		return errors.New("boom")
	})

	require.Error(t, err)

	var count int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM test
	`).Scan(&count)

	require.NoError(t, err)
	require.Equal(t, 0, count)
}

func TestRun_Panic_Rollback_OneOperation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := setupPgxPool(t)

	runner := transaction.NewPgxTransactionRunner(pool)

	err := runner.Run(ctx, func(ctx context.Context) error {
		tx, ok := transaction.FromContext(ctx)
		require.True(t, ok)

		_, err := tx.Exec(ctx, `
			CREATE TABLE IF NOT EXISTS test (
			    id INTEGER PRIMARY KEY
			)
		`)
		require.NoError(t, err)

		panic("boom")
	})

	require.Error(t, err)

	var exists bool
	row := pool.QueryRow(ctx, `
		SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_name = 'test'
			)
	`)

	require.NoError(t, row.Scan(&exists))
	require.False(t, exists)
}

func TestRun_Panic_Rollback_MultipleOperations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := setupPgxPool(t)

	runner := transaction.NewPgxTransactionRunner(pool)

	_, err := pool.Exec(ctx, `
			CREATE TABLE IF NOT EXISTS test (
			    id INTEGER PRIMARY KEY
			)
		`)
	require.NoError(t, err)

	err = runner.Run(ctx, func(ctx context.Context) error {
		tx, ok := transaction.FromContext(ctx)
		require.True(t, ok)

		_, err = tx.Exec(ctx, `
			INSERT INTO test 
			(id) 
			VALUES ($1)
		`, 1)
		require.NoError(t, err)

		_, err = tx.Exec(ctx, `
			INSERT INTO test 
			(id) 
			VALUES ($1)
		`, 2)
		require.NoError(t, err)

		panic("boom")
	})

	require.Error(t, err)

	var count int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM test
	`).Scan(&count)

	require.NoError(t, err)
	require.Equal(t, 0, count)
}
