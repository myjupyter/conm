package params

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

const mysqlDefaultPort = 3306

const mysqlLoginPathFlag = "--login-path"

type mysqlKey = string

const (
	myHost     mysqlKey = "host"
	myPort     mysqlKey = "port"
	myUser     mysqlKey = "user"
	myPassword mysqlKey = "password"
	myDatabase mysqlKey = "dbname"
	myTLS      mysqlKey = "tls"
	mySocket   mysqlKey = "socket"
)

var mysqlURIRegexp = regexp.MustCompile(`^(?i:mysql|mariadb|maria|percona|aurora|my)://`)

var mysqlFlags = map[string]mysqlKey{
	"-h":         myHost,
	"--host":     myHost,
	"-P":         myPort,
	"--port":     myPort,
	"-u":         myUser,
	"--user":     myUser,
	"-p":         myPassword,
	"--password": myPassword,
	"-D":         myDatabase,
	"--database": myDatabase,
	"--ssl-mode": myTLS,
	"-S":         mySocket,
	"--socket":   mySocket,
}

var mysqlIgnoredFlags = map[string]bool{
	"-e":                      true,
	"--execute":               true,
	"--protocol":              true,
	"--bind-address":          true,
	"--connect-timeout":       true,
	"--default-character-set": true,
	"--charset":               true,
	"--pager":                 true,
	"--prompt":                true,
	"--tee":                   true,
	"--myclirc":               true,
	"--logfile":               true,
	"--row-limit":             true,
	dsnFlag:                   true,
	mysqlLoginPathFlag:        true,
	"--defaults-file":         true,
	"--defaults-extra-file":   true,
	"--defaults-group-suffix": true,
}

var mysqlTLSRefusals = map[string]bool{
	"--ssl-ca":           true,
	"--ssl-capath":       true,
	"--ssl-cert":         true,
	"--ssl-key":          true,
	"--ssl-crl":          true,
	"--ssl-crlpath":      true,
	"--ssl-cipher":       true,
	"--tls-ciphersuites": true,
	"--tls-version":      true,
}

var mysqlDefaultsFiles = map[string]bool{
	"--defaults-file":       true,
	"--defaults-extra-file": true,
}

var mysqlEnv = map[string]mysqlKey{
	"MYSQL_HOST":     myHost,
	"MYSQL_TCP_PORT": myPort,
	"MYSQL_PWD":      myPassword,
}

var mysqlSSLModes = map[string]config.MySQLTLSMode{
	"DISABLED":        config.MySQLTLSModeDisable,
	"PREFERRED":       config.MySQLTLSModePreferred,
	"REQUIRED":        config.MySQLTLSModeSkipVerify,
	"VERIFY_CA":       config.MySQLTLSModeVerify,
	"VERIFY_IDENTITY": config.MySQLTLSModeVerify,
}

type mysqlValues map[mysqlKey]string

func parseMySQL(req request) (Result, error) {
	values := mysqlValues{}
	for name, value := range req.env {
		if key, ok := mysqlEnv[name]; ok {
			values.set(key, value)
		}
	}

	flags, positional := scanFlags(req.args, mysqlArity(req.client))
	uri, syntax, operands := mysqlConninfo(positional)

	res := Result{Syntax: FlagSyntax, Client: req.client}
	if syntax != UnknownSyntax {
		res.Syntax = syntax
	}

	if len(operands) > 0 {
		values.set(myDatabase, operands[0])
	}
	if syntax == URISyntax {
		res.Warnings = append(res.Warnings, applyMySQLURI(values, uri)...)
	}

	flagWarnings, err := applyMySQLFlags(values, flags)
	if err != nil {
		return Result{}, err
	}
	res.Warnings = append(res.Warnings, flagWarnings...)

	if err := foreignScheme(req.kind, values[myDatabase]); err != nil {
		return Result{}, err
	}

	conn, warnings := buildMySQL(values)
	res.Warnings = append(res.Warnings, warnings...)
	res.Conn = conn

	return res, nil
}

func mysqlArity(client string) func(string) flagArity {
	return func(name string) flagArity {
		if key, ok := mysqlFlags[name]; ok {
			if key == myPassword && client != config.MyCLI {
				return attachedValueFlag
			}

			return valueFlag
		}
		if mysqlIgnoredFlags[name] || mysqlTLSRefusals[name] {
			return valueFlag
		}

		return noValueFlag
	}
}

