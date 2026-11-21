package importer

import (
	"encoding/base64"
	"errors"
	"fmt"
	"slices"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/parser/params"
)

const (
	mssqlDefaultDatabase = "master"
	mssqlNoEncryption    = "none"
)

var errNoEndpoint = errors.New("context points at no endpoint")

func readMSSQL(path string) ([]Entry, error) {
	contexts, err := config.ReadSQLConfig(path)
	if err != nil {
		return nil, err
	}

	entries := make([]Entry, 0, len(contexts))
	for _, context := range contexts {
		entries = append(entries, mssqlEntry(context))
	}

	return entries, nil
}

func mssqlEntry(context config.ForeignSection) Entry {
	entry := Entry{Name: context.Name}

	args, warnings, err := mssqlArgs(values(context.Params))
	if err != nil {
		entry.Err = err
		return entry
	}

	res, err := params.ParseArgs(config.MSSQLConnType, args)
	if err != nil {
		entry.Err = err
		return entry
	}

	ms, ok := res.Conn.(config.MSSQL)
	if !ok {
		entry.Err = fmt.Errorf("context was read as a %T", res.Conn)
		return entry
	}

	if ms.DBName == "" {
		ms.DBName = mssqlDefaultDatabase
		warnings = append(warnings, "no database in a context, kept "+mssqlDefaultDatabase)
	}
	ms.Metadata.Name = context.Name

	entry.Conn, entry.Password = withoutPassword(ms, config.MSSQLConnType, context.Name)
	entry.Warnings = slices.Concat(warnings, res.Warnings)

	return entry
}

func mssqlArgs(context map[string]string) (args, warnings []string, err error) {
	server := context[config.SQLConfigAddress]
	if server == "" {
		return nil, nil, errNoEndpoint
	}
	if port := context[config.SQLConfigPort]; port != "" {
		server += "," + port
	}
	args = append(args, "--server="+server)

	if user := context[config.SQLConfigUsername]; user != "" {
		args = append(args, "--user-name="+user)
	}

	stored := context[config.SQLConfigPassword]
	if stored == "" {
		return args, nil, nil
	}

	if method := context[config.SQLConfigPasswordEncryption]; method != "" && method != mssqlNoEncryption {
		return args, []string{fmt.Sprintf("password is encrypted with %s and was left out", method)}, nil
	}

	password, ok := decoded(stored)
	if !ok {
		return args, []string{"password is not base64 and was left out"}, nil
	}

	return append(args, "--password="+password), nil, nil
}

func decoded(stored string) (string, bool) {
	password, err := base64.StdEncoding.DecodeString(stored)

	return string(password), err == nil
}
