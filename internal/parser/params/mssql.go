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

const mssqlDefaultPort = 1433

const (
	mssqlTCPProtocol   = "tcp"
	mssqlInstanceSplit = `\`
	mssqlPortSplit     = ","
	mssqlTrustEnabled  = "true"
	mssqlTrustParam    = "trustservercertificate"
	mssqlDatabaseParam = "database"
	mssqlEncryptParam  = "encrypt"
)

type mssqlKey = string

const (
	msServer   mssqlKey = "server"
	msHost     mssqlKey = "host"
	msPort     mssqlKey = "port"
	msUser     mssqlKey = "user"
	msPassword mssqlKey = "password"
	msDatabase mssqlKey = "dbname"
	msEncrypt  mssqlKey = "encrypt"
	msTrust    mssqlKey = "trust"
)

var mssqlFlags = map[string]mssqlKey{
	"-S":                         msServer,
	"--server":                   msServer,
	"-U":                         msUser,
	"--user-name":                msUser,
	"-P":                         msPassword,
	passwordFlag:                 msPassword,
	"-d":                         msDatabase,
	databaseFlag:                 msDatabase,
	"-N":                         msEncrypt,
	"--encrypt-connection":       msEncrypt,
	"-C":                         msTrust,
	"--trust-server-certificate": msTrust,
}

var mssqlOptionalValueFlags = map[string]bool{
	"-N": true,
	"-r": true,
	"-X": true,
}

var mssqlIgnoredFlags = map[string]bool{
	"-Q": true,
	"-q": true,
	"-i": true,
	"-o": true,
	"-l": true,
	"-t": true,
	"-h": true,
	"-s": true,
	"-w": true,
	"-a": true,
	"-V": true,
	"-m": true,
	"-K": true,
	"-y": true,
	"-Y": true,
	"-c": true,
	"-v": true,
	"-f": true,
}

var mssqlAuthFlags = map[string]string{
	"-E": "a trusted connection",
	"-G": "Entra authentication",
}

var mssqlTLSRefusals = map[string]bool{
	"certificate":           true,
	"hostnameincertificate": true,
	"tlsmin":                true,
}

var mssqlProtocols = map[string]bool{
	mssqlTCPProtocol: true,
	"np":             true,
	"lpc":            true,
	"admin":          true,
}

var mssqlEncryptModes = map[string]config.MSSQLEncryptMode{
	"s":         config.MSSQLEncryptStrict,
	"strict":    config.MSSQLEncryptStrict,
	"m":         config.MSSQLEncryptRequire,
	"mandatory": config.MSSQLEncryptRequire,
	"o":         config.MSSQLEncryptLogin,
	"optional":  config.MSSQLEncryptLogin,
}

var mssqlEnv = map[string]mssqlKey{
	"SQLCMDSERVER":   msServer,
	"SQLCMDUSER":     msUser,
	"SQLCMDPASSWORD": msPassword,
	"SQLCMDDBNAME":   msDatabase,
}

type mssqlValues map[mssqlKey]string

func parseMSSQL(req request) (Result, error) {
	values := mssqlValues{}

	var warnings []string
	for _, name := range slices.Sorted(maps.Keys(req.env)) {
		key, ok := mssqlEnv[name]
		if !ok {
			continue
		}
		if key == msServer {
			warnings = append(warnings, setMSSQLServer(values, req.env[name])...)
			continue
		}
		values.set(key, req.env[name])
	}

	flags, positional := scanFlags(req.args, mssqlArity)
	uri, syntax := mssqlConninfo(positional)

	if err := foreignScheme(req.kind, uri); err != nil {
		return Result{}, err
	}

	res := Result{Syntax: FlagSyntax, Client: req.client, Warnings: warnings}
	if syntax != UnknownSyntax {
		res.Syntax = syntax

		uriWarnings, err := applyMSSQLURI(values, uri)
		if err != nil {
			return Result{}, err
		}
		res.Warnings = append(res.Warnings, uriWarnings...)
	}

	res.Warnings = append(res.Warnings, applyMSSQLFlags(values, flags)...)

	conn, buildWarnings := buildMSSQL(values)
	res.Warnings = append(res.Warnings, buildWarnings...)
	res.Conn = conn

	return res, nil
}

func mssqlArity(name string) flagArity {
	if mssqlOptionalValueFlags[name] {
		return attachedValueFlag
	}
	if key, ok := mssqlFlags[name]; ok {
		if key == msTrust {
			return noValueFlag
		}

		return valueFlag
	}
	if mssqlIgnoredFlags[name] {
		return valueFlag
	}

	return noValueFlag
}

func mssqlConninfo(positional []string) (string, Syntax) {
	if len(positional) > 0 && strings.Contains(positional[0], "://") {
		return positional[0], URISyntax
	}

	return "", UnknownSyntax
}

func applyMSSQLFlags(values mssqlValues, flags []argFlag) []string {
	var warnings []string

	for _, flag := range flags {
		switch {
		case mssqlAuthFlags[flag.name] != "":
			warnings = append(warnings, mssqlAuthWarning(mssqlAuthFlags[flag.name]))
		case mssqlFlags[flag.name] == msServer:
			warnings = append(warnings, setMSSQLServer(values, flag.value)...)
		case mssqlFlags[flag.name] == msEncrypt:
			values.set(msEncrypt, mssqlEncryptMode(flag.value))
		case mssqlFlags[flag.name] == msTrust:
			values.set(msTrust, mssqlTrustEnabled)
		default:
			if key, ok := mssqlFlags[flag.name]; ok {
				values.set(key, flag.value)
			}
		}
	}

	return warnings
}

func applyMSSQLURI(values mssqlValues, raw string) ([]string, error) {
	var warnings []string

	_, rest, _ := strings.Cut(raw, "://")

	authority := rest
	tail := ""
	if i := strings.IndexAny(rest, "/?#"); i >= 0 {
		authority, tail = rest[:i], rest[i:]
	}

	if i := strings.LastIndex(authority, "@"); i >= 0 {
		user, password, hasPassword := strings.Cut(authority[:i], ":")
		values.set(msUser, unescape(user))
		if hasPassword {
			values.set(msPassword, unescape(password))
		}
		authority = authority[i+1:]
	}

	if hosts := splitHosts(authority); len(hosts) > 0 {
		host, port := splitHostPort(hosts[0])
		values.set(msHost, unescape(host))
		values.set(msPort, port)
		if len(hosts) > 1 {
			warnings = append(warnings, hostListWarning(hosts))
		}
	}

	path, rawQuery, _ := strings.Cut(tail, "?")
	rawQuery, _, _ = strings.Cut(rawQuery, "#")
	if instance := strings.TrimPrefix(path, "/"); instance != "" {
		warnings = append(warnings, instanceWarning(unescape(instance)))
	}

	query, queryWarnings := queryParams(rawQuery)
	warnings = append(warnings, queryWarnings...)

	for _, key := range slices.Sorted(maps.Keys(query)) {
		value := query.Get(key)
		switch name := strings.ToLower(key); {
		case mssqlTLSRefusals[name]:
			return nil, tlsRefusal(key)
		case name == mssqlDatabaseParam:
			values.set(msDatabase, value)
		case name == mssqlEncryptParam:
			values.set(msEncrypt, value)
		case name == mssqlTrustParam:
			values.set(msTrust, value)
		default:
			warnings = append(warnings, fmt.Sprintf("ignored parameter %s=%s", key, value))
		}
	}

	return warnings, nil
}

func setMSSQLServer(values mssqlValues, raw string) []string {
	var warnings []string

	server := raw
	if protocol, rest, ok := strings.Cut(server, ":"); ok && mssqlProtocols[strings.ToLower(protocol)] {
		if !strings.EqualFold(protocol, mssqlTCPProtocol) {
			warnings = append(warnings, fmt.Sprintf("transport %q has no field to keep it in", protocol))
		}
		server = rest
	}

	server, port, hasPort := strings.Cut(server, mssqlPortSplit)
	if hasPort {
		values.set(msPort, port)
	}

	if host, instance, hasInstance := strings.Cut(server, mssqlInstanceSplit); hasInstance {
		warnings = append(warnings, instanceWarning(instance))
		server = host
	}

	values.set(msHost, strings.Trim(server, "[]"))

	return warnings
}

func mssqlEncryptMode(value string) string {
	if value == "" {
		return config.MSSQLEncryptRequire
	}
	if mode, ok := mssqlEncryptModes[strings.ToLower(value)]; ok {
		return mode
	}

	return value
}

func mssqlAuthWarning(kind string) string {
	return kind + " keeps no username or password"
}

func instanceWarning(value string) string {
	return fmt.Sprintf("named instance %q has no field to keep it in", value)
}

func buildMSSQL(values mssqlValues) (config.MSSQL, []string) {
	var warnings []string

	port := mssqlDefaultPort
	if raw, ok := values[msPort]; ok {
		number, err := strconv.Atoi(raw)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("port %q is not a number, kept %d", raw, port))
		} else {
			port = number
		}
	}

	encrypt := values[msEncrypt]
	if err := config.ValidateMSSQLEncryptMode(encrypt); err != nil {
		warnings = append(warnings, fmt.Sprintf("%s, kept %s", err, config.MSSQLEncryptLogin))
		encrypt = ""
	}
	if encrypt == "" {
		encrypt = config.MSSQLEncryptLogin
	}

	trust := false
	if raw, ok := values[msTrust]; ok {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s %q is not a boolean, kept %t", mssqlTrustParam, raw, trust))
		} else {
			trust = parsed
		}
	}

	conn := config.MSSQL{
		Hostname:    values[msHost],
		PortNumber:  port,
		User:        values[msUser],
		DBName:      values[msDatabase],
		EncryptMode: encrypt,
		TrustCert:   trust,
	}

	if password, ok := values[msPassword]; ok {
		conn.Password = secret.Ref(secret.Literal, password)
	}

	return conn, warnings
}

func (v mssqlValues) set(key mssqlKey, value string) {
	if value == "" {
		return
	}
	v[key] = value
}
