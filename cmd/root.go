package cmd

import (
	"errors"
	"os"
	"os/exec"
	"runtime/debug"

	"github.com/spf13/cobra"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui"
)

var rootCmd = &cobra.Command{
	Use:           "conm",
	Short:         "Terminal connection manager",
	Version:       version(),
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		conm, err := config.ReadConm(config.ConmPath())
		if err != nil {
			return err
		}

		return ui.Run(conm)
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
	if exitErr := (*exec.ExitError)(nil); errors.As(err, &exitErr) {
		os.Exit(exitErr.ExitCode())
	}
	if err != nil {
		rootCmd.PrintErrln(rootCmd.ErrPrefix(), err)
		os.Exit(1)
	}
}