func mysqlConninfo(positional []string) (string, Syntax, []string) {
	if len(positional) > 0 && mysqlURIRegexp.MatchString(positional[0]) {
		return positional[0], URISyntax, positional[1:]
	}

	return "", UnknownSyntax, positional
}

func applyMySQLFlags(values mysqlValues, flags []argFlag) ([]string, error) {
	var warnings []string

	for _, flag := range flags {
		switch {
		case mysqlTLSRefusals[flag.name]:
			return nil, tlsRefusal(flag.name)
		case flag.name == dsnFlag, flag.name == mysqlLoginPathFlag:
			warnings = append(warnings, namedDSNWarning(flag.value))
		case mysqlDefaultsFiles[flag.name]:
			warnings = append(warnings, defaultsFileWarning(flag.value))
		case mysqlFlags[flag.name] == mySocket:
			warnings = append(warnings, socketWarning(flag.value))
		case mysqlFlags[flag.name] == myTLS:
			values.set(myTLS, mysqlTLSMode(flag.value))
		default:
			if key, ok := mysqlFlags[flag.name]; ok {
				values.set(key, flag.value)
			}
		}
	}

	return warnings, nil
}

func applyMySQLURI(values mysqlValues, raw string) []string {
	var warnings []string

	_, rest, _ := strings.Cut(raw, "://")

	authority := rest
	tail := ""
	if i := strings.IndexAny(rest, "/?#"); i >= 0 {
		authority, tail = rest[:i], rest[i:]
	}

	if i := strings.LastIndex(authority, "@"); i >= 0 {
		user, password, hasPassword := strings.Cut(authority[:i], ":")
		values.set(myUser, unescape(user))
		if hasPassword {
			values.set(myPassword, unescape(password))
		}
		authority = authority[i+1:]
	}

	if hosts := splitHosts(authority); len(hosts) > 0 {
		host, port := splitHostPort(hosts[0])
		values.set(myHost, unescape(host))
		values.set(myPort, port)
		if len(hosts) > 1 {
			warnings = append(warnings, hostListWarning(hosts))
		}
	}

	path, rawQuery, _ := strings.Cut(tail, "?")
	rawQuery, _, _ = strings.Cut(rawQuery, "#")
	values.set(myDatabase, unescape(strings.TrimPrefix(path, "/")))

	query, queryWarnings := queryParams(rawQuery)
	warnings = append(warnings, queryWarnings...)

	for _, key := range slices.Sorted(maps.Keys(query)) {
		value := query.Get(key)
		if key == myTLS {
			values.set(myTLS, value)
			continue
		}
		warnings = append(warnings, fmt.Sprintf("ignored parameter %s=%s", key, value))
	}

	return warnings
}

func defaultsFileWarning(value string) string {
	return fmt.Sprintf("defaults file %q can't be read from here", value)
}

func mysqlTLSMode(value string) string {
	if mode, ok := mysqlSSLModes[strings.ToUpper(value)]; ok {
		return mode
	}

	return value
}

func buildMySQL(values mysqlValues) (config.MySQL, []string) {
	var warnings []string

	port := mysqlDefaultPort
	if raw, ok := values[myPort]; ok {
		number, err := strconv.Atoi(raw)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("port %q is not a number, kept %d", raw, port))
		} else {
			port = number
		}
	}

	tlsMode := values[myTLS]
	if err := config.ValidateMySQLTLSMode(tlsMode); err != nil {
		warnings = append(warnings, fmt.Sprintf("%s, kept %s", err, config.MySQLTLSModePreferred))
		tlsMode = ""
	}
	if tlsMode == "" {
		tlsMode = config.MySQLTLSModePreferred
	}

	conn := config.MySQL{
		Hostname:   values[myHost],
		PortNumber: port,
		User:       values[myUser],
		DBName:     values[myDatabase],
		TLSMode:    tlsMode,
	}

	if password, ok := values[myPassword]; ok {
		conn.Password = secret.Ref(secret.Literal, password)
	}

	return conn, warnings
}

func (v mysqlValues) set(key mysqlKey, value string) {
	if value == "" {
		return
	}
	v[key] = value
}
