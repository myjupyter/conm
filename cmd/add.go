package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/parser/params"
	"github.com/myjupyter/conm/internal/ui"
	"github.com/myjupyter/conm/internal/ui/spec"
)

var addCmd = &cobra.Command{
	Use:   "add [database] [command line]",
	Short: "Add a database, or a connection to one",
	Long: `Add a database, or a connection to one.

Without arguments it opens the setup table, where a database is enabled or
disabled, and enter leaves for the connections of the database under the cursor.
With a database it opens an empty connection form, and with a client command
line after it the form opens filled in from what that command line describes; a
saved connection leaves for the connections of that database:

  conm add
  conm add postgres
  conm add postgres "psql -h localhost -p 5432 -U me mydb"
  conm add postgres "postgres://me@localhost:5432/mydb?sslmode=require"
  conm add postgres -- pgcli --host localhost --dbname mydb --user me`,
	Args:         cobra.ArbitraryArgs,
	ValidArgs:    config.DatabaseNames(),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return ui.RunInit()
		}

		t, err := config.ParseConnType(args[0])
		if err != nil {
			return fmt.Errorf("%w\nknown databases: %s", err, strings.Join(config.DatabaseNames(), ", "))
		}

		initial, warnings, err := parseConnection(t, args[1:])
		if err != nil {
			return err
		}

		conm, err := readConm()
		if err != nil {
			return err
		}

		saved, err := ui.RunAddForm(conm, t, initial, warnings)
		if err != nil || !saved {
			return err
		}

		return ui.RunOn(conm, t)
	},
}

func readConm() (config.Conm, error) {
	c, err := config.OpenConfig[*config.ConmConfigWrapper](config.ConmPath())
	if err != nil {
		return config.Conm{}, fmt.Errorf("failed to open conm config: %w", err)
	}
	defer c.Close()

	return c.Get(0), nil
}

func parseConnection(
	t config.ConnType,
	args []string,
) (map[spec.FormFieldKey]spec.FormFieldValue, []string, error) {
	if len(args) == 0 {
		return nil, nil, nil
	}

	var (
		res params.Result
		err error
	)
	if len(args) == 1 {
		res, err = params.Parse(t, args[0])
	} else {
		res, err = params.ParseArgs(t, args)
	}
	if err != nil {
		return nil, nil, err
	}

	formSpec, ok := spec.FormSpecs[t]
	if !ok || formSpec.SeedFunc == nil {
		return nil, nil, fmt.Errorf("no add form to fill in for %q", t)
	}

	initial := formSpec.SeedFunc(res.Conn)
	if initial == nil {
		return nil, nil, fmt.Errorf("cannot fill a %q form in from a %T", t, res.Conn)
	}

	return initial, res.Warnings, nil
}

func init() {
	rootCmd.AddCommand(addCmd)
}
