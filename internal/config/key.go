package config

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	errUnknownModifier  = errors.New("unknown modifier")
	errRepeatedModifier = errors.New("repeated modifier")
	errUnknownKey       = errors.New("unknown key")
	errShiftedSymbol    = errors.New("shift with a symbol depends on the keyboard layout, write the symbol itself")
)

var namedKeys = func() []string {
	names := []string{
		"enter", "esc", "tab", "space", "backspace", "delete", "insert",
		"up", "down", "left", "right", "home", "end", "pgup", "pgdown",
	}
	for n := 1; n <= 12; n++ {
		names = append(names, "f"+strconv.Itoa(n))
	}
	return names
}()

func NamedKeys() []string {
	return slices.Clone(namedKeys)
}

type Mod uint8

const (
	ModCtrl Mod = 1 << iota
	ModAlt
	ModShift
	ModMeta
	ModHyper
	ModSuper
)

type modifier struct {
	mod  Mod
	name string
}

var modifiers = []modifier{
	{ModCtrl, "ctrl"},
	{ModAlt, "alt"},
	{ModShift, "shift"},
	{ModMeta, "meta"},
	{ModHyper, "hyper"},
	{ModSuper, "super"},
}

func (m Mod) String() string {
	names := make([]string, 0, len(modifiers))
	for _, held := range modifiers {
		if m&held.mod != 0 {
			names = append(names, held.name)
		}
	}
	return strings.Join(names, "+")
}

type Key struct {
	Mod  Mod
	Name string
}

func ParseKey(raw string) (Key, error) {
	held, name := splitKey(raw)

	key, err := withModifiers(held)
	if err != nil {
		return Key{}, fmt.Errorf("%q: %w", raw, err)
	}

	if slices.Contains(namedKeys, name) {
		key.Name = name
		return key, nil
	}

	char, size := utf8.DecodeRuneInString(name)
	if size == 0 || size != len(name) || char == ' ' || !unicode.IsPrint(char) {
		return Key{}, fmt.Errorf("%q: %w %q", raw, errUnknownKey, name)
	}

	if unicode.IsUpper(char) {
		key.Mod |= ModShift
		char = unicode.ToLower(char)
	}
	if key.Mod == ModShift && !unicode.IsLetter(char) {
		return Key{}, fmt.Errorf("%q: %w", raw, errShiftedSymbol)
	}

	key.Name = string(char)
	return key, nil
}

func splitKey(raw string) (held []string, name string) {
	if raw == "" {
		return nil, ""
	}

	cut := strings.LastIndex(raw[:len(raw)-1], "+")
	if cut < 0 {
		return nil, raw
	}
	return strings.Split(raw[:cut], "+"), raw[cut+1:]
}

func withModifiers(names []string) (Key, error) {
	var key Key
	for _, name := range names {
		i := slices.IndexFunc(modifiers, func(m modifier) bool { return m.name == name })
		if i < 0 {
			return Key{}, fmt.Errorf("%w %q", errUnknownModifier, name)
		}

		held := modifiers[i].mod
		if key.Mod&held != 0 {
			return Key{}, fmt.Errorf("%w %q", errRepeatedModifier, name)
		}
		key.Mod |= held
	}
	return key, nil
}
