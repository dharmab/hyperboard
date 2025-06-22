package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dharmab/hyperboard/internal/db"
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

// server represents the base command when called without any subcommands
var server = &cobra.Command{
	Use:   "hyperboard-api",
	Short: "Hyperboard API server",
	Run:   func(cmd *cobra.Command, args []string) { startServer() },
}

func init() {
	server.Flags().StringVar(&port, "port", "8080", "Port to bind for the API server")
	server.Flags().StringVar(&postgresqlSSLMode, "postgresql-ssl-mode", "disable", "PostgreSQL SSL mode")
	server.Flags().StringVar(&postgresqlHost, "postgresql-host", "localhost:5432", "PostgreSQL host")
	server.Flags().StringVar(&postgresqlUser, "postgresql-user", "postgres", "PostgreSQL user")
	server.Flags().StringVar(&postgresqlPassword, "postgresql-password", "", "PostgreSQL password")
	server.Flags().StringVar(&postgresqlDatabase, "postgresql-database", "postgres", "PostgreSQL database name")
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func imagesHandler(w http.ResponseWriter, r *http.Request) {
	// Handle /images/<id> path parameter
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) > 2 && parts[1] == "images" && parts[2] != "" {
		// This is /images/<id>
		// id := parts[2]
		panic("not implemented")
	}

	panic("not implemented")
}

func tagsHandler(w http.ResponseWriter, r *http.Request) {
	// Handle /tags/<n> path parameter
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) > 2 && parts[1] == "tags" && parts[2] != "" {
		// This is /tags/<n>
		// name := parts[2]
		panic("not implemented")
	}

	panic("not implemented")
}

func tagCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	// Handle /tagCategories/<n> path parameter
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) > 2 && parts[1] == "tagCategories" && parts[2] != "" {
		// This is /tagCategories/<n>
		// name := parts[2]
		panic("not implemented")
	}

	panic("not implemented")
}

func startServer() {
	log.Info().Msg("Running database migrations...")
	if err := migrateDatabase(); err != nil {
		log.Fatal().Err(err).Msg("Failed to run database migrations")
	}

	log.Info().Msg("Registering API routes...")
	http.HandleFunc("/healthz", healthzHandler)
	http.HandleFunc("/images", imagesHandler)
	http.HandleFunc("/tags", tagsHandler)
	http.HandleFunc("/tagCategories", tagCategoriesHandler)

	log.Info().Msg("Starting API server...")
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal().Err(err).Msg("Failed to start API server")
	}
}

func main() {
	log.Info().Msg("Running database migrations...")
	if err := server.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func migrateDatabase() error {
	url := fmt.Sprintf(
		"postgresql://%s:%s@%s/%s?sslmode=%s",
		postgresqlUser,
		postgresqlPassword,
		postgresqlHost,
		postgresqlDatabase,
		postgresqlSSLMode,
	)
	if err := db.Migrate(url); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	return nil
}
