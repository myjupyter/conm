package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/config"
)

const sshAskpassScript = "#!/bin/sh\ncat \"$(dirname \"$0\")/password\"\n"

const sshKeyMaterialPrefix = "-----BEGIN "

func sshCommand(cfg config.Connection, password string) (command, error) {
	s, ok := cfg.(config.SSH)
	if !ok {
		return command{}, fmt.Errorf("%w: %s cannot run a %s connection", ErrConnType, SSH, cfg.ConnType())
	}

	var (
		args    []string
		cleanup func()
	)
	if s.PortNumber != 0 {
		args = append(args, "-p", strconv.Itoa(s.PortNumber))
	}
	switch s.Auth {
	case config.SSHAuthKey:
		identity, remove, err := sshIdentity(password)
		if err != nil {
			return command{}, err
		}
		args = append(args, "-i", identity)
		cleanup = remove
	case config.SSHAuthPassword:
		args = append(args,
			"-o", "PreferredAuthentications=password",
			"-o", "PubkeyAuthentication=no",
			"-o", "StrictHostKeyChecking=accept-new",
		)
	}
	switch {
	case s.Jump != "" && sshUsesAskpass(s.Auth, password):
		args = append(args, "-o", "ProxyCommand="+sshJumpWithoutAskpass(s.Jump))
	case s.Jump != "":
		args = append(args, "-J", s.Jump)
	}
	if s.ForwardAgent {
		args = append(args, "-A")
	}
	if s.KeepAlive > 0 {
		args = append(args, "-o", "ServerAliveInterval="+strconv.Itoa(s.KeepAlive))
	}
	if s.LocalForward != "" {
		args = append(args, "-L", s.LocalForward)
	}

	args = append(args, s.User+"@"+s.Hostname)

	var env []string
	if sshUsesAskpass(s.Auth, password) {
		askpass, remove, err := sshAskpass(password)
		if err != nil {
			return command{}, err
		}
		env = append(os.Environ(), "SSH_ASKPASS="+askpass, "SSH_ASKPASS_REQUIRE=force")
		cleanup = remove
	}

	return command{args: args, env: env, cleanup: cleanup}, nil
}

func sshIdentity(secret string) (path string, remove func(), err error) {
	if !strings.HasPrefix(secret, sshKeyMaterialPrefix) {
		return secret, nil, nil
	}

	f, err := os.CreateTemp("", "conm-identity-*")
	if err != nil {
		return "", nil, err
	}
	remove = func() { _ = os.Remove(f.Name()) }

	if _, err := f.WriteString(strings.TrimRight(secret, "\n") + "\n"); err != nil {
		f.Close()
		remove()
		return "", nil, err
	}
	if err := f.Close(); err != nil {
		remove()
		return "", nil, err
	}

	return f.Name(), remove, nil
}

func sshUsesAskpass(auth config.SSHAuth, password string) bool {
	return auth == config.SSHAuthPassword && password != ""
}

func sshJumpWithoutAskpass(jump string) string {
	hops := strings.Split(jump, ",")
	for i, hop := range hops {
		hops[i] = strings.TrimSpace(hop)
	}
	last := len(hops) - 1

	proxy := []string{"env", "-u", "SSH_ASKPASS", "-u", "SSH_ASKPASS_REQUIRE", "ssh"}
	if last > 0 {
		proxy = append(proxy, "-J", strings.Join(hops[:last], ","))
	}
	proxy = append(proxy, "-W", "%h:%p", "ssh://"+hops[last])
	return strings.Join(proxy, " ")
}

func sshAskpass(password string) (path string, remove func(), err error) {
	dir, err := os.MkdirTemp("", "conm-askpass-*")
	if err != nil {
		return "", nil, err
	}
	remove = func() { _ = os.RemoveAll(dir) }

	if err := os.WriteFile(filepath.Join(dir, "password"), []byte(password), 0o600); err != nil {
		remove()
		return "", nil, err
	}
	path = filepath.Join(dir, "askpass")
	if err := os.WriteFile(path, []byte(sshAskpassScript), 0o600); err != nil {
		remove()
		return "", nil, err
	}
	if err := os.Chmod(path, 0o700); err != nil {
		remove()
		return "", nil, err
	}

	return path, remove, nil
}
