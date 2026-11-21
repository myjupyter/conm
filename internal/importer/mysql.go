package importer

import (
	"fmt"
	"slices"
	"strings"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/parser/params"
)

const (
	mysqlClientGroup = "client"
	mysqlToolGroup   = "mysql"
	mysqlHostOption  = "host"
)

var mysqlSharedGroups = []string{"client-server", "client-mariadb"}

func readMySQL(path string) ([]Entry, error) {
	groups, err := config.ReadINI(path)
	if err != nil {
		return nil, err
	}

	options := func(names ...string) []config.ForeignParam {
		var found []config.ForeignParam
		for _, group := range groups {
			if slices.Contains(names, group.Name) {
				found = append(found, group.Params...)
			}
		}

		return found
	}

	base := options(mysqlClientGroup, mysqlToolGroup)

	var (
		entries []Entry
		seen    []string
		names   = []string{mysqlClientGroup}
	)
	for _, group := range groups {
		suffix, ok := strings.CutPrefix(group.Name, mysqlClientGroup)
		if !ok || suffix == "" || slices.Contains(mysqlSharedGroups, group.Name) || slices.Contains(seen, suffix) {
			continue
		}
		seen = append(seen, suffix)

		name := strings.TrimLeft(suffix, "_-.")
		if name == "" || slices.Contains(names, name) {
			name = suffix
		}
		names = append(names, name)

		own := options(mysqlClientGroup+suffix, mysqlToolGroup+suffix)
		entries = append(entries, mysqlEntry(name, slices.Concat(base, own)))
	}

	if len(base) > 0 && (len(entries) == 0 || values(base)[mysqlHostOption] != "") {
		entries = slices.Insert(entries, 0, mysqlEntry(mysqlClientGroup, base))
	}

	return entries, nil
}

func mysqlEntry(name string, options []config.ForeignParam) Entry {
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

	my.Metadata.Name = name

	entry.Conn, entry.Password = withoutPassword(my, config.MySQLConnType, name)
	entry.Warnings = res.Warnings

	return entry
}

func mysqlArgs(options []config.ForeignParam) []string {
	var (
		args  []string
		index = map[string]int{}
	)

	for _, option := range options {
		if strings.HasPrefix(option.Key, "!") {
			continue
		}

		name := "--" + strings.ReplaceAll(option.Key, "_", "-")
		arg := name + "=" + unquoted(option.Value)

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
