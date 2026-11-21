package config

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/adrg/xdg"
	"go.yaml.in/yaml/v3"
)

const (
	pgServiceFileName       = ".pg_service.conf"
	pgPassFileName          = ".pgpass"
	mysqlOptionFileName     = ".my.cnf"
	sshClientConfigFileName = "config"
	sshConfigDirName        = ".ssh"
	sqlConfigFileName       = "sqlconfig"
	sqlConfigDirName        = ".sqlcmd"

	clickhouseLocalConfigName = "clickhouse-client"
	clickhouseConfigName      = "config"
	clickhouseConfigDirName   = "clickhouse"
	clickhouseHomeDirName     = ".clickhouse-client"
	clickhouseNameKey         = "name"
	yamlNullTag               = "!!null"

	sshHostKeyword  = "host"
	sshMatchKeyword = "match"
	sshAnyHost      = "*"
	pgPassFields    = 5
	pgPassWildcard  = "*"
)

type ForeignSection struct {
	Name   string
	Params []ForeignParam
}

type ForeignParam struct {
	Key   string
	Value string
}

var clickhouseConfigExtensions = []string{".xml", ".yaml", ".yml"}

var clickhouseSystemDir = "/etc/clickhouse-client"

func PGServicePath() string {
	if path := os.Getenv("PGSERVICEFILE"); path != "" {
		return path
	}

	return filepath.Join(xdg.Home, pgServiceFileName)
}

func MySQLOptionPath() string {
	return filepath.Join(xdg.Home, mysqlOptionFileName)
}

func SSHConfigPath() string {
	return filepath.Join(xdg.Home, sshConfigDirName, sshClientConfigFileName)
}

func SQLConfigPath() string {
	return filepath.Join(xdg.Home, sqlConfigDirName, sqlConfigFileName)
}

func ClickHouseClientPath() string {
	userConfigDir := os.Getenv("XDG_CONFIG_HOME")
	if userConfigDir == "" {
		userConfigDir = filepath.Join(xdg.Home, ".config")
	}

	stems := []string{
		clickhouseLocalConfigName,
		filepath.Join(userConfigDir, clickhouseConfigDirName, clickhouseConfigName),
		filepath.Join(xdg.Home, clickhouseHomeDirName, clickhouseConfigName),
		filepath.Join(clickhouseSystemDir, clickhouseConfigName),
	}
	for _, stem := range stems {
		for _, ext := range clickhouseConfigExtensions {
			if isFile(stem + ext) {
				return stem + ext
			}
		}
	}

	return filepath.Join(xdg.Home, clickhouseHomeDirName, clickhouseConfigName+clickhouseConfigExtensions[0])
}

func isFile(path string) bool {
	info, err := os.Stat(filepath.Clean(path))

	return err == nil && info.Mode().IsRegular()
}

func PGPassPath() string {
	if path := os.Getenv("PGPASSFILE"); path != "" {
		return path
	}

	return filepath.Join(xdg.Home, pgPassFileName)
}

type PGPass [][]string

func ReadPGPass(path string) (PGPass, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var rows PGPass
	for line := range strings.Lines(string(raw)) {
		line = strings.TrimRight(line, "\r\n")
		if strings.HasPrefix(line, "#") {
			continue
		}
		if fields := splitPGPassRow(line); len(fields) == pgPassFields {
			rows = append(rows, fields)
		}
	}

	return rows, nil
}

func (p PGPass) Password(host, port, database, user string) string {
	want := []string{host, port, database, user}

rows:
	for _, row := range p {
		for i, value := range want {
			if row[i] != pgPassWildcard && row[i] != value {
				continue rows
			}
		}

		return row[pgPassFields-1]
	}

	return ""
}

func splitPGPassRow(line string) []string {
	var (
		fields  []string
		field   strings.Builder
		escaped bool
	)

	for _, r := range line {
		switch {
		case escaped:
			field.WriteRune(r)
			escaped = false
		case r == '\\':
			escaped = true
		case r == ':' && len(fields) < pgPassFields-1:
			fields = append(fields, field.String())
			field.Reset()
		default:
			field.WriteRune(r)
		}
	}

	return append(fields, field.String())
}

type clickhouseXML struct {
	Connections []struct {
		Fields []struct {
			XMLName xml.Name
			Value   string `xml:",chardata"`
		} `xml:",any"`
	} `xml:"connections_credentials>connection"`
}

type clickhouseYAML struct {
	Credentials struct {
		Connection yaml.Node `yaml:"connection"`
	} `yaml:"connections_credentials"`
}

func ReadClickHouseConnections(path string) ([]ForeignSection, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var sections []ForeignSection
	if bytes.HasPrefix(bytes.TrimSpace(raw), []byte("<")) {
		sections, err = clickhouseXMLConnections(raw)
	} else {
		sections, err = clickhouseYAMLConnections(raw)
	}
	if err != nil {
		return nil, fmt.Errorf("couldn't parse file %q: %w", path, err)
	}

	return sections, nil
}

