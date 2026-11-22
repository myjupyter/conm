package importer

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/adrg/xdg"
	"go.yaml.in/yaml/v3"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/parser/params"
)

const (
	clickhouseHostKey     = "hostname"
	clickhouseSecureKey   = "secure"
	clickhouseInsecureKey = "accept-invalid-certificate"
	clickhouseDefault     = "default"
)

var clickhouseFlags = map[string]string{
	clickhouseHostKey: "--host",
	"port":            "--port",
	"user":            "--user",
	"password":        "--password",
	"database":        "--database",
}

var clickhouseSwitches = map[string]string{
	clickhouseSecureKey:   "--secure",
	clickhouseInsecureKey: "--accept-invalid-certificate",
}

func readClickHouse(path string) ([]Entry, error) {
	connections, err := ReadClickHouseConnections(path)
	if err != nil {
		return nil, err
	}

	entries := make([]Entry, 0, len(connections))
	for _, connection := range connections {
		entries = append(entries, clickhouseEntry(connection))
	}

	return entries, nil
}

func clickhouseEntry(connection Section) Entry {
	entry := Entry{Name: connection.Name}

	args, warnings := clickhouseArgs(connection)

	res, err := params.ParseOptions(config.ClickHouseConnType, args)
	if err != nil {
		entry.Err = err
		return entry
	}

	ch, ok := res.Conn.(config.ClickHouse)
	if !ok {
		entry.Err = fmt.Errorf("connection was read as a %T", res.Conn)
		return entry
	}

	if ch.User == "" {
		ch.User = clickhouseDefault
		warnings = append(warnings, "no user, kept the client's own: "+clickhouseDefault)
	}
	if ch.DBName == "" {
		ch.DBName = clickhouseDefault
		warnings = append(warnings, "no database, kept the client's own: "+clickhouseDefault)
	}
	ch.Metadata.Name = connection.Name

	entry.Conn, entry.Warnings = ch, slices.Concat(warnings, res.Warnings)

	return entry
}

func clickhouseArgs(connection Section) (args, warnings []string) {
	if _, named := Values(connection.Params)[clickhouseHostKey]; !named {
		args = append(args, clickhouseFlags[clickhouseHostKey]+"="+connection.Name)
	}

	for _, param := range connection.Params {
		flag, isSwitch := clickhouseSwitches[param.Key]
		switch {
		case isSwitch:
			if on, _ := strconv.ParseBool(param.Value); on {
				args = append(args, flag)
			}
		case clickhouseFlags[param.Key] != "":
			args = append(args, clickhouseFlags[param.Key]+"="+param.Value)
		default:
			warnings = append(warnings, fmt.Sprintf("ignored parameter %s=%s", param.Key, param.Value))
		}
	}

	return args, warnings
}

const (
	clickhouseLocalConfigName = "clickhouse-client"
	clickhouseConfigName      = "config"
	clickhouseConfigDirName   = "clickhouse"
	clickhouseHomeDirName     = ".clickhouse-client"
	ClickHouseNameKey         = "name"
	YAMLNullTag               = "!!null"
)

var clickhouseConfigExtensions = []string{".xml", ".yaml", ".yml"}

var clickhouseSystemDir = "/etc/clickhouse-client"

func ClickHouseClientPath() string {
	return clickhouseConfigFile(slices.Concat(
		[]string{clickhouseLocalConfigName},
		clickhouseUserStems(),
		[]string{filepath.Join(clickhouseSystemDir, clickhouseConfigName)},
	))
}

func ClickHouseUserPath() string {
	return clickhouseConfigFile(clickhouseUserStems())
}

func clickhouseUserStems() []string {
	userConfigDir := os.Getenv("XDG_CONFIG_HOME")
	if userConfigDir == "" {
		userConfigDir = filepath.Join(xdg.Home, ".config")
	}

	return []string{
		filepath.Join(userConfigDir, clickhouseConfigDirName, clickhouseConfigName),
		filepath.Join(xdg.Home, clickhouseHomeDirName, clickhouseConfigName),
	}
}

func clickhouseConfigFile(stems []string) string {
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

func ReadClickHouseConnections(path string) ([]Section, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var sections []Section
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

func clickhouseXMLConnections(raw []byte) ([]Section, error) {
	var file clickhouseXML
	if err := xml.Unmarshal(raw, &file); err != nil {
		return nil, err
	}

	sections := make([]Section, 0, len(file.Connections))
	for _, conn := range file.Connections {
		var section Section
		for _, field := range conn.Fields {
			section.addClickHouseField(field.XMLName.Local, field.Value)
		}
		sections = append(sections, section)
	}

	return sections, nil
}

func clickhouseYAMLConnections(raw []byte) ([]Section, error) {
	var file clickhouseYAML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		return nil, err
	}

	nodes := []*yaml.Node{&file.Credentials.Connection}
	if file.Credentials.Connection.Kind == yaml.SequenceNode {
		nodes = file.Credentials.Connection.Content
	}

	var sections []Section
	for _, node := range nodes {
		if node.Kind != yaml.MappingNode {
			continue
		}

		var section Section
		for i := 0; i+1 < len(node.Content); i += 2 {
			if key, value := node.Content[i], node.Content[i+1]; value.Tag != YAMLNullTag {
				section.addClickHouseField(key.Value, value.Value)
			}
		}
		sections = append(sections, section)
	}

	return sections, nil
}

func (s *Section) addClickHouseField(key, value string) {
	key, value = strings.TrimSpace(key), strings.TrimSpace(value)
	if key == ClickHouseNameKey {
		s.Name = value
		return
	}

	s.Params = append(s.Params, Param{Key: key, Value: value})
}
