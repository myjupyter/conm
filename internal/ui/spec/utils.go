package spec

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/config"
)

const StringArraySeparator = ","

func metadataSection(nameKey, descriptionKey, tagsKey FormFieldKey) FormSection {
	return FormSection{
		Title:  MetadataSectionTitle,
		Fields: []FormFieldKey{nameKey, descriptionKey, tagsKey},
	}
}

func intValue(values map[FormFieldKey]FormFieldValue, key FormFieldKey, fallback int) (int, error) {
	raw, ok := values[key]
	if !ok || raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s value: %w", key, err)
	}

	return value, nil
}

func buildMeta(values map[FormFieldKey]FormFieldValue, nameKey, descriptionKey, tagsKey FormFieldKey) config.ConnMeta {
	return config.ConnMeta{
		Name:        values[nameKey],
		Description: values[descriptionKey],
		Tags:        SplitStringArray(values[tagsKey]),
		Links:       LinksFrom(values),
	}
}

func JoinStringArray(items []string) FormFieldValue {
	return strings.Join(items, StringArraySeparator+" ")
}

func SplitStringArray(raw FormFieldValue) []string {
	var items []string
	for item := range strings.SplitSeq(raw, StringArraySeparator) {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}
