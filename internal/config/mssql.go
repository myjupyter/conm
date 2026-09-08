package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/pelletier/go-toml/v2"
)

type MSSQLEncryptMode = string

const (
	MSSQLEncryptDisable MSSQLEncryptMode = "disable"
	MSSQLEncryptLogin   MSSQLEncryptMode = "false"
	MSSQLEncryptRequire MSSQLEncryptMode = "true"
	MSSQLEncryptStrict  MSSQLEncryptMode = "strict"
)

const mssqlScheme = "sqlserver"

var _ Connection = (*MSSQL)(nil)

type MSSQL struct {
	Metadata    ConnMeta `toml:"meta"`
	Hostname    string   `toml:"host" json:"host"`
	PortNumber  int      `toml:"port" json:"port"`
	User        string   `toml:"username" json:"username"`
	Password    string   `toml:"password" json:"password"`
	DBName      string   `toml:"dbname" json:"dbname"`
	EncryptMode string   `toml:"encrypt,omitempty" json:"encrypt,omitempty"`
	TrustCert   bool     `toml:"trust_server_certificate,omitempty" json:"trust_server_certificate,omitempty"`

	validationErrs []error
}

var _ ConfigWrapper[Connection] = (*MSSQLConfigWrapper)(nil)

type MSSQLConfigWrapper struct {
	Conns []MSSQL `toml:"mssql"`
}

func ValidateMSSQLHost(host string) error {
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

func ValidateMSSQLPort(port int) error {
	if port < portMinValue || port > portMaxValue {
		return fmt.Errorf("port must be between %d and %d", portMinValue, portMaxValue)
	}
	return nil
}

func ValidateMSSQLUsername(user string) error {
	if user == "" {
		return errors.New("username can't be empty")
	}
	if n := len(user); n < usernameMinLength || n > usernameMaxLength {
		return fmt.Errorf("username must be between %d and %d characters long", usernameMinLength, usernameMaxLength)
	}
	return nil
}

func ValidateMSSQLDatabase(db string) error {
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

func ValidateMSSQLEncryptMode(mode string) error {
	switch mode {
	case "",
		MSSQLEncryptDisable,
		MSSQLEncryptLogin,
		MSSQLEncryptRequire,
		MSSQLEncryptStrict:
		return nil
	}
	return fmt.Errorf("invalid encrypt mode: %s", mode)
}

func ValidateMSSQLName(name string) error {
	if len(name) > nameMaxLength {
		return fmt.Errorf("name must be at most %d characters long", nameMaxLength)
	}
	return nil
}

func ValidateMSSQLDescription(desc string) error {
	if len(desc) > descriptionMaxLength {
		return fmt.Errorf("description must be at most %d characters long", descriptionMaxLength)
	}
	return nil
}

func (w *MSSQLConfigWrapper) Add(conn Connection) {
	ms, ok := conn.(MSSQL)
	if !ok {
		return
	}

	w.Conns = append(w.Conns, ms)
}

func (w *MSSQLConfigWrapper) Validate() {
	for i := range w.Conns {
		w.Conns[i].validationErrs = w.Conns[i].Validate()
	}
}

func (w *MSSQLConfigWrapper) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	return toml.Unmarshal(data, w)
}

func (w *MSSQLConfigWrapper) Marshal() ([]byte, error) {
	return toml.Marshal(w)
}

func (w *MSSQLConfigWrapper) Len() int {
	return len(w.Conns)
}

func (w *MSSQLConfigWrapper) Get(i int) Connection {
	return w.Conns[i]
}

func (w *MSSQLConfigWrapper) Put(i int, conn Connection) {
	ms, ok := conn.(MSSQL)
	if !ok {
		return
	}

	w.Conns[i] = ms
}

func (w *MSSQLConfigWrapper) Remove(i int) {
	if i < 0 || i >= len(w.Conns) {
		return
	}
	w.Conns = append(w.Conns[:i], w.Conns[i+1:]...)
}

func (m MSSQL) ConnType() ConnType {
	return MSSQLConnType
}

func (m MSSQL) Meta() ConnMeta {
	return m.Metadata
}

func (m MSSQL) Host() string {
	return m.Hostname
}

func (m MSSQL) Port() int {
	return m.PortNumber
}

func (m MSSQL) Database() string {
	return m.DBName
}

func (m MSSQL) Schema() string {
	return ""
}

func (m MSSQL) Username() string {
	return m.User
}

func (m MSSQL) SecretRef() string {
	return m.Password
}

func (m *MSSQL) SetSecretRef(secret string) {
	m.Password = secret
}

func (m MSSQL) ConnectionString(secret string) string {
	return m.buildDSN(secret)
}

func (m MSSQL) buildDSN(password string) string {
	host := m.Hostname
	if m.PortNumber != 0 {
		host = net.JoinHostPort(m.Hostname, strconv.Itoa(m.PortNumber))
	}

	userInfo := url.User(m.User)
	if password != "" {
		userInfo = url.UserPassword(m.User, password)
	}

	q := url.Values{}
	q.Set("database", m.DBName)
	if m.EncryptMode != "" {
		q.Set("encrypt", m.EncryptMode)
	}
	if m.TrustCert {
		q.Set("trustservercertificate", "true")
	}

	u := url.URL{
		Scheme:   mssqlScheme,
		User:     userInfo,
		Host:     host,
		RawQuery: q.Encode(),
	}

	return u.String()
}

func (m MSSQL) Validate() []error {
	return validationErrors(
		ValidateMSSQLHost(m.Hostname),
		ValidateMSSQLPort(m.PortNumber),
		ValidateMSSQLUsername(m.User),
		ValidateMSSQLDatabase(m.DBName),
		ValidateMSSQLEncryptMode(m.EncryptMode),
		ValidateMSSQLName(m.Metadata.Name),
		ValidateMSSQLDescription(m.Metadata.Description),
	)
}

func (m MSSQL) IsValid() bool {
	return len(m.validationErrs) == 0
}