func clickhouseXMLConnections(raw []byte) ([]ForeignSection, error) {
	var file clickhouseXML
	if err := xml.Unmarshal(raw, &file); err != nil {
		return nil, err
	}

	sections := make([]ForeignSection, 0, len(file.Connections))
	for _, conn := range file.Connections {
		var section ForeignSection
		for _, field := range conn.Fields {
			section.addClickHouseField(field.XMLName.Local, field.Value)
		}
		sections = append(sections, section)
	}

	return sections, nil
}

func clickhouseYAMLConnections(raw []byte) ([]ForeignSection, error) {
	var file clickhouseYAML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		return nil, err
	}

	nodes := []*yaml.Node{&file.Credentials.Connection}
	if file.Credentials.Connection.Kind == yaml.SequenceNode {
		nodes = file.Credentials.Connection.Content
	}

	var sections []ForeignSection
	for _, node := range nodes {
		if node.Kind != yaml.MappingNode {
			continue
		}

		var section ForeignSection
		for i := 0; i+1 < len(node.Content); i += 2 {
			if key, value := node.Content[i], node.Content[i+1]; value.Tag != yamlNullTag {
				section.addClickHouseField(key.Value, value.Value)
			}
		}
		sections = append(sections, section)
	}

	return sections, nil
}

func (s *ForeignSection) addClickHouseField(key, value string) {
	key, value = strings.TrimSpace(key), strings.TrimSpace(value)
	if key == clickhouseNameKey {
		s.Name = value
		return
	}

	s.Params = append(s.Params, ForeignParam{Key: key, Value: value})
}

type sqlConfig struct {
	Endpoints []struct {
		Name     string `yaml:"name"`
		Endpoint struct {
			Address string `yaml:"address"`
			Port    int    `yaml:"port"`
		} `yaml:"endpoint"`
	} `yaml:"endpoints"`
	Contexts []struct {
		Name    string `yaml:"name"`
		Context struct {
			Endpoint string `yaml:"endpoint"`
			User     string `yaml:"user"`
		} `yaml:"context"`
	} `yaml:"contexts"`
	Users []struct {
		Name      string `yaml:"name"`
		BasicAuth struct {
			Username           string `yaml:"username"`
			Password           string `yaml:"password"`
			PasswordEncryption string `yaml:"password-encryption"`
		} `yaml:"basic-auth"`
	} `yaml:"users"`
}

const (
	SQLConfigAddress            = "address"
	SQLConfigPort               = "port"
	SQLConfigUsername           = "username"
	SQLConfigPassword           = "password"
	SQLConfigPasswordEncryption = "password-encryption"
)

func ReadSQLConfig(path string) ([]ForeignSection, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var file sqlConfig
	if err := yaml.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("couldn't parse file %q: %w", path, err)
	}

	sections := make([]ForeignSection, 0, len(file.Contexts))
	for _, context := range file.Contexts {
		section := ForeignSection{Name: context.Name}

		for _, endpoint := range file.Endpoints {
			if endpoint.Name != context.Context.Endpoint {
				continue
			}
			section.Params = append(section.Params, ForeignParam{Key: SQLConfigAddress, Value: endpoint.Endpoint.Address})
			if port := endpoint.Endpoint.Port; port != 0 {
				section.Params = append(section.Params, ForeignParam{Key: SQLConfigPort, Value: strconv.Itoa(port)})
			}

			break
		}

		for _, user := range file.Users {
			if user.Name != context.Context.User {
				continue
			}
			section.Params = append(section.Params,
				ForeignParam{Key: SQLConfigUsername, Value: user.BasicAuth.Username},
				ForeignParam{Key: SQLConfigPassword, Value: user.BasicAuth.Password},
				ForeignParam{Key: SQLConfigPasswordEncryption, Value: user.BasicAuth.PasswordEncryption},
			)

			break
		}

		sections = append(sections, section)
	}

	return sections, nil
}

func ReadSSHConfig(path string) ([]ForeignSection, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	sections := []ForeignSection{{Name: sshAnyHost}}
	skipping := false
	for line := range strings.Lines(string(raw)) {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}

		keyword, value := SplitSSHOption(line)
		switch strings.ToLower(keyword) {
		case sshHostKeyword:
			sections, skipping = append(sections, ForeignSection{Name: value}), false
		case sshMatchKeyword:
			skipping = true
		default:
			if skipping {
				continue
			}
			last := &sections[len(sections)-1]
			last.Params = append(last.Params, ForeignParam{Key: keyword, Value: value})
		}
	}

	return sections, nil
}

func SplitSSHOption(line string) (keyword, value string) {
	line = strings.TrimSpace(line)
	at := strings.IndexAny(line, "= \t")
	if at < 0 {
		return line, ""
	}

	value = strings.TrimSpace(line[at:])
	value = strings.TrimSpace(strings.TrimPrefix(value, "="))
	if len(value) > 1 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
	}

	return line[:at], value
}
