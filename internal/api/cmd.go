package api

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
	"github.com/dharmab/hyperboard/internal/db/store"
	"github.com/dharmab/hyperboard/internal/middleware/auth"
	"github.com/dharmab/hyperboard/internal/middleware/logging"
	s3storage "github.com/dharmab/hyperboard/internal/storage/s3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configPath is the path to the configuration file, set via the --config CLI flag.
var configPath string

// NewCommand returns the cobra command for the hyperboard-api server.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hyperboard-api",
		Short: "Hyperboard API server",
		RunE:  func(cmd *cobra.Command, args []string) error { return run(cmd.Context()) },
	}
	cmd.Flags().StringVar(&configPath, "config", "", "Path to configuration file")
	bindConfig(cmd)
	cobra.OnInitialize(initConfig)
	return cmd
}

// initConfig reads the configuration file if configPath is set.
func initConfig() {
	if configPath != "" {
		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			log.Fatal().Err(err).Str("config", configPath).Msg("Failed to read config file")
		}
	}
}

// run initializes and starts the API server with the given context.
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
		cfg.SQLStore.User,
		cfg.SQLStore.Password,
		cfg.SQLStore.Host,
		cfg.SQLStore.Database,
		cfg.SQLStore.SSLMode,
	)

	log.Info().Msg("Running database migrations...")
	if err := migrations.Migrate(dsn); err != nil {
		return fmt.Errorf("failed to run database migrations: %w", err)
	}

	log.Info().Msg("Starting API server...")
	return serveAPI(ctx, cfg, dsn)
}

// serveAPI sets up the HTTP server, storage backends, database connection, and starts serving the API.
func serveAPI(ctx context.Context, cfg *config, dsn string) error {
	objStorage, err := s3storage.New(
		ctx,
		cfg.ObjectStore.Endpoint,
		cfg.ObjectStore.Bucket,
		cfg.ObjectStore.AccessKey,
		cfg.ObjectStore.SecretKey,
		cfg.ObjectStore.Region,
		cfg.ObjectStore.UsePathStyle,
	)
	if err != nil {
		return fmt.Errorf("failed to create S3 storage: %w", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("failed to create database pool: %w", err)
	}
	defer pool.Close()

	db := stdlib.OpenDBFromPool(pool)
	s := store.NewPostgresSQLStore(db, cfg.SimilarityThreshold)
	apiServer := NewServer(s, objStorage)
	mux := http.NewServeMux()
	HandlerFromMux(apiServer, mux)
	mux.HandleFunc("/media/", apiServer.HandleMedia)
	authMiddleware := auth.BasicAuthMiddleware(cfg.AdminPassword, "/healthz", "/readyz", "/metrics")
	httpServer := &http.Server{
		Handler:           logging.RequestLoggingMiddleware(authMiddleware(mux)),
		Addr:              ":" + cfg.Port,
		ReadHeaderTimeout: 30 * time.Second,
	}

	shutdownCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Info().Str("port", cfg.Port).Msg("Starting API server")
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
		//nolint:contextcheck // intentional: using fresh context for graceful shutdown after signal cancellation
		if err := httpServer.Shutdown(timeoutCtx); err != nil {
			return fmt.Errorf("failed to shut down API server: %w", err)
		}
		log.Info().Msg("API server stopped")
		return nil
	}
}
