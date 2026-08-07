package cmd

import (
	"strings"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Inits conm",
	Long:  `Inits conm. It will create a conm directory, conm.toml and postgres.toml files.`,
	RunE:  initPostgresCmd.RunE,
}

var initPostgresCmd = &cobra.Command{
	Use:   "postgres",
	Short: "Inits postgres config",
	Long:  `Inits postgres config. It will create a postgres.toml file.`,
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
			ConnType:     config.PostgresConnType,
			CLIFlagValue: cliValue,
			PostgresConfig: &ui.InitPostgresScreenConfig{
				UsePGPassImportValue: usePgPassValue,
			},
		})
	},
}

func init() {
	initCmd.AddCommand(
		initPostgresCmd,
	)

	rootCmd.AddCommand(initCmd)

	pgClientsHelp := "Postgres client to use; skips interactive selection (one of: " + strings.Join(config.PGClients, ", ") + ")"

	initCmd.Flags().Bool("pgpass", false, "Import from pgpass")
	initCmd.Flags().String("cli", "", pgClientsHelp)

	initPostgresCmd.Flags().Bool("pgpass", false, "Import from pgpass")
	initPostgresCmd.Flags().String("cli", "", pgClientsHelp)
}
