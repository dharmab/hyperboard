package api

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// config holds the configuration for the API server.
type config struct {
	Port                string
	AdminPassword       string
	LogLevel            string
	SimilarityThreshold int
	SQLStore            sqlStoreConfig
	ObjectStore         objectStoreConfig
}

// sqlStoreConfig holds PostgreSQL connection configuration.
type sqlStoreConfig struct {
	Host     string
	User     string
	Password string
	Database string
	SSLMode  string
}

// objectStoreConfig holds S3-compatible object store configuration.
type objectStoreConfig struct {
	Endpoint     string
	Bucket       string
	AccessKey    string
	SecretKey    string
	Region       string
	UsePathStyle bool
}

// bindConfig registers CLI flags and environment variable bindings for the API server configuration.
func bindConfig(cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.String("port", "8080", "Port to listen on")
	flags.String("admin-password", "", "Admin password for basic auth")
	flags.String("log-level", "info", "Log level (trace, debug, info, warn, error, fatal, panic)")
	flags.Int("similarity-threshold", 10, "Maximum Hamming distance for perceptual hash similarity (0-64, lower is stricter)")

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

	viper.SetEnvPrefix("HYPERBOARD_API")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	_ = viper.BindPFlags(flags)
}

// loadConfig reads configuration values from viper and returns a populated config struct.
func loadConfig() *config {
	return &config{
		Port:                viper.GetString("port"),
		AdminPassword:       viper.GetString("admin-password"),
		LogLevel:            viper.GetString("log-level"),
		SimilarityThreshold: viper.GetInt("similarity-threshold"),
		SQLStore: sqlStoreConfig{
			Host:     viper.GetString("postgresql-host"),
			User:     viper.GetString("postgresql-user"),
			Password: viper.GetString("postgresql-password"),
			Database: viper.GetString("postgresql-database"),
			SSLMode:  viper.GetString("postgresql-ssl-mode"),
		},
		ObjectStore: objectStoreConfig{
			Endpoint:     viper.GetString("storage-endpoint"),
			Bucket:       viper.GetString("storage-bucket"),
			AccessKey:    viper.GetString("storage-access-key"),
			SecretKey:    viper.GetString("storage-secret-key"),
			Region:       viper.GetString("storage-region"),
			UsePathStyle: viper.GetBool("storage-use-path-style"),
		},
	}
}
