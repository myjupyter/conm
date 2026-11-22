package exporter

import (
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/importer"
)

func postgresSections(conn config.Connection, password string) ([]importer.Section, []string) {
	pg, _ := conn.(config.Postgres)

	var service params
	service.add("host", pg.Hostname)
	service.addPort("port", pg.PortNumber)
	service.add("user", pg.User)
	service.add("dbname", pg.DBName)
	service.add("sslmode", pg.SSLMode)
	if pg.SchemaName != "" {
		service.add("options", "-c search_path="+pg.SchemaName)
	}
	service.add("password", password)

	return []importer.Section{{Name: pg.Metadata.Name, Params: service}}, transportWarning(conn)
}
