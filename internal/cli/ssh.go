package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/myjupyter/conm/internal/config"
)

const sshAskpassScript = "#!/bin/sh\nprintf '%s' \"$CONM_SSH_PASSWORD\"\n"

func sshCommand(cfg config.Connection, password string) (command, error) {
	s, ok := cfg.(config.SSH)
	if !ok {
		return command{}, fmt.Errorf("%w: %s cannot run a %s connection", ErrConnType, SSH, cfg.ConnType())
	}

	var args []string
	if s.PortNumber != 0 {
		args = append(args, "-p", strconv.Itoa(s.PortNumber))
	}
	switch s.Auth {
	case config.SSHAuthKey:
		if s.Identity != "" {
			args = append(args, "-i", s.Identity)
		}
	case config.SSHAuthPassword:
		args = append(args, "-o", "PreferredAuthentications=password", "-o", "PubkeyAuthentication=no")
	}
	if s.Jump != "" {
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
	if s.RemoteCommand != "" {
		args = append(args, s.RemoteCommand)
	}

	env, err := sshPasswordEnv(s.Auth, password)
	if err != nil {
		return command{}, err
	}

	return command{args: args, env: env}, nil
}

func sshPasswordEnv(auth config.SSHAuth, password string) ([]string, error) {
	if auth != config.SSHAuthPassword || password == "" {
		return nil, nil
	}

	askpass, err := sshAskpass()
	if err != nil {
		return nil, err
	}

	return append(
		passwordEnv("CONM_SSH_PASSWORD", password),
		"SSH_ASKPASS="+askpass,
		"SSH_ASKPASS_REQUIRE=force",
	), nil
}

func sshAskpass() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir = filepath.Join(dir, "conm")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}

	path := filepath.Join(dir, "askpass.sh")
	if err := os.WriteFile(path, []byte(sshAskpassScript), 0o600); err != nil {
		return "", err
	}
	if err := os.Chmod(path, 0o700); err != nil {
		return "", err
	}

	return path, nil
}
