package migrations

import (
	"embed"

	"github.com/rs/zerolog/log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed data/*.sql
var data embed.FS

func Migrate(dsn string) error {
	migrationDriver, err := iofs.New(data, "data")
	if err != nil {
		return err
	}
	defer func() {
		if err := migrationDriver.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close migration driver")
		}
	}()
	migrator, err := migrate.NewWithSourceInstance("iofs", migrationDriver, dsn)
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
