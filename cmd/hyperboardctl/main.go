package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	configPath   string
	outputFormat string
	cfg          *Config
)

var rootCmd = &cobra.Command{
	Use:   "hyperboardctl",
	Short: "Hyperboard CLI",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = loadConfig()
		return err
	},
}

var (
	getCmd = &cobra.Command{
		Use:   "get",
		Short: "Get resources",
	}
	createCmd = &cobra.Command{
		Use:   "create",
		Short: "Create resources",
	}
	editCmd = &cobra.Command{
		Use:   "edit",
		Short: "Edit resources",
	}
	deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete resources",
	}
	replaceCmd = &cobra.Command{
		Use:   "replace",
		Short: "Replace post content or thumbnail",
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to config file")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "output", "table", "Output format (table, yaml, json)")

	bindConfig(rootCmd)

	rootCmd.AddCommand(getCmd, createCmd, editCmd, deleteCmd, replaceCmd)
}

func main() {
	cobra.OnInitialize(initConfig)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initConfig() {
	if configPath != "" {
		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			// Config file is optional for hyperboardctl — env vars or flags can suffice
			os.Stderr.WriteString("Warning: failed to read config file: " + err.Error() + "\n")
		}
	}
}
