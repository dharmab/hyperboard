package db

import (
	"embed"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"golang.org/x/net/context"

	"github.com/jackc/pgx/v5/pgxpool"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
)

//go:embed migrations/*.sql
var migrations embed.FS

func Migrate(url string) error {
	migrationDriver, err := iofs.New(migrations, "migrations")
	if err != nil {
		return err
	}
	defer migrationDriver.Close()
	migrator, err := migrate.NewWithSourceInstance("iofs", migrationDriver, url)
	if err != nil {
		return err
	}
	defer migrator.Close()
	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

type Database interface {
	Close()
}

type database struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, url string) (*database, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	return &database{
		pool: pool,
	}, nil
}

func (db *database) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}
