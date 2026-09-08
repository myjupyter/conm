package spec

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/config"
)

func metadataSection(nameKey, descriptionKey, tagsKey FormFieldKey) FormSection {
	return FormSection{
		Title:  MetadataSectionTitle,
		Note:   metadataSectionNote,
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
		Tags:        parseTags(values[tagsKey]),
		Links:       LinksFrom(values),
	}
}

func parseTags(raw string) []string {
	var tags []string
	for t := range strings.SplitSeq(raw, ",") {
		if t = strings.TrimSpace(t); t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}
