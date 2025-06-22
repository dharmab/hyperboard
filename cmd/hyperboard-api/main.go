package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dharmab/hyperboard/internal/db/migrations"
	"github.com/dharmab/hyperboard/pkg/api"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	port               string
	postgresqlHost     string
	postgresqlUser     string
	postgresqlPassword string
	postgresqlDatabase string
	postgresqlSSLMode  string
)

// cmd represents the base command when called without any subcommands
var cmd = &cobra.Command{
	Use:   "hyperboard-api",
	Short: "Hyperboard API server",
	RunE:  func(cmd *cobra.Command, args []string) error { return run(cmd.Context()) },
}

func init() {
	cmd.Flags().StringVar(&port, "port", "8080", "Port to bind for the API server")
	cmd.Flags().StringVar(&postgresqlSSLMode, "postgresql-ssl-mode", "disable", "PostgreSQL SSL mode")
	cmd.Flags().StringVar(&postgresqlHost, "postgresql-host", "localhost:5432", "PostgreSQL host")
	cmd.Flags().StringVar(&postgresqlUser, "postgresql-user", "postgres", "PostgreSQL user")
	cmd.Flags().StringVar(&postgresqlPassword, "postgresql-password", "", "PostgreSQL password")
	cmd.Flags().StringVar(&postgresqlDatabase, "postgresql-database", "postgres", "PostgreSQL database name")
}

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to run server")
	}
}

func run(ctx context.Context) error {
	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s/%s?sslmode=%s",
		postgresqlUser,
		postgresqlPassword,
		postgresqlHost,
		postgresqlDatabase,
		postgresqlSSLMode,
	)

	log.Info().Msg("Running database migrations...")
	if err := migrateDatabase(dsn); err != nil {
		return fmt.Errorf("failed to run database migrations: %w", err)
	}

	log.Info().Msg("Starting API server...")
	if err := serveAPI(ctx, dsn); err != nil {
		return fmt.Errorf("failed to serve API: %w", err)
	}
	return nil
}

func migrateDatabase(dsn string) error {
	if err := migrations.Migrate(dsn); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	return nil
}

func serveAPI(ctx context.Context, dsn string) error {
	apiServer, err := api.NewServer(ctx, dsn)
	if err != nil {
		return fmt.Errorf("failed to create API server: %w", err)
	}
	httpServer := &http.Server{
		Handler: api.HandlerFromMux(
			apiServer,
			http.NewServeMux(),
		),
		Addr: ":" + port,
	}
	if err := httpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to serve API: %w", err)
	}
	return nil
}
