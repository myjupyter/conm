/*
Copyright © 2026 Kirill Tkachuk <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/myjupyter/conm/internal/config"
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

		var confs []config.Postgres
		if ok, err := cmd.Flags().GetBool("pgpass"); err != nil {
			return err
		} else if ok {
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
			// TODO: come up with idea how to pass postgres cli
			PostgresCli: "pgcli",
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
