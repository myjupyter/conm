/*
Copyright © 2026 Kirill Tkachuk <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		conmDirConfigPath, err := config.ConmDirPath()
		if err != nil {
			return err
		}

		if err := os.Mkdir(conmDirConfigPath, 0744); err != nil {
			if !os.IsExist(err) {
				return err
			}
		}

		clients := config.DetectPGClients()
		hasClient := false
		for _, c := range clients {
			if c.Exists {
				hasClient = true
				break
			}
		}
		if !hasClient {
			return fmt.Errorf("no known postgres client found; install one of: psql, pgcli, usql")
		}

		postgresCli, chosen, err := ui.SelectCLI(clients)
		if err != nil {
			return err
		}
		if !chosen {
			return fmt.Errorf("no postgres client selected")
		}

		pgpassFlag, err := cmd.Flags().GetBool("pgpass")
		if err != nil {
			return err
		}

		importPGPass := pgpassFlag
		if !pgpassFlag {
			doImport, pgpassChosen, err := ui.SelectPGPassImport()
			if err != nil {
				return err
			}
			if !pgpassChosen {
				return fmt.Errorf("init cancelled")
			}
			importPGPass = doImport
		}

		var confs []config.Postgres
		if importPGPass {
			cfgs, exists := config.ImportFromPGPass()
			if !exists {
				return fmt.Errorf(".pgpass file not found")
			}
			confs = cfgs
		}

		conmConfigFilePath, err := config.ConmConfigPath()
		if err != nil {
			return err
		}

		err = config.CreateConm(conmConfigFilePath, config.Conm{
			PostgresCli: postgresCli,
		})
		if err != nil {
			return err
		}

		pgConfigFilePath, err := config.PGFilePath()
		if err != nil {
			return err
		}

		return config.CreatePG(pgConfigFilePath, confs)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().Bool("pgpass", false, "Import from pgpass")
}
