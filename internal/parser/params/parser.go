package params

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/myjupyter/conm/internal/config"
)

type Syntax int

const (
	UnknownSyntax Syntax = iota
	URISyntax
	KeywordSyntax
	FlagSyntax
)

var (
	ErrEmptyInput       = errors.New("nothing to parse")
	ErrTypeMismatch     = errors.New("the command line describes another database")
	ErrUnsupported      = errors.New("database is not supported by the parser yet")
	ErrUnterminated     = errors.New("unterminated quote")
	ErrMalformedKeyword = errors.New("malformed keyword/value pair")
)

var assignmentRegexp = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)

var keywordRegexp = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*\s*=`)

type Result struct {
	Syntax   Syntax
	Client   string
	Conn     config.Connection
	Warnings []string
}

type request struct {
	kind   config.ConnType
	client string
	env    map[string]string
	args   []string
}

type parseFunc func(request) (Result, error)

var parsers = map[config.ConnType]parseFunc{
	config.PostgresConnType: parsePostgres,
}

var schemes = map[string]config.ConnType{
	"postgres":   config.PostgresConnType,
	"postgresql": config.PostgresConnType,
}

var knownClients = func() map[string]bool {
	names := make(map[string]bool)
	for t := config.PostgresConnType; config.Clients(t) != nil; t++ {
		for _, name := range config.Clients(t) {
			names[name] = true
		}
	}

	return names
}()

func Parse(t config.ConnType, input string) (Result, error) {
	args, err := splitArgs(input)
	if err != nil {
		return Result{}, err
	}

	return ParseArgs(t, args)
}

func ParseArgs(t config.ConnType, args []string) (Result, error) {
	parse, ok := parsers[t]
	if !ok {
		return Result{}, fmt.Errorf("%w: %s", ErrUnsupported, t)
	}

	req, err := newRequest(t, args)
	if err != nil {
		return Result{}, err
	}

	return parse(req)
}

func newRequest(t config.ConnType, args []string) (request, error) {
	if len(args) == 0 {
		return request{}, ErrEmptyInput
	}

	env, args := splitEnv(args)

	req := request{kind: t, env: env, args: args}
	if len(args) == 0 {
		return req, nil
	}

	name := clientName(args[0])
	if !isClient(name) {
		return req, nil
	}
	if !slices.Contains(config.Clients(t), name) {
		return request{}, fmt.Errorf("%w: %s is a %s client", ErrTypeMismatch, name, clientDatabases(name))
	}

	req.client = name
	req.args = args[1:]

	return req, nil
}

func foreignScheme(t config.ConnType, value string) error {
	scheme, _, ok := strings.Cut(value, "://")
	if !ok {
		return nil
	}
	if schemes[strings.ToLower(scheme)] == t {
		return nil
	}

	return fmt.Errorf("%w: %s:// is not a %s connection string", ErrTypeMismatch, scheme, t)
}

func splitEnv(args []string) (map[string]string, []string) {
	var n int
	for n < len(args) && assignmentRegexp.MatchString(args[n]) {
		n++
	}
	if n == 0 || n == len(args) || !isClient(clientName(args[n])) {
		return nil, args
	}

	env := make(map[string]string, n)
	for _, arg := range args[:n] {
		name, value, _ := strings.Cut(arg, "=")
		env[name] = value
	}

	return env, args[n:]
}

func clientName(arg string) string {
	if arg == "" || strings.HasPrefix(arg, "-") {
		return ""
	}

	return filepath.Base(arg)
}

func isClient(name string) bool {
	return knownClients[name]
}

func clientDatabases(name string) string {
	var names []string
	for _, t := range config.Databases {
		if slices.Contains(config.Clients(t), name) {
			names = append(names, t.String())
		}
	}

	return strings.Join(names, "/")
}

func (s Syntax) String() string {
	switch s {
	case URISyntax:
		return "uri"
	case KeywordSyntax:
		return "keyword"
	case FlagSyntax:
		return "flags"
	default:
		return "unknown"
	}
}
