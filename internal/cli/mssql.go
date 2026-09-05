package cli

import (
	"fmt"
	"strconv"

	"github.com/myjupyter/conm/internal/config"
)

func mssqlCommand(cfg config.Connection, password string) (command, error) {
	ms, ok := cfg.(config.MSSQL)
	if !ok {
		return command{}, fmt.Errorf("%w: %s cannot run a %s connection", ErrConnType, SQLCmd, cfg.ConnType())
	}

	args := []string{"-S", mssqlServer(ms), "-U", ms.User, "-d", ms.DBName}
	if ms.EncryptMode == config.MSSQLEncryptRequire || ms.EncryptMode == config.MSSQLEncryptStrict {
		args = append(args, "-N")
	}
	if ms.TrustCert {
		args = append(args, "-C")
	}

	return command{args: args, env: passwordEnv("SQLCMDPASSWORD", password)}, nil
}

func mssqlServer(ms config.MSSQL) string {
	if ms.PortNumber == 0 {
		return "tcp:" + ms.Hostname
	}

	return "tcp:" + ms.Hostname + "," + strconv.Itoa(ms.PortNumber)
}
