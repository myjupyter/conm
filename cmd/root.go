/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/conn"
	"github.com/myjupyter/conm/internal/ui"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "conm",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		conmConfigPath, err := config.ConmConfigPath()
		if err != nil {
			return err
		}

		conmCfg, err := config.ReadConm(conmConfigPath)
		if err != nil {
			return err
		}

		confs, err := config.ReadPG()
		if err != nil {
			return err
		}

		conns := make([]ui.Connection, 0, len(confs))
		for i := range confs {
			conn, err := conn.NewPGClient(conmCfg, confs[i])
			if err != nil {
				return err
			}

			conns = append(conns, conn)
		}

		return ui.Run(conns)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.conm.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
