package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/exporter"
	"github.com/myjupyter/conm/internal/ui"
)

func exportConnectionsCmd(t config.ConnType) *cobra.Command {
	var opts ui.ExportOptions

	cmd := &cobra.Command{
		Use:   "export [name...]",
		Short: fmt.Sprintf("Export %s connections to the client's own config file", t),
		Long: fmt.Sprintf(`Export %[1]s connections to the client's own config file.

Without names every connection is exported, otherwise only the named ones. The
entries are added to the file, which is created when it is not there; nothing
in it is replaced. When a connection cannot be exported - the file already has
an entry of its name - nothing is: --skip-errors exports the rest. Passwords
stay in conm unless --with-passwords writes them to the file in plain text.

  conm %[1]s export
  conm %[1]s export orders billing
  conm %[1]s export --show
  conm %[1]s export --path ./for-a-teammate --with-passwords`, t),
		Args:         cobra.ArbitraryArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Path == "" {
				opts.Path = exporter.DefaultPath(t)
			}

			conm, err := config.ReadConm(config.ConmPath())
			if err != nil {
				return err
			}

			results, err := ui.ExportConnections(cmd.Context(), conm, t, args, opts)
			ui.RenderImport(cmd.OutOrStdout(), results)

			failed := 0
			for _, res := range results {
				if res.Status == ui.ImportFailed {
					failed++
				}
			}
			if err == nil && failed > 0 && !opts.Show && !opts.SkipErrors {
				err = fmt.Errorf("nothing was exported to %s: %d of %d can't be · --skip-errors exports the rest", opts.Path, failed, len(results))
			}

			return err
		},
	}
	cmd.Flags().StringVar(&opts.Path, "path", "", "config file to export to (default: the client's own)")
	cmd.Flags().BoolVar(&opts.SkipErrors, "skip-errors", false, "export the connections that can be and report the rest")
	cmd.Flags().BoolVar(&opts.Show, "show", false, "print what would be exported and change nothing")
	cmd.Flags().BoolVar(&opts.WithPasswords, "with-passwords", false, "write the passwords to the file in plain text")
	cmd.MarkFlagsMutuallyExclusive("skip-errors", "show")

	return cmd
}
