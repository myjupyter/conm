package params

import (
	"errors"
	"reflect"
	"testing"
)

func TestSplitArgs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{name: "empty", input: ""},
		{name: "only spaces", input: "   \t\n "},
		{
			name:  "plain words",
			input: `psql -h localhost`,
			want:  []string{"psql", "-h", "localhost"},
		},
		{
			name:  "runs of whitespace separate one token",
			input: "psql\t-h  \n localhost",
			want:  []string{"psql", "-h", "localhost"},
		},
		{
			name:  "double quotes hold a space",
			input: `psql "host=localhost port=5432"`,
			want:  []string{"psql", "host=localhost port=5432"},
		},
		{
			name:  "single quotes hold a space",
			input: `psql 'host=localhost port=5432'`,
			want:  []string{"psql", "host=localhost port=5432"},
		},
		{
			name:  "quotes glue to the surrounding token",
			input: `--host="local host"`,
			want:  []string{"--host=local host"},
		},
		{
			name:  "an empty quoted token is still a token",
			input: `psql ""`,
			want:  []string{"psql", ""},
		},
		{
			name:  "single quotes survive inside double quotes",
			input: `"options='-c search_path=public'"`,
			want:  []string{"options='-c search_path=public'"},
		},
		{
			name:  "double quotes survive inside single quotes",
			input: `'say "hi"'`,
			want:  []string{`say "hi"`},
		},
		{
			name:  "a backslash escapes a separator",
			input: `my\ db`,
			want:  []string{"my db"},
		},
		{
			name:  "a backslash escapes a quote inside double quotes",
			input: `"a\"b"`,
			want:  []string{`a"b`},
		},
		{
			name:  "a backslash is literal inside single quotes",
			input: `'a\b'`,
			want:  []string{`a\b`},
		},
		{
			name:  "a trailing backslash is literal",
			input: `a\`,
			want:  []string{`a\`},
		},
		{
			name:  "quotes may open and close several times in one token",
			input: `-h"local"'host'`,
			want:  []string{"-hlocalhost"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := splitArgs(test.input)
			if err != nil {
				t.Fatalf("splitArgs(%q) failed: %v", test.input, err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("splitArgs(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestSplitArgsUnterminatedQuote(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "double", input: `psql "postgres://h/db`},
		{name: "single", input: `psql 'postgres://h/db`},
		{name: "closed by the other kind", input: `psql "abc'`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := splitArgs(test.input)
			if !errors.Is(err, ErrUnterminated) {
				t.Fatalf("splitArgs(%q) error = %v, want %v", test.input, err, ErrUnterminated)
			}
			if got != nil {
				t.Errorf("splitArgs(%q) = %q, want no args alongside the error", test.input, got)
			}
		})
	}
}

func TestScanFlags(t *testing.T) {
	valued := func(name string) bool {
		switch name {
		case "-h", "--host", "-p", "-d", "--dbname", "-c":
			return true
		default:
			return false
		}
	}

	tests := []struct {
		name       string
		args       []string
		flags      []argFlag
		positional []string
	}{
		{name: "no args"},
		{
			name:  "short flag takes the next arg",
			args:  []string{"-h", "localhost"},
			flags: []argFlag{{name: "-h", value: "localhost"}},
		},
		{
			name:  "short flag takes an attached value",
			args:  []string{"-hlocalhost"},
			flags: []argFlag{{name: "-h", value: "localhost"}},
		},
		{
			name:  "long flag takes the next arg",
			args:  []string{"--host", "localhost"},
			flags: []argFlag{{name: "--host", value: "localhost"}},
		},
		{
			name:  "long flag takes an inline value",
			args:  []string{"--host=localhost"},
			flags: []argFlag{{name: "--host", value: "localhost"}},
		},
		{
			name:       "an inline empty value is not filled from the next arg",
			args:       []string{"--dbname=", "mydb"},
			flags:      []argFlag{{name: "--dbname", value: ""}},
			positional: []string{"mydb"},
		},
		{
			name:       "an unknown long flag carries no value",
			args:       []string{"--single-connection", "mydb"},
			flags:      []argFlag{{name: "--single-connection", value: ""}},
			positional: []string{"mydb"},
		},
		{
			name:  "valueless short flags cluster",
			args:  []string{"-lw"},
			flags: []argFlag{{name: "-l"}, {name: "-w"}},
		},
		{
			name:  "a cluster may end in a valued flag",
			args:  []string{"-lh", "localhost"},
			flags: []argFlag{{name: "-l"}, {name: "-h", value: "localhost"}},
		},
		{
			name:  "a valued flag at the end of the args gets nothing",
			args:  []string{"-h"},
			flags: []argFlag{{name: "-h", value: ""}},
		},
		{
			name:       "a consumed value is never a positional",
			args:       []string{"-c", "select 1", "mydb"},
			flags:      []argFlag{{name: "-c", value: "select 1"}},
			positional: []string{"mydb"},
		},
		{
			name:       "positionals keep their order",
			args:       []string{"mydb", "me"},
			positional: []string{"mydb", "me"},
		},
		{
			name:       "a lone dash is a positional",
			args:       []string{"-"},
			positional: []string{"-"},
		},
		{
			name:       "everything after a double dash is positional",
			args:       []string{"-h", "localhost", "--", "-p", "5432"},
			flags:      []argFlag{{name: "-h", value: "localhost"}},
			positional: []string{"-p", "5432"},
		},
		{
			name: "a trailing double dash adds nothing",
			args: []string{"--"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			flags, positional := scanFlags(test.args, valued)
			if !reflect.DeepEqual(flags, test.flags) {
				t.Errorf("flags = %+v, want %+v", flags, test.flags)
			}
			if !reflect.DeepEqual(positional, test.positional) {
				t.Errorf("positional = %q, want %q", positional, test.positional)
			}
		})
	}
}

