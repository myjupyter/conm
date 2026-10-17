package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui"
)

func connTypeCmd(t config.ConnType) *cobra.Command {
	cmd := &cobra.Command{
		Use:          t.String(),
		Short:        fmt.Sprintf("Open the %s connections", t),
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			conm, err := readConm()
			if err != nil {
				return err
			}

			if settings, ok := conm.Connection(t); !ok || !settings.Enabled {
				return fmt.Errorf("%q is not enabled\nrun 'conm add' to enable it", t)
			}

			return ui.RunOn(conm, t)
		},
	}
	cmd.AddCommand(addConnectionCmd(t))

	return cmd
}

func init() {
	for _, t := range config.ConnTypes {
		rootCmd.AddCommand(connTypeCmd(t))
	}
}
