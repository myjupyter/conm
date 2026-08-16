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
		if _, err := os.Stat(config.PostgresPath()); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("postgres config file not found\nrun 'conm init' first\n")
			}
		}

		c, err := config.OpenConfig[*config.ConmConfigWrapper](config.ConmPath())
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
