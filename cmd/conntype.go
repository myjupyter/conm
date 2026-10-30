package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui"
)

func connTypeCmd(t config.ConnType) *cobra.Command {
	cmd := &cobra.Command{
		Use:          t.String() + " [name]",
		Short:        fmt.Sprintf("Open the %s connections, or connect to one", t),
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			conm, err := config.ReadConm(config.ConmPath())
			if err != nil {
				return err
			}

			if len(args) == 1 {
				return ui.Connect(cmd.Context(), conm, t, args[0])
			}

			return ui.RunOn(conm, t)
		},
	}
	cmd.AddCommand(
		addConnectionCmd(t),
		editConnectionCmd(t),
		removeConnectionCmd(t),
		pingConnectionCmd(t),
		listConnectionsCmd(t),
		execConnectionCmd(t),
	)

	return cmd
}

func init() {
	for _, t := range config.ConnTypes {
		rootCmd.AddCommand(connTypeCmd(t))
	}
}
