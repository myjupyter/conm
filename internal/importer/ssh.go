package importer

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/adrg/xdg"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/parser/params"
)

const (
	sshPatternMarks   = "*?!"
	sshNegation       = "!"
	sshHostNameOption = "HostName"
)

func readSSH(file string) ([]Entry, error) {
	blocks, err := ReadSSHConfig(file)
	if err != nil {
		return nil, err
	}

	var (
		entries []Entry
		seen    []string
	)
	for _, block := range blocks {
		for alias := range strings.FieldsSeq(block.Name) {
			if strings.ContainsAny(alias, sshPatternMarks) || slices.Contains(seen, alias) {
				continue
			}
			seen = append(seen, alias)
			entries = append(entries, sshEntry(alias, blocks))
		}
	}

	return entries, nil
}

func sshEntry(alias string, blocks []Section) Entry {
	entry := Entry{Name: alias}

	hostTokens := strings.NewReplacer("%%", "%", "%h", alias)

	var args []string
	for _, block := range blocks {
		if !sshMatches(block.Name, alias) {
			continue
		}
		for _, option := range block.Params {
			value := option.Value
			if strings.EqualFold(option.Key, sshHostNameOption) {
				value = hostTokens.Replace(value)
			}
			args = append(args, "-o", option.Key+"="+value)
		}
	}
	args = append(args, "-o", sshHostNameOption+"="+alias)

	res, err := params.ParseOptions(config.SSHConnType, args)
	if err != nil {
		entry.Err = err
		return entry
	}

	ssh, ok := res.Conn.(config.SSH)
	if !ok {
		entry.Err = fmt.Errorf("host was read as a %T", res.Conn)
		return entry
	}
	ssh.Metadata.Name = alias

	entry.Conn, entry.Warnings = ssh, res.Warnings

	return entry
}

func sshMatches(patterns, alias string) bool {
	matched := false
	for pattern := range strings.FieldsSeq(patterns) {
		pattern, negated := strings.CutPrefix(pattern, sshNegation)
		if ok, _ := path.Match(pattern, alias); ok {
			if negated {
				return false
			}
			matched = true
		}
	}

	return matched
}

const (
	sshClientConfigFileName = "config"
	sshConfigDirName        = ".ssh"
	sshHostKeyword          = "host"
	sshMatchKeyword         = "match"
	sshAnyHost              = "*"
)

func SSHConfigPath() string {
	return filepath.Join(xdg.Home, sshConfigDirName, sshClientConfigFileName)
}

func ReadSSHConfig(file string) ([]Section, error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	sections := []Section{{Name: sshAnyHost}}
	skipping := false
	for line := range strings.Lines(string(raw)) {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}

		keyword, value := config.SplitSSHOption(line)
		switch strings.ToLower(keyword) {
		case sshHostKeyword:
			sections, skipping = append(sections, Section{Name: value}), false
		case sshMatchKeyword:
			skipping = true
		default:
			if skipping {
				continue
			}
			last := &sections[len(sections)-1]
			last.Params = append(last.Params, Param{Key: keyword, Value: value})
		}
	}

	return sections, nil
}
