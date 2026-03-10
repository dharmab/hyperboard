package api

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/dharmab/hyperboard/internal/db/migrations"
	"github.com/dharmab/hyperboard/internal/storage/memory"
	embedpg "github.com/fergusstrange/embedded-postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stephenafamo/bob"
)

var testDB bob.DB

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

	dsn := fmt.Sprintf("postgresql://postgres:postgres@localhost:%d/postgres?sslmode=disable", port)
	if err := migrations.Migrate(dsn); err != nil {
		_ = postgres.Stop()
		fmt.Fprintf(os.Stderr, "failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		_ = postgres.Stop()
		fmt.Fprintf(os.Stderr, "failed to create pool: %v\n", err)
		os.Exit(1)
	}

	testDB = bob.NewDB(stdlib.OpenDBFromPool(pool))

	code := m.Run()

	pool.Close()
	_ = postgres.Stop()
	os.Exit(code)
}

func freePort() (uint32, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	port := uint32(l.Addr().(*net.TCPAddr).Port)
	_ = l.Close()
	return port, nil
}

func newTestServer(t *testing.T) *Server {
	t.Helper()
	// No cleanup needed: tests use unique random data (UUIDs, random tag names)
	// and query by specific IDs/names, so they don't interfere with each other.
	return NewServer(testDB, memory.New(), 5)
}
