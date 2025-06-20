package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	port string
)

// server represents the base command when called without any subcommands
var server = &cobra.Command{
	Use:   "hyperboard-api",
	Short: "Hyperboard API server",
	Run:   func(cmd *cobra.Command, args []string) { startServer() },
}

func init() {
	server.PersistentFlags().StringVarP(&port, "port", "p", "8080", "Port to run the server on")
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
	// Register handlers
	http.HandleFunc("/healthz", healthzHandler)
	http.HandleFunc("/images/", imagesHandler)
	http.HandleFunc("/images", imagesHandler)
	http.HandleFunc("/tags/", tagsHandler)
	http.HandleFunc("/tags", tagsHandler)
	http.HandleFunc("/tagCategories/", tagCategoriesHandler)
	http.HandleFunc("/tagCategories", tagCategoriesHandler)

	// Start server
	log.Printf("Starting API server on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func main() {
	if err := server.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
