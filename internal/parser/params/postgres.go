package params

import (
	"fmt"
	"maps"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

const postgresDefaultPort = 5432

type postgresKey = string

const (
	pgHost     postgresKey = "host"
	pgPort     postgresKey = "port"
	pgUser     postgresKey = "user"
	pgPassword postgresKey = "password"
	pgDatabase postgresKey = "dbname"
	pgSchema   postgresKey = "search_path"
	pgSSLMode  postgresKey = "sslmode"
	pgOptions  postgresKey = "options"
)

var postgresURIRegexp = regexp.MustCompile(`^(?i:postgres(?:ql)?)://`)

var searchPathRegexp = regexp.MustCompile(`(?:-c\s*|--)search_path\s*=\s*(\S+)`)

var postgresFlags = map[string]postgresKey{
	"-h":         pgHost,
	"--host":     pgHost,
	"-p":         pgPort,
	"--port":     pgPort,
	"-U":         pgUser,
	"-u":         pgUser,
	"--user":     pgUser,
	"--username": pgUser,
	"-d":         pgDatabase,
	"--dbname":   pgDatabase,
	"--database": pgDatabase,
	"--sslmode":  pgSSLMode,
}

var postgresIgnoredFlags = map[string]bool{
	"-c":                 true,
	"--command":          true,
	"-f":                 true,
	"--file":             true,
	"-F":                 true,
	"--field-separator":  true,
	"-L":                 true,
	"--log-file":         true,
	"-o":                 true,
	"--output":           true,
	"-P":                 true,
	"--pset":             true,
	"-R":                 true,
	"--record-separator": true,
	"-T":                 true,
	"--table-attr":       true,
	"-v":                 true,
	"--set":              true,
	"--variable":         true,
	"-D":                 true,
	"--dsn":              true,
	"--row-limit":        true,
	"--prompt":           true,
	"--prompt-dsn":       true,
	"--ssh-tunnel":       true,
	"--log-level":        true,
}

var postgresEnv = map[string]postgresKey{
	"PGHOST":     pgHost,
	"PGPORT":     pgPort,
	"PGUSER":     pgUser,
	"PGPASSWORD": pgPassword,
	"PGDATABASE": pgDatabase,
	"PGSSLMODE":  pgSSLMode,
	"PGOPTIONS":  pgOptions,
}

var postgresKeywords = map[string]postgresKey{
	"host":     pgHost,
	"port":     pgPort,
	"user":     pgUser,
	"password": pgPassword,
	"dbname":   pgDatabase,
	"database": pgDatabase,
	"sslmode":  pgSSLMode,
	"options":  pgOptions,
}

type postgresValues map[postgresKey]string

func ParsePostgres(input string) (Result, error) {
	req, err := newRequest(input)
	if err != nil {
		return Result{}, err
	}

	return parsePostgres(req)
}

func parsePostgres(req request) (Result, error) {
	values := postgresValues{}
	for name, value := range req.env {
		if key, ok := postgresEnv[name]; ok {
			values.set(key, value)
		}
	}

	flags, positional := scanFlags(req.args, postgresValued)
	keywords, operands := splitPositional(positional)

	conninfo, syntax, flags, operands := postgresConninfo(flags, operands, keywords)

	res := Result{Syntax: FlagSyntax, Client: req.client}
	if syntax != UnknownSyntax {
		res.Syntax = syntax
	}

	if len(operands) > 0 {
		values.set(pgDatabase, operands[0])
	}
	if len(operands) > 1 {
		values.set(pgUser, operands[1])
	}

	switch syntax {
	case URISyntax:
		res.Warnings = append(res.Warnings, applyPostgresURI(values, conninfo)...)
	case KeywordSyntax:
		warnings, err := applyPostgresKeywords(values, conninfo)
		if err != nil {
			return Result{}, err
		}
		res.Warnings = append(res.Warnings, warnings...)
	case UnknownSyntax, FlagSyntax:
	}

	res.Warnings = append(res.Warnings, applyPostgresFlags(values, flags)...)

	conn, warnings := buildPostgres(values)
	res.Warnings = append(res.Warnings, warnings...)
	res.Conn = conn

	return res, nil
}

func postgresValued(name string) bool {
	if _, ok := postgresFlags[name]; ok {
		return true
	}

	return postgresIgnoredFlags[name]
}

func splitPositional(positional []string) (keywords, operands []string) {
	for _, arg := range positional {
		if keywordRegexp.MatchString(arg) {
			keywords = append(keywords, arg)
			continue
		}
		operands = append(operands, arg)
	}

	return keywords, operands
}

func postgresConninfo(
	flags []argFlag,
	operands []string,
	keywords []string,
) (string, Syntax, []argFlag, []string) {
	var (
		conninfo string
		syntax   = UnknownSyntax
		kept     []argFlag
	)

	for _, flag := range flags {
		if syntax == UnknownSyntax && postgresFlags[flag.name] == pgDatabase {
			if s := postgresSyntax(flag.value); s != UnknownSyntax {
				conninfo, syntax = flag.value, s
				continue
			}
		}
		kept = append(kept, flag)
	}

	if syntax == UnknownSyntax && len(operands) > 0 {
		if s := postgresSyntax(operands[0]); s != UnknownSyntax {
			conninfo, syntax = operands[0], s
			operands = operands[1:]
		}
	}

	if syntax == UnknownSyntax && len(keywords) > 0 {
		conninfo, syntax = strings.Join(keywords, " "), KeywordSyntax
	}

	return conninfo, syntax, kept, operands
}

func postgresSyntax(value string) Syntax {
	switch {
	case postgresURIRegexp.MatchString(value):
		return URISyntax
	case keywordRegexp.MatchString(value):
		return KeywordSyntax
	default:
		return UnknownSyntax
	}
}

func hasPostgresKeywords(args []string) bool {
	for _, arg := range args {
		if !keywordRegexp.MatchString(arg) {
			continue
		}
		key, _, _ := strings.Cut(arg, "=")
		if _, ok := postgresKeywords[strings.ToLower(strings.TrimSpace(key))]; ok {
			return true
		}
	}

	return false
}

func applyPostgresFlags(values postgresValues, flags []argFlag) []string {
	var warnings []string

	for _, flag := range flags {
		if flag.name == "-D" || flag.name == "--dsn" {
			warnings = append(warnings, fmt.Sprintf("named dsn %q can't be resolved from here", flag.value))
			continue
		}
		if key, ok := postgresFlags[flag.name]; ok {
			values.set(key, flag.value)
		}
	}

	return warnings
}

func applyPostgresURI(values postgresValues, raw string) []string {
	var warnings []string

	_, rest, _ := strings.Cut(raw, "://")

	authority := rest
	tail := ""
	if i := strings.IndexAny(rest, "/?#"); i >= 0 {
		authority, tail = rest[:i], rest[i:]
	}

	if i := strings.LastIndex(authority, "@"); i >= 0 {
		user, password, hasPassword := strings.Cut(authority[:i], ":")
		values.set(pgUser, unescape(user))
		if hasPassword {
			values.set(pgPassword, unescape(password))
		}
		authority = authority[i+1:]
	}

	if hosts := splitHosts(authority); len(hosts) > 0 {
		host, port := splitHostPort(hosts[0])
		values.set(pgHost, unescape(host))
		values.set(pgPort, port)
		if len(hosts) > 1 {
			warnings = append(warnings, fmt.Sprintf("kept only the first of %d hosts: %s", len(hosts), hosts[0]))
		}
	}

	path, rawQuery, _ := strings.Cut(tail, "?")
	rawQuery, _, _ = strings.Cut(rawQuery, "#")
	values.set(pgDatabase, unescape(strings.TrimPrefix(path, "/")))

	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		return append(warnings, fmt.Sprintf("ignored unreadable parameters %q", rawQuery))
	}

	for _, key := range slices.Sorted(maps.Keys(query)) {
		value := query.Get(key)
		switch key {
		case pgHost:
			warnings = append(warnings, setPostgresHostList(values, value)...)
		case pgPort, pgUser, pgPassword, pgDatabase, pgSSLMode, pgOptions, pgSchema:
			values.set(key, value)
		default:
			warnings = append(warnings, fmt.Sprintf("ignored parameter %s=%s", key, value))
		}
	}

	return warnings
}

