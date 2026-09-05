package params

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

const (
	redisDefaultPort     = 6379
	redisDefaultDatabase = 0
)

const redisTLSScheme = "rediss"

type redisKey = string

const (
	rdHost     redisKey = "host"
	rdPort     redisKey = "port"
	rdUser     redisKey = "user"
	rdPassword redisKey = "password"
	rdDatabase redisKey = "db"
	rdTLS      redisKey = "tls"
	rdURL      redisKey = "url"
	rdSocket   redisKey = "socket"
)

var redisFlags = map[string]redisKey{
	"-h":         rdHost,
	hostFlag:     rdHost,
	"-p":         rdPort,
	portFlag:     rdPort,
	userFlag:     rdUser,
	"--username": rdUser,
	"-a":         rdPassword,
	"--pass":     rdPassword,
	passwordFlag: rdPassword,
	"-n":         rdDatabase,
	"-u":         rdURL,
	"--url":      rdURL,
	"-s":         rdSocket,
	"--socket":   rdSocket,
}

var redisIgnoredFlags = map[string]bool{
	"-r":                  true,
	"-i":                  true,
	"-t":                  true,
	"-X":                  true,
	"-d":                  true,
	"-D":                  true,
	dsnFlag:               true,
	"--pattern":           true,
	"--quoted-pattern":    true,
	"--count":             true,
	"--rdb":               true,
	"--functions-rdb":     true,
	"--pipe-timeout":      true,
	"--eval":              true,
	"--lru-test":          true,
	"--memkeys-samples":   true,
	"--keystats-samples":  true,
	"--intrinsic-latency": true,
	"--cluster":           true,
	"--decode":            true,
	"--client-name":       true,
	"--iredisrc":          true,
	"--pager":             true,
	"--prompt":            true,
}

var redisTLSRefusals = map[string]bool{
	"--cacert":     true,
	"--cacertdir":  true,
	"--cert":       true,
	"--key":        true,
	"--sni":        true,
	"--verify-ssl": true,
}

var redisEnv = map[string]redisKey{
	"REDISCLI_AUTH":  rdPassword,
	"VALKEYCLI_AUTH": rdPassword,
}

type redisValues map[redisKey]string

func parseRedis(req request) (Result, error) {
	values := redisValues{}
	for name, value := range req.env {
		if key, ok := redisEnv[name]; ok {
			values.set(key, value)
		}
	}

	flags, positional := scanFlags(req.args, redisArity)
	uri, syntax, flags := redisConninfo(req.client, flags, positional)

	if err := foreignScheme(req.kind, uri); err != nil {
		return Result{}, err
	}

	res := Result{Syntax: FlagSyntax, Client: req.client}
	if syntax != UnknownSyntax {
		res.Syntax = syntax
		res.Warnings = append(res.Warnings, applyRedisURI(values, uri)...)
	}

	flagWarnings, err := applyRedisFlags(values, req.client, flags)
	if err != nil {
		return Result{}, err
	}
	res.Warnings = append(res.Warnings, flagWarnings...)

	conn, warnings := buildRedis(values)
	res.Warnings = append(res.Warnings, warnings...)
	res.Conn = conn

	return res, nil
}

func redisArity(name string) flagArity {
	if key, ok := redisFlags[name]; ok {
		if key == rdTLS {
			return noValueFlag
		}

		return valueFlag
	}
	if redisIgnoredFlags[name] || redisTLSRefusals[name] {
		return valueFlag
	}

	return noValueFlag
}

func redisConninfo(client string, flags []argFlag, positional []string) (string, Syntax, []argFlag) {
	var (
		uri    string
		syntax = UnknownSyntax
		kept   []argFlag
	)

	for _, flag := range flags {
		if redisFlags[flag.name] == rdURL {
			if syntax == UnknownSyntax && flag.value != "" {
				uri, syntax = flag.value, URISyntax
			}
			continue
		}
		kept = append(kept, flag)
	}

	if syntax == UnknownSyntax && client == "" && len(positional) > 0 && strings.Contains(positional[0], "://") {
		uri, syntax = positional[0], URISyntax
	}

	return uri, syntax, kept
}

