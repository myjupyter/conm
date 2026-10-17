package cmd

import (
	"fmt"
	"net"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui"
)

func listConnectionsCmd(t config.ConnType) *cobra.Command {
	return &cobra.Command{
		Use:          "list",
		Short:        fmt.Sprintf("List the %s connections", t),
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			conm, err := readConm()
			if err != nil {
				return err
			}

			conns, err := ui.Connections(conm, t)
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			for _, c := range conns {
				fmt.Fprintf(w, "%s\t%s\t%s\n", c.Meta().Name, address(c), database(c))
			}

			return w.Flush()
		},
	}
}

func address(c config.Connection) string {
	hostPort := net.JoinHostPort(c.Host(), strconv.Itoa(c.Port()))
	if c.Username() == "" {
		return hostPort
	}

	return c.Username() + "@" + hostPort
}

func database(c config.Connection) string {
	if db, ok := c.(config.DBConnection); ok {
		return db.Database()
	}

	return ""
}
