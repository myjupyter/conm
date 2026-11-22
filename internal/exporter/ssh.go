package exporter

import (
	"slices"
	"strings"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/importer"
	"github.com/myjupyter/conm/internal/secret"
)

func sshSections(conn config.Connection, _ string) ([]importer.Section, []string) {
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
	case ssh.Auth == config.SSHAuthPassword && ssh.Password != "":
		warnings = append(warnings, "the password is not exported: an ssh config holds none")
	}

	host.add("ProxyJump", config.NormalizeSSHJump(ssh.Jump))
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
	sshHostLine  = "Host"
	sshMatchLine = "Match"
	sshAnyHost   = "*"
	sshIndent    = "    "
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
	header := func(alias string) string { return sshHostLine + " " + alias }

	at, global := sshFirstOption(raw)
	if at < 0 {
		return appended(raw, sections, header, sshIndent, " ")
	}

	added, err := textBlocks(sections, header, sshIndent, " ")
	if err != nil {
		return nil, err
	}

	added += "\n"
	if global {
		added += sshHostLine + " " + sshAnyHost + "\n"
	}

	return slices.Concat(raw[:at], []byte(added), raw[at:]), nil
}

func sshFirstOption(raw []byte) (at int, global bool) {
	for line := range strings.Lines(string(raw)) {
		keyword, _ := config.SplitSSHOption(line)
		if keyword != "" && !strings.HasPrefix(keyword, "#") {
			return at, !strings.EqualFold(keyword, sshHostLine) && !strings.EqualFold(keyword, sshMatchLine)
		}
		at += len(line)
	}

	return -1, false
}