func applyRedisFlags(values redisValues, client string, flags []argFlag) ([]string, error) {
	var warnings []string

	for _, flag := range flags {
		switch {
		case redisTLSRefusals[flag.name]:
			return nil, tlsRefusal(flag.name)
		case flag.name == "--tls":
			values.set(rdTLS, config.RedisTLSModeRequire)
		case flag.name == dsnFlag, flag.name == "-d" && client == cli.IRedis:
			warnings = append(warnings, namedDSNWarning(flag.value))
		case redisFlags[flag.name] == rdSocket:
			warnings = append(warnings, socketWarning(flag.value))
		default:
			if key, ok := redisFlags[flag.name]; ok {
				values.set(key, flag.value)
			}
		}
	}

	return warnings, nil
}

func applyRedisURI(values redisValues, raw string) []string {
	var warnings []string

	scheme, rest, _ := strings.Cut(raw, "://")
	if strings.EqualFold(scheme, redisTLSScheme) {
		values.set(rdTLS, config.RedisTLSModeRequire)
	}

	authority := rest
	tail := ""
	if i := strings.IndexAny(rest, "/?#"); i >= 0 {
		authority, tail = rest[:i], rest[i:]
	}

	if i := strings.LastIndex(authority, "@"); i >= 0 {
		user, password, hasPassword := strings.Cut(authority[:i], ":")
		values.set(rdUser, unescape(user))
		if hasPassword {
			values.set(rdPassword, unescape(password))
		}
		authority = authority[i+1:]
	}

	if hosts := splitHosts(authority); len(hosts) > 0 {
		host, port := splitHostPort(hosts[0])
		values.set(rdHost, unescape(host))
		values.set(rdPort, port)
		if len(hosts) > 1 {
			warnings = append(warnings, hostListWarning(hosts))
		}
	}

	path, rawQuery, _ := strings.Cut(tail, "?")
	rawQuery, _, _ = strings.Cut(rawQuery, "#")
	values.set(rdDatabase, unescape(strings.TrimPrefix(path, "/")))

	query, queryWarnings := queryParams(rawQuery)
	warnings = append(warnings, queryWarnings...)

	for _, key := range slices.Sorted(maps.Keys(query)) {
		value := query.Get(key)
		if key == rdDatabase {
			values.set(rdDatabase, value)
			continue
		}
		warnings = append(warnings, fmt.Sprintf("ignored parameter %s=%s", key, value))
	}

	return warnings
}

func buildRedis(values redisValues) (config.Redis, []string) {
	var warnings []string

	port := redisDefaultPort
	if raw, ok := values[rdPort]; ok {
		number, err := strconv.Atoi(raw)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("port %q is not a number, kept %d", raw, port))
		} else {
			port = number
		}
	}

	database := redisDefaultDatabase
	if raw, ok := values[rdDatabase]; ok {
		number, err := strconv.Atoi(raw)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("database %q is not a number, kept %d", raw, database))
		} else if err := config.ValidateRedisDatabase(number); err != nil {
			warnings = append(warnings, fmt.Sprintf("%s, kept %d", err, database))
		} else {
			database = number
		}
	}

	tlsMode := values[rdTLS]
	if tlsMode == "" {
		tlsMode = config.RedisTLSModeDisable
	}

	conn := config.Redis{
		Hostname:   values[rdHost],
		PortNumber: port,
		User:       values[rdUser],
		DBIndex:    database,
		TLSMode:    tlsMode,
	}

	if password, ok := values[rdPassword]; ok {
		conn.Password = secret.Ref(secret.Literal, password)
	}

	return conn, warnings
}

func (v redisValues) set(key redisKey, value string) {
	if value == "" {
		return
	}
	v[key] = value
}
