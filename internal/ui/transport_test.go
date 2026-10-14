package ui

import (
	"context"
	"errors"
	"maps"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/network"
	"github.com/myjupyter/conm/internal/repository"
	"github.com/myjupyter/conm/internal/transport"
	"github.com/myjupyter/conm/internal/ui/spec"
	"github.com/myjupyter/conm/internal/ui/view"
)

var errTransportInUse = errors.New(`transport "ssh:prod" is still in use by 1 connection(s): postgres/orders`)

type stubTransports struct {
	records []config.Transport
	inUse   string
}

func (s *stubTransports) Resolve(config.Connection) (transport.Transport, error) {
	return nil, errTransportInUse
}

func (s *stubTransports) Len() int { return len(s.records) }

func (s *stubTransports) Get(i int) (config.Transport, bool) {
	if i < 0 || i >= len(s.records) {
		return config.Transport{}, false
	}
	return s.records[i], true
}

func (s *stubTransports) UsagesAt(i int) []repository.Usage {
	if s.records[i].Name == s.inUse {
		return []repository.Usage{{Kind: "postgres", Owner: "orders"}}
	}
	return nil
}

func (s *stubTransports) Add(t config.Transport) error {
	s.records = append(s.records, t)
	return nil
}

func (s *stubTransports) Edit(i int, t config.Transport) error {
	s.records[i] = t
	return nil
}

func (s *stubTransports) Remove(i int) error {
	if s.records[i].Name == s.inUse {
		return errTransportInUse
	}
	s.records = append(s.records[:i], s.records[i+1:]...)
	return nil
}

func newTestTransportForm(initial formValues, ctx transportFormContext) transportFormModel {
	return newTransportFormModel(spec.TransportFormSpec, "add", initial, initial != nil, ctx)
}

func TestNewTransportFormModel(t *testing.T) {
	t.Parallel()

	type args struct {
		initial formValues
		sources spec.Sources
	}
	tests := []struct {
		name     string
		args     args
		wantKind string
		wantVia  string
	}{
		{
			name:     "a new transport starts on ssh and the first named bastion",
			args:     args{sources: spec.Sources{"ssh": {"bastion", "jump"}}},
			wantKind: config.SSHTransportKind,
			wantVia:  "bastion",
		},
		{
			name: "an edited transport keeps its bastion",
			args: args{
				initial: spec.TransportFormSpec.SeedFunc(config.Transport{Name: "prod", Kind: config.SSHTransportKind, Via: "jump"}),
				sources: spec.Sources{"ssh": {"bastion", "jump"}},
			},
			wantKind: config.SSHTransportKind,
			wantVia:  "jump",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := newTestTransportForm(tt.args.initial, transportFormContext{sources: tt.args.sources})

			assert.Equal(t, tt.wantKind, got.vals["kind"])
			assert.Equal(t, tt.wantVia, got.vals[spec.TransportViaKey])
			assert.NotEmpty(t, got.render())
		})
	}
}

func TestTransportFormModel_submit(t *testing.T) {
	t.Parallel()

	type args struct {
		values  formValues
		sources spec.Sources
	}
	tests := []struct {
		name       string
		args       args
		wantSaved  bool
		wantStatus string
		want       config.Transport
	}{
		{
			name: "a complete transport is saved",
			args: args{
				values:  formValues{"name": "prod", spec.TransportViaKey: "bastion", "keepalive": "30"},
				sources: spec.Sources{"ssh": {"bastion"}},
			},
			wantSaved: true,
			want:      config.Transport{Name: "prod", Kind: config.SSHTransportKind, Via: "bastion", KeepAlive: 30},
		},
		{
			name: "a bastion that is gone blocks saving",
			args: args{
				values:  formValues{"name": "prod", spec.TransportViaKey: "old-bastion"},
				sources: spec.Sources{"ssh": {"bastion"}},
			},
			wantStatus: "1 field need attention",
		},
		{
			name: "a name is required",
			args: args{
				values:  formValues{spec.TransportViaKey: "bastion"},
				sources: spec.Sources{"ssh": {"bastion"}},
			},
			wantStatus: "1 field need attention",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newTestTransportForm(nil, transportFormContext{sources: tt.args.sources})
			maps.Copy(m.vals, tt.args.values)

			model, _ := m.submit()
			got, ok := model.(transportFormModel)
			require.True(t, ok)

			assert.Equal(t, tt.wantSaved, got.submitted)
			if !tt.wantSaved {
				assert.Equal(t, tt.wantStatus, got.status)
				assert.NotEmpty(t, got.render())
				return
			}
			rec, err := got.spec.BuildFunc(got.values())
			require.NoError(t, err)
			assert.Equal(t, tt.want, rec)
		})
	}
}

func TestTransportFormModel_pingViaCmd(t *testing.T) {
	t.Parallel()

	type args struct {
		via     string
		pingErr error
	}
	tests := []struct {
		name       string
		args       args
		wantStatus string
		wantKind   statusKind
	}{
		{
			name:       "the bastion answers",
			args:       args{via: "bastion"},
			wantStatus: "pong · bastion answered in 12ms",
			wantKind:   kindOK,
		},
		{
			name:       "the bastion does not answer",
			args:       args{via: "bastion", pingErr: errors.New("connection refused")},
			wantStatus: "ping failed · bastion · connection refused",
			wantKind:   kindErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var pinged string
			ctx := transportFormContext{
				sources: spec.Sources{"ssh": {tt.args.via}},
				pingVia: func(_ context.Context, name string) (network.PingResult, error) {
					pinged = name
					return network.PingResult{PingTime: 12 * time.Millisecond}, tt.args.pingErr
				},
			}
			m := newTestTransportForm(nil, ctx)

			model, cmd := m.pingViaCmd()
			require.NotNil(t, cmd)
			pinging, ok := model.(transportFormModel)
			require.True(t, ok)
			next, _ := pinging.Update(cmd())
			got, ok := next.(transportFormModel)
			require.True(t, ok)

			is := assert.New(t)
			is.Equal(tt.args.via, pinged)
			is.Equal(tt.wantStatus, got.status)
			is.Equal(tt.wantKind, got.statusKind)
		})
	}
}

func TestTransportModel_removeTransportCmd(t *testing.T) {
	t.Parallel()

	type args struct {
		cursor int
	}
	tests := []struct {
		name     string
		args     args
		wantLen  int
		wantKind statusKind
	}{
		{name: "an unused transport goes", args: args{cursor: 1}, wantLen: 1, wantKind: kindOK},
		{name: "a transport in use stays and says who uses it", args: args{cursor: 0}, wantLen: 2, wantKind: kindErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &stubTransports{
				records: []config.Transport{
					{Name: "prod", Kind: config.SSHTransportKind, Via: "bastion"},
					{Name: "staging", Kind: config.SSHTransportKind, Via: "jump"},
				},
				inUse: "prod",
			}
			m := newTransportModel(view.NewTransports(repo), view.NewConnections())
			require.True(t, m.transports.SetCursor(tt.args.cursor))

			next, _ := m.Update(m.removeTransportCmd()())
			got, ok := next.(transportModel)
			require.True(t, ok)

			assert.Equal(t, tt.wantLen, got.transports.Len())
			assert.Equal(t, tt.wantKind, got.statusKind)
			assert.NotEmpty(t, got.render())
		})
	}
}

var _ tea.Model = transportModel{}
