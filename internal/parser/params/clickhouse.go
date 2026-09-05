package params

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

const (
	clickhouseDefaultPort       = 9000
	clickhouseDefaultSecurePort = 9440
)

const (
	clickhouseClientSubcommand = "client"
	clickhouseTLSScheme        = "clickhouses"
	clickhouseConnectionFlag   = "--connection"
	clickhouseSecureEnabled    = "true"
)

type clickhouseKey = string

const (
	chHost     clickhouseKey = "host"
	chPort     clickhouseKey = "port"
	chUser     clickhouseKey = "user"
	chPassword clickhouseKey = "password"
	chDatabase clickhouseKey = "database"
	chSecure   clickhouseKey = "secure"
)

var clickhouseFlags = map[string]clickhouseKey{
	"-h":         chHost,
	hostFlag:     chHost,
	portFlag:     chPort,
	"-u":         chUser,
	userFlag:     chUser,
	passwordFlag: chPassword,
	"-d":         chDatabase,
	databaseFlag: chDatabase,
	"-s":         chSecure,
	"--secure":   chSecure,
}

var clickhouseIgnoredFlags = map[string]bool{
	"-q":                     true,
	"--query":                true,
	"--queries-file":         true,
	"--query_id":             true,
	"--format":               true,
	"--config-file":          true,
	"--history_file":         true,
	"--log-level":            true,
	"--send_logs_level":      true,
	"--server_logs_file":     true,
	"--user_files_path":      true,
	"--stage":                true,
	"--pager":                true,
	clickhouseConnectionFlag: true,
}

var clickhouseTLSRefusals = map[string]bool{
	"--accept-invalid-certificate": true,
	"skip_verify":                  true,
}

var clickhouseEnv = map[string]clickhouseKey{
	"CLICKHOUSE_HOST":     chHost,
	"CLICKHOUSE_PORT":     chPort,
	"CLICKHOUSE_USER":     chUser,
	"CLICKHOUSE_PASSWORD": chPassword,
	"CLICKHOUSE_DATABASE": chDatabase,
}

type clickhouseValues map[clickhouseKey]string

func parseClickHouse(req request) (Result, error) {
	values := clickhouseValues{}
	for name, value := range req.env {
		if key, ok := clickhouseEnv[name]; ok {
			values.set(key, value)
		}
	}

	flags, positional := scanFlags(clickhouseArgs(req), clickhouseArity)
	uri, syntax := clickhouseConninfo(positional)

	if err := foreignScheme(req.kind, uri); err != nil {
		return Result{}, err
	}

	res := Result{Syntax: FlagSyntax, Client: req.client}
	if syntax != UnknownSyntax {
		res.Syntax = syntax

		warnings, err := applyClickHouseURI(values, uri)
		if err != nil {
			return Result{}, err
		}
		res.Warnings = append(res.Warnings, warnings...)
	}

	flagWarnings, err := applyClickHouseFlags(values, flags)
	if err != nil {
		return Result{}, err
	}
	res.Warnings = append(res.Warnings, flagWarnings...)

	conn, warnings := buildClickHouse(values)
	res.Warnings = append(res.Warnings, warnings...)
	res.Conn = conn

	return res, nil
}

func clickhouseArgs(req request) []string {
	if req.client == config.ClickHouseCLI && len(req.args) > 0 && req.args[0] == clickhouseClientSubcommand {
		return req.args[1:]
	}

	return req.args
}

func clickhouseArity(name string) flagArity {
	if key, ok := clickhouseFlags[name]; ok {
		switch key {
		case chSecure:
			return noValueFlag
		case chPassword:
			return attachedValueFlag
		default:
			return valueFlag
		}
	}
	if clickhouseIgnoredFlags[name] || clickhouseTLSRefusals[name] {
		return valueFlag
	}

	return noValueFlag
}

func clickhouseConninfo(positional []string) (string, Syntax) {
	if len(positional) > 0 && strings.Contains(positional[0], "://") {
		return positional[0], URISyntax
	}

	return "", UnknownSyntax
}

func applyClickHouseFlags(values clickhouseValues, flags []argFlag) ([]string, error) {
	var warnings []string

	for _, flag := range flags {
		switch {
		case clickhouseTLSRefusals[flag.name]:
			return nil, tlsRefusal(flag.name)
		case flag.name == clickhouseConnectionFlag:
			warnings = append(warnings, namedDSNWarning(flag.value))
		case clickhouseFlags[flag.name] == chSecure:
			values.set(chSecure, clickhouseSecureEnabled)
		default:
			if key, ok := clickhouseFlags[flag.name]; ok {
				values.set(key, flag.value)
			}
		}
	}

	return warnings, nil
}

func applyClickHouseURI(values clickhouseValues, raw string) ([]string, error) {
	var warnings []string

	scheme, rest, _ := strings.Cut(raw, "://")
	if strings.EqualFold(scheme, clickhouseTLSScheme) {
		values.set(chSecure, clickhouseSecureEnabled)
	}

	authority := rest
	tail := ""
	if i := strings.IndexAny(rest, "/?#"); i >= 0 {
		authority, tail = rest[:i], rest[i:]
	}

	if i := strings.LastIndex(authority, "@"); i >= 0 {
		user, password, hasPassword := strings.Cut(authority[:i], ":")
		values.set(chUser, unescape(user))
		if hasPassword {
			values.set(chPassword, unescape(password))
		}
		authority = authority[i+1:]
	}

	if hosts := splitHosts(authority); len(hosts) > 0 {
		host, port := splitHostPort(hosts[0])
		values.set(chHost, unescape(host))
		values.set(chPort, port)
		if len(hosts) > 1 {
			warnings = append(warnings, hostListWarning(hosts))
		}
	}

	path, rawQuery, _ := strings.Cut(tail, "?")
	rawQuery, _, _ = strings.Cut(rawQuery, "#")
	values.set(chDatabase, unescape(strings.TrimPrefix(path, "/")))

	query, queryWarnings := queryParams(rawQuery)
	warnings = append(warnings, queryWarnings...)

	for _, key := range slices.Sorted(maps.Keys(query)) {
		value := query.Get(key)
		switch {
		case clickhouseTLSRefusals[key]:
			return nil, tlsRefusal(key)
		case key == chSecure:
			values.set(chSecure, value)
		default:
			warnings = append(warnings, fmt.Sprintf("ignored parameter %s=%s", key, value))
		}
	}

	return warnings, nil
}

func buildClickHouse(values clickhouseValues) (config.ClickHouse, []string) {
	var warnings []string

	secure := false
	if raw, ok := values[chSecure]; ok {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("secure %q is not a boolean, kept %t", raw, secure))
		} else {
			secure = parsed
		}
	}

	port := clickhouseDefaultPort
	if secure {
		port = clickhouseDefaultSecurePort
	}
	if raw, ok := values[chPort]; ok {
		number, err := strconv.Atoi(raw)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("port %q is not a number, kept %d", raw, port))
		} else {
			port = number
		}
	}

	conn := config.ClickHouse{
		Hostname:   values[chHost],
		PortNumber: port,
		User:       values[chUser],
		DBName:     values[chDatabase],
		Secure:     secure,
	}

	if password, ok := values[chPassword]; ok {
		conn.Password = secret.Ref(secret.Literal, password)
	}

	return conn, warnings
}

func (v clickhouseValues) set(key clickhouseKey, value string) {
	if value == "" {
		return
	}
	v[key] = value
}
