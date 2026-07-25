package ui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/myjupyter/conm/internal/config"
)

type InitScreenConfig struct {
	CLIFlagValue         string
	UsePGPassImportValue bool
}

func RunInitScreen(cfg InitScreenConfig) error {
	if err := config.CreateConmConfigPath(); err != nil {
		return err
	}

	clies := config.DetectPGCLI()

	var conmConfig config.Conm
	switch {
	case cfg.CLIFlagValue != "":
		var finalCLI *config.CLIInfo
		for _, cli := range clies {
			if cli.Name == cfg.CLIFlagValue {
				finalCLI = &cli
				break
			}
		}
		if finalCLI == nil {
			return fmt.Errorf("unknown postgres client %q; choose one of: %s", cfg.CLIFlagValue, strings.Join(config.KnownPGClients, ", "))
		}
		if !finalCLI.Exists {
			return fmt.Errorf("postgres client %q is not installed", cfg.CLIFlagValue)
		}
		conmConfig.PostgresCli = cfg.CLIFlagValue

	default:
		hasClient := false
		for _, cli := range clies {
			if cli.Exists {
				hasClient = true
				break
			}
		}
		if !hasClient {
			return fmt.Errorf("no known postgres client found; install one of: %s", strings.Join(config.KnownPGClients, ", "))
		}

		cli, chosen, err := RunSelectCLIForm(clies)
		if err != nil {
			return err
		}
		if !chosen {
			return errors.New("no postgres client selected")
		}

		conmConfig.PostgresCli = cli.Name
	}

	var confs []config.Postgres
	switch cfg.UsePGPassImportValue {
	case true:
		pgPassCreds, err := config.ImportFromPGPass()
		if err != nil {
			return fmt.Errorf("couldn't import from .pgpass: %w", err)
		}

		confs = pgPassCreds
	default:
		opt, chosen, err := RunSelectImport()
		if err != nil {
			return err
		}
		if !chosen {
			return errors.New("init cancelled")
		}

		if opt == importSelectPGPass {
			pgPassCreds, err := config.ImportFromPGPass()
			if err != nil {
				return fmt.Errorf("couldn't import from .pgpass: %w", err)
			}

			confs = pgPassCreds
		}
	}

	if err := config.CreateConm(conmConfig); err != nil {
		return err
	}

	return config.CreatePG(confs)
}
