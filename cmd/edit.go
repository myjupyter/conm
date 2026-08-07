/*
Copyright © 2026 Tkachuk Kirill <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"os/exec"

	"github.com/myjupyter/conm/internal/config"
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

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.Flags().Bool("postgres", false, "Edit postgres.toml")
}
