package exporter

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strconv"

	"go.yaml.in/yaml/v3"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/importer"
)

const (
	mssqlDefaultDatabase = "master"
	mssqlNoEncryption    = "none"
)

func mssqlSections(conn config.Connection, password string) ([]importer.Section, []string) {
	ms, _ := conn.(config.MSSQL)

	var context params
	context.add(importer.SQLConfigAddress, ms.Hostname)
	context.addPort(importer.SQLConfigPort, ms.PortNumber)
	context.add(importer.SQLConfigUsername, ms.User)
	if password != "" {
		context.add(importer.SQLConfigPassword, base64.StdEncoding.EncodeToString([]byte(password)))
		context.add(importer.SQLConfigPasswordEncryption, mssqlNoEncryption)
	}

	var warnings []string
	if ms.DBName != "" && ms.DBName != mssqlDefaultDatabase {
		warnings = append(warnings, fmt.Sprintf("database %s is not exported: a context holds none", ms.DBName))
	}
	if (ms.EncryptMode != "" && ms.EncryptMode != config.MSSQLEncryptLogin) || ms.TrustCert {
		warnings = append(warnings, "encryption settings are not exported: a context holds none")
	}

	return []importer.Section{{Name: ms.Metadata.Name, Params: context}}, append(warnings, transportWarning(conn)...)
}

const (
	sqlConfigVersionKey   = "version"
	sqlConfigVersion      = "v1"
	sqlConfigEndpointsKey = "endpoints"
	sqlConfigContextsKey  = "contexts"
	sqlConfigUsersKey     = "users"
	sqlConfigBasicAuth    = "basic"
)

type sqlConfigNamed struct {
	Name string `yaml:"name"`
}

type sqlConfigNames struct {
	Endpoints []sqlConfigNamed `yaml:"endpoints"`
	Contexts  []sqlConfigNamed `yaml:"contexts"`
	Users     []sqlConfigNamed `yaml:"users"`
}

type sqlConfigEndpoint struct {
	Name     string `yaml:"name"`
	Endpoint struct {
		Address string `yaml:"address"`
		Port    int    `yaml:"port,omitempty"`
	} `yaml:"endpoint"`
}

type sqlConfigContext struct {
	Name    string `yaml:"name"`
	Context struct {
		Endpoint string `yaml:"endpoint"`
		User     string `yaml:"user,omitempty"`
	} `yaml:"context"`
}

type sqlConfigUser struct {
	Name               string `yaml:"name"`
	AuthenticationType string `yaml:"authentication-type"`
	BasicAuth          struct {
		Username           string `yaml:"username"`
		PasswordEncryption string `yaml:"password-encryption,omitempty"`
		Password           string `yaml:"password,omitempty"`
	} `yaml:"basic-auth"`
}

func sqlConfigHeld(path string) ([]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	var file sqlConfigNames
	if err := yaml.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("couldn't parse file %q: %w", path, err)
	}

	held := make([]string, 0, len(file.Endpoints)+len(file.Contexts)+len(file.Users))
	for _, named := range slices.Concat(file.Endpoints, file.Contexts, file.Users) {
		held = append(held, named.Name)
	}

	return held, nil
}

func sqlConfigMerged(_ string, raw []byte, sections []importer.Section) ([]byte, error) {
	doc, root, err := yamlMappingRoot(raw)
	if err != nil {
		return nil, err
	}

	if len(root.Content) == 0 {
		root.Content = append(root.Content, yamlScalar(sqlConfigVersionKey), yamlScalar(sqlConfigVersion))
	}
	endpoints := yamlChild(root, sqlConfigEndpointsKey, yaml.SequenceNode)
	contexts := yamlChild(root, sqlConfigContextsKey, yaml.SequenceNode)
	users := yamlChild(root, sqlConfigUsersKey, yaml.SequenceNode)

	for _, section := range sections {
		values := make(map[string]string, len(section.Params))
		for _, param := range section.Params {
			values[param.Key] = param.Value
		}

		endpoint := sqlConfigEndpoint{Name: section.Name}
		endpoint.Endpoint.Address = values[importer.SQLConfigAddress]
		endpoint.Endpoint.Port, _ = strconv.Atoi(values[importer.SQLConfigPort])

		context := sqlConfigContext{Name: section.Name}
		context.Context.Endpoint = section.Name

		added := []error{yamlAppend(endpoints, endpoint)}
		if username := values[importer.SQLConfigUsername]; username != "" {
			user := sqlConfigUser{Name: section.Name, AuthenticationType: sqlConfigBasicAuth}
			user.BasicAuth.Username = username
			user.BasicAuth.Password = values[importer.SQLConfigPassword]
			user.BasicAuth.PasswordEncryption = values[importer.SQLConfigPasswordEncryption]

			context.Context.User = section.Name
			added = append(added, yamlAppend(users, user))
		}

		if err := errors.Join(append(added, yamlAppend(contexts, context))...); err != nil {
			return nil, err
		}
	}

	return yamlEncoded(doc)
}
