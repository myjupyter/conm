package cmd

import (
	"fmt"
	"os"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new connection",
	RunE:  addPostgresCmd.RunE,
}

var addPostgresCmd = &cobra.Command{
	Use:   "postgres",
	Short: "Add a new postgres connection",
	RunE: func(cmd *cobra.Command, args []string) error {
		postgresConfigPath, err := config.PGFilePath()
		if err != nil {
			return fmt.Errorf("failed to get postgres config path: %w", err)
		}

		if _, err := os.Stat(postgresConfigPath); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("postgres config file not found\nrun 'conm init' first\n")
			}
		}

		conmConfigPath, err := config.ConmFilePath()
		if err != nil {
			return fmt.Errorf("failed to get conm config path: %w", err)
		}

		c, err := config.OpenConfig[*config.ConmConfigWrapper](conmConfigPath)
		if err != nil {
			return fmt.Errorf("failed to open conm config: %w", err)
		}
		defer c.Close()

		if _, err := ui.RunAddForm(c.Get(0), config.PostgresConnType); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	addCmd.AddCommand(
		addPostgresCmd,
	)

	rootCmd.AddCommand(addCmd)
}
