package cmd

import (
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui"
)

var rootCmd = &cobra.Command{
	Use:     "conm",
	Short:   "Terminal connection manager",
	Version: version(),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.OpenConfig[*config.ConmConfigWrapper](config.ConmPath())
		if err != nil {
			return err
		}
		defer c.Close()

		return ui.Run(c.Get(0))
	},
}

func version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	return info.Main.Version
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
