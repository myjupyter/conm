package importer

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/adrg/xdg"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/parser/params"
)

const (
	mysqlClientGroup = "client"
	mysqlToolGroup   = "mysql"
	mysqlHostOption  = "host"
)

var mysqlSharedGroups = []string{"client-server", "client-mariadb"}

const mysqlLoosePrefix = "loose-"

var mysqlValueUnescaper = strings.NewReplacer(
	`\\`, `\`, `\"`, `"`, `\'`, `'`, `\n`, "\n", `\r`, "\r", `\t`, "\t", `\b`, "\b", `\s`, " ",
)

func readMySQL(path string) ([]Entry, error) {
	groups, err := ReadINI(path)
	if err != nil {
		return nil, err
	}

	options := func(names ...string) []Param {
		var found []Param
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

	if len(base) > 0 && (len(entries) == 0 || Values(base)[mysqlHostOption] != "") {
		entries = slices.Insert(entries, 0, mysqlEntry(mysqlClientGroup, base))
	}

	return entries, nil
}

func mysqlEntry(name string, options []Param) Entry {
	entry := Entry{Name: name}

	res, err := params.ParseOptions(config.MySQLConnType, mysqlArgs(options))
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

	entry.Conn, entry.Warnings = my, res.Warnings

	return entry
}

func mysqlArgs(options []Param) []string {
	var (
		args  []string
		index = map[string]int{}
	)

	for _, option := range options {
		if strings.HasPrefix(option.Key, "!") {
			continue
		}

		name := "--" + strings.TrimPrefix(strings.ReplaceAll(option.Key, "_", "-"), mysqlLoosePrefix)
		arg := name + "=" + mysqlValueUnescaper.Replace(unquoted(option.Value))

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

const mysqlOptionFileName = ".my.cnf"

func MySQLOptionPath() string {
	return filepath.Join(xdg.Home, mysqlOptionFileName)
}
