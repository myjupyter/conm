package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

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

const linkNameMaxLength = 64

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
