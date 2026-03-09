package api

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/dharmab/hyperboard/internal/db/migrations"
	"github.com/dharmab/hyperboard/internal/storage"
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
	t.Cleanup(func() { cleanTestDB(t) })
	return NewServer(testDB, storage.NewFakeStorage(), 5)
}

func cleanTestDB(t *testing.T) {
	t.Helper()
	ctx := t.Context()
	for _, table := range []string{"posts_tags", "tag_aliases", "tags", "posts", "notes", "tag_categories"} {
		if _, err := testDB.ExecContext(ctx, "DELETE FROM "+table); err != nil {
			t.Logf("warning: failed to clean table %s: %v", table, err)
		}
	}
}
