package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

var simpleHostRegexp = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

var simpleDatabaseRegexp = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_.-]*$`)

var simpleSchemaRegexp = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_$]*$`)

const (
	hostMinLength = 1
	hostMaxLength = 255
)

const (
	portMinValue = 1
	portMaxValue = 65535
)

const (
	usernameMinLength = 1
	usernameMaxLength = 63
)

const databaseMaxLength = 63

const schemaMaxLength = 63

const (
	nameMaxLength        = 63
	descriptionMaxLength = 255
)

type PostgresSSLMode = string

const (
	PostgresSSLModeDisable    PostgresSSLMode = "disable"
	PostgresSSLModeAllow      PostgresSSLMode = "allow"
	PostgresSSLModePrefer     PostgresSSLMode = "prefer"
	PostgresSSLModeRequire    PostgresSSLMode = "require"
	PostgresSSLModeVerifyCA   PostgresSSLMode = "verify-ca"
	PostgresSSLModeVerifyFull PostgresSSLMode = "verify-full"
)

type PostgresFormField = string

const (
	PostgresFormFieldName        PostgresFormField = "Name"
	PostgresFormFieldDescription PostgresFormField = "Description"
	PostgresFormFieldTags        PostgresFormField = "Tags"
	PostgresFormFieldHost        PostgresFormField = "Host"
	PostgresFormFieldPort        PostgresFormField = "Port"
	PostgresFormFieldUsername    PostgresFormField = "Username"
	PostgresFormFieldPassword    PostgresFormField = "Password"
	PostgresFormFieldDatabase    PostgresFormField = "Database"
	PostgresFormFieldSchema      PostgresFormField = "Schema"
	PostgresFormFieldSSLMode     PostgresFormField = "SSLMode"
)

type Postgres struct {
	Meta       ConnMeta `toml:"meta"`
	Hostname   string   `toml:"host" json:"host"`
	PortNumber int      `toml:"port" json:"port"`
	User       string   `toml:"username" json:"username"`
	Password   string   `toml:"password" json:"password"`
	DBName     string   `toml:"dbname" json:"dbname"`
	SchemaName string   `toml:"schema,omitempty" json:"schema,omitempty"`
	SSLMode    string   `toml:"sslmode,omitempty" json:"sslmode,omitempty"`

	validationErrs []error
}

type PostgresConfigWrapper struct {
	Conns []Postgres `toml:"postgres"`
}

func ImportFromPGPass() ([]Postgres, error) {
	pgPassPath := os.Getenv("PGPASSFILE")
	if pgPassPath == "" {
		homeDir := os.Getenv("HOME")
		const pgpassFile = ".pgpass"
		pgPassPath = filepath.Join(homeDir, pgpassFile)
	}

	raw, err := os.ReadFile(pgPassPath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(raw), "\n")

	var confs []Postgres
	for _, line := range lines {
		values := strings.Split(line, ":")
		if skipPGPassRow(values) {
			continue
		}

		port, err := strconv.Atoi(values[1])
		if err != nil {
			continue
		}

		// .pgpass format: hostname:port:database:username:password
		conf := Postgres{
			Hostname:   values[0],
			PortNumber: port,
			DBName:     values[2],
			User:       values[3],
			Password:   values[4],
			Meta: ConnMeta{
				Name: values[0],
			},
		}

		confs = append(confs, conf)
	}

	return confs, nil
}

func ValidatePostgresHost(host string) error {
	if host == "" {
		return fmt.Errorf("hostname can't be empty")
	}
	if len(host) < hostMinLength || len(host) > hostMaxLength {
		return fmt.Errorf("hostname must be between %d and %d characters long", hostMinLength, hostMaxLength)
	}
	if net.ParseIP(host) != nil {
		return nil
	}
	if !simpleHostRegexp.MatchString(host) {
		return fmt.Errorf("hostname is not a valid host or IP address")
	}
	return nil
}

func ValidatePostgresPort(port int) error {
	if port < portMinValue || port > portMaxValue {
		return fmt.Errorf("port must be between %d and %d", portMinValue, portMaxValue)
	}
	return nil
}

func ValidatePostgresUsername(user string) error {
	if user == "" {
		return fmt.Errorf("username can't be empty")
	}
	if n := len(user); n < usernameMinLength || n > usernameMaxLength {
		return fmt.Errorf("username must be between %d and %d characters long", usernameMinLength, usernameMaxLength)
	}
	return nil
}

func ValidatePostgresDatabase(db string) error {
	if db == "" {
		return fmt.Errorf("database can't be empty")
	}
	if len(db) > databaseMaxLength {
		return fmt.Errorf("database must be at most %d characters long", databaseMaxLength)
	}
	if !simpleDatabaseRegexp.MatchString(db) {
		return fmt.Errorf("database contains invalid characters")
	}
	return nil
}

