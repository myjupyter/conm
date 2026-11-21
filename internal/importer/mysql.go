package importer

import (
	"fmt"
	"slices"
	"strings"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/parser/params"
	"github.com/myjupyter/conm/internal/secret"
)

const (
	mysqlClientGroup = "client"
	mysqlToolGroup   = "mysql"
)

var mysqlSharedGroups = []string{"client-server", "client-mariadb"}

func readMySQL(path string) ([]Entry, error) {
	groups, err := config.ReadINI(path)
	if err != nil {
		return nil, err
	}

	options := func(names ...string) []config.INIParam {
		var found []config.INIParam
		for _, group := range groups {
			if slices.Contains(names, group.Name) {
				found = append(found, group.Params...)
			}
		}

		return found
	}

	base := options(mysqlClientGroup, mysqlToolGroup)

	var entries []Entry
	if len(base) > 0 {
		entries = append(entries, mysqlEntry(mysqlClientGroup, base))
	}

	var seen []string
	for _, group := range groups {
		suffix, ok := strings.CutPrefix(group.Name, mysqlClientGroup)
		if !ok || suffix == "" || slices.Contains(mysqlSharedGroups, group.Name) || slices.Contains(seen, suffix) {
			continue
		}
		seen = append(seen, suffix)

		own := options(mysqlClientGroup+suffix, mysqlToolGroup+suffix)
		entries = append(entries, mysqlEntry(strings.TrimLeft(suffix, "_-."), slices.Concat(base, own)))
	}

	return entries, nil
}

func mysqlEntry(name string, options []config.INIParam) Entry {
	entry := Entry{Name: name}

	res, err := params.ParseArgs(config.MySQLConnType, mysqlArgs(options))
	if err != nil {
		entry.Err = err
		return entry
	}

	my, ok := res.Conn.(config.MySQL)
	if !ok {
		entry.Err = fmt.Errorf("group was read as a %T", res.Conn)
		return entry
	}

	_, password, _ := secret.ParseRef(my.Password)
	my.Password = ""
	if password != "" {
		my.Password = secret.Ref(secret.Keyring, SecretLocation(config.MySQLConnType, name))
	}
	my.Metadata.Name = name

	entry.Conn, entry.Password, entry.Warnings = my, password, res.Warnings

	return entry
}

func mysqlArgs(options []config.INIParam) []string {
	var (
		args  []string
		index = map[string]int{}
	)

	for _, option := range options {
		if strings.HasPrefix(option.Key, "!") {
			continue
		}

		name := "--" + strings.ReplaceAll(option.Key, "_", "-")
		arg := name
		if option.Value != "" {
			arg += "=" + unquoted(option.Value)
		}

		if i, repeated := index[name]; repeated {
			args[i] = arg
			continue
		}
		index[name] = len(args)
		args = append(args, arg)
	}

	return args
}

func unquoted(value string) string {
	if len(value) < 2 {
		return value
	}

	if quote := value[0]; (quote == '"' || quote == '\'') && value[len(value)-1] == quote {
		return value[1 : len(value)-1]
	}

	return value
}
