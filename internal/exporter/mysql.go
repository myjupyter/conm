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

var mysqlValueEscaper = strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`)

func mysqlQuoted(value string) string {
	if value == "" {
		return ""
	}

	return `"` + mysqlValueEscaper.Replace(value) + `"`
}

func mysqlSections(conn config.Connection, password string) ([]importer.Section, []string) {
	my, _ := conn.(config.MySQL)

	var client params
	client.add("host", mysqlQuoted(my.Hostname))
	client.addPort("port", my.PortNumber)
	client.add("user", mysqlQuoted(my.User))
	client.add("loose-ssl-mode", mysqlSSLModes[my.TLSMode])
	client.add("password", mysqlQuoted(password))

	sections := []importer.Section{{Name: mysqlClientGroup + my.Metadata.Name, Params: client}}
	if my.DBName != "" {
		sections = append(sections, importer.Section{
			Name:   mysqlToolGroup + my.Metadata.Name,
			Params: params{{Key: "database", Value: mysqlQuoted(my.DBName)}},
		})
	}

	return sections, transportWarning(conn)
}