func ValidatePostgresSchema(schema string) error {
	if schema == "" {
		return nil
	}
	if len(schema) > schemaMaxLength {
		return fmt.Errorf("schema must be at most %d characters long", schemaMaxLength)
	}
	if !simpleSchemaRegexp.MatchString(schema) {
		return fmt.Errorf("schema contains invalid characters")
	}
	return nil
}

func ValidatePostgresSSLMode(mode string) error {
	switch mode {
	case "",
		PostgresSSLModeDisable,
		PostgresSSLModeAllow,
		PostgresSSLModePrefer,
		PostgresSSLModeRequire,
		PostgresSSLModeVerifyCA,
		PostgresSSLModeVerifyFull:
		return nil
	}
	return fmt.Errorf("invalid SSL mode: %s", mode)
}

func ValidatePostgresName(name string) error {
	if len(name) > nameMaxLength {
		return fmt.Errorf("name must be at most %d characters long", nameMaxLength)
	}
	return nil
}

func ValidatePostgresDescription(desc string) error {
	if len(desc) > descriptionMaxLength {
		return fmt.Errorf("description must be at most %d characters long", descriptionMaxLength)
	}
	return nil
}

func (w *PostgresConfigWrapper) Add(conn Postgres) {
	w.Conns = append(w.Conns, conn)
}

func (w *PostgresConfigWrapper) Validate() {
	for i := range w.Conns {
		w.Conns[i].validationErrs = w.Conns[i].Validate()
	}
}

func (w *PostgresConfigWrapper) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	return toml.Unmarshal(data, w)
}

func (w *PostgresConfigWrapper) Marshal() ([]byte, error) {
	return toml.Marshal(w)
}

func (w *PostgresConfigWrapper) Len() int {
	return len(w.Conns)
}

func (w *PostgresConfigWrapper) Get(i int) Postgres {
	return w.Conns[i]
}

func (w *PostgresConfigWrapper) Put(i int, conn Postgres) {
	w.Conns[i] = conn
}

func (w *PostgresConfigWrapper) Remove(i int) {
	if i < 0 || i >= len(w.Conns) {
		return
	}
	w.Conns = append(w.Conns[:i], w.Conns[i+1:]...)
}

func (p Postgres) ConnType() ConnType {
	return PostgresConnType
}

func (p Postgres) ConnMeta() ConnMeta {
	return p.Meta
}

func (p Postgres) Name() string {
	return p.Meta.Name
}

func (p Postgres) Description() string {
	return p.Meta.Description
}

func (p Postgres) Tags() []string {
	return p.Meta.Tags
}

func (p Postgres) Host() string {
	return p.Hostname
}

func (p Postgres) Port() int {
	return p.PortNumber
}

func (p Postgres) Database() string {
	return p.DBName
}

func (p Postgres) Schema() string {
	return p.SchemaName
}

func (p Postgres) Username() string {
	return p.User
}

func (p Postgres) URL() string {
	host := p.Hostname
	if p.PortNumber != 0 {
		host = net.JoinHostPort(p.Hostname, strconv.Itoa(p.PortNumber))
	}
	u := url.URL{
		Scheme: "postgresql",
		User:   url.UserPassword(p.User, p.Password),
		Host:   host,
		Path:   "/" + p.DBName,
	}

	q := url.Values{}
	if p.SSLMode != "" {
		q.Set("sslmode", p.SSLMode)
	}
	if p.SchemaName != "" {
		q.Set("options", "-c search_path="+p.SchemaName)
	}
	if len(q) > 0 {
		u.RawQuery = strings.ReplaceAll(q.Encode(), "+", "%20")
	}

	return u.String()
}

func (p Postgres) Validate() []error {
	var errs []error
	if err := ValidatePostgresHost(p.Hostname); err != nil {
		errs = append(errs, err)
	}
	if err := ValidatePostgresPort(p.PortNumber); err != nil {
		errs = append(errs, err)
	}
	if err := ValidatePostgresUsername(p.User); err != nil {
		errs = append(errs, err)
	}
	if err := ValidatePostgresDatabase(p.DBName); err != nil {
		errs = append(errs, err)
	}
	if err := ValidatePostgresSchema(p.SchemaName); err != nil {
		errs = append(errs, err)
	}
	if err := ValidatePostgresSSLMode(p.SSLMode); err != nil {
		errs = append(errs, err)
	}
	if err := ValidatePostgresName(p.Meta.Name); err != nil {
		errs = append(errs, err)
	}
	if err := ValidatePostgresDescription(p.Meta.Description); err != nil {
		errs = append(errs, err)
	}
	return errs
}

func (p Postgres) IsValid() bool {
	return len(p.validationErrs) == 0
}

func skipPGPassRow(values []string) bool {
	if len(values) != 5 {
		return true
	}

	return values[0] == "*" || values[1] == "*" || values[2] == "*" || values[3] == "*" || values[4] == "*"
}
