package cmd

import (
	"os"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "conm",
	Short: "Terminal connection manager",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.OpenConfig[*config.ConmConfigWrapper](config.ConmPath())
		if err != nil {
			return err
		}
		defer c.Close()

		return ui.Run(c.Get(0))
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
