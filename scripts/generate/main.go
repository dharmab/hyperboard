package main

import (
	"context"
	"io/fs"

	"github.com/dharmab/hyperboard/internal/db/migrations"
	embedpg "github.com/fergusstrange/embedded-postgres"
	"github.com/rs/zerolog/log"
	bobgen "github.com/stephenafamo/bob/gen"
	helpers "github.com/stephenafamo/bob/gen/bobgen-helpers"
	"github.com/stephenafamo/bob/gen/bobgen-psql/driver"
)

//go:generate go run main.go

func main() {
	ctx := context.Background()
	if err := generate(ctx); err != nil {
		log.Fatal().Err(err).Msg("an error occurred")
	}
}

func generate(ctx context.Context) error {
	postgres := embedpg.NewDatabase()
	log.Info().Msg("starting embedded postgres database...")
	if err := postgres.Start(); err != nil {
		return err
	}
	defer func() {
		log.Info().Msg("stopping embedded postgres database...")
		if err := postgres.Stop(); err != nil {
			log.Error().Err(err).Msg("failed to stop embedded postgres")
		}
	}()

	log.Info().Msg("running database migrations...")
	url := "postgresql://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	if err := migrations.Migrate(url); err != nil {
		return err
	}

	packageName := "models"
	state := bobgen.State[any]{
		Config: bobgen.Config[any]{},
		Outputs: helpers.DefaultOutputs(
			"../../internal/db/"+packageName,
			packageName,
			false,
			&helpers.Templates{Models: []fs.FS{bobgen.PSQLModelTemplates}},
		),
	}
	driver := driver.New(driver.Config{
		Config: helpers.Config{
			Driver: "github.com/jackc/pgx/v5/stdlib",
			Dsn:    url,
			Except: map[string][]string{
				"schema_migrations": {},
			},
		},
	})
	log.Info().Msg("generating code...")
	if err := bobgen.Run(ctx, &state, driver); err != nil {
		return err
	}

	return nil
}
