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

func TestRootCommandNamesAndAliasesAreUnique(t *testing.T) {
	t.Parallel()

	seen := map[string]string{}
	for _, cmd := range rootCmd.Commands() {
		for _, name := range append([]string{cmd.Name()}, cmd.Aliases...) {
			assert.NotContains(t, seen, name, "%s and %s", seen[name], cmd.Name())
			seen[name] = cmd.Name()
		}
	}
}
