package exporter

import (
	"strings"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/importer"
)

const (
	mysqlClientGroup = "client_"
	mysqlToolGroup   = "mysql_"
)

var mysqlSSLModes = map[config.MySQLTLSMode]string{
	config.MySQLTLSModeDisable:    "DISABLED",
	config.MySQLTLSModeSkipVerify: "REQUIRED",
	config.MySQLTLSModeVerify:     "VERIFY_IDENTITY",
}

var mysqlValueEscaper = strings.NewReplacer(`\`, `\\`, `"`, `\"`)

func mysqlSections(conn config.Connection, password string) ([]importer.Section, []string) {
	my, _ := conn.(config.MySQL)

	var client params
	client.add("host", my.Hostname)
	client.addPort("port", my.PortNumber)
	client.add("user", my.User)
	client.add("ssl-mode", mysqlSSLModes[my.TLSMode])
	if password != "" {
		client.add("password", `"`+mysqlValueEscaper.Replace(password)+`"`)
	}

	sections := []importer.Section{{Name: mysqlClientGroup + my.Metadata.Name, Params: client}}
	if my.DBName != "" {
		sections = append(sections, importer.Section{
			Name:   mysqlToolGroup + my.Metadata.Name,
			Params: params{{Key: "database", Value: my.DBName}},
		})
	}

	return sections, transportWarning(conn)
}
