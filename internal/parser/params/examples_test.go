package params

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/myjupyter/conm/internal/config"
)

func TestEveryTypeHasAParsingExample(t *testing.T) {
	t.Parallel()

	for _, kind := range config.ConnTypes {
		assert.NotEmpty(t, Examples(kind), "%s has no example", kind)
		for _, input := range Examples(kind) {
			_, err := Parse(kind, input)
			assert.NoError(t, err, "%s: %s", kind, input)
		}
	}
}
