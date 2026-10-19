package view

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/network"
	"github.com/myjupyter/conm/internal/repository"
	"github.com/myjupyter/conm/internal/transport"
)

var errInUse = errors.New("in use")

type fakeTransports struct {
	records []config.Transport
	used    map[string][]repository.Usage
}

func (f *fakeTransports) Resolve(config.Connection) (transport.Transport, error) {
	return nil, errInUse
}

func (f *fakeTransports) Len() int { return len(f.records) }

func (f *fakeTransports) Get(i int) (config.Transport, bool) {
	if i < 0 || i >= len(f.records) {
		return config.Transport{}, false
	}
	return f.records[i], true
}

func (f *fakeTransports) UsagesAt(i int) []repository.Usage {
	return f.used[f.records[i].Name]
}

func (f *fakeTransports) Add(t config.Transport) error {
	f.records = append(f.records, t)
	return nil
}

func (f *fakeTransports) Edit(i int, t config.Transport) error {
	f.records[i] = t
	return nil
}

func (f *fakeTransports) Remove(i int) error {
	if len(f.used[f.records[i].Name]) > 0 {
		return errInUse
	}
	f.records = append(f.records[:i], f.records[i+1:]...)
	return nil
}

func twoTransports() *fakeTransports {
	return &fakeTransports{
		records: []config.Transport{
			{Name: "prod", Kind: config.SSHTransportKind, Via: "bastion", Description: "eu-west"},
			{Name: "staging", Kind: config.SSHTransportKind, Via: "jump"},
		},
		used: map[string][]repository.Usage{"prod": {{Kind: "postgres", Owner: "orders"}}},
	}
}

func TestTransports_Search(t *testing.T) {
	t.Parallel()

	type args struct {
		query string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{name: "no query shows every transport", args: args{query: ""}, want: []string{"prod", "staging"}},
		{name: "by the connection it goes through", args: args{query: "jump"}, want: []string{"staging"}},
		{name: "by description", args: args{query: "eu-west"}, want: []string{"prod"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v := NewTransports(twoTransports())
			v.Search(tt.args.query)

			got := make([]string, 0, v.Len())
			for i := range v.Len() {
				tr, ok := v.At(i)
				require.True(t, ok)
				got = append(got, tr.Name)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTransports_Remove(t *testing.T) {
	t.Parallel()

	type args struct {
		cursor int
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
		wantErr error
	}{
		{name: "an unused transport goes", args: args{cursor: 1}, wantLen: 1},
		{name: "a transport in use stays", args: args{cursor: 0}, wantLen: 2, wantErr: errInUse},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v := NewTransports(twoTransports())
			require.True(t, v.SetCursor(tt.args.cursor))

			err := v.Remove()

			assert.Equal(t, tt.wantLen, v.Len())
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

type fakeConnections struct {
	kind    config.ConnType
	configs []config.Connection
}

func (f fakeConnections) Kind() config.ConnType                     { return f.kind }
func (fakeConnections) ClientInfo(context.Context) cli.Info         { return cli.Info{} }
func (f fakeConnections) Len() int                                  { return len(f.configs) }
func (fakeConnections) IndexOf(string) (int, bool)                  { return -1, false }
func (fakeConnections) ConnectionAt(int) (network.Connection, bool) { return nil, false }
func (f fakeConnections) ConfigAt(i int) (config.Connection, bool)  { return f.configs[i], true }
func (fakeConnections) Add(config.Connection) error                 { return nil }
func (fakeConnections) Edit(int, config.Connection) error           { return nil }
func (fakeConnections) Remove(int) error                            { return nil }

func TestConnections_NamesOf(t *testing.T) {
	t.Parallel()

	type args struct {
		kind config.ConnType
	}
	ssh := fakeConnections{kind: config.SSHConnType, configs: []config.Connection{
		config.SSH{Metadata: config.ConnMeta{Name: "bastion"}},
		config.SSH{Hostname: "unnamed.local"},
		config.SSH{Metadata: config.ConnMeta{Name: "jump"}},
	}}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{name: "named connections of the type", args: args{kind: config.SSHConnType}, want: []string{"bastion", "jump"}},
		{name: "a type that is not open", args: args{kind: config.PostgresConnType}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v := NewConnections(ssh)

			assert.Equal(t, tt.want, v.NamesOf(tt.args.kind))
		})
	}
}
