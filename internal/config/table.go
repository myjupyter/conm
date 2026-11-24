package config

import (
	"fmt"
	"slices"
)

const (
	NameColumn     = "name"
	UsernameColumn = "username"
	HostColumn     = "host"
	PortColumn     = "port"
	DatabaseColumn = "database"
	PingColumn     = "ping"
	TagsColumn     = "tags"
)

var DefaultColumns = []string{NameColumn, UsernameColumn, HostColumn, PortColumn, DatabaseColumn, PingColumn}

var TableColumns = append(slices.Clone(DefaultColumns), TagsColumn)

func (t ConnType) Columns() []string {
	if t.Kind() == DatabaseConnKind {
		return slices.Clone(TableColumns)
	}
	return slices.DeleteFunc(slices.Clone(TableColumns), func(c string) bool { return c == DatabaseColumn })
}

func (t ConnType) DefaultColumns() []string {
	return slices.DeleteFunc(t.Columns(), func(c string) bool { return !slices.Contains(DefaultColumns, c) })
}

func (s ConnectionSettings) VisibleColumns() []string {
	if s.Columns == nil {
		return s.Type.DefaultColumns()
	}
	return s.Columns
}

func (s ConnectionSettings) ResolvedTabColor() string {
	if s.TabColor == "" {
		return s.Type.String()
	}
	return s.TabColor
}

func IsTabColor(value string) bool {
	_, err := ParseConnType(value)
	return err == nil || IsColor(value)
}

func (s ConnectionSettings) ValidateTable() error {
	available := s.Type.Columns()
	for i, c := range s.Columns {
		if !slices.Contains(available, c) {
			return fmt.Errorf("%s has no %q column", s.Type, c)
		}
		if slices.Contains(s.Columns[:i], c) {
			return fmt.Errorf("%s lists the %q column twice", s.Type, c)
		}
	}
	if s.Columns != nil && !slices.Contains(s.Columns, NameColumn) {
		return fmt.Errorf("%s must show the %q column", s.Type, NameColumn)
	}
	if s.TabColor != "" && !IsTabColor(s.TabColor) {
		return fmt.Errorf("%s tab colour %q is neither a type, a colour name nor #rrggbb", s.Type, s.TabColor)
	}
	return nil
}

func (s *ConnectionSettings) completeTable() {
	if s.Columns != nil {
		available := s.Type.Columns()
		var kept []string
		for _, c := range s.Columns {
			if slices.Contains(available, c) && !slices.Contains(kept, c) {
				kept = append(kept, c)
			}
		}
		if !slices.Contains(kept, NameColumn) {
			kept = append([]string{NameColumn}, kept...)
		}
		s.Columns = kept
	}
	if s.TabColor != "" && !IsTabColor(s.TabColor) {
		s.TabColor = ""
	}
}
