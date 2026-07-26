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
	RunE: func(cmd *cobra.Command, args []string) error {
		conmConfigPath, err := config.ConmConfigPath()
		if err != nil {
			return err
		}

		if _, err := os.Stat(conmConfigPath); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("conm config file not found\nrun 'conm init' first\n")
			}
		}

		cfg, ok, err := ui.RunAddForm()
		if err != nil {
			return err
		}

		if !ok {
			return nil
		}

		cfgs, err := config.ReadPG()
		if err != nil {
			return err
		}

		cfgs = append(cfgs, cfg)
		if err := config.CreatePG(cfgs); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
