package exporter

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/importer"
)

var entryNameRegexp = regexp.MustCompile(`^[A-Za-z0-9_.@-]+$`)

var (
	errNameUnfit = errors.New("name can't be an entry of the client's file: only letters, digits and _ . @ - can")
	errLineBreak = errors.New("a line break can't be written to this file")
)

type Entry struct {
	Name     string
	Sections []importer.Section
	Warnings []string
	Err      error
}

type target struct {
	path      func() string
	held      func(path string) ([]string, error)
	merged    func(path string, raw []byte, sections []importer.Section) ([]byte, error)
	sections  func(conn config.Connection, password string) ([]importer.Section, []string)
	lineBased bool
	passwords bool
}

var targets = map[config.ConnType]target{
	config.PostgresConnType: {
		path:      importer.PGServicePath,
		held:      heldBy(importer.ReadINI),
		merged:    iniMerged,
		sections:  postgresSections,
		lineBased: true,
		passwords: true,
	},
	config.MySQLConnType: {
		path:      importer.MySQLOptionPath,
		held:      heldBy(importer.ReadINI),
		merged:    iniMerged,
		sections:  mysqlSections,
		lineBased: true,
		passwords: true,
	},
	config.ClickHouseConnType: {
		path:      importer.ClickHouseUserPath,
		held:      heldBy(importer.ReadClickHouseConnections),
		merged:    clickhouseMerged,
		sections:  clickhouseSections,
		passwords: true,
	},
	config.MSSQLConnType: {
		path:      importer.SQLConfigPath,
		held:      sqlConfigHeld,
		merged:    sqlConfigMerged,
		sections:  mssqlSections,
		passwords: true,
	},
	config.SSHConnType: {
		path:      importer.SSHConfigPath,
		held:      sshHeld,
		merged:    sshMerged,
		sections:  sshSections,
		lineBased: true,
	},
}

func Supported(t config.ConnType) bool {
	_, ok := targets[t]
	return ok
}

func DefaultPath(t config.ConnType) string {
	return targets[t].path()
}

func WritesPasswords(t config.ConnType) bool {
	return targets[t].passwords
}

func Plan(t config.ConnType, path string, conns []config.Connection, passwords []string) ([]Entry, error) {
	to, ok := targets[t]
	if !ok {
		return nil, fmt.Errorf("%s has no config file to export to", t)
	}

	held, err := to.held(path)
	if err != nil {
		return nil, err
	}

	entries := make([]Entry, 0, len(conns))
	for i, conn := range conns {
		entry := Entry{Name: conn.Meta().Name}
		if !entryNameRegexp.MatchString(entry.Name) {
			entry.Err = errNameUnfit
			entries = append(entries, entry)

			continue
		}

		entry.Sections, entry.Warnings = to.sections(conn, passwords[i])
		if to.lineBased && slices.ContainsFunc(entry.Sections, hasLineBreak) {
			entry.Err = errLineBreak
		}
		for _, section := range entry.Sections {
			if slices.Contains(held, section.Name) {
				entry.Err = fmt.Errorf("%s already has %s", path, section.Name)
			}
			held = append(held, section.Name)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func Write(t config.ConnType, path string, entries []Entry) error {
	var sections []importer.Section
	for _, entry := range entries {
		if entry.Err == nil {
			sections = append(sections, entry.Sections...)
		}
	}
	if len(sections) == 0 {
		return nil
	}

	raw, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	merged, err := targets[t].merged(path, raw, sections)
	if err != nil {
		return fmt.Errorf("couldn't merge into %q: %w", path, err)
	}

	return config.WriteForeign(path, merged)
}

func heldBy(read func(string) ([]importer.Section, error)) func(string) ([]string, error) {
	return func(path string) ([]string, error) {
		sections, err := read(path)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}

		names := make([]string, 0, len(sections))
		for _, section := range sections {
			names = append(names, section.Name)
		}

		return names, nil
	}
}

func appended(raw []byte, sections []importer.Section, header func(name string) string, indent, separator string) ([]byte, error) {
	added, err := textBlocks(sections, header, indent, separator)
	if err != nil {
		return nil, err
	}

	if len(raw) > 0 {
		if !bytes.HasSuffix(raw, []byte("\n")) {
			raw = append(slices.Clone(raw), '\n')
		}
		raw = append(slices.Clone(raw), '\n')
	}

	return append(raw, added...), nil
}

func textBlocks(sections []importer.Section, header func(name string) string, indent, separator string) (string, error) {
	blocks := make([]string, 0, len(sections))
	for _, section := range sections {
		if hasLineBreak(section) {
			return "", fmt.Errorf("%s: %w", section.Name, errLineBreak)
		}

		var block strings.Builder
		block.WriteString(header(section.Name) + "\n")
		for _, param := range section.Params {
			block.WriteString(indent + param.Key + separator + param.Value + "\n")
		}

		blocks = append(blocks, block.String())
	}

	return strings.Join(blocks, "\n"), nil
}

func hasLineBreak(section importer.Section) bool {
	const lineBreaks = "\r\n"

	return strings.ContainsAny(section.Name, lineBreaks) || slices.ContainsFunc(section.Params, func(param importer.Param) bool {
		return strings.ContainsAny(param.Key+param.Value, lineBreaks)
	})
}

type params []importer.Param

func (p *params) add(key, value string) {
	if value != "" {
		*p = append(*p, importer.Param{Key: key, Value: value})
	}
}

func (p *params) addPort(key string, port int) {
	if port != 0 {
		p.add(key, strconv.Itoa(port))
	}
}

func transportWarning(conn config.Connection) []string {
	routed, ok := conn.(interface{ TransportRef() string })
	if !ok || routed.TransportRef() == "" {
		return nil
	}

	return []string{fmt.Sprintf("transport %s is not exported: the client would connect directly", routed.TransportRef())}
}
