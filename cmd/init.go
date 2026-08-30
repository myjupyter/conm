package cmd

import (
	"github.com/spf13/cobra"

	"github.com/myjupyter/conm/internal/ui"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Inits conm",
	Long:  `Inits conm. It will create a conm directory and pick the databases to manage.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return ui.RunInit()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
