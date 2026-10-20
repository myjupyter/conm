package i18n

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const russian = `
"Add a %s connection" = "Добавить подключение %s"

["Removed %d connections"]
one   = "Удалено %d подключение"
few   = "Удалено %d подключения"
many  = "Удалено %d подключений"
other = "Удалено %d подключения"
`

const english = `
["Removed %d connections"]
one   = "Removed %d connection"
other = "Removed %d connections"
`

func locale(files map[string]string) fstest.MapFS {
	fsys := fstest.MapFS{}
	for name, content := range files {
		fsys["locales/"+name] = &fstest.MapFile{Data: []byte(content)}
	}

	return fsys
}

func Test_newPrinter(t *testing.T) {
	t.Parallel()

	files := map[string]string{"ru.toml": russian, "en.toml": english}
	type args struct {
		env    map[string]string
		format string
		args   []any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no locale set",
			args: args{format: "Add a %s connection", args: []any{"postgres"}},
			want: "Add a postgres connection",
		},
		{
			name: "language from LANG with region and encoding",
			args: args{env: map[string]string{"LANG": "ru_RU.UTF-8"}, format: "Add a %s connection", args: []any{"postgres"}},
			want: "Добавить подключение postgres",
		},
		{
			name: "language with modifier",
			args: args{env: map[string]string{"LANG": "ru_RU.UTF-8@euro"}, format: "Add a %s connection", args: []any{"redis"}},
			want: "Добавить подключение redis",
		},
		{
			name: "LC_MESSAGES wins over LANG",
			args: args{env: map[string]string{"LC_MESSAGES": "ru_RU", "LANG": "en_US"}, format: "Add a %s connection", args: []any{"mysql"}},
			want: "Добавить подключение mysql",
		},
		{
			name: "LC_ALL wins over LC_MESSAGES",
			args: args{env: map[string]string{"LC_ALL": "C.UTF-8", "LC_MESSAGES": "ru_RU"}, format: "Add a %s connection", args: []any{"mysql"}},
			want: "Add a mysql connection",
		},
		{
			name: "POSIX locale",
			args: args{env: map[string]string{"LANG": "POSIX"}, format: "Add a %s connection", args: []any{"mysql"}},
			want: "Add a mysql connection",
		},
		{
			name: "language without a locale file",
			args: args{env: map[string]string{"LANG": "de_DE.UTF-8"}, format: "Add a %s connection", args: []any{"mysql"}},
			want: "Add a mysql connection",
		},
		{
			name: "untranslated text",
			args: args{env: map[string]string{"LANG": "ru_RU"}, format: "Ping a %s connection", args: []any{"mysql"}},
			want: "Ping a mysql connection",
		},
		{
			name: "russian plural one",
			args: args{env: map[string]string{"LANG": "ru_RU"}, format: "Removed %d connections", args: []any{21}},
			want: "Удалено 21 подключение",
		},
		{
			name: "russian plural few",
			args: args{env: map[string]string{"LANG": "ru_RU"}, format: "Removed %d connections", args: []any{3}},
			want: "Удалено 3 подключения",
		},
		{
			name: "russian plural many",
			args: args{env: map[string]string{"LANG": "ru_RU"}, format: "Removed %d connections", args: []any{11}},
			want: "Удалено 11 подключений",
		},
		{
			name: "english plural one",
			args: args{format: "Removed %d connections", args: []any{1}},
			want: "Removed 1 connection",
		},
		{
			name: "english plural under a language without a locale file",
			args: args{env: map[string]string{"LANG": "de_DE"}, format: "Removed %d connections", args: []any{2}},
			want: "Removed 2 connections",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p, err := newPrinter(locale(files), func(name string) string { return tt.args.env[name] })
			require.NoError(t, err)
			assert.Equal(t, tt.want, p.Sprintf(tt.args.format, tt.args.args...))
		})
	}
}

func Test_load(t *testing.T) {
	t.Parallel()

	type args struct {
		files map[string]string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{name: "no locale files", args: args{}},
		{name: "plain and plural translations", args: args{files: map[string]string{"ru.toml": russian}}},
		{
			name:    "plural table without other",
			args:    args{files: map[string]string{"ru.toml": "[\"%d rows\"]\none = \"%d строка\"\n"}},
			wantErr: errMissingOtherForm,
		},
		{
			name:    "unknown plural form",
			args:    args{files: map[string]string{"ru.toml": "[\"%d rows\"]\nseveral = \"%d строки\"\nother = \"%d строк\"\n"}},
			wantErr: errUnknownPluralForm,
		},
		{
			name:    "plural form that is not a string",
			args:    args{files: map[string]string{"ru.toml": "[\"%d rows\"]\none = 1\nother = \"%d строк\"\n"}},
			wantErr: errTranslationType,
		},
		{
			name:    "translation that is not a string",
			args:    args{files: map[string]string{"ru.toml": "\"rows\" = 1\n"}},
			wantErr: errTranslationType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := load(locale(tt.args.files))
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func Test_load_malformedFile(t *testing.T) {
	t.Parallel()

	type args struct {
		files map[string]string
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "file name that is not a language", args: args{files: map[string]string{"not a language.toml": ""}}},
		{name: "invalid TOML", args: args{files: map[string]string{"ru.toml": "\"rows\" = \n"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := load(locale(tt.args.files))
			assert.Error(t, err)
		})
	}
}

func TestLocales(t *testing.T) {
	t.Parallel()

	_, _, err := load(locales)
	require.NoError(t, err)
	used := literalFormats(t, filepath.Join("..", ".."))
	files, err := fs.Glob(locales, localesGlob)
	require.NoError(t, err)

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			t.Parallel()

			translations, err := readTranslations(locales, file)
			require.NoError(t, err)
			for key := range translations {
				assert.Contains(t, used, key, "translated text is never passed to i18n.Sprintf")
			}
		})
	}
}

func literalFormats(t *testing.T, root string) map[string]bool {
	t.Helper()

	used := map[string]bool{}
	fset := token.NewFileSet()
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && strings.HasPrefix(entry.Name(), ".") && path != root {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok || !isI18nSprintf(call) || len(call.Args) == 0 {
				return true
			}
			literal, ok := call.Args[0].(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				assert.Fail(t, "i18n.Sprintf needs a string literal format", fset.Position(call.Pos()).String())
				return true
			}
			format, err := strconv.Unquote(literal.Value)
			require.NoError(t, err)
			used[format] = true

			return true
		})

		return nil
	})
	require.NoError(t, err)

	return used
}

func isI18nSprintf(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Sprintf" {
		return false
	}
	pkg, ok := selector.X.(*ast.Ident)

	return ok && pkg.Name == "i18n"
}
