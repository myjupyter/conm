package params

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

const sshDefaultPort = 22

const sshScheme = "ssh"

type sshKey = string

const (
	sshHost         sshKey = "host"
	sshPort         sshKey = "port"
	sshUser         sshKey = "user"
	sshIdentity     sshKey = "identity"
	sshJump         sshKey = "jump"
	sshForwardAgent sshKey = "forward_agent"
	sshKeepAlive    sshKey = "keepalive"
	sshLocalForward sshKey = "local_forward"
	sshOption       sshKey = "option"
)

var sshFlags = map[string]sshKey{
	"-p": sshPort,
	"-l": sshUser,
	"-i": sshIdentity,
	"-J": sshJump,
	"-A": sshForwardAgent,
	"-L": sshLocalForward,
	"-o": sshOption,
}

var sshValueFlags = map[string]bool{
	"-B": true, "-b": true, "-c": true, "-D": true, "-E": true, "-e": true, "-F": true,
	"-I": true, "-m": true, "-O": true, "-P": true, "-Q": true, "-R": true, "-S": true,
	"-W": true, "-w": true,
}

var sshOptions = map[string]sshKey{
	"user":                sshUser,
	"port":                sshPort,
	"hostname":            sshHost,
	"identityfile":        sshIdentity,
	"proxyjump":           sshJump,
	"forwardagent":        sshForwardAgent,
	"serveraliveinterval": sshKeepAlive,
	"localforward":        sshLocalForward,
}

type sshValues map[sshKey]string

func parseSSH(req request) (Result, error) {
	values := sshValues{}
	res := Result{Syntax: FlagSyntax, Client: req.client}

	flags, positional := scanFlags(req.args, sshArity)
	for _, flag := range flags {
		res.Warnings = append(res.Warnings, applySSHFlag(values, flag)...)
	}

	if len(positional) > 0 {
		destination := positional[0]
		if err := foreignScheme(req.kind, destination); err != nil {
			return Result{}, err
		}
		if strings.Contains(destination, "://") {
			res.Syntax = URISyntax
		}
		applySSHDestination(values, destination)
		if len(positional) > 1 {
			res.Warnings = append(res.Warnings, fmt.Sprintf("dropped remote command %q", strings.Join(positional[1:], " ")))
		}
	}

	conn, warnings := buildSSH(values)
	res.Warnings = append(res.Warnings, warnings...)
	res.Conn = conn

	return res, nil
}

func sshArity(name string) flagArity {
	if key, ok := sshFlags[name]; ok && key != sshForwardAgent {
		return valueFlag
	}
	if sshValueFlags[name] {
		return valueFlag
	}

	return noValueFlag
}

func applySSHFlag(values sshValues, flag argFlag) []string {
	key, ok := sshFlags[flag.name]
	switch {
	case !ok:
		return []string{"ignored flag " + flag.name}
	case key == sshForwardAgent:
		values.set(sshForwardAgent, "yes")
	case key == sshOption:
		name, value, _ := strings.Cut(flag.value, "=")
		if optionKey, known := sshOptions[strings.ToLower(name)]; known {
			values.set(optionKey, value)
			return nil
		}
		return []string{"ignored option " + flag.value}
	default:
		values.set(key, flag.value)
	}

	return nil
}

func applySSHDestination(values sshValues, destination string) {
	_, rest, isURI := strings.Cut(destination, "://")
	if !isURI {
		rest = destination
	}
	rest, _, _ = strings.Cut(rest, "/")

	if i := strings.LastIndex(rest, "@"); i >= 0 {
		values.set(sshUser, unescape(rest[:i]))
		rest = rest[i+1:]
	}

	host, port := splitHostPort(rest)
	if !isURI && port != "" {
		host, port = rest, ""
	}
	values.set(sshHost, unescape(host))
	values.set(sshPort, port)
}

func buildSSH(values sshValues) (config.SSH, []string) {
	var warnings []string

	port := sshDefaultPort
	if raw, ok := values[sshPort]; ok {
		number, err := strconv.Atoi(raw)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("port %q is not a number, kept %d", raw, port))
		} else {
			port = number
		}
	}

	keepAlive := 0
	if raw, ok := values[sshKeepAlive]; ok {
		number, err := strconv.Atoi(raw)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("keepalive %q is not a number, dropped", raw))
		} else {
			keepAlive = number
		}
	}

	auth, password := config.SSHAuthAgent, ""
	if identity := values[sshIdentity]; identity != "" {
		auth, password = config.SSHAuthKey, secret.Ref(secret.Filepath, identity)
	}

	return config.SSH{
		Hostname:     values[sshHost],
		PortNumber:   port,
		User:         values[sshUser],
		Auth:         auth,
		Password:     password,
		Jump:         values[sshJump],
		ForwardAgent: strings.EqualFold(values[sshForwardAgent], "yes"),
		KeepAlive:    keepAlive,
		LocalForward: values[sshLocalForward],
	}, warnings
}

func (v sshValues) set(key sshKey, value string) {
	if value == "" {
		return
	}
	v[key] = value
}
