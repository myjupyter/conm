package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui"
)

func workspaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "workspace",
		Aliases:      []string{"w"},
		Short:        "Open the workspaces, or manage them with a subcommand",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ui.RunWorkspaces()
		},
	}
	cmd.AddCommand(addWorkspaceCmd(), useWorkspaceCmd())

	return cmd
}

func addWorkspaceCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "add [name]",
		Short:        "Create a workspace; without a name it opens the form, then the workspaces",
		Args:         cobra.MaximumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				saved, err := ui.RunAddWorkspaceForm()
				if err != nil || !saved {
					return err
				}
				return ui.RunWorkspaces()
			}

			if err := config.CreateWorkspace(args[0]); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "added workspace %q\n", args[0])

			return nil
		},
	}
}

func useWorkspaceCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "use <name>",
		Short:        fmt.Sprintf("Open this workspace from now on; %q goes back to the default", config.DefaultWorkspace),
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.SetStartupWorkspace(args[0]); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "conm now opens workspace %q\n", args[0])

			return nil
		},
	}
}

func init() {
	rootCmd.AddCommand(workspaceCmd())
}
