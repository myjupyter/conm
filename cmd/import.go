package cmd

import (
	"fmt"
	"slices"

	"github.com/spf13/cobra"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/importer"
	"github.com/myjupyter/conm/internal/ui"
)

func importConnectionsCmd(t config.ConnType) *cobra.Command {
	var (
		path        string
		interactive bool
		skipErrors  bool
		show        bool
	)

	cmd := &cobra.Command{
		Use:   "import [name...]",
		Short: fmt.Sprintf("Import %s connections from the client's own config file", t),
		Long: fmt.Sprintf(`Import %[1]s connections from the client's own config file.

Without names every connection in the file is imported, otherwise only the
named ones. A connection that is already in conm is left alone. When one cannot
be imported nothing is: --skip-errors imports the rest, -i opens it in the add
form to be fixed. Passwords found in the file go to the system keyring.

  conm %[1]s import
  conm %[1]s import orders billing
  conm %[1]s import --show
  conm %[1]s import --path ./pg_service.conf --workspace staging`, t),
		Args:         cobra.ArbitraryArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := importer.Read(t, path)
			if err != nil {
				return err
			}

			entries, err = pickEntries(entries, args)
			if err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}

			conm, err := config.ReadConm(config.ConmPath())
			if err != nil {
				return err
			}

			mode := ui.ImportStrict
			switch {
			case show:
				mode = ui.ImportShow
			case interactive:
				mode = ui.ImportInteractive
			case skipErrors:
				mode = ui.ImportSkipErrors
			}

			results, err := ui.ImportConnections(cmd.Context(), conm, t, entries, mode)
			ui.RenderImport(cmd.OutOrStdout(), results)

			failed := 0
			for _, res := range results {
				if res.Status == ui.ImportFailed {
					failed++
				}
			}
			if err == nil && failed > 0 && mode == ui.ImportStrict {
				err = fmt.Errorf("nothing was imported: %d of %d can't be · --skip-errors imports the rest, -i fixes them in a form", failed, len(results))
			}

			return err
		},
	}
	cmd.Flags().StringVar(&path, "path", importer.DefaultPath(t), "config file to import from")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "open a connection that cannot be imported in the add form")
	cmd.Flags().BoolVar(&skipErrors, "skip-errors", false, "import the connections that can be and report the rest")
	cmd.Flags().BoolVar(&show, "show", false, "print what would be imported and change nothing")
	cmd.MarkFlagsMutuallyExclusive("interactive", "skip-errors", "show")

	return cmd
}

func pickEntries(entries []importer.Entry, names []string) ([]importer.Entry, error) {
	if len(names) == 0 {
		return entries, nil
	}

	picked := make([]importer.Entry, 0, len(names))
	for _, name := range names {
		i := slices.IndexFunc(entries, func(e importer.Entry) bool { return e.Name == name })
		if i < 0 {
			return nil, fmt.Errorf("no connection named %q", name)
		}
		picked = append(picked, entries[i])
	}

	return picked, nil
}
