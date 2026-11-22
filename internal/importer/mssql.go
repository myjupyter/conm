package importer

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/adrg/xdg"
	"go.yaml.in/yaml/v3"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/parser/params"
)

const (
	mssqlDefaultDatabase = "master"
	MSSQLNoEncryption    = "none"
)

var errNoEndpoint = errors.New("context points at no endpoint")

func readMSSQL(path string) ([]Entry, error) {
	contexts, err := ReadSQLConfig(path)
	if err != nil {
		return nil, err
	}

	entries := make([]Entry, 0, len(contexts))
	for _, context := range contexts {
		entries = append(entries, mssqlEntry(context))
	}

	return entries, nil
}

func mssqlEntry(context Section) Entry {
	entry := Entry{Name: context.Name}

	args, warnings, err := mssqlArgs(Values(context.Params))
	if err != nil {
		entry.Err = err
		return entry
	}

	res, err := params.ParseOptions(config.MSSQLConnType, args)
	if err != nil {
		entry.Err = err
		return entry
	}

	ms, ok := res.Conn.(config.MSSQL)
	if !ok {
		entry.Err = fmt.Errorf("context was read as a %T", res.Conn)
		return entry
	}

	if ms.DBName == "" {
		ms.DBName = mssqlDefaultDatabase
		warnings = append(warnings, "no database in a context, kept "+mssqlDefaultDatabase)
	}
	ms.Metadata.Name = context.Name

	entry.Conn, entry.Warnings = ms, slices.Concat(warnings, res.Warnings)

	return entry
}

func mssqlArgs(context map[string]string) (args, warnings []string, err error) {
	server := context[SQLConfigAddress]
	if server == "" {
		return nil, nil, errNoEndpoint
	}
	if port := context[SQLConfigPort]; port != "" {
		server += "," + port
	}
	args = append(args, "--server="+server)

	if user := context[SQLConfigUsername]; user != "" {
		args = append(args, "--user-name="+user)
	}

	stored := context[SQLConfigPassword]
	if stored == "" {
		return args, nil, nil
	}

	if method := context[SQLConfigPasswordEncryption]; method != "" && method != MSSQLNoEncryption {
		return args, []string{fmt.Sprintf("password is encrypted with %s and was left out", method)}, nil
	}

	password, ok := decoded(stored)
	if !ok {
		return args, []string{"password is not base64 and was left out"}, nil
	}

	return append(args, "--password="+password), nil, nil
}

func decoded(stored string) (string, bool) {
	password, err := base64.StdEncoding.DecodeString(stored)

	return string(password), err == nil
}

const (
	sqlConfigFileName = "sqlconfig"
	sqlConfigDirName  = ".sqlcmd"
)

func SQLConfigPath() string {
	return filepath.Join(xdg.Home, sqlConfigDirName, sqlConfigFileName)
}

type SQLConfig struct {
	Endpoints []SQLConfigEndpoint `yaml:"endpoints"`
	Contexts  []SQLConfigContext  `yaml:"contexts"`
	Users     []SQLConfigUser     `yaml:"users"`
}

type SQLConfigEndpoint struct {
	Name     string `yaml:"name"`
	Endpoint struct {
		Address string `yaml:"address"`
		Port    int    `yaml:"port,omitempty"`
	} `yaml:"endpoint"`
}

type SQLConfigContext struct {
	Name    string `yaml:"name"`
	Context struct {
		Endpoint string `yaml:"endpoint"`
		User     string `yaml:"user,omitempty"`
	} `yaml:"context"`
}

type SQLConfigUser struct {
	Name               string `yaml:"name"`
	AuthenticationType string `yaml:"authentication-type"`
	BasicAuth          struct {
		Username           string `yaml:"username"`
		PasswordEncryption string `yaml:"password-encryption,omitempty"`
		Password           string `yaml:"password,omitempty"`
	} `yaml:"basic-auth"`
}

func ReadSQLConfig(path string) ([]Section, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var file SQLConfig
	if err := yaml.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("couldn't parse file %q: %w", path, err)
	}

	sections := make([]Section, 0, len(file.Contexts))
	for _, context := range file.Contexts {
		section := Section{Name: context.Name}

		for _, endpoint := range file.Endpoints {
			if endpoint.Name != context.Context.Endpoint {
				continue
			}
			section.Params = append(section.Params, Param{Key: SQLConfigAddress, Value: endpoint.Endpoint.Address})
			if port := endpoint.Endpoint.Port; port != 0 {
				section.Params = append(section.Params, Param{Key: SQLConfigPort, Value: strconv.Itoa(port)})
			}

			break
		}

		for _, user := range file.Users {
			if user.Name != context.Context.User {
				continue
			}
			section.Params = append(section.Params,
				Param{Key: SQLConfigUsername, Value: user.BasicAuth.Username},
				Param{Key: SQLConfigPassword, Value: user.BasicAuth.Password},
				Param{Key: SQLConfigPasswordEncryption, Value: user.BasicAuth.PasswordEncryption},
			)

			break
		}

		sections = append(sections, section)
	}

	return sections, nil
}

const (
	SQLConfigAddress            = "address"
	SQLConfigPort               = "port"
	SQLConfigUsername           = "username"
	SQLConfigPassword           = "password"
	SQLConfigPasswordEncryption = "password-encryption"
)
