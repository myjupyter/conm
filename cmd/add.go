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

func addConnectionCmd(t config.ConnType) *cobra.Command {
	return &cobra.Command{
		Use:          "add [command line]",
		Short:        fmt.Sprintf("Add a %s connection", t),
		Long:         addConnectionLong(t),
		Args:         cobra.ArbitraryArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			initial, warnings, err := parseConnection(t, args)
			if err != nil {
				return err
			}

			conm, err := config.ReadConm(config.ConmPath())
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
}

func addConnectionLong(t config.ConnType) string {
	var b strings.Builder
	fmt.Fprintf(&b, `Add a %[1]s connection.

Without arguments it opens an empty form. Given a real %[1]s command line or
URL, it is parsed and the form opens filled in from it; quote it as one
argument, or pass it after --. A saved connection leaves for the %[1]s
connections:

  conm %[1]s add
`, t)
	for _, example := range params.Examples(t) {
		fmt.Fprintf(&b, "  conm %s add %q\n", t, example)
	}

	return strings.TrimSuffix(b.String(), "\n")
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
