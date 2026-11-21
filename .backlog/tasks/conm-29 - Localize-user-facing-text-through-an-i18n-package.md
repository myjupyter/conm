---
id: CONM-29
title: Localize user-facing text through an i18n package
status: In Progress
assignee: []
created_date: '2026-10-20 16:40'
updated_date: '2026-10-20 17:34'
labels: []
dependencies: []
type: feature
ordinal: 25000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
User-facing text is hardcoded English scattered across the TUI and CLI, so conm cannot be shown in another language. Decided design: a leaf package internal/i18n built on golang.org/x/text/message; the English text itself is the message key (gettext style); translations live in embedded per-language TOML files (locales/<lang>.toml) with plain string values or a plural table (one/few/many/other); the language comes from the locale environment; a test guards that every translated key is still used in code. First step is the package and the guard; migrating screens follows.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 internal/i18n exposes a printf-style function that returns the translation for the detected language, or the formatted English text when none exists
- [x] #2 Translations load from embedded locales/<lang>.toml files; a string value is a plain translation, a table of CLDR plural forms selects on the first argument
- [x] #3 Language is detected from LC_ALL, LC_MESSAGES, LANG; C/POSIX/unset/unknown falls back to English
- [x] #4 A test fails when a locale contains a key that no i18n call in the code uses as a literal
- [x] #5 internal/i18n imports no other internal package; tests cover loading, plurals, fallback and detection
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Package internal/i18n: Sprintf over a message.Printer whose catalog is built at init from embedded locales/*.toml (no gotext/codegen). Plural tables map to plural.Selectf(1, "%d", ...). Language: first non-empty of LC_ALL/LC_MESSAGES/LANG, encoding/modifier stripped, matched against English + locale files; no match → English. TestLocales walks the module's non-test Go files, requires i18n.Sprintf's format to be a string literal, and fails on translated keys no call uses. First call sites: cobra Short texts in cmd/, translated in locales/ru.toml. Cobra's own template text (Usage:, Available Commands:) is not covered. Remaining: migrate TUI/form/error strings.

TUI/form migration: internal/ui and internal/ui/spec user-facing text now goes through i18n.Sprintf (305 keys in ru.toml, full coverage; en.toml holds the English plural forms). Translated values that act as identifiers (section titles, keyhint group titles, ConnectionEnabled/Disabled, statusReady) became vars compared through the same var. plural() helper removed in favour of plural tables. transport_test now expects the grammatical '1 field needs attention'. Still English: config.Validate* errors (config layer), network/repository errors, developer-only 'unexpected model' errors, cobra's own templates.
<!-- SECTION:NOTES:END -->
