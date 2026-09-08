package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/pelletier/go-toml/v2"
)

const clickhouseScheme = "clickhouse"

var _ Connection = (*ClickHouse)(nil)

type ClickHouse struct {
	Metadata   ConnMeta `toml:"meta"`
	Hostname   string   `toml:"host" json:"host"`
	PortNumber int      `toml:"port" json:"port"`
	User       string   `toml:"username" json:"username"`
	Password   string   `toml:"password" json:"password"`
	DBName     string   `toml:"dbname" json:"dbname"`
	Secure     bool     `toml:"secure,omitempty" json:"secure,omitempty"`

	validationErrs []error
}

var _ ConfigWrapper[Connection] = (*ClickHouseConfigWrapper)(nil)

type ClickHouseConfigWrapper struct {
	Conns []ClickHouse `toml:"clickhouse"`
}

func ValidateClickHouseHost(host string) error {
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

func ValidateClickHousePort(port int) error {
	if port < portMinValue || port > portMaxValue {
		return fmt.Errorf("port must be between %d and %d", portMinValue, portMaxValue)
	}
	return nil
}

func ValidateClickHouseUsername(user string) error {
	if user == "" {
		return errors.New("username can't be empty")
	}
	if n := len(user); n < usernameMinLength || n > usernameMaxLength {
		return fmt.Errorf("username must be between %d and %d characters long", usernameMinLength, usernameMaxLength)
	}
	return nil
}

func ValidateClickHouseDatabase(db string) error {
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

func ValidateClickHouseName(name string) error {
	if len(name) > nameMaxLength {
		return fmt.Errorf("name must be at most %d characters long", nameMaxLength)
	}
	return nil
}

func ValidateClickHouseDescription(desc string) error {
	if len(desc) > descriptionMaxLength {
		return fmt.Errorf("description must be at most %d characters long", descriptionMaxLength)
	}
	return nil
}

func (w *ClickHouseConfigWrapper) Add(conn Connection) {
	ch, ok := conn.(ClickHouse)
	if !ok {
		return
	}

	w.Conns = append(w.Conns, ch)
}

func (w *ClickHouseConfigWrapper) Validate() {
	for i := range w.Conns {
		w.Conns[i].validationErrs = w.Conns[i].Validate()
	}
}

func (w *ClickHouseConfigWrapper) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	return toml.Unmarshal(data, w)
}

func (w *ClickHouseConfigWrapper) Marshal() ([]byte, error) {
	return toml.Marshal(w)
}

func (w *ClickHouseConfigWrapper) Len() int {
	return len(w.Conns)
}

func (w *ClickHouseConfigWrapper) Get(i int) Connection {
	return w.Conns[i]
}

func (w *ClickHouseConfigWrapper) Put(i int, conn Connection) {
	ch, ok := conn.(ClickHouse)
	if !ok {
		return
	}

	w.Conns[i] = ch
}

func (w *ClickHouseConfigWrapper) Remove(i int) {
	if i < 0 || i >= len(w.Conns) {
		return
	}
	w.Conns = append(w.Conns[:i], w.Conns[i+1:]...)
}

func (c ClickHouse) ConnType() ConnType {
	return ClickHouseConnType
}

func (c ClickHouse) Meta() ConnMeta {
	return c.Metadata
}

func (c ClickHouse) Host() string {
	return c.Hostname
}

func (c ClickHouse) Port() int {
	return c.PortNumber
}

func (c ClickHouse) Database() string {
	return c.DBName
}

func (c ClickHouse) Schema() string {
	return ""
}

func (c ClickHouse) Username() string {
	return c.User
}

func (c ClickHouse) SecretRef() string {
	return c.Password
}

func (c *ClickHouse) SetSecretRef(secret string) {
	c.Password = secret
}

func (c ClickHouse) ConnectionString(secret string) string {
	return c.buildDSN(secret)
}

func (c ClickHouse) buildDSN(password string) string {
	host := c.Hostname
	if c.PortNumber != 0 {
		host = net.JoinHostPort(c.Hostname, strconv.Itoa(c.PortNumber))
	}

	userInfo := url.User(c.User)
	if password != "" {
		userInfo = url.UserPassword(c.User, password)
	}

	u := url.URL{
		Scheme: clickhouseScheme,
		User:   userInfo,
		Host:   host,
		Path:   "/" + c.DBName,
	}

	if c.Secure {
		q := url.Values{}
		q.Set("secure", "true")
		u.RawQuery = q.Encode()
	}

	return u.String()
}

func (c ClickHouse) Validate() []error {
	return validationErrors(
		ValidateClickHouseHost(c.Hostname),
		ValidateClickHousePort(c.PortNumber),
		ValidateClickHouseUsername(c.User),
		ValidateClickHouseDatabase(c.DBName),
		ValidateClickHouseName(c.Metadata.Name),
		ValidateClickHouseDescription(c.Metadata.Description),
	)
}

func (c ClickHouse) IsValid() bool {
	return len(c.validationErrs) == 0
}
