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

		cfg, ok, err := ui.RunAddForm()
		if err != nil {
			return err
		}

		// If the user didn't choose to add a new connection, we don't need to do anything
		// and just exit
		if !ok {
			return nil
		}

		c, err := config.OpenConfig[*config.PostgresConfigWrapper, config.Postgres](postgresConfigPath)
		if err != nil {
			return fmt.Errorf("failed to open postgres config: %w", err)
		}

		defer c.Close()

		c.Add(cfg)

		if err := c.Save(); err != nil {
			return fmt.Errorf("failed to save postgres config: %w", err)
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
