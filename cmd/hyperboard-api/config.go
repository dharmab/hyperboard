package main

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	Port          string
	AdminPassword string
	LogLevel      string
	PostgreSQL    PGConfig
	Storage       S3Config
}

type PGConfig struct {
	Host     string
	User     string
	Password string
	Database string
	SSLMode  string
}

type S3Config struct {
	Endpoint     string
	Bucket       string
	AccessKey    string
	SecretKey    string
	Region       string
	UsePathStyle bool
}

func bindConfig(cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.String("port", "8080", "Port to listen on")
	flags.String("admin-password", "", "Admin password for basic auth")
	flags.String("log-level", "info", "Log level (trace, debug, info, warn, error, fatal, panic)")

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

	viper.BindPFlags(flags)
}

func loadConfig() *Config {
	return &Config{
		Port:          viper.GetString("port"),
		AdminPassword: viper.GetString("admin-password"),
		LogLevel:      viper.GetString("log-level"),
		PostgreSQL: PGConfig{
			Host:     viper.GetString("postgresql-host"),
			User:     viper.GetString("postgresql-user"),
			Password: viper.GetString("postgresql-password"),
			Database: viper.GetString("postgresql-database"),
			SSLMode:  viper.GetString("postgresql-ssl-mode"),
		},
		Storage: S3Config{
			Endpoint:     viper.GetString("storage-endpoint"),
			Bucket:       viper.GetString("storage-bucket"),
			AccessKey:    viper.GetString("storage-access-key"),
			SecretKey:    viper.GetString("storage-secret-key"),
			Region:       viper.GetString("storage-region"),
			UsePathStyle: viper.GetBool("storage-use-path-style"),
		},
	}
}
