package spec

import (
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/config"
)

const (
	linkKeyPrefix  = "link."
	linkNameSuffix = ".name"
	linkURLSuffix  = ".url"

	linkNameLabel = "name"
	linkURLLabel  = "url"

	linkNameExample = "grafana"
	linkURLExample  = "https://grafana.internal/d/db"
)

func LinkNameKey(i int) FormFieldKey {
	return linkKeyPrefix + strconv.Itoa(i) + linkNameSuffix
}

func LinkURLKey(i int) FormFieldKey {
	return linkKeyPrefix + strconv.Itoa(i) + linkURLSuffix
}

func IsLinkKey(key FormFieldKey) bool {
	_, ok := LinkIndexOf(key)
	return ok
}

func LinkIndexOf(key FormFieldKey) (int, bool) {
	rest, found := strings.CutPrefix(key, linkKeyPrefix)
	if !found {
		return 0, false
	}

	digits, isName := strings.CutSuffix(rest, linkNameSuffix)
	if !isName {
		var isURL bool
		if digits, isURL = strings.CutSuffix(rest, linkURLSuffix); !isURL {
			return 0, false
		}
	}

	i, err := strconv.Atoi(digits)
	if err != nil {
		return 0, false
	}
	return i, true
}

func LinkCount(values map[FormFieldKey]FormFieldValue) int {
	n := 0
	for {
		if _, ok := values[LinkNameKey(n)]; !ok {
			return n
		}
		n++
	}
}

func LinkNameField(i int) FormField {
	return FormField{
		Key:      LinkNameKey(i),
		Label:    linkNameLabel,
		Example:  linkNameExample,
		Property: OptionalFieldProperty,
	}
}

func LinkURLField(i int) FormField {
	return FormField{
		Key:      LinkURLKey(i),
		Label:    linkURLLabel,
		Example:  linkURLExample,
		Property: OptionalFieldProperty,
	}
}

func SeedLinks(links []config.WebLink, values map[FormFieldKey]FormFieldValue) {
	for i, l := range links {
		values[LinkNameKey(i)] = l.Name
		values[LinkURLKey(i)] = l.URL
	}
}

func LinksFrom(values map[FormFieldKey]FormFieldValue) []config.WebLink {
	var links []config.WebLink
	for i := range LinkCount(values) {
		name := strings.TrimSpace(values[LinkNameKey(i)])
		url := strings.TrimSpace(values[LinkURLKey(i)])
		if name == "" && url == "" {
			continue
		}
		links = append(links, config.WebLink{Name: name, URL: url})
	}
	return links
}

func ValidateLink(name, url string) (nameErr, urlErr error) {
	name, url = strings.TrimSpace(name), strings.TrimSpace(url)
	if name == "" && url == "" {
		return nil, nil
	}
	return config.ValidateWebLinkName(name), config.ValidateWebLinkURL(url)
}
