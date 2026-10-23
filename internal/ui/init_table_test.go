package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
)

func TestInitModel_pick(t *testing.T) {
	tests := []struct {
		name       string
		cursor     int
		wantPicked bool
		wantChosen config.ConnType
		wantStatus statusKind
	}{
		{name: "available type is picked", cursor: 0, wantPicked: true, wantChosen: config.PostgresConnType, wantStatus: kindIdle},
		{name: "type without a client is refused", cursor: 1, wantStatus: kindErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := initModel{
				types: []config.ConnType{config.PostgresConnType, config.MySQLConnType},
				clients: map[config.ConnType][]cli.Info{
					config.PostgresConnType: {{Name: cli.Psql, Installed: true}},
					config.MySQLConnType:    {{Name: cli.MySQL}},
				},
				cursor: tt.cursor,
			}

			got, _ := m.pick()
			gotModel, ok := got.(initModel)
			assert.True(t, ok)
			assert.Equal(t, tt.wantPicked, gotModel.picked)
			assert.Equal(t, tt.wantChosen, gotModel.chosen)
			assert.Equal(t, tt.wantStatus, gotModel.statusKind)
		})
	}
}
