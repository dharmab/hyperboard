package db

import (
	"embed"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/rs/zerolog/log"
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
	defer func() {
		if err := migrationDriver.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close migration driver")
		}
	}()
	migrator, err := migrate.NewWithSourceInstance("iofs", migrationDriver, url)
	if err != nil {
		return err
	}
	defer func() {
		sourceErr, databaseErr := migrator.Close()
		if sourceErr != nil {
			log.Error().Err(err).Msg("Error closing migration source")
		}
		if databaseErr != nil {
			log.Error().Err(err).Msg("Error closing migration database")
		}
	}()
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
