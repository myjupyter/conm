/*
Copyright © 2026 Kirill Tkachuk <EMAIL ADDRESS>
*/
package cmd

import (
	"strings"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Inits conm",
	Long:  `Inits conm. It will create a conm directory, conm.toml and postgres.toml files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cliValue, err := cmd.Flags().GetString("cli")
		if err != nil {
			return err
		}

		usePgPassValue, err := cmd.Flags().GetBool("pgpass")
		if err != nil {
			return err
		}

		return ui.RunInitScreen(ui.InitScreenConfig{
			CLIFlagValue:         cliValue,
			UsePGPassImportValue: usePgPassValue,
		})
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().Bool("pgpass", false, "Import from pgpass")
	initCmd.Flags().String("cli", "", "Postgres client to use; skips interactive selection (one of: "+strings.Join(config.KnownPGClients, ", ")+")")
}
