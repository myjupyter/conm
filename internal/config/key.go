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

type Key struct {
	Ctrl  bool
	Alt   bool
	Shift bool
	Name  string
}

func ParseKey(raw string) (Key, error) {
	modifiers, name := splitKey(raw)

	key, err := withModifiers(modifiers)
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
		key.Shift = true
		char = unicode.ToLower(char)
	}
	if key.Shift && !key.Ctrl && !key.Alt && !unicode.IsLetter(char) {
		return Key{}, fmt.Errorf("%q: %w", raw, errShiftedSymbol)
	}

	key.Name = string(char)
	return key, nil
}

func splitKey(raw string) (modifiers []string, name string) {
	if raw == "" {
		return nil, ""
	}

	cut := strings.LastIndex(raw[:len(raw)-1], "+")
	if cut < 0 {
		return nil, raw
	}
	return strings.Split(raw[:cut], "+"), raw[cut+1:]
}

func withModifiers(modifiers []string) (Key, error) {
	var key Key
	for _, modifier := range modifiers {
		var held *bool
		switch modifier {
		case "ctrl":
			held = &key.Ctrl
		case "alt":
			held = &key.Alt
		case "shift":
			held = &key.Shift
		default:
			return Key{}, fmt.Errorf("%w %q", errUnknownModifier, modifier)
		}

		if *held {
			return Key{}, fmt.Errorf("%w %q", errRepeatedModifier, modifier)
		}
		*held = true
	}
	return key, nil
}
