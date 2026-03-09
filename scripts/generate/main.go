package main

import (
	"context"
	"fmt"
	"io/fs"
	"net"

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

// freePort asks the OS for an available TCP port.
func freePort() (uint32, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	port := uint32(l.Addr().(*net.TCPAddr).Port)
	_ = l.Close()
	return port, nil
}

func generate(ctx context.Context) error {
	port, err := freePort()
	if err != nil {
		return err
	}

	config := embedpg.DefaultConfig().Port(port)
	postgres := embedpg.NewDatabase(config)
	log.Info().Uint32("port", port).Msg("starting embedded postgres database...")
	if err := postgres.Start(); err != nil {
		return err
	}
	defer func() {
		log.Info().Msg("stopping embedded postgres database...")
		if err := postgres.Stop(); err != nil {
			log.Error().Err(err).Msg("failed to stop embedded postgres")
		}
	}()

	url := fmt.Sprintf("postgresql://postgres:postgres@localhost:%d/postgres?sslmode=disable", port)
	log.Info().Msg("running database migrations...")
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
	d := driver.New(driver.Config{
		Config: helpers.Config{
			Driver: "github.com/jackc/pgx/v5/stdlib",
			Dsn:    url,
			Except: map[string][]string{
				"schema_migrations": {},
			},
		},
	})
	log.Info().Msg("generating code...")
	if err := bobgen.Run(ctx, &state, d); err != nil {
		return err
	}

	return nil
}
