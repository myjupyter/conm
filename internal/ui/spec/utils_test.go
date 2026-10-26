package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitStringArray(t *testing.T) {
	tests := map[string]struct {
		raw  FormFieldValue
		want []string
	}{
		"empty":                   {raw: "", want: nil},
		"one":                     {raw: "prod", want: []string{"prod"}},
		"whitespace is trimmed":   {raw: " prod ,  eu-west ", want: []string{"prod", "eu-west"}},
		"blank items are dropped": {raw: "prod,, ,eu-west,", want: []string{"prod", "eu-west"}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, SplitStringArray(tt.raw))
		})
	}
}

func TestStringArrayRoundTrip(t *testing.T) {
	items := []string{"prod", "eu-west", "analytics"}
	assert.Equal(t, items, SplitStringArray(JoinStringArray(items)))
}
