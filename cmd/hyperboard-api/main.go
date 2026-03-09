package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dharmab/hyperboard/internal/db/migrations"
	"github.com/dharmab/hyperboard/pkg/api"
	"github.com/dharmab/hyperboard/pkg/authmw"
	"github.com/dharmab/hyperboard/pkg/httplog"
	"github.com/dharmab/hyperboard/pkg/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stephenafamo/bob"
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
	objStorage, err := storage.NewS3Storage(
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

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("failed to create database pool: %w", err)
	}
	defer pool.Close()

	db := bob.NewDB(stdlib.OpenDBFromPool(pool))
	apiServer := api.NewServer(db, objStorage, cfg.SimilarityThreshold)
	mux := http.NewServeMux()
	api.HandlerFromMux(apiServer, mux)
	mux.HandleFunc("/media/", apiServer.HandleMedia)
	authMiddleware := authmw.BasicAuthMiddleware(cfg.AdminPassword, "/healthz", "/readyz", "/metrics")
	httpServer := &http.Server{
		Handler: httplog.RequestLoggingMiddleware(authMiddleware(mux)),
		Addr:    ":" + cfg.Port,
	}

	shutdownCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("failed to serve API: %w", err)
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-shutdownCtx.Done():
		log.Info().Msg("Shutting down API server...")
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(timeoutCtx); err != nil {
			return fmt.Errorf("failed to shut down API server: %w", err)
		}
		log.Info().Msg("API server stopped")
		return nil
	}
}
