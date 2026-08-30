package async

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dharmab/hyperboard/internal/heartbeat"
	"github.com/dharmab/hyperboard/internal/leaderelection"
	storages3 "github.com/dharmab/hyperboard/internal/storage/s3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewCommand returns the cobra command for the hyperboard-async controller.
func NewCommand() *cobra.Command {
	values := viper.New()
	var configPath string
	cmd := &cobra.Command{
		Use:          "hyperboard-async",
		Short:        "Hyperboard async controller",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			if configPath == "" {
				return nil
			}
			values.SetConfigFile(configPath)
			if err := values.ReadInConfig(); err != nil {
				return fmt.Errorf("read config file %q: %w", configPath, err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := loadConfig(values)
			if err != nil {
				return fmt.Errorf("load configuration: %w", err)
			}
			return run(cmd.Context(), cfg)
		},
	}
	cmd.Flags().StringVar(&configPath, "config", "", "Path to configuration file")
	if err := bindConfig(cmd, values); err != nil {
		panic(err)
	}
	return cmd
}

func run(ctx context.Context, cfg *config) error {
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("parse log level: %w", err)
	}
	zerolog.SetGlobalLevel(level)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	shutdownCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(shutdownCtx, cfg.SQLStore.dsn())
	if err != nil {
		return fmt.Errorf("create database pool: %w", err)
	}
	defer pool.Close()
	if err := pool.Ping(shutdownCtx); err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}

	mediaStore, err := storages3.New(
		shutdownCtx,
		cfg.ObjectStore.Endpoint,
		cfg.ObjectStore.Bucket,
		cfg.ObjectStore.AccessKey,
		cfg.ObjectStore.SecretKey,
		cfg.ObjectStore.Region,
		cfg.ObjectStore.UsePathStyle,
	)
	if err != nil {
		return fmt.Errorf("connect to object storage: %w", err)
	}

	deletions := deletionController{
		repository: postgresDeletionRepository{pool: pool},
		mediaStore: mediaStore,
	}
	fileSizes := fileSizeController{
		repository: postgresFileSizeRepository{pool: pool},
		mediaStore: mediaStore,
	}
	reconcile := func(reconcileCtx context.Context) (bool, error) {
		didWork, err := deletions.Reconcile(reconcileCtx)
		if err != nil || didWork {
			return didWork, err
		}
		return fileSizes.Reconcile(reconcileCtx)
	}
	leaderTask := func(leaderCtx context.Context) error {
		return heartbeat.Run(
			leaderCtx,
			cfg.Controller.MinInterval,
			cfg.Controller.MaxInterval,
			reconcile,
		)
	}

	log.Info().Str("backend", string(cfg.LeaderElection.Backend)).Msg("Starting async controller")
	return leaderelection.Run(shutdownCtx, pool, cfg.LeaderElection, leaderTask)
}
