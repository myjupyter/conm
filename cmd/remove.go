package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui"
)

func removeConnectionCmd(t config.ConnType) *cobra.Command {
	return &cobra.Command{
		Use:          "remove <name>",
		Short:        fmt.Sprintf("Remove a %s connection", t),
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			conm, err := readConm()
			if err != nil {
				return err
			}

			if err := ui.RemoveConnection(conm, t, args[0]); err != nil {
				return err
			}

			cmd.Printf("removed %s connection %q\n", t, args[0])

			return nil
		},
	}
}
