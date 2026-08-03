/*
Copyright © 2026 Tkachuk Kirill <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui"
	"github.com/spf13/cobra"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		const defaultEditor = "vim"
		editorCmd := os.Getenv("EDITOR")
		if editorCmd == "" {
			editorCmd = defaultEditor
		}

		var configPath string
		switch {
		case cmd.Flags().Changed("postgres"):
			cp, err := config.PGFilePath()
			if err != nil {
				return err
			}
			configPath = cp
		default:
			cp, err := config.ConmFilePath()
			if err != nil {
				return err
			}
			configPath = cp
		}

		shellCmd := exec.
			CommandContext(
				cmd.Context(),
				editorCmd,
				configPath,
			)

		shellCmd.Stdin = os.Stdin
		shellCmd.Stdout = os.Stdout
		shellCmd.Stderr = os.Stderr

		if err := shellCmd.Run(); err != nil {
			return err
		}

		return nil
	},
}

var editPostgresCmd = &cobra.Command{
	Use:   "postgres",
	Short: "Edit an existing postgres connection",
	RunE: func(cmd *cobra.Command, args []string) error {
		postgresConfigPath, err := config.PGFilePath()
		if err != nil {
			return fmt.Errorf("failed to get postgres config path: %w", err)
		}

		if _, err := os.Stat(postgresConfigPath); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("postgres config file not found\nrun 'conm init' first\n")
			}
			return err
		}

		c, err := config.OpenConfig[*config.PostgresConfigWrapper](postgresConfigPath)
		if err != nil {
			return fmt.Errorf("failed to open postgres config: %w", err)
		}

		defer c.Close()

		if c.Len() == 0 {
			return fmt.Errorf("no postgres connections to edit\nrun 'conm add postgres' first\n")
		}

		conns := make([]config.ConnectionConfig, c.Len())
		for i := 0; i < c.Len(); i++ {
			conns[i] = c.Get(i)
		}

		idx, ok, err := ui.RunSelectConnForm("Choose a connection to edit", conns)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}

		cfg, ok, err := ui.RunEditForm(config.PostgresConnType, c.Get(idx))
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}

		pg, ok := cfg.(config.Postgres)
		if !ok {
			return fmt.Errorf("unexpected connection type %T for postgres edit form", cfg)
		}

		if errs := pg.Validate(); len(errs) > 0 {
			return fmt.Errorf("connection is invalid: %w", errors.Join(errs...))
		}

		c.Put(idx, pg)

		if err := c.Save(); err != nil {
			return fmt.Errorf("failed to save postgres config: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.AddCommand(editPostgresCmd)

	editCmd.Flags().Bool("postgres", false, "Edit postgres.toml")
}