func TestSplitKeywords(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []keywordPair
	}{
		{name: "empty", input: ""},
		{name: "only spaces", input: "  \t "},
		{
			name:  "pairs",
			input: `host=localhost port=5432 dbname=mydb`,
			want: []keywordPair{
				{key: "host", value: "localhost"},
				{key: "port", value: "5432"},
				{key: "dbname", value: "mydb"},
			},
		},
		{
			name:  "spaces around the equals sign",
			input: `host = localhost   port =5432`,
			want: []keywordPair{
				{key: "host", value: "localhost"},
				{key: "port", value: "5432"},
			},
		},
		{
			name:  "keys are folded to lower case",
			input: `HOST=localhost DBName=mydb`,
			want: []keywordPair{
				{key: "host", value: "localhost"},
				{key: "dbname", value: "mydb"},
			},
		},
		{
			name:  "a quoted value holds spaces and an equals sign",
			input: `host=h options='-c search_path=public' port=5432`,
			want: []keywordPair{
				{key: "host", value: "h"},
				{key: "options", value: "-c search_path=public"},
				{key: "port", value: "5432"},
			},
		},
		{
			name:  "an empty value",
			input: `host=localhost password=`,
			want: []keywordPair{
				{key: "host", value: "localhost"},
				{key: "password", value: ""},
			},
		},
		{
			name:  "an empty quoted value",
			input: `password='' host=h`,
			want: []keywordPair{
				{key: "password", value: ""},
				{key: "host", value: "h"},
			},
		},
		{
			name:  "a backslash escapes a separator",
			input: `password=a\ b host=h`,
			want: []keywordPair{
				{key: "password", value: "a b"},
				{key: "host", value: "h"},
			},
		},
		{
			name:  "a backslash escapes the closing quote",
			input: `password='a\'b' host=h`,
			want: []keywordPair{
				{key: "password", value: "a'b"},
				{key: "host", value: "h"},
			},
		},
		{
			name:  "leading and trailing whitespace is ignored",
			input: "  host=localhost  ",
			want:  []keywordPair{{key: "host", value: "localhost"}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := splitKeywords(test.input)
			if err != nil {
				t.Fatalf("splitKeywords(%q) failed: %v", test.input, err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("splitKeywords(%q) = %+v, want %+v", test.input, got, test.want)
			}
		})
	}
}

func TestSplitKeywordsErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  error
	}{
		{name: "a key without a value", input: `host`, want: ErrMalformedKeyword},
		{name: "a trailing key without a value", input: `host=h port`, want: ErrMalformedKeyword},
		{name: "a key whose equals sign never arrives", input: `host h`, want: ErrMalformedKeyword},
		{name: "an unterminated quoted value", input: `password='abc`, want: ErrUnterminated},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := splitKeywords(test.input)
			if !errors.Is(err, test.want) {
				t.Fatalf("splitKeywords(%q) error = %v, want %v", test.input, err, test.want)
			}
			if got != nil {
				t.Errorf("splitKeywords(%q) = %+v, want no pairs alongside the error", test.input, got)
			}
		})
	}
}

