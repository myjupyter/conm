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
)

var keyModifiers = []string{"ctrl", "alt", "shift"}

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

func validateKey(key string) error {
	modifiers, base := splitKey(key)

	if err := validateModifiers(modifiers); err != nil {
		return fmt.Errorf("%q: %w", key, err)
	}

	if !isKeyName(base) {
		return fmt.Errorf("%q: %w %q", key, errUnknownKey, base)
	}
	return nil
}

func splitKey(key string) (modifiers []string, base string) {
	if key == "" {
		return nil, ""
	}

	cut := strings.LastIndex(key[:len(key)-1], "+")
	if cut < 0 {
		return nil, key
	}
	return strings.Split(key[:cut], "+"), key[cut+1:]
}

func validateModifiers(modifiers []string) error {
	for i, modifier := range modifiers {
		if !slices.Contains(keyModifiers, modifier) {
			return fmt.Errorf("%w %q", errUnknownModifier, modifier)
		}
		if slices.Contains(modifiers[:i], modifier) {
			return fmt.Errorf("%w %q", errRepeatedModifier, modifier)
		}
	}
	return nil
}

func isKeyName(base string) bool {
	if slices.Contains(namedKeys, base) {
		return true
	}

	char, size := utf8.DecodeRuneInString(base)
	return size > 0 && size == len(base) && char != ' ' && unicode.IsPrint(char)
}
