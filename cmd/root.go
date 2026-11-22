package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime/debug"

	"github.com/spf13/cobra"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui"
)

func NewRootCmd() *cobra.Command {
	var workspaceName string

	root := &cobra.Command{
		Use:           "conm",
		Short:         "Terminal connection manager",
		Version:       version(),
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if workspaceName == "" {
				name, err := config.StartupWorkspace()
				if err != nil {
					return err
				}
				return config.UseWorkspace(name)
			}

			exists, err := config.WorkspaceExists(workspaceName)
			if err != nil {
				return err
			}
			if !exists {
				return fmt.Errorf("no workspace named %q · 'conm workspace add %s' creates it", workspaceName, workspaceName)
			}

			return config.UseWorkspace(workspaceName)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			conm, err := config.ReadConm(config.ConmPath())
			if err != nil {
				return err
			}

			return ui.Run(conm)
		},
	}
	root.PersistentFlags().StringVarP(&workspaceName, "workspace", "w", "", "workspace to open instead of the one set by 'conm workspace use'")

	for _, t := range config.ConnTypes {
		root.AddCommand(connTypeCmd(t))
	}
	root.AddCommand(workspaceCmd(), doctorCmd())

	return root
}

func version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	return info.Main.Version
}

func Execute() {
	rootCmd := NewRootCmd()

	err := rootCmd.Execute()
	if exitErr := (*exec.ExitError)(nil); errors.As(err, &exitErr) {
		os.Exit(exitErr.ExitCode())
	}
	if errors.Is(err, ui.ErrDoctorFound) {
		os.Exit(1)
	}
	if err != nil {
		rootCmd.PrintErrln(rootCmd.ErrPrefix(), err)
		os.Exit(1)
	}
}
