package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new connection",
	RunE:  addPostgresCmd.RunE,
}

var addPostgresCmd = &cobra.Command{
	Use:   "postgres",
	Short: "Add a new postgres connection",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(config.PostgresPath()); err != nil {
			if os.IsNotExist(err) {
				return errors.New("postgres config file not found\nrun 'conm init' first")
			}
		}

		c, err := config.OpenConfig[*config.ConmConfigWrapper](config.ConmPath())
		if err != nil {
			return fmt.Errorf("failed to open conm config: %w", err)
		}
		defer c.Close()

		if _, err := ui.RunAddForm(c.Get(0), config.PostgresConnType); err != nil {
			return err
		}

		return nil
	},
}

var addMySQLCmd = &cobra.Command{
	Use:   "mysql",
	Short: "Add a new mysql connection",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(config.MySQLPath()); err != nil {
			if os.IsNotExist(err) {
				return errors.New("mysql config file not found\nrun 'conm init' first")
			}
		}

		c, err := config.OpenConfig[*config.ConmConfigWrapper](config.ConmPath())
		if err != nil {
			return fmt.Errorf("failed to open conm config: %w", err)
		}
		defer c.Close()

		if _, err := ui.RunAddForm(c.Get(0), config.MySQLConnType); err != nil {
			return err
		}

		return nil
	},
}

var addMSSQLCmd = &cobra.Command{
	Use:   "mssql",
	Short: "Add a new mssql connection",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(config.MSSQLPath()); err != nil {
			if os.IsNotExist(err) {
				return errors.New("mssql config file not found\nrun 'conm init' first")
			}
		}

		c, err := config.OpenConfig[*config.ConmConfigWrapper](config.ConmPath())
		if err != nil {
			return fmt.Errorf("failed to open conm config: %w", err)
		}
		defer c.Close()

		if _, err := ui.RunAddForm(c.Get(0), config.MSSQLConnType); err != nil {
			return err
		}

		return nil
	},
}

var addClickHouseCmd = &cobra.Command{
	Use:   "clickhouse",
	Short: "Add a new clickhouse connection",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(config.ClickHousePath()); err != nil {
			if os.IsNotExist(err) {
				return errors.New("clickhouse config file not found\nrun 'conm init' first")
			}
		}

		c, err := config.OpenConfig[*config.ConmConfigWrapper](config.ConmPath())
		if err != nil {
			return fmt.Errorf("failed to open conm config: %w", err)
		}
		defer c.Close()

		if _, err := ui.RunAddForm(c.Get(0), config.ClickHouseConnType); err != nil {
			return err
		}

		return nil
	},
}

var addRedisCmd = &cobra.Command{
	Use:   "redis",
	Short: "Add a new redis connection",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(config.RedisPath()); err != nil {
			if os.IsNotExist(err) {
				return errors.New("redis config file not found\nrun 'conm init' first")
			}
		}

		c, err := config.OpenConfig[*config.ConmConfigWrapper](config.ConmPath())
		if err != nil {
			return fmt.Errorf("failed to open conm config: %w", err)
		}
		defer c.Close()

		if _, err := ui.RunAddForm(c.Get(0), config.RedisConnType); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	addCmd.AddCommand(
		addPostgresCmd,
		addMySQLCmd,
		addMSSQLCmd,
		addClickHouseCmd,
		addRedisCmd,
	)

	rootCmd.AddCommand(addCmd)
}
