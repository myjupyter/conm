package cli

import (
	"fmt"
	"strconv"

	"github.com/myjupyter/conm/internal/config"
)

func mysqlCommand(cfg config.Connection, password string) (command, error) {
	my, ok := cfg.(config.MySQL)
	if !ok {
		return command{}, fmt.Errorf("%w: %s cannot run a %s connection", ErrConnType, MySQL, cfg.ConnType())
	}

	args := []string{"--protocol=TCP", "--host=" + my.Hostname, "--user=" + my.User}
	if my.PortNumber != 0 {
		args = append(args, "--port="+strconv.Itoa(my.PortNumber))
	}
	if mode := mysqlSSLMode(my.TLSMode); mode != "" {
		args = append(args, "--ssl-mode="+mode)
	}

	return command{
		args: append(args, my.DBName),
		env:  passwordEnv("MYSQL_PWD", password),
	}, nil
}

func mysqlSSLMode(tls config.MySQLTLSMode) string {
	switch tls {
	case config.MySQLTLSModeDisable:
		return "DISABLED"
	case config.MySQLTLSModePreferred:
		return "PREFERRED"
	case config.MySQLTLSModeSkipVerify:
		return "REQUIRED"
	case config.MySQLTLSModeVerify:
		return "VERIFY_CA"
	}

	return ""
}
