package ui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui/view"
)

func TestModel_renderFitsWindow(t *testing.T) {
	type args struct {
		rows  int
		size  tea.WindowSizeMsg
		downs int
	}
	tests := []struct {
		name       string
		args       args
		wantWidth  int
		wantHeight int
		wantRow    string
	}{
		{
			name:       "no size yet keeps the minimum width and shows every row",
			args:       args{rows: 30},
			wantWidth:  tableInner + 2,
			wantHeight: 38,
			wantRow:    "conn-29",
		},
		{
			name:       "a wide window stretches the frame",
			args:       args{rows: 3, size: tea.WindowSizeMsg{Width: 160, Height: 40}},
			wantWidth:  160,
			wantHeight: 11,
			wantRow:    "conn-2",
		},
		{
			name:       "a narrow window keeps the minimum width",
			args:       args{rows: 3, size: tea.WindowSizeMsg{Width: 60, Height: 40}},
			wantWidth:  tableInner + 2,
			wantHeight: 11,
			wantRow:    "conn-2",
		},
		{
			name:       "a short window scrolls the rows to the cursor",
			args:       args{rows: 30, size: tea.WindowSizeMsg{Width: 120, Height: 20}, downs: 25},
			wantWidth:  120,
			wantHeight: 20,
			wantRow:    "conn-25",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confs := make([]config.Connection, tt.args.rows)
			for i := range confs {
				confs[i] = config.Postgres{Metadata: config.ConnMeta{Name: fmt.Sprintf("conn-%d", i)}}
			}
			var model tea.Model = New(view.NewConnections(stubConnections{confs: confs}), nil, nil)
			if tt.args.size.Height > 0 {
				model, _ = model.Update(tt.args.size)
			}
			for range tt.args.downs {
				model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
			}
			m, ok := model.(Model)
			require.True(t, ok)

			out := m.render()
			lines := strings.Split(out, "\n")

			assert.Len(t, lines, tt.wantHeight)
			assert.Equal(t, tt.wantWidth, lipgloss.Width(out))
			assert.Contains(t, out, tt.wantRow)
		})
	}
}
