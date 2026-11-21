package exporter

import (
	"strings"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/importer"
	"github.com/myjupyter/conm/internal/secret"
)

func sshSections(conn config.Connection, password string) ([]importer.Section, []string) {
	ssh, _ := conn.(config.SSH)

	var (
		host     params
		warnings []string
	)
	host.add("HostName", ssh.Hostname)
	host.addPort("Port", ssh.PortNumber)
	host.add("User", ssh.User)

	scheme, location, _ := secret.ParseRef(ssh.Password)
	switch {
	case ssh.Auth == config.SSHAuthKey && scheme == secret.Filepath:
		host.add("IdentityFile", quoted(location))
	case ssh.Auth == config.SSHAuthKey:
		warnings = append(warnings, "the key is not a file and is not exported: ssh would fall back to its default identities")
	case ssh.Auth == config.SSHAuthPassword && password != "":
		warnings = append(warnings, "the password is not exported: an ssh config holds none")
	}

	host.add("ProxyJump", ssh.Jump)
	if ssh.ForwardAgent {
		host.add("ForwardAgent", "yes")
	}
	host.addPort("ServerAliveInterval", ssh.KeepAlive)
	host.add("LocalForward", strings.Replace(ssh.LocalForward, ":", " ", 1))

	return []importer.Section{{Name: ssh.Metadata.Name, Params: host}}, warnings
}

func quoted(value string) string {
	if strings.ContainsAny(value, " \t") {
		return `"` + value + `"`
	}

	return value
}

const (
	sshHostLine = "Host"
	sshIndent   = "    "
)

func sshHeld(path string) ([]string, error) {
	lines, err := heldBy(importer.ReadSSHConfig)(path)

	aliases := make([]string, 0, len(lines))
	for _, line := range lines {
		aliases = append(aliases, strings.Fields(line)...)
	}

	return aliases, err
}

func sshMerged(_ string, raw []byte, sections []importer.Section) ([]byte, error) {
	return appended(raw, sections, func(alias string) string { return sshHostLine + " " + alias }, sshIndent, " ")
}
