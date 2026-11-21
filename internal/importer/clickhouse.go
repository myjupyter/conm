package importer

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/parser/params"
)

const (
	clickhouseHostKey     = "hostname"
	clickhouseSecureKey   = "secure"
	clickhouseInsecureKey = "accept-invalid-certificate"
	clickhouseDefault     = "default"
)

var clickhouseFlags = map[string]string{
	clickhouseHostKey: "--host",
	"port":            "--port",
	"user":            "--user",
	"password":        "--password",
	"database":        "--database",
}

var clickhouseSwitches = map[string]string{
	clickhouseSecureKey:   "--secure",
	clickhouseInsecureKey: "--accept-invalid-certificate",
}

func readClickHouse(path string) ([]Entry, error) {
	connections, err := config.ReadClickHouseConnections(path)
	if err != nil {
		return nil, err
	}

	entries := make([]Entry, 0, len(connections))
	for _, connection := range connections {
		entries = append(entries, clickhouseEntry(connection))
	}

	return entries, nil
}

func clickhouseEntry(connection config.ForeignSection) Entry {
	entry := Entry{Name: connection.Name}

	args, warnings := clickhouseArgs(connection)

	res, err := params.ParseOptions(config.ClickHouseConnType, args)
	if err != nil {
		entry.Err = err
		return entry
	}

	ch, ok := res.Conn.(config.ClickHouse)
	if !ok {
		entry.Err = fmt.Errorf("connection was read as a %T", res.Conn)
		return entry
	}

	if ch.User == "" {
		ch.User = clickhouseDefault
		warnings = append(warnings, "no user, kept the client's own: "+clickhouseDefault)
	}
	if ch.DBName == "" {
		ch.DBName = clickhouseDefault
		warnings = append(warnings, "no database, kept the client's own: "+clickhouseDefault)
	}
	ch.Metadata.Name = connection.Name

	entry.Conn, entry.Warnings = ch, slices.Concat(warnings, res.Warnings)

	return entry
}

func clickhouseArgs(connection config.ForeignSection) (args, warnings []string) {
	if _, named := values(connection.Params)[clickhouseHostKey]; !named {
		args = append(args, clickhouseFlags[clickhouseHostKey]+"="+connection.Name)
	}

	for _, param := range connection.Params {
		flag, isSwitch := clickhouseSwitches[param.Key]
		switch {
		case isSwitch:
			if on, _ := strconv.ParseBool(param.Value); on {
				args = append(args, flag)
			}
		case clickhouseFlags[param.Key] != "":
			args = append(args, clickhouseFlags[param.Key]+"="+param.Value)
		default:
			warnings = append(warnings, fmt.Sprintf("ignored parameter %s=%s", param.Key, param.Value))
		}
	}

	return args, warnings
}
