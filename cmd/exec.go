package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui"
)

func execConnectionCmd(t config.ConnType) *cobra.Command {
	return &cobra.Command{
		Use:   "exec <name> -- <query>",
		Short: fmt.Sprintf("Run a single query on a %s connection", t),
		Long: fmt.Sprintf(`Run a single query on a %[1]s connection.

The query goes to the client picked for %[1]s, which prints its result on
stdout and its errors on stderr; conm exits with the client's exit code.
Quote the query, or the shell expands * and ends the command at ;:

  conm %[1]s exec <name> -- "<query>"`, t),
		Args:         cobra.MinimumNArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			conm, err := config.ReadConm(config.ConmPath())
			if err != nil {
				return err
			}

			return ui.Connect(cmd.Context(), conm, t, args[0], args[1:]...)
		},
	}
}
