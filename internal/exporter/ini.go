package exporter

import "github.com/myjupyter/conm/internal/importer"

func iniMerged(_ string, raw []byte, sections []importer.Section) ([]byte, error) {
	return appended(raw, sections, func(name string) string { return "[" + name + "]" }, "", "=")
}
