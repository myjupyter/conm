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

const mongodbDefaultPort = 27017

const (
	mongodbSRVScheme  = "mongodb+srv"
	mongodbTLSEnabled = "true"
	mongodbSRVWarning = "mongodb+srv resolves its hosts and port from dns, kept the name and the default port"
)

type mongodbKey = string

const (
	mgHost       mongodbKey = "host"
	mgPort       mongodbKey = "port"
	mgUser       mongodbKey = "user"
	mgPassword   mongodbKey = "password"
	mgDatabase   mongodbKey = "dbname"
	mgAuthSource mongodbKey = "authsource"
	mgTLS        mongodbKey = "tls"
)

var mongodbFlags = map[string]mongodbKey{
	hostFlag:                   mgHost,
	portFlag:                   mgPort,
	"-u":                       mgUser,
	"--username":               mgUser,
	"-p":                       mgPassword,
	passwordFlag:               mgPassword,
	"--authenticationDatabase": mgAuthSource,
	"--tls":                    mgTLS,
	"--ssl":                    mgTLS,
}

var mongodbOptionalValueFlags = map[string]bool{
	"--json":        true,
	"--retryWrites": true,
}

var mongodbIgnoredFlags = map[string]bool{
	"--eval":                         true,
	"--file":                         true,
	"--authenticationMechanism":      true,
	"--apiVersion":                   true,
	"--gssapiServiceName":            true,
	"--gssapiHostName":               true,
	"--sspiHostnameCanonicalization": true,
	"--sspiRealmOverride":            true,
	"--awsAccessKeyId":               true,
	"--awsSecretAccessKey":           true,
	"--awsSessionToken":              true,
	"--awsIamSessionToken":           true,
	"--oidcRedirectUri":              true,
	"--oidcFlows":                    true,
	"--browser":                      true,
	"--jsContext":                    true,
}

var mongodbTLSRefusals = map[string]bool{
	"--tlsCAFile":                          true,
	"--tlsCertificateKeyFile":              true,
	"--tlsCertificateKeyFilePassword":      true,
	"--tlsCertificateSelector":             true,
	"--tlsCRLFile":                         true,
	"--tlsDisabledProtocols":               true,
	"--tlsAllowInvalidCertificates":        true,
	"--tlsAllowInvalidHostnames":           true,
	"--sslCAFile":                          true,
	"--sslPEMKeyFile":                      true,
	"--sslPEMKeyPassword":                  true,
	"--sslCRLFile":                         true,
	"--sslDisabledProtocols":               true,
	"--sslAllowInvalidCertificates":        true,
	"--sslAllowInvalidHostnames":           true,
	"tlscafile":                            true,
	"tlscertificatekeyfile":                true,
	"tlscertificatekeyfilepassword":        true,
	"tlscrlfile":                           true,
	"tlsinsecure":                          true,
	"tlsallowinvalidcertificates":          true,
	"tlsallowinvalidhostnames":             true,
	"tlsdisableocspendpointcheck":          true,
	"tlsdisablecertificaterevocationcheck": true,
}

var mongodbTLSParams = map[string]bool{
	"tls": true,
	"ssl": true,
}

const mongodbReplicaSetParam = "replicaset"

type mongodbValues map[mongodbKey]string

func parseMongoDB(req request) (Result, error) {
	values := mongodbValues{}

	flags, positional := scanFlags(req.args, mongodbArity(req.client))
	uri, syntax := mongodbConninfo(positional)

	if err := foreignScheme(req.kind, uri); err != nil {
		return Result{}, err
	}

	res := Result{Syntax: FlagSyntax, Client: req.client}
	switch {
	case syntax == URISyntax:
		res.Syntax = syntax

		uriWarnings, err := applyMongoDBURI(values, uri)
		if err != nil {
			return Result{}, err
		}
		res.Warnings = append(res.Warnings, uriWarnings...)
	case len(positional) > 0:
		res.Warnings = append(res.Warnings, applyMongoDBAddress(values, positional[0])...)
	}

	flagWarnings, err := applyMongoDBFlags(values, flags)
	if err != nil {
		return Result{}, err
	}
	res.Warnings = append(res.Warnings, flagWarnings...)

	conn, buildWarnings := buildMongoDB(values)
	res.Warnings = append(res.Warnings, buildWarnings...)
	res.Conn = conn

	return res, nil
}

func mongodbArity(client string) func(string) flagArity {
	return func(name string) flagArity {
		if mongodbOptionalValueFlags[name] {
			return attachedValueFlag
		}
		if key, ok := mongodbFlags[name]; ok {
			switch {
			case key == mgTLS:
				return noValueFlag
			case key == mgPassword && client == config.MongoCLI:
				return attachedValueFlag
			default:
				return valueFlag
			}
		}
		if mongodbIgnoredFlags[name] || mongodbTLSRefusals[name] {
			return valueFlag
		}

		return noValueFlag
	}
}

