package pathname

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{name: "plain", args: args{name: "corp"}},
		{name: "spaces inside", args: args{name: "My Servers"}},
		{name: "non-latin", args: args{name: "работа"}},
		{name: "dot inside", args: args{name: "prod.eu"}},
		{name: "leading dot", args: args{name: ".hidden"}},
		{name: "reserved word as prefix", args: args{name: "console"}},
		{name: "reserved word with suffix", args: args{name: "com10"}},
		{name: "single character", args: args{name: "a"}},
		{name: "max length", args: args{name: strings.Repeat("a", folderMaxName)}},
		{name: "max length in multibyte runes", args: args{name: strings.Repeat("ж", folderMaxName/2)}},
		{name: "empty", args: args{name: ""}, wantErr: ErrEmptyName},
		{name: "one past max length", args: args{name: strings.Repeat("a", folderMaxName+1)}, wantErr: ErrMaxLengthExceeded},
		{name: "multibyte runes past max length", args: args{name: strings.Repeat("ж", folderMaxName/2+1)}, wantErr: ErrMaxLengthExceeded},
		{name: "dot", args: args{name: "."}, wantErr: ErrDotName},
		{name: "double dot", args: args{name: ".."}, wantErr: ErrDotName},
		{name: "slash", args: args{name: "a/b"}, wantErr: ErrForbiddenChar},
		{name: "backslash", args: args{name: `a\b`}, wantErr: ErrForbiddenChar},
		{name: "colon", args: args{name: "a:b"}, wantErr: ErrForbiddenChar},
		{name: "angle brackets", args: args{name: "<a>"}, wantErr: ErrForbiddenChar},
		{name: "double quote", args: args{name: `a"b`}, wantErr: ErrForbiddenChar},
		{name: "pipe", args: args{name: "a|b"}, wantErr: ErrForbiddenChar},
		{name: "question mark", args: args{name: "a?"}, wantErr: ErrForbiddenChar},
		{name: "asterisk", args: args{name: "a*"}, wantErr: ErrForbiddenChar},
		{name: "tab", args: args{name: "a\tb"}, wantErr: ErrControlChar},
		{name: "nul", args: args{name: "a\x00b"}, wantErr: ErrControlChar},
		{name: "trailing dot", args: args{name: "corp."}, wantErr: ErrTrailingDotSpace},
		{name: "trailing space", args: args{name: "corp "}, wantErr: ErrTrailingDotSpace},
		{name: "reserved device", args: args{name: "aux"}, wantErr: ErrReservedName},
		{name: "reserved uppercase", args: args{name: "CON"}, wantErr: ErrReservedName},
		{name: "reserved port", args: args{name: "Com1"}, wantErr: ErrReservedName},
		{name: "reserved with extension", args: args{name: "nul.txt"}, wantErr: ErrReservedName},
		{name: "reserved with space before extension", args: args{name: "lpt9 .log"}, wantErr: ErrReservedName},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.ErrorIs(t, Validate(tt.args.name), tt.wantErr)
		})
	}
}
