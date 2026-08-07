/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: addPostgresCmd.RunE,
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
