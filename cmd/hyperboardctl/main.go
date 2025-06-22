package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cli = &cobra.Command{
	Use:   "hyperboardctl",
	Short: "Hyperboard CLI",
	Long:  "Command line interface for the Hyperboard media board",
	Run: func(cmd *cobra.Command, args []string) {
		// This is the default action when no subcommand is specified
		fmt.Println("Hyperboard CLI - use --help for available commands")
	},
}

func init() {
}

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
