package params

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
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
	ErrUnknownDatabase  = errors.New("couldn't tell which database this connects to")
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

const ambiguousClient config.ConnType = 0

var knownClients = func() map[string]config.ConnType {
	names := make(map[string]config.ConnType)
	for t := config.PostgresConnType; config.Clients(t) != nil; t++ {
		for _, name := range config.Clients(t) {
			if _, seen := names[name]; seen {
				names[name] = ambiguousClient
				continue
			}
			names[name] = t
		}
	}

	return names
}()

func Parse(input string) (Result, error) {
	req, err := newRequest(input)
	if err != nil {
		return Result{}, err
	}

	t, ok := requestType(req)
	if !ok {
		return Result{}, ErrUnknownDatabase
	}

	parse, ok := parsers[t]
	if !ok {
		return Result{}, fmt.Errorf("%w: %s", ErrUnsupported, t)
	}

	return parse(req)
}

func newRequest(input string) (request, error) {
	args, err := splitArgs(input)
	if err != nil {
		return request{}, err
	}
	if len(args) == 0 {
		return request{}, ErrEmptyInput
	}

	env, args := splitEnv(args)

	req := request{env: env, args: args}
	if name := clientName(args[0]); isClient(name) {
		req.client = name
		req.args = args[1:]
	}

	return req, nil
}

func requestType(req request) (config.ConnType, bool) {
	if t, ok := clientType(req.client); ok {
		return t, true
	}

	for _, arg := range req.args {
		if t, ok := schemeType(arg); ok {
			return t, true
		}
	}

	if hasPostgresKeywords(req.args) {
		return config.PostgresConnType, true
	}

	return ambiguousClient, false
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
	_, ok := knownClients[name]

	return ok
}

func clientType(name string) (config.ConnType, bool) {
	t, ok := knownClients[name]

	return t, ok && t != ambiguousClient
}

func schemeType(arg string) (config.ConnType, bool) {
	scheme, _, ok := strings.Cut(arg, "://")
	if !ok {
		return ambiguousClient, false
	}
	t, ok := schemes[strings.ToLower(scheme)]

	return t, ok
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
