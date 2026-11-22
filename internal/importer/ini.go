package importer

import (
	"os"
	"strings"
)

func ReadINI(path string) ([]Section, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var sections []Section
	for line := range strings.Lines(string(raw)) {
		line = strings.TrimSpace(line)
		switch {
		case line == "", line[0] == '#', line[0] == ';':
		case line[0] == '[' && line[len(line)-1] == ']':
			sections = append(sections, Section{Name: strings.TrimSpace(line[1 : len(line)-1])})
		case len(sections) > 0:
			key, value, _ := strings.Cut(line, "=")
			last := &sections[len(sections)-1]
			last.Params = append(last.Params, Param{Key: strings.TrimSpace(key), Value: strings.TrimSpace(value)})
		}
	}

	return sections, nil
}
