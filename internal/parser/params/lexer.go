package params

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

type flagArity int

const (
	noValueFlag flagArity = iota
	valueFlag
	attachedValueFlag
)

type argFlag struct {
	name  string
	value string
}

type keywordPair struct {
	key   string
	value string
}

func splitArgs(input string) ([]string, error) {
	var (
		args    []string
		token   strings.Builder
		started bool
		quote   rune
	)

	runes := []rune(input)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		switch {
		case r == '\\' && quote != '\'' && i+1 < len(runes):
			i++
			token.WriteRune(runes[i])
			started = true
		case quote == 0 && (r == '\'' || r == '"'):
			quote = r
			started = true
		case quote == r:
			quote = 0
		case quote == 0 && unicode.IsSpace(r):
			if started {
				args = append(args, token.String())
				token.Reset()
				started = false
			}
		default:
			token.WriteRune(r)
			started = true
		}
	}

	if quote != 0 {
		return nil, fmt.Errorf("%w: %c", ErrUnterminated, quote)
	}
	if started {
		args = append(args, token.String())
	}

	return args, nil
}

func scanFlags(args []string, arity func(string) flagArity) ([]argFlag, []string) {
	var (
		flags      []argFlag
		positional []string
	)

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--":
			return flags, append(positional, args[i+1:]...)
		case strings.HasPrefix(arg, "--"):
			name, value, inline := strings.Cut(arg, "=")
			if !inline && arity(name) == valueFlag && i+1 < len(args) {
				i++
				value = args[i]
			}
			flags = append(flags, argFlag{name: name, value: value})
		case len(arg) > 1 && strings.HasPrefix(arg, "-"):
			rest := arg[1:]
			for rest != "" {
				name := "-" + rest[:1]
				rest = rest[1:]
				kind := arity(name)
				if kind == noValueFlag {
					flags = append(flags, argFlag{name: name})
					continue
				}

				value := rest
				rest = ""
				if value == "" && kind == valueFlag && i+1 < len(args) {
					i++
					value = args[i]
				}
				flags = append(flags, argFlag{name: name, value: value})
			}
		default:
			positional = append(positional, arg)
		}
	}

	return flags, positional
}

func splitKeywords(input string) ([]keywordPair, error) {
	var pairs []keywordPair

	runes := []rune(input)
	for i := skipSpace(runes, 0); i < len(runes); i = skipSpace(runes, i) {
		start := i
		for i < len(runes) && !unicode.IsSpace(runes[i]) && runes[i] != '=' {
			i++
		}
		key := string(runes[start:i])

		i = skipSpace(runes, i)
		if i >= len(runes) || runes[i] != '=' {
			return nil, fmt.Errorf("%w: %q has no value", ErrMalformedKeyword, key)
		}

		value, next, err := readKeywordValue(runes, skipSpace(runes, i+1))
		if err != nil {
			return nil, err
		}
		i = next

		pairs = append(pairs, keywordPair{key: strings.ToLower(key), value: value})
	}

	return pairs, nil
}

func readKeywordValue(runes []rune, i int) (string, int, error) {
	var value strings.Builder

	quoted := i < len(runes) && runes[i] == '\''
	if quoted {
		i++
	}

	for i < len(runes) {
		r := runes[i]
		switch {
		case r == '\\' && i+1 < len(runes):
			i++
			value.WriteRune(runes[i])
		case quoted && r == '\'':
			return value.String(), i + 1, nil
		case !quoted && unicode.IsSpace(r):
			return value.String(), i, nil
		default:
			value.WriteRune(r)
		}
		i++
	}

	if quoted {
		return "", i, fmt.Errorf("%w: '", ErrUnterminated)
	}

	return value.String(), i, nil
}

func skipSpace(runes []rune, i int) int {
	for i < len(runes) && unicode.IsSpace(runes[i]) {
		i++
	}

	return i
}

func splitHosts(authority string) []string {
	if authority == "" {
		return nil
	}

	var (
		hosts   []string
		start   int
		bracket bool
	)
	for i, r := range authority {
		switch {
		case r == '[':
			bracket = true
		case r == ']':
			bracket = false
		case r == ',' && !bracket:
			hosts = append(hosts, authority[start:i])
			start = i + 1
		}
	}

	return append(hosts, authority[start:])
}

func splitHostPort(host string) (string, string) {
	if strings.HasPrefix(host, "[") {
		if end := strings.Index(host, "]"); end >= 0 {
			port := strings.TrimPrefix(host[end+1:], ":")

			return host[1:end], port
		}
	}

	i := strings.LastIndex(host, ":")
	if i < 0 || !isDigits(host[i+1:]) {
		return host, ""
	}

	return host[:i], host[i+1:]
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}

func unescape(s string) string {
	decoded, err := url.PathUnescape(s)
	if err != nil {
		return s
	}

	return decoded
}
