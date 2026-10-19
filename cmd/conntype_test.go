package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/myjupyter/conm/internal/config"
)

func TestEverySubcommandIsAReservedName(t *testing.T) {
	t.Parallel()

	for _, kind := range config.ConnTypes {
		for _, sub := range connTypeCmd(kind).Commands() {
			assert.Contains(t, config.ReservedNames, sub.Name(), "%s %s", kind, sub.Name())
		}
	}
}
