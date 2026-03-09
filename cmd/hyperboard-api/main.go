package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/dharmab/hyperboard/internal/db/migrations"
	"github.com/dharmab/hyperboard/pkg/api"
	"github.com/dharmab/hyperboard/pkg/httplog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configPath string

var cmd = &cobra.Command{
	Use:   "hyperboard-api",
	Short: "Hyperboard API server",
	RunE:  func(cmd *cobra.Command, args []string) error { return run(cmd.Context()) },
}

func init() {
	cmd.Flags().StringVar(&configPath, "config", "", "Path to configuration file")
	bindConfig(cmd)
}

func main() {
	cobra.OnInitialize(initConfig)
	if err := cmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to run server")
	}
}

func initConfig() {
	if configPath != "" {
		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			log.Fatal().Err(err).Str("config", configPath).Msg("Failed to read config file")
		}
	}
}

func run(ctx context.Context) error {
	cfg := loadConfig()

	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=%s",
		cfg.PostgreSQL.User,
		cfg.PostgreSQL.Password,
		cfg.PostgreSQL.Host,
		cfg.PostgreSQL.Database,
		cfg.PostgreSQL.SSLMode,
	)

	log.Info().Msg("Running database migrations...")
	if err := migrateDatabase(dsn); err != nil {
		return fmt.Errorf("failed to run database migrations: %w", err)
	}

	log.Info().Msg("Starting API server...")
	if err := serveAPI(ctx, cfg, dsn); err != nil {
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

func serveAPI(ctx context.Context, cfg *Config, dsn string) error {
	storage, err := api.NewS3Storage(
		ctx,
		cfg.Storage.Endpoint,
		cfg.Storage.Bucket,
		cfg.Storage.AccessKey,
		cfg.Storage.SecretKey,
		cfg.Storage.Region,
		cfg.Storage.UsePathStyle,
	)
	if err != nil {
		return fmt.Errorf("failed to create S3 storage: %w", err)
	}

	apiServer, err := api.NewServer(ctx, dsn, storage, cfg.SimilarityThreshold)
	if err != nil {
		return fmt.Errorf("failed to create API server: %w", err)
	}
	mux := http.NewServeMux()
	api.HandlerFromMux(apiServer, mux)
	mux.HandleFunc("/media/", apiServer.HandleMedia)
	authMiddleware := api.BasicAuthMiddleware(cfg.AdminPassword, "/healthz", "/readyz", "/metrics")
	httpServer := &http.Server{
		Handler: httplog.RequestLoggingMiddleware(authMiddleware(mux)),
		Addr:    ":" + cfg.Port,
	}
	if err := httpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to serve API: %w", err)
	}
	return nil
}
