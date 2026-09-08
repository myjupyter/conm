package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"

	"github.com/pelletier/go-toml/v2"
)

const (
	mongodbScheme        = "mongodb"
	mongodbAuthSourceKey = "authSource"
	mongodbTLSKey        = "tls"
	mongodbTLSEnabled    = "true"
)

var mongodbDatabaseRegexp = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

var _ Connection = (*MongoDB)(nil)

type MongoDB struct {
	Metadata   ConnMeta `toml:"meta"`
	Hostname   string   `toml:"host" json:"host"`
	PortNumber int      `toml:"port" json:"port"`
	User       string   `toml:"username" json:"username"`
	Password   string   `toml:"password" json:"password"`
	DBName     string   `toml:"dbname" json:"dbname"`
	AuthSource string   `toml:"authsource,omitempty" json:"authsource,omitempty"`
	TLS        bool     `toml:"tls,omitempty" json:"tls,omitempty"`

	validationErrs []error
}

var _ ConfigWrapper[Connection] = (*MongoDBConfigWrapper)(nil)

type MongoDBConfigWrapper struct {
	Conns []MongoDB `toml:"mongodb"`
}

func ValidateMongoDBHost(host string) error {
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

func ValidateMongoDBPort(port int) error {
	if port < portMinValue || port > portMaxValue {
		return fmt.Errorf("port must be between %d and %d", portMinValue, portMaxValue)
	}
	return nil
}

func ValidateMongoDBUsername(user string) error {
	if user == "" {
		return nil
	}
	if len(user) > usernameMaxLength {
		return fmt.Errorf("username must be at most %d characters long", usernameMaxLength)
	}
	return nil
}

func ValidateMongoDBDatabase(db string) error {
	if db == "" {
		return errors.New("database can't be empty")
	}
	if len(db) > databaseMaxLength {
		return fmt.Errorf("database must be at most %d characters long", databaseMaxLength)
	}
	if !mongodbDatabaseRegexp.MatchString(db) {
		return errors.New("database contains invalid characters")
	}
	return nil
}

func ValidateMongoDBAuthSource(source string) error {
	if source == "" {
		return nil
	}
	return ValidateMongoDBDatabase(source)
}

func ValidateMongoDBName(name string) error {
	if len(name) > nameMaxLength {
		return fmt.Errorf("name must be at most %d characters long", nameMaxLength)
	}
	return nil
}

func ValidateMongoDBDescription(desc string) error {
	if len(desc) > descriptionMaxLength {
		return fmt.Errorf("description must be at most %d characters long", descriptionMaxLength)
	}
	return nil
}

func (w *MongoDBConfigWrapper) Add(conn Connection) {
	mg, ok := conn.(MongoDB)
	if !ok {
		return
	}

	w.Conns = append(w.Conns, mg)
}

func (w *MongoDBConfigWrapper) Validate() {
	for i := range w.Conns {
		w.Conns[i].validationErrs = w.Conns[i].Validate()
	}
}

func (w *MongoDBConfigWrapper) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	return toml.Unmarshal(data, w)
}

func (w *MongoDBConfigWrapper) Marshal() ([]byte, error) {
	return toml.Marshal(w)
}

func (w *MongoDBConfigWrapper) Len() int {
	return len(w.Conns)
}

func (w *MongoDBConfigWrapper) Get(i int) Connection {
	return w.Conns[i]
}

func (w *MongoDBConfigWrapper) Put(i int, conn Connection) {
	mg, ok := conn.(MongoDB)
	if !ok {
		return
	}

	w.Conns[i] = mg
}

func (w *MongoDBConfigWrapper) Remove(i int) {
	if i < 0 || i >= len(w.Conns) {
		return
	}
	w.Conns = append(w.Conns[:i], w.Conns[i+1:]...)
}

func (m MongoDB) ConnType() ConnType {
	return MongoDBConnType
}

func (m MongoDB) Meta() ConnMeta {
	return m.Metadata
}

func (m MongoDB) Host() string {
	return m.Hostname
}

func (m MongoDB) Port() int {
	return m.PortNumber
}

func (m MongoDB) Database() string {
	return m.DBName
}

func (m MongoDB) Schema() string {
	return ""
}

func (m MongoDB) Username() string {
	return m.User
}

func (m MongoDB) SecretRef() string {
	return m.Password
}

func (m *MongoDB) SetSecretRef(secret string) {
	m.Password = secret
}

func (m MongoDB) ConnectionString(secret string) string {
	return m.buildDSN(secret)
}

func (m MongoDB) buildDSN(password string) string {
	host := m.Hostname
	if m.PortNumber != 0 {
		host = net.JoinHostPort(m.Hostname, strconv.Itoa(m.PortNumber))
	}

	var userInfo *url.Userinfo
	switch {
	case password != "":
		userInfo = url.UserPassword(m.User, password)
	case m.User != "":
		userInfo = url.User(m.User)
	}

	u := url.URL{
		Scheme: mongodbScheme,
		User:   userInfo,
		Host:   host,
		Path:   "/" + m.DBName,
	}

	q := url.Values{}
	if m.AuthSource != "" {
		q.Set(mongodbAuthSourceKey, m.AuthSource)
	}
	if m.TLS {
		q.Set(mongodbTLSKey, mongodbTLSEnabled)
	}
	u.RawQuery = q.Encode()

	return u.String()
}

func (m MongoDB) Validate() []error {
	return validationErrors(
		ValidateMongoDBHost(m.Hostname),
		ValidateMongoDBPort(m.PortNumber),
		ValidateMongoDBUsername(m.User),
		ValidateMongoDBDatabase(m.DBName),
		ValidateMongoDBAuthSource(m.AuthSource),
		ValidateMongoDBName(m.Metadata.Name),
		ValidateMongoDBDescription(m.Metadata.Description),
	)
}

func (m MongoDB) IsValid() bool {
	return len(m.validationErrs) == 0
}
