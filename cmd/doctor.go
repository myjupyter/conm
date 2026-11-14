package cmd

import (
	"github.com/spf13/cobra"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:          "doctor",
		Short:        "Check the installed clients and the state of the config files",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			conm, err := config.ReadConm(config.ConmPath())

			return ui.Doctor(cmd.Context(), cmd.OutOrStdout(), cmd.Root().Version, conm, err)
		},
	})
}
