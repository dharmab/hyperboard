package api

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"uuid"

	"github.com/dharmab/hyperboard/internal/db/migrations"
	"github.com/dharmab/hyperboard/internal/db/store"
	"github.com/dharmab/hyperboard/internal/storage/memory"
	embedpg "github.com/fergusstrange/embedded-postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

var (
	testAdminPool *pgxpool.Pool
	testDSN       string
	testSQLStore  store.SQLStore
)

func TestMain(m *testing.M) {
	port, err := freePort()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to find free port: %v\n", err)
		os.Exit(1)
	}

	config := embedpg.DefaultConfig().Port(port)
	postgres := embedpg.NewDatabase(config)
	if err := postgres.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start embedded postgres: %v\n", err)
		os.Exit(1)
	}

	testDSN = fmt.Sprintf("postgresql://postgres:postgres@localhost:%d/postgres?sslmode=disable", port)
	if err := migrations.Migrate(testDSN); err != nil {
		if stopErr := postgres.Stop(); stopErr != nil {
			fmt.Fprintf(os.Stderr, "failed to stop embedded postgres after migration failure: %v\n", stopErr)
		}
		fmt.Fprintf(os.Stderr, "failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	pool, err := pgxpool.New(context.Background(), testDSN)
	if err != nil {
		if stopErr := postgres.Stop(); stopErr != nil {
			fmt.Fprintf(os.Stderr, "failed to stop embedded postgres after pool creation failure: %v\n", stopErr)
		}
		fmt.Fprintf(os.Stderr, "failed to create pool: %v\n", err)
		os.Exit(1)
	}

	lockPool, err := pgxpool.New(context.Background(), testDSN)
	if err != nil {
		pool.Close()
		if stopErr := postgres.Stop(); stopErr != nil {
			fmt.Fprintf(os.Stderr, "failed to stop embedded postgres after lock pool creation failure: %v\n", stopErr)
		}
		fmt.Fprintf(os.Stderr, "failed to create lock pool: %v\n", err)
		os.Exit(1)
	}

	testAdminPool = pool
	db := stdlib.OpenDBFromPool(pool)
	lockDB := stdlib.OpenDBFromPool(lockPool)
	testSQLStore = store.NewPostgresSQLStore(db, lockDB, 5)

	code := m.Run()

	lockPool.Close()
	pool.Close()
	if err := postgres.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to stop embedded postgres: %v\n", err)
		code = 1
	}
	os.Exit(code)
}

func freePort() (uint32, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	port := uint32(l.Addr().(*net.TCPAddr).Port)
	if err := l.Close(); err != nil {
		return 0, fmt.Errorf("close temporary listener: %w", err)
	}
	return port, nil
}

func newTestServer(t *testing.T) *Server {
	t.Helper()
	return NewServer(testSQLStore, memory.New())
}

func newTestStore(t *testing.T) store.SQLStore {
	t.Helper()
	sqlStore, _ := newTestStoreWithPool(t)
	return sqlStore
}

func newTestStoreWithPool(t *testing.T) (store.SQLStore, *pgxpool.Pool) {
	t.Helper()

	databaseName := "test_" + strings.ReplaceAll(uuid.NewV4().String(), "-", "")
	_, err := testAdminPool.Exec(context.Background(), `CREATE DATABASE "`+databaseName+`"`)
	require.NoError(t, err, "create isolated test database")

	dsn := strings.Replace(testDSN, "/postgres?", "/"+databaseName+"?", 1)
	err = migrations.Migrate(dsn)
	if err != nil {
		dropTestDatabase(t, databaseName)
	}
	require.NoError(t, err, "migrate isolated test database")

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		dropTestDatabase(t, databaseName)
	}
	require.NoError(t, err, "connect to isolated test database")
	lockPool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		pool.Close()
		dropTestDatabase(t, databaseName)
	}
	require.NoError(t, err, "connect post mutation lock pool to isolated test database")
	t.Cleanup(func() {
		lockPool.Close()
		pool.Close()
		dropTestDatabase(t, databaseName)
	})

	db := stdlib.OpenDBFromPool(pool)
	lockDB := stdlib.OpenDBFromPool(lockPool)
	return store.NewPostgresSQLStore(db, lockDB, 5), pool
}

func dropTestDatabase(t *testing.T, databaseName string) {
	t.Helper()
	_, err := testAdminPool.Exec(context.Background(), `DROP DATABASE "`+databaseName+`" WITH (FORCE)`)
	assert.NoError(t, err, "drop isolated test database")
}
