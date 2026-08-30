package async

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/dharmab/hyperboard/internal/leaderelection"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type config struct {
	LogLevel       string
	LeaderElection leaderelection.Config
	SQLStore       sqlStoreConfig
	ObjectStore    objectStoreConfig
	Controller     controllerConfig
}

type sqlStoreConfig struct {
	Host     string
	User     string
	Password string
	Database string
	SSLMode  string
}

type objectStoreConfig struct {
	Endpoint     string
	Bucket       string
	AccessKey    string
	SecretKey    string
	Region       string
	UsePathStyle bool
}

type controllerConfig struct {
	MinInterval time.Duration
	MaxInterval time.Duration
}

func bindConfig(cmd *cobra.Command, values *viper.Viper) error {
	flags := cmd.Flags()

	flags.String("log-level", "info", "Log level (trace, debug, info, warn, error, fatal, panic)")
	flags.String("leader-election", string(leaderelection.Postgres), "Leader election backend (kubernetes or postgres)")
	flags.String("leader-election-identity", "", "Unique identity for this controller instance (defaults to hostname)")
	flags.String("leader-election-lock-name", "hyperboard-async", "Name of the leader election lock")
	flags.String("leader-election-namespace", "default", "Kubernetes namespace containing the leader election Lease")
	flags.String("kubeconfig", "", "Path to a kubeconfig file (defaults to in-cluster configuration)")
	flags.Duration("leader-election-lease-duration", 15*time.Second, "Duration before a Kubernetes Lease can be taken over")
	flags.Duration("leader-election-renew-deadline", 10*time.Second, "Deadline for renewing a Kubernetes Lease")
	flags.Duration("leader-election-retry-period", 2*time.Second, "Delay between leader election attempts")

	flags.String("postgresql-host", "localhost", "PostgreSQL host")
	flags.String("postgresql-user", "hyperboard", "PostgreSQL user")
	flags.String("postgresql-password", "", "PostgreSQL password")
	flags.String("postgresql-database", "hyperboard", "PostgreSQL database name")
	flags.String("postgresql-ssl-mode", "disable", "PostgreSQL SSL mode")

	flags.String("storage-endpoint", "", "S3-compatible storage endpoint")
	flags.String("storage-bucket", "", "S3 bucket name")
	flags.String("storage-access-key", "", "S3 access key")
	flags.String("storage-secret-key", "", "S3 secret key")
	flags.String("storage-region", "", "S3 region")
	flags.Bool("storage-use-path-style", false, "Use path-style S3 URLs")

	flags.Duration("controller-min-interval", 100*time.Millisecond, "Fastest controller heartbeat interval")
	flags.Duration("controller-max-interval", time.Hour, "Slowest controller heartbeat interval")

	values.SetEnvPrefix("HYPERBOARD_ASYNC")
	values.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	values.AutomaticEnv()

	if err := values.BindPFlags(flags); err != nil {
		return fmt.Errorf("bind configuration flags: %w", err)
	}
	return nil
}

func loadConfig(values *viper.Viper) (*config, error) {
	cfg := &config{
		LogLevel: values.GetString("log-level"),
		LeaderElection: leaderelection.Config{
			Backend:       leaderelection.Backend(values.GetString("leader-election")),
			Identity:      values.GetString("leader-election-identity"),
			LockName:      values.GetString("leader-election-lock-name"),
			Namespace:     values.GetString("leader-election-namespace"),
			Kubeconfig:    values.GetString("kubeconfig"),
			LeaseDuration: values.GetDuration("leader-election-lease-duration"),
			RenewDeadline: values.GetDuration("leader-election-renew-deadline"),
			RetryPeriod:   values.GetDuration("leader-election-retry-period"),
		},
		SQLStore: sqlStoreConfig{
			Host:     values.GetString("postgresql-host"),
			User:     values.GetString("postgresql-user"),
			Password: values.GetString("postgresql-password"),
			Database: values.GetString("postgresql-database"),
			SSLMode:  values.GetString("postgresql-ssl-mode"),
		},
		ObjectStore: objectStoreConfig{
			Endpoint:     values.GetString("storage-endpoint"),
			Bucket:       values.GetString("storage-bucket"),
			AccessKey:    values.GetString("storage-access-key"),
			SecretKey:    values.GetString("storage-secret-key"),
			Region:       values.GetString("storage-region"),
			UsePathStyle: values.GetBool("storage-use-path-style"),
		},
		Controller: controllerConfig{
			MinInterval: values.GetDuration("controller-min-interval"),
			MaxInterval: values.GetDuration("controller-max-interval"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (cfg *config) validate() error {
	if err := cfg.LeaderElection.Validate(); err != nil {
		return err
	}
	if cfg.Controller.MinInterval <= 0 {
		return errors.New("controller minimum interval must be positive")
	}
	if cfg.Controller.MaxInterval < cfg.Controller.MinInterval {
		return errors.New("controller maximum interval must be greater than or equal to minimum interval")
	}
	return nil
}

func (cfg sqlStoreConfig) dsn() string {
	dsn := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   cfg.Host,
		Path:   cfg.Database,
	}
	query := dsn.Query()
	query.Set("sslmode", cfg.SSLMode)
	dsn.RawQuery = query.Encode()
	return dsn.String()
}
