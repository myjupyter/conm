package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type SSHAuth = string

const (
	SSHAuthAgent    SSHAuth = "agent"
	SSHAuthKey      SSHAuth = "key"
	SSHAuthPassword SSHAuth = "password"
)

const sshScheme = "ssh"

const (
	sshNoneSecretScheme     = "none"
	sshFilepathSecretScheme = "filepath"
)

var sshJumpHopRegexp = regexp.MustCompile(`^([A-Za-z0-9._-]+@)?[A-Za-z0-9._-]+(:\d+)?$`)

var sshLocalForwardRegexp = regexp.MustCompile(`^\d+:[^:\s]+:\d+$`)

var _ Connection = (*SSH)(nil)

type SSH struct {
	Metadata     ConnMeta `toml:"meta"`
	Hostname     string   `toml:"host"`
	PortNumber   int      `toml:"port"`
	User         string   `toml:"username"`
	Auth         SSHAuth  `toml:"auth"`
	Password     string   `toml:"password,omitempty"`
	Jump         string   `toml:"jump,omitempty"`
	ForwardAgent bool     `toml:"forward_agent,omitempty"`
	KeepAlive    int      `toml:"keepalive,omitempty"`
	LocalForward string   `toml:"local_forward,omitempty"`

	validationErrs []error
}

var _ ConfigWrapper[Connection] = (*SSHConfigWrapper)(nil)

type SSHConfigWrapper struct {
	Conns []SSH `toml:"ssh"`
}

func ValidateSSHHost(host string) error {
	if host == "" {
		return errors.New("hostname can't be empty")
	}
	if len(host) < hostMinLength || len(host) > hostMaxLength {
		return fmt.Errorf("hostname must be between %d and %d characters long", hostMinLength, hostMaxLength)
	}
	if net.ParseIP(host) != nil {
		return nil
	}
	if !simpleHostRegexp.MatchString(host) {
		return errors.New("hostname is not a valid host or IP address")
	}
	return nil
}

func ValidateSSHPort(port int) error {
	if port < portMinValue || port > portMaxValue {
		return fmt.Errorf("port must be between %d and %d", portMinValue, portMaxValue)
	}
	return nil
}

func ValidateSSHUsername(user string) error {
	if n := len(user); n < usernameMinLength || n > usernameMaxLength {
		return fmt.Errorf("username must be between %d and %d characters long", usernameMinLength, usernameMaxLength)
	}
	if !simpleHostRegexp.MatchString(user) {
		return errors.New("username contains invalid characters")
	}
	return nil
}

func ValidateSSHAuth(auth string) error {
	switch auth {
	case SSHAuthAgent, SSHAuthKey, SSHAuthPassword:
		return nil
	}
	return fmt.Errorf("invalid auth: %s", auth)
}

func ValidateSSHIdentity(identity string) error {
	if !strings.HasPrefix(identity, "~") && !strings.HasPrefix(identity, "/") {
		return errors.New("identity must be a path — ~/.ssh/… or /…")
	}
	return nil
}

func ValidateSSHCredential(auth SSHAuth, ref string) error {
	scheme, location, _ := strings.Cut(ref, ":")
	switch auth {
	case SSHAuthKey:
		if ref == "" || scheme == sshNoneSecretScheme {
			return errors.New("key auth needs the key: a filepath, or its material in a store")
		}
		if scheme == sshFilepathSecretScheme {
			return ValidateSSHIdentity(location)
		}
	case SSHAuthAgent:
		if ref != "" && scheme != sshNoneSecretScheme {
			return errors.New("agent auth takes no secret")
		}
	case SSHAuthPassword:
		if scheme == sshFilepathSecretScheme {
			return errors.New("password auth takes a password, not a file")
		}
	}
	return nil
}

func ValidateSSHJump(jump string) error {
	if jump == "" {
		return nil
	}
	for hop := range strings.SplitSeq(jump, ",") {
		if !sshJumpHopRegexp.MatchString(strings.TrimSpace(hop)) {
			return fmt.Errorf("jump hop %q must be [user@]host[:port]", hop)
		}
	}
	return nil
}

func ValidateSSHKeepAlive(seconds int) error {
	if seconds < 0 {
		return errors.New("keepalive can't be negative")
	}
	return nil
}

func ValidateSSHLocalForward(forward string) error {
	if forward == "" || sshLocalForwardRegexp.MatchString(forward) {
		return nil
	}
	return errors.New("local forward must be port:host:port — e.g. 5432:localhost:5432")
}

func ValidateSSHName(name string) error {
	if len(name) > nameMaxLength {
		return fmt.Errorf("name must be at most %d characters long", nameMaxLength)
	}
	return nil
}

func ValidateSSHDescription(desc string) error {
	if len(desc) > descriptionMaxLength {
		return fmt.Errorf("description must be at most %d characters long", descriptionMaxLength)
	}
	return nil
}

func (w *SSHConfigWrapper) Add(conn Connection) {
	s, ok := conn.(SSH)
	if !ok {
		return
	}

	w.Conns = append(w.Conns, s)
}

func (w *SSHConfigWrapper) Validate() {
	for i := range w.Conns {
		w.Conns[i].validationErrs = w.Conns[i].Validate()
	}
}

func (w *SSHConfigWrapper) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	return toml.Unmarshal(data, w)
}

func (w *SSHConfigWrapper) Marshal() ([]byte, error) {
	return toml.Marshal(w)
}

func (w *SSHConfigWrapper) Len() int {
	return len(w.Conns)
}

func (w *SSHConfigWrapper) Get(i int) Connection {
	return w.Conns[i]
}

func (w *SSHConfigWrapper) Put(i int, conn Connection) {
	s, ok := conn.(SSH)
	if !ok {
		return
	}

	w.Conns[i] = s
}

func (w *SSHConfigWrapper) Remove(i int) {
	if i < 0 || i >= len(w.Conns) {
		return
	}
	w.Conns = append(w.Conns[:i], w.Conns[i+1:]...)
}

func (s SSH) ConnType() ConnType {
	return SSHConnType
}

func (s SSH) Meta() ConnMeta {
	return s.Metadata
}

func (s SSH) Host() string {
	return s.Hostname
}

func (s SSH) Port() int {
	return s.PortNumber
}

func (s SSH) Username() string {
	return s.User
}

func (s SSH) SecretRef() string {
	return s.Password
}

func (s *SSH) SetSecretRef(secret string) {
	s.Password = secret
}

func (s SSH) ConnectionString(string) string {
	host := s.Hostname
	if s.PortNumber != 0 {
		host = net.JoinHostPort(s.Hostname, strconv.Itoa(s.PortNumber))
	}

	var userInfo *url.Userinfo
	if s.User != "" {
		userInfo = url.User(s.User)
	}

	u := url.URL{
		Scheme: sshScheme,
		User:   userInfo,
		Host:   host,
	}

	return u.String()
}

func (s SSH) Validate() []error {
	return validationErrors(
		ValidateSSHHost(s.Hostname),
		ValidateSSHPort(s.PortNumber),
		ValidateSSHUsername(s.User),
		ValidateSSHAuth(s.Auth),
		ValidateSSHCredential(s.Auth, s.Password),
		ValidateSSHJump(s.Jump),
		ValidateSSHKeepAlive(s.KeepAlive),
		ValidateSSHLocalForward(s.LocalForward),
		ValidateSSHName(s.Metadata.Name),
		ValidateSSHDescription(s.Metadata.Description),
	)
}

func (s SSH) IsValid() bool {
	return len(s.validationErrs) == 0
}