func applyPostgresKeywords(values postgresValues, raw string) ([]string, error) {
	pairs, err := splitKeywords(raw)
	if err != nil {
		return nil, err
	}

	var warnings []string
	for _, pair := range pairs {
		key, ok := postgresKeywords[pair.key]
		if !ok {
			warnings = append(warnings, fmt.Sprintf("ignored parameter %s=%s", pair.key, pair.value))
			continue
		}

		switch key {
		case pgHost:
			warnings = append(warnings, setPostgresHostList(values, pair.value)...)
		case pgPort:
			warnings = append(warnings, setPostgresPortList(values, pair.value)...)
		default:
			values.set(key, pair.value)
		}
	}

	return warnings, nil
}

func setPostgresHostList(values postgresValues, raw string) []string {
	hosts := splitHosts(raw)
	if len(hosts) == 0 {
		return nil
	}

	host, port := splitHostPort(hosts[0])
	values.set(pgHost, unescape(host))
	values.set(pgPort, port)

	if len(hosts) == 1 {
		return nil
	}

	return []string{fmt.Sprintf("kept only the first of %d hosts: %s", len(hosts), hosts[0])}
}

func setPostgresPortList(values postgresValues, raw string) []string {
	ports := strings.Split(raw, ",")
	values.set(pgPort, ports[0])

	if len(ports) == 1 {
		return nil
	}

	return []string{fmt.Sprintf("kept only the first of %d ports: %s", len(ports), ports[0])}
}

func buildPostgres(values postgresValues) (config.Postgres, []string) {
	var warnings []string

	port := postgresDefaultPort
	if raw, ok := values[pgPort]; ok {
		number, err := strconv.Atoi(raw)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("port %q is not a number, kept %d", raw, port))
		} else {
			port = number
		}
	}

	sslMode := values[pgSSLMode]
	if err := config.ValidatePostgresSSLMode(sslMode); err != nil {
		warnings = append(warnings, fmt.Sprintf("%s, kept %s", err, config.PostgresSSLModePrefer))
		sslMode = ""
	}
	if sslMode == "" {
		sslMode = config.PostgresSSLModePrefer
	}

	schema := values[pgSchema]
	if schema == "" {
		schema = searchPath(values[pgOptions])
	}

	conn := config.Postgres{
		Hostname:   values[pgHost],
		PortNumber: port,
		User:       values[pgUser],
		DBName:     values[pgDatabase],
		SchemaName: schema,
		SSLMode:    sslMode,
	}

	if password, ok := values[pgPassword]; ok {
		conn.Password = secret.Ref(secret.Literal, password)
	}

	return conn, warnings
}

func searchPath(options string) string {
	match := searchPathRegexp.FindStringSubmatch(options)
	if len(match) < 2 {
		return ""
	}

	first, _, _ := strings.Cut(match[1], ",")

	return strings.Trim(first, `"'`)
}

func (v postgresValues) set(key postgresKey, value string) {
	if value == "" {
		return
	}
	v[key] = value
}
