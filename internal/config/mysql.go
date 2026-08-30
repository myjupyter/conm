package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/pelletier/go-toml/v2"
)

type MySQLTLSMode = string

const (
	MySQLTLSModeDisable    MySQLTLSMode = "false"
	MySQLTLSModePreferred  MySQLTLSMode = "preferred"
	MySQLTLSModeSkipVerify MySQLTLSMode = "skip-verify"
	MySQLTLSModeVerify     MySQLTLSMode = "true"
)

const mysqlScheme = "mysql"

var _ Connection = (*MySQL)(nil)

type MySQL struct {
	Meta       ConnMeta `toml:"meta"`
	Hostname   string   `toml:"host" json:"host"`
	PortNumber int      `toml:"port" json:"port"`
	User       string   `toml:"username" json:"username"`
	Password   string   `toml:"password" json:"password"`
	DBName     string   `toml:"dbname" json:"dbname"`
	TLSMode    string   `toml:"tls,omitempty" json:"tls,omitempty"`

	validationErrs []error
}

var _ ConfigWrapper[Connection] = (*MySQLConfigWrapper)(nil)

type MySQLConfigWrapper struct {
	Conns []MySQL `toml:"mysql"`
}

func ValidateMySQLHost(host string) error {
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

func ValidateMySQLPort(port int) error {
	if port < portMinValue || port > portMaxValue {
		return fmt.Errorf("port must be between %d and %d", portMinValue, portMaxValue)
	}
	return nil
}

func ValidateMySQLUsername(user string) error {
	if user == "" {
		return errors.New("username can't be empty")
	}
	if n := len(user); n < usernameMinLength || n > usernameMaxLength {
		return fmt.Errorf("username must be between %d and %d characters long", usernameMinLength, usernameMaxLength)
	}
	return nil
}

func ValidateMySQLDatabase(db string) error {
	if db == "" {
		return errors.New("database can't be empty")
	}
	if len(db) > databaseMaxLength {
		return fmt.Errorf("database must be at most %d characters long", databaseMaxLength)
	}
	if !simpleDatabaseRegexp.MatchString(db) {
		return errors.New("database contains invalid characters")
	}
	return nil
}

func ValidateMySQLTLSMode(mode string) error {
	switch mode {
	case "",
		MySQLTLSModeDisable,
		MySQLTLSModePreferred,
		MySQLTLSModeSkipVerify,
		MySQLTLSModeVerify:
		return nil
	}
	return fmt.Errorf("invalid TLS mode: %s", mode)
}

func ValidateMySQLName(name string) error {
	if len(name) > nameMaxLength {
		return fmt.Errorf("name must be at most %d characters long", nameMaxLength)
	}
	return nil
}

func ValidateMySQLDescription(desc string) error {
	if len(desc) > descriptionMaxLength {
		return fmt.Errorf("description must be at most %d characters long", descriptionMaxLength)
	}
	return nil
}

func (w *MySQLConfigWrapper) Add(conn Connection) {
	my, ok := conn.(MySQL)
	if !ok {
		return
	}

	w.Conns = append(w.Conns, my)
}

func (w *MySQLConfigWrapper) Validate() {
	for i := range w.Conns {
		w.Conns[i].validationErrs = w.Conns[i].Validate()
	}
}

func (w *MySQLConfigWrapper) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	return toml.Unmarshal(data, w)
}

func (w *MySQLConfigWrapper) Marshal() ([]byte, error) {
	return toml.Marshal(w)
}

func (w *MySQLConfigWrapper) Len() int {
	return len(w.Conns)
}

func (w *MySQLConfigWrapper) Get(i int) Connection {
	return w.Conns[i]
}

func (w *MySQLConfigWrapper) Put(i int, conn Connection) {
	my, ok := conn.(MySQL)
	if !ok {
		return
	}

	w.Conns[i] = my
}

func (w *MySQLConfigWrapper) Remove(i int) {
	if i < 0 || i >= len(w.Conns) {
		return
	}
	w.Conns = append(w.Conns[:i], w.Conns[i+1:]...)
}

func (m MySQL) ConnType() ConnType {
	return MySQLConnType
}

func (m MySQL) ConnMeta() ConnMeta {
	return m.Meta
}

func (m MySQL) Name() string {
	return m.Meta.Name
}

func (m MySQL) Description() string {
	return m.Meta.Description
}

func (m MySQL) Tags() []string {
	return m.Meta.Tags
}

func (m MySQL) Host() string {
	return m.Hostname
}

func (m MySQL) Port() int {
	return m.PortNumber
}

func (m MySQL) Database() string {
	return m.DBName
}

func (m MySQL) Schema() string {
	return ""
}

func (m MySQL) Username() string {
	return m.User
}

func (m MySQL) SecretRef() string {
	return m.Password
}

func (m *MySQL) SetSecretRef(secret string) {
	m.Password = secret
}

func (m MySQL) ConnectionString(secret string) string {
	return m.buildDSN(secret)
}

func (m MySQL) buildDSN(password string) string {
	host := m.Hostname
	if m.PortNumber != 0 {
		host = net.JoinHostPort(m.Hostname, strconv.Itoa(m.PortNumber))
	}

	userInfo := url.User(m.User)
	if password != "" {
		userInfo = url.UserPassword(m.User, password)
	}

	u := url.URL{
		Scheme: mysqlScheme,
		User:   userInfo,
		Host:   host,
		Path:   "/" + m.DBName,
	}

	if m.TLSMode != "" {
		q := url.Values{}
		q.Set("tls", m.TLSMode)
		u.RawQuery = q.Encode()
	}

	return u.String()
}

func (m MySQL) Validate() []error {
	return validationErrors(
		ValidateMySQLHost(m.Hostname),
		ValidateMySQLPort(m.PortNumber),
		ValidateMySQLUsername(m.User),
		ValidateMySQLDatabase(m.DBName),
		ValidateMySQLTLSMode(m.TLSMode),
		ValidateMySQLName(m.Meta.Name),
		ValidateMySQLDescription(m.Meta.Description),
	)
}

func (m MySQL) IsValid() bool {
	return len(m.validationErrs) == 0
}
