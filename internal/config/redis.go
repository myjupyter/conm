package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/pelletier/go-toml/v2"
)

const redisDatabaseMaxIndex = 255

type RedisTLSMode = string

const (
	RedisTLSModeDisable RedisTLSMode = "disable"
	RedisTLSModeRequire RedisTLSMode = "require"
)

const (
	redisScheme    = "redis"
	redisTLSScheme = "rediss"
)

var _ Connection = (*Redis)(nil)

type Redis struct {
	Meta       ConnMeta `toml:"meta"`
	Hostname   string   `toml:"host" json:"host"`
	PortNumber int      `toml:"port" json:"port"`
	User       string   `toml:"username,omitempty" json:"username,omitempty"`
	Password   string   `toml:"password" json:"password"`
	DBIndex    int      `toml:"db" json:"db"`
	TLSMode    string   `toml:"tls,omitempty" json:"tls,omitempty"`

	validationErrs []error
}

var _ ConfigWrapper[Connection] = (*RedisConfigWrapper)(nil)

type RedisConfigWrapper struct {
	Conns []Redis `toml:"redis"`
}

func ValidateRedisHost(host string) error {
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

func ValidateRedisPort(port int) error {
	if port < portMinValue || port > portMaxValue {
		return fmt.Errorf("port must be between %d and %d", portMinValue, portMaxValue)
	}
	return nil
}

func ValidateRedisUsername(user string) error {
	if user == "" {
		return nil
	}
	if len(user) > usernameMaxLength {
		return fmt.Errorf("username must be at most %d characters long", usernameMaxLength)
	}
	if !simpleHostRegexp.MatchString(user) {
		return errors.New("username contains invalid characters")
	}
	return nil
}

func ValidateRedisDatabase(db int) error {
	if db < 0 || db > redisDatabaseMaxIndex {
		return fmt.Errorf("database must be between %d and %d", 0, redisDatabaseMaxIndex)
	}
	return nil
}

func ValidateRedisTLSMode(mode string) error {
	switch mode {
	case "", RedisTLSModeDisable, RedisTLSModeRequire:
		return nil
	}
	return fmt.Errorf("invalid TLS mode: %s", mode)
}

func ValidateRedisName(name string) error {
	if len(name) > nameMaxLength {
		return fmt.Errorf("name must be at most %d characters long", nameMaxLength)
	}
	return nil
}

func ValidateRedisDescription(desc string) error {
	if len(desc) > descriptionMaxLength {
		return fmt.Errorf("description must be at most %d characters long", descriptionMaxLength)
	}
	return nil
}

func (w *RedisConfigWrapper) Add(conn Connection) {
	rd, ok := conn.(Redis)
	if !ok {
		return
	}

	w.Conns = append(w.Conns, rd)
}

func (w *RedisConfigWrapper) Validate() {
	for i := range w.Conns {
		w.Conns[i].validationErrs = w.Conns[i].Validate()
	}
}

func (w *RedisConfigWrapper) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	return toml.Unmarshal(data, w)
}

func (w *RedisConfigWrapper) Marshal() ([]byte, error) {
	return toml.Marshal(w)
}

func (w *RedisConfigWrapper) Len() int {
	return len(w.Conns)
}

func (w *RedisConfigWrapper) Get(i int) Connection {
	return w.Conns[i]
}

func (w *RedisConfigWrapper) Put(i int, conn Connection) {
	rd, ok := conn.(Redis)
	if !ok {
		return
	}

	w.Conns[i] = rd
}

func (w *RedisConfigWrapper) Remove(i int) {
	if i < 0 || i >= len(w.Conns) {
		return
	}
	w.Conns = append(w.Conns[:i], w.Conns[i+1:]...)
}

func (r Redis) ConnType() ConnType {
	return RedisConnType
}

func (r Redis) ConnMeta() ConnMeta {
	return r.Meta
}

func (r Redis) Name() string {
	return r.Meta.Name
}

func (r Redis) Description() string {
	return r.Meta.Description
}

func (r Redis) Tags() []string {
	return r.Meta.Tags
}

func (r Redis) Host() string {
	return r.Hostname
}

func (r Redis) Port() int {
	return r.PortNumber
}

func (r Redis) Database() string {
	return strconv.Itoa(r.DBIndex)
}

func (r Redis) Schema() string {
	return ""
}

func (r Redis) Username() string {
	return r.User
}

func (r Redis) SecretRef() string {
	return r.Password
}

func (r *Redis) SetSecretRef(secret string) {
	r.Password = secret
}

func (r Redis) ConnectionString(secret string) string {
	return r.buildDSN(secret)
}

func (r Redis) scheme() string {
	if r.TLSMode == RedisTLSModeRequire {
		return redisTLSScheme
	}
	return redisScheme
}

func (r Redis) buildDSN(password string) string {
	host := r.Hostname
	if r.PortNumber != 0 {
		host = net.JoinHostPort(r.Hostname, strconv.Itoa(r.PortNumber))
	}

	var userInfo *url.Userinfo
	switch {
	case password != "":
		userInfo = url.UserPassword(r.User, password)
	case r.User != "":
		userInfo = url.User(r.User)
	}

	u := url.URL{
		Scheme: r.scheme(),
		User:   userInfo,
		Host:   host,
		Path:   "/" + strconv.Itoa(r.DBIndex),
	}

	return u.String()
}

func (r Redis) Validate() []error {
	return validationErrors(
		ValidateRedisHost(r.Hostname),
		ValidateRedisPort(r.PortNumber),
		ValidateRedisUsername(r.User),
		ValidateRedisDatabase(r.DBIndex),
		ValidateRedisTLSMode(r.TLSMode),
		ValidateRedisName(r.Meta.Name),
		ValidateRedisDescription(r.Meta.Description),
	)
}

func (r Redis) IsValid() bool {
	return len(r.validationErrs) == 0
}
