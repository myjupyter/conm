package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui"
)

func editConnectionCmd(t config.ConnType) *cobra.Command {
	return &cobra.Command{
		Use:   "edit <name>",
		Short: fmt.Sprintf("Edit a %s connection", t),
		Long: fmt.Sprintf(`Edit a %[1]s connection.
		
It opens the edit form of the %[1]s connection with that name; a saved
connection leaves for the %[1]s connections:
		
conm %[1]s edit <name>`, t),
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			conm, err := readConm()
			if err != nil {
				return err
			}

			saved, err := ui.RunEditForm(conm, t, args[0])
			if err != nil || !saved {
				return err
			}

			return ui.RunOn(conm, t)
		},
	}
}
