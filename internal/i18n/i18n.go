package i18n

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

const localesGlob = "locales/*.toml"

//go:embed locales/*.toml
var locales embed.FS

var (
	localeVariables = []string{"LC_ALL", "LC_MESSAGES", "LANG"}
	pluralForms     = []string{"zero", "one", "two", "few", "many", "other"}
	printer         = mustPrinter(locales, os.Getenv)
)

func Sprintf(format string, args ...any) string {
	return printer.Sprintf(format, args...)
}

func mustPrinter(fsys fs.FS, getenv func(string) string) *message.Printer {
	p, err := newPrinter(fsys, getenv)
	if err != nil {
		panic(err)
	}

	return p
}

func newPrinter(fsys fs.FS, getenv func(string) string) (*message.Printer, error) {
	messages, tags, err := load(fsys)
	if err != nil {
		return nil, err
	}

	return message.NewPrinter(detect(tags, getenv), message.Catalog(messages)), nil
}

func load(fsys fs.FS) (*catalog.Builder, []language.Tag, error) {
	files, err := fs.Glob(fsys, localesGlob)
	if err != nil {
		return nil, nil, err
	}

	messages := catalog.NewBuilder()
	tags := []language.Tag{language.English}
	for _, file := range files {
		tag, err := loadLocale(fsys, file, messages)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", file, err)
		}
		if tag != language.English {
			tags = append(tags, tag)
		}
	}

	return messages, tags, nil
}

func loadLocale(fsys fs.FS, file string, messages *catalog.Builder) (language.Tag, error) {
	tag, err := language.Parse(strings.TrimSuffix(path.Base(file), path.Ext(file)))
	if err != nil {
		return language.Und, err
	}

	translations, err := readTranslations(fsys, file)
	if err != nil {
		return language.Und, err
	}

	for key, translation := range translations {
		if err := setTranslation(messages, tag, key, translation); err != nil {
			return language.Und, fmt.Errorf("%q: %w", key, err)
		}
	}

	return tag, nil
}

func readTranslations(fsys fs.FS, file string) (map[string]any, error) {
	data, err := fs.ReadFile(fsys, file)
	if err != nil {
		return nil, err
	}

	var translations map[string]any
	if err := toml.Unmarshal(data, &translations); err != nil {
		return nil, err
	}

	return translations, nil
}

func setTranslation(messages *catalog.Builder, tag language.Tag, key string, translation any) error {
	switch translation := translation.(type) {
	case string:
		return messages.SetString(tag, key, translation)
	case map[string]any:
		cases, err := pluralCases(translation)
		if err != nil {
			return err
		}

		return messages.Set(tag, key, plural.Selectf(1, "%d", cases...))
	default:
		return errTranslationType
	}
}

func pluralCases(forms map[string]any) ([]any, error) {
	if _, ok := forms["other"]; !ok {
		return nil, errMissingOtherForm
	}

	cases := make([]any, 0, 2*len(forms))
	for _, form := range pluralForms {
		translation, ok := forms[form]
		if !ok {
			continue
		}
		text, ok := translation.(string)
		if !ok {
			return nil, errTranslationType
		}
		cases = append(cases, form, text)
	}
	if len(cases) != 2*len(forms) {
		return nil, errUnknownPluralForm
	}

	return cases, nil
}

func detect(tags []language.Tag, getenv func(string) string) language.Tag {
	for _, variable := range localeVariables {
		if locale := getenv(variable); locale != "" {
			return match(tags, locale)
		}
	}

	return language.English
}

func match(tags []language.Tag, locale string) language.Tag {
	locale, _, _ = strings.Cut(locale, ".")
	locale, _, _ = strings.Cut(locale, "@")
	wanted, err := language.Parse(strings.ReplaceAll(locale, "_", "-"))
	if err != nil {
		return language.English
	}

	_, index, confidence := language.NewMatcher(tags).Match(wanted)
	if confidence == language.No {
		return language.English
	}

	return tags[index]
}
