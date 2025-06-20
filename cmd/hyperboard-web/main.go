package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var (
	port string
)

// server represents the base command when called without any subcommands
var server = &cobra.Command{
	Use:   "hyperboard-web",
	Short: "Hyperboard web server",
	Run:   func(cmd *cobra.Command, args []string) { startServer() },
}

func init() {
	server.PersistentFlags().StringVarP(&port, "port", "p", "8080", "Port to run the server on")
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello  world"))
}

func startServer() {
	// Register handlers
	http.HandleFunc("/healthz", healthzHandler)
	http.HandleFunc("/", rootHandler)

	// Start server
	log.Printf("Starting hypermedia server on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func main() {
	if err := server.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
