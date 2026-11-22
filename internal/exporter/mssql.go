package exporter

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"

	"go.yaml.in/yaml/v3"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/importer"
)

const mssqlDefaultDatabase = "master"

func mssqlSections(conn config.Connection, password string) ([]importer.Section, []string) {
	ms, _ := conn.(config.MSSQL)

	var (
		context  params
		warnings []string
	)
	context.add(importer.SQLConfigAddress, ms.Hostname)
	context.addPort(importer.SQLConfigPort, ms.PortNumber)
	context.add(importer.SQLConfigUsername, ms.User)
	switch {
	case password == "":
	case ms.User == "":
		warnings = append(warnings, "the password is not exported: a context without a user holds none")
	default:
		context.add(importer.SQLConfigPassword, base64.StdEncoding.EncodeToString([]byte(password)))
		context.add(importer.SQLConfigPasswordEncryption, importer.MSSQLNoEncryption)
	}

	if ms.DBName != "" && ms.DBName != mssqlDefaultDatabase {
		warnings = append(warnings, fmt.Sprintf("database %s is not exported: a context holds none", ms.DBName))
	}
	if (ms.EncryptMode != "" && ms.EncryptMode != config.MSSQLEncryptLogin) || ms.TrustCert {
		warnings = append(warnings, "encryption settings are not exported: a context holds none")
	}

	return []importer.Section{{Name: ms.Metadata.Name, Params: context}}, append(warnings, transportWarning(conn)...)
}

const (
	sqlConfigVersionKey        = "version"
	sqlConfigVersion           = "v1"
	sqlConfigEndpointsKey      = "endpoints"
	sqlConfigContextsKey       = "contexts"
	sqlConfigUsersKey          = "users"
	sqlConfigCurrentContextKey = "currentcontext"
	sqlConfigBasicAuth         = "basic"
)

func sqlConfigHeld(path string) ([]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	var file importer.SQLConfig
	if err := yaml.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("couldn't parse file %q: %w", path, err)
	}

	held := make([]string, 0, len(file.Endpoints)+len(file.Contexts)+len(file.Users))
	for _, endpoint := range file.Endpoints {
		held = append(held, endpoint.Name)
	}
	for _, context := range file.Contexts {
		held = append(held, context.Name)
	}
	for _, user := range file.Users {
		held = append(held, user.Name)
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
		values := importer.Values(section.Params)

		endpoint := importer.SQLConfigEndpoint{Name: section.Name}
		endpoint.Endpoint.Address = values[importer.SQLConfigAddress]
		endpoint.Endpoint.Port, _ = strconv.Atoi(values[importer.SQLConfigPort])

		context := importer.SQLConfigContext{Name: section.Name}
		context.Context.Endpoint = section.Name

		added := []error{yamlAppend(endpoints, endpoint)}
		if username := values[importer.SQLConfigUsername]; username != "" {
			user := importer.SQLConfigUser{Name: section.Name, AuthenticationType: sqlConfigBasicAuth}
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

	if current := yamlChild(root, sqlConfigCurrentContextKey, yaml.ScalarNode); current.Kind == yaml.ScalarNode && current.Value == "" {
		*current = *yamlScalar(sections[0].Name)
	}

	return yamlEncoded(doc)
}