func mongodbConninfo(positional []string) (string, Syntax) {
	if len(positional) > 0 && strings.Contains(positional[0], "://") {
		return positional[0], URISyntax
	}

	return "", UnknownSyntax
}

func applyMongoDBAddress(values mongodbValues, address string) []string {
	hosts, database, hasDatabase := strings.Cut(address, "/")
	if !hasDatabase {
		if !strings.Contains(address, ":") {
			values.set(mgDatabase, address)

			return nil
		}
		hosts, database = address, ""
	}

	values.set(mgDatabase, database)

	return setMongoDBHosts(values, hosts)
}

func applyMongoDBFlags(values mongodbValues, flags []argFlag) ([]string, error) {
	var warnings []string

	for _, flag := range flags {
		switch {
		case mongodbTLSRefusals[flag.name]:
			return nil, tlsRefusal(flag.name)
		case mongodbFlags[flag.name] == mgHost:
			warnings = append(warnings, setMongoDBHosts(values, flag.value)...)
		case mongodbFlags[flag.name] == mgTLS:
			values.set(mgTLS, mongodbTLSEnabled)
		default:
			if key, ok := mongodbFlags[flag.name]; ok {
				values.set(key, flag.value)
			}
		}
	}

	return warnings, nil
}

func applyMongoDBURI(values mongodbValues, raw string) ([]string, error) {
	var warnings []string

	scheme, rest, _ := strings.Cut(raw, "://")
	if strings.EqualFold(scheme, mongodbSRVScheme) {
		values.set(mgTLS, mongodbTLSEnabled)
		warnings = append(warnings, mongodbSRVWarning)
	}

	authority := rest
	tail := ""
	if i := strings.IndexAny(rest, "/?#"); i >= 0 {
		authority, tail = rest[:i], rest[i:]
	}

	if i := strings.LastIndex(authority, "@"); i >= 0 {
		user, password, hasPassword := strings.Cut(authority[:i], ":")
		values.set(mgUser, unescape(user))
		if hasPassword {
			values.set(mgPassword, unescape(password))
		}
		authority = authority[i+1:]
	}

	warnings = append(warnings, setMongoDBHosts(values, authority)...)

	path, rawQuery, _ := strings.Cut(tail, "?")
	rawQuery, _, _ = strings.Cut(rawQuery, "#")
	values.set(mgDatabase, unescape(strings.TrimPrefix(path, "/")))

	query, queryWarnings := queryParams(rawQuery)
	warnings = append(warnings, queryWarnings...)

	for _, key := range slices.Sorted(maps.Keys(query)) {
		value := query.Get(key)
		switch name := strings.ToLower(key); {
		case mongodbTLSRefusals[name]:
			return nil, tlsRefusal(key)
		case name == mgAuthSource:
			values.set(mgAuthSource, value)
		case mongodbTLSParams[name]:
			values.set(mgTLS, value)
		case name == mongodbReplicaSetParam:
			warnings = append(warnings, replicaSetWarning(value))
		default:
			warnings = append(warnings, fmt.Sprintf("ignored parameter %s=%s", key, value))
		}
	}

	return warnings, nil
}

func setMongoDBHosts(values mongodbValues, raw string) []string {
	var warnings []string

	if name, rest, ok := strings.Cut(raw, "/"); ok {
		warnings = append(warnings, replicaSetWarning(name))
		raw = rest
	}

	hosts := splitHosts(raw)
	if len(hosts) == 0 {
		return warnings
	}

	host, port := splitHostPort(hosts[0])
	values.set(mgHost, unescape(host))
	values.set(mgPort, port)
	if len(hosts) > 1 {
		warnings = append(warnings, hostListWarning(hosts))
	}

	return warnings
}

func replicaSetWarning(name string) string {
	return fmt.Sprintf("replica set %q has no field to keep it in", name)
}

func buildMongoDB(values mongodbValues) (config.MongoDB, []string) {
	var warnings []string

	port := mongodbDefaultPort
	if raw, ok := values[mgPort]; ok {
		number, err := strconv.Atoi(raw)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("port %q is not a number, kept %d", raw, port))
		} else {
			port = number
		}
	}

	tls := false
	if raw, ok := values[mgTLS]; ok {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("tls %q is not a boolean, kept %t", raw, tls))
		} else {
			tls = parsed
		}
	}

	conn := config.MongoDB{
		Hostname:   values[mgHost],
		PortNumber: port,
		User:       values[mgUser],
		DBName:     values[mgDatabase],
		AuthSource: values[mgAuthSource],
		TLS:        tls,
	}

	if password, ok := values[mgPassword]; ok {
		conn.Password = secret.Ref(secret.Literal, password)
	}

	return conn, warnings
}

func (v mongodbValues) set(key mongodbKey, value string) {
	if value == "" {
		return
	}
	v[key] = value
}
