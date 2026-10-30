package config

import (
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
)

var ReservedNames = []string{"add", "edit", "remove", "ping", "list", "exec", "help"}

type WebLink struct {
	Name string `toml:"name"`
	URL  string `toml:"url"`
}

type ConnMeta struct {
	Name        string    `toml:"name,omitempty"`
	Description string    `toml:"description,omitempty"`
	Tags        []string  `toml:"tags,omitempty"`
	Links       []WebLink `toml:"links,omitempty"`
}

const (
	linkNameMaxLength = 64
	tagMaxLength      = 32
)

func ValidateWebLinkName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("link name can't be empty")
	}
	if len(name) > linkNameMaxLength {
		return fmt.Errorf("link name must be at most %d characters long", linkNameMaxLength)
	}
	return nil
}

func ValidateWebLinkURL(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return errors.New("link url can't be empty")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return errors.New("link url is not a valid url")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("link url must start with http:// or https://")
	}
	if u.Host == "" {
		return errors.New("link url must have a host")
	}
	return nil
}

func ValidateTags(tags []string) error {
	seen := make(map[string]bool, len(tags))
	for _, tag := range tags {
		if len(tag) > tagMaxLength {
			return fmt.Errorf("tag '%s' is longer that %d symbols", tag, tagMaxLength)
		}
		if seen[tag] {
			return fmt.Errorf("tag '%s' is set more than once", tag)
		}
		seen[tag] = true
	}
	return nil
}

func ValidateReservedName(name string) error {
	if slices.Contains(ReservedNames, name) {
		return fmt.Errorf("name '%s' is reserved and cannot be used", name)
	}
	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("name '%s' cannot start with '-'", name)
	}
	return nil
}
