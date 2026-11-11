package names

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandom(t *testing.T) {
	tests := []struct {
		name  string
		draws int
	}{
		{name: "every draw is two lowercase words joined by a dash", draws: 1000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for range tt.draws {
				got := Random()

				assert.Regexp(t, `^[a-z]+-[a-z]+$`, got)
			}
		})
	}
}
