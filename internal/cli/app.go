package cli

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// App holds CLI application state including configuration, commands, and output format.
type App struct {
	Config       *Config
	OutputFormat string

	RootCmd    *cobra.Command
	GetCmd     *cobra.Command
	CreateCmd  *cobra.Command
	EditCmd    *cobra.Command
	DeleteCmd  *cobra.Command
	ReplaceCmd *cobra.Command

	configPath string
}

// NewApp creates and configures a new CLI application with all root-level commands.
func NewApp() *App {
	a := &App{}

	a.RootCmd = &cobra.Command{
		Use:   "hyperboardctl",
		Short: "Hyperboard CLI",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			a.Config, err = loadConfig()
			return err
		},
	}

	a.GetCmd = &cobra.Command{Use: "get", Short: "Get resources"}
	a.CreateCmd = &cobra.Command{Use: "create", Short: "Create resources"}
	a.EditCmd = &cobra.Command{Use: "edit", Short: "Edit resources"}
	a.DeleteCmd = &cobra.Command{Use: "delete", Short: "Delete resources"}
	a.ReplaceCmd = &cobra.Command{Use: "replace", Short: "Replace post content or thumbnail"}

	a.RootCmd.PersistentFlags().StringVar(&a.configPath, "config", "", "Path to config file")
	a.RootCmd.PersistentFlags().StringVar(&a.OutputFormat, "output", "table", "Output format (table, yaml, json)")

	bindConfig(a.RootCmd)

	a.RootCmd.AddCommand(a.GetCmd, a.CreateCmd, a.EditCmd, a.DeleteCmd, a.ReplaceCmd)

	cobra.OnInitialize(func() {
		if a.configPath != "" {
			viper.SetConfigFile(a.configPath)
			if err := viper.ReadInConfig(); err != nil {
				_, _ = os.Stderr.WriteString("Warning: failed to read config file: " + err.Error() + "\n")
			}
		}
	})

	return a
}