func TestReadKeywordValue(t *testing.T) {
	tests := []struct {
		name  string
		input string
		from  int
		want  string
		next  int
	}{
		{name: "unquoted to the end", input: "localhost", want: "localhost", next: 9},
		{name: "unquoted stops at a space", input: "localhost port=5432", want: "localhost", next: 9},
		{name: "quoted stops after the closing quote", input: "'a b' rest", want: "a b", next: 5},
		{name: "reads from the given index", input: "host=h", from: 5, want: "h", next: 6},
		{name: "an empty unquoted value", input: " port=5432", want: "", next: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, next, err := readKeywordValue([]rune(test.input), test.from)
			if err != nil {
				t.Fatalf("readKeywordValue(%q, %d) failed: %v", test.input, test.from, err)
			}
			if value != test.want {
				t.Errorf("value = %q, want %q", value, test.want)
			}
			if next != test.next {
				t.Errorf("next = %d, want %d", next, test.next)
			}
		})
	}
}

func TestSkipSpace(t *testing.T) {
	tests := []struct {
		name  string
		input string
		from  int
		want  int
	}{
		{name: "already on a word", input: "host", want: 0},
		{name: "skips to the word", input: "   host", want: 3},
		{name: "skips tabs and newlines", input: " \t\n host", want: 4},
		{name: "skips to the end", input: "host   ", from: 4, want: 7},
		{name: "past the end stays put", input: "host", from: 9, want: 9},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := skipSpace([]rune(test.input), test.from); got != test.want {
				t.Errorf("skipSpace(%q, %d) = %d, want %d", test.input, test.from, got, test.want)
			}
		})
	}
}

func TestSplitHosts(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{name: "empty"},
		{name: "one host", input: "localhost", want: []string{"localhost"}},
		{name: "one host with a port", input: "localhost:5432", want: []string{"localhost:5432"}},
		{
			name:  "a seed list",
			input: "host1:5432,host2:5432",
			want:  []string{"host1:5432", "host2:5432"},
		},
		{
			name:  "a seed list without ports",
			input: "host1,host2,host3",
			want:  []string{"host1", "host2", "host3"},
		},
		{
			name:  "a comma inside brackets does not split",
			input: "[::1]:5432",
			want:  []string{"[::1]:5432"},
		},
		{
			name:  "bracketed hosts still split on the outer comma",
			input: "[::1]:5432,[fe80::1]:5433",
			want:  []string{"[::1]:5432", "[fe80::1]:5433"},
		},
		{
			name:  "a trailing comma leaves an empty host",
			input: "host1,",
			want:  []string{"host1", ""},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := splitHosts(test.input); !reflect.DeepEqual(got, test.want) {
				t.Errorf("splitHosts(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestSplitHostPort(t *testing.T) {
	tests := []struct {
		name  string
		input string
		host  string
		port  string
	}{
		{name: "empty"},
		{name: "host only", input: "localhost", host: "localhost"},
		{name: "host and port", input: "localhost:5432", host: "localhost", port: "5432"},
		{name: "bracketed ipv6 with a port", input: "[::1]:5432", host: "::1", port: "5432"},
		{name: "bracketed ipv6 without a port", input: "[::1]", host: "::1"},
		{name: "ipv4 with a port", input: "127.0.0.1:5432", host: "127.0.0.1", port: "5432"},
		{
			name:  "a non numeric tail is not a port",
			input: "localhost:main",
			host:  "localhost:main",
		},
		{
			name:  "a socket directory has no port",
			input: "/var/run/postgresql",
			host:  "/var/run/postgresql",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			host, port := splitHostPort(test.input)
			if host != test.host || port != test.port {
				t.Errorf("splitHostPort(%q) = (%q, %q), want (%q, %q)", test.input, host, port, test.host, test.port)
			}
		})
	}
}

func TestIsDigits(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "empty is not a number", input: ""},
		{name: "digits", input: "5432", want: true},
		{name: "zero", input: "0", want: true},
		{name: "letters", input: "abc"},
		{name: "digits and letters", input: "54a2"},
		{name: "a sign is not a digit", input: "-1"},
		{name: "spaces are not digits", input: " 5432"},
		{name: "non ascii digits do not count", input: "٥٤٣٢"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isDigits(test.input); got != test.want {
				t.Errorf("isDigits(%q) = %t, want %t", test.input, got, test.want)
			}
		})
	}
}

func TestUnescape(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty"},
		{name: "nothing to decode", input: "localhost", want: "localhost"},
		{name: "a space", input: "my%20user", want: "my user"},
		{name: "an at sign", input: "p%40ss", want: "p@ss"},
		{name: "an equals sign", input: "-c%20search_path%3Dpublic", want: "-c search_path=public"},
		{name: "a socket path", input: "%2Fvar%2Frun", want: "/var/run"},
		{name: "a plus stays a plus", input: "a+b", want: "a+b"},
		{name: "an unreadable escape is kept as written", input: "100%zz", want: "100%zz"},
		{name: "a truncated escape is kept as written", input: "abc%2", want: "abc%2"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := unescape(test.input); got != test.want {
				t.Errorf("unescape(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}
