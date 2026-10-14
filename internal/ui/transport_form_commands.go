package ui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/network"
	"github.com/myjupyter/conm/internal/ui/spec"
)

type transportFormContext struct {
	sources spec.Sources
	pingVia func(ctx context.Context, name string) (network.PingResult, error)
}

type transportPingMsg struct {
	via    string
	result network.PingResult
	err    error
}

func transportForm(isEdit bool, ctx transportFormContext, save func(config.Transport) error) formLoop[config.Transport] {
	title := spec.TransportFormSpec.AddTitle
	if isEdit {
		title = spec.TransportFormSpec.EditTitle
	}

	return formLoop[config.Transport]{
		open: func(initial formValues, status formStatus) (formResult[config.Transport], error) {
			model := newTransportFormModel(spec.TransportFormSpec, title, initial, isEdit, ctx)
			if status.text != "" {
				model.setStatus(status.text, status.kind)
			}

			return runTransportForm(model)
		},
		save: save,
	}
}

func runAddTransport(ctx transportFormContext, save func(config.Transport) error) (bool, error) {
	return transportForm(false, ctx, save).run(nil, formStatus{})
}

func runEditTransport(existing config.Transport, ctx transportFormContext, save func(config.Transport) error) (bool, error) {
	initial := spec.TransportFormSpec.SeedFunc(existing)
	return transportForm(true, ctx, save).run(initial, formStatus{})
}

func (m transportFormModel) pingViaCmd() (tea.Model, tea.Cmd) {
	via := m.vals[spec.TransportViaKey]
	if via == "" || m.pingVia == nil {
		m.attempted = true
		m.setStatus("pick a connection to go through before pinging", kindErr)
		return m, nil
	}

	m.setStatus("pinging "+via+" …", kindPending)
	pingVia := m.pingVia
	return m, func() tea.Msg {
		result, err := pingVia(context.Background(), via)
		return transportPingMsg{via: via, result: result, err: err}
	}
}

func runTransportForm(model transportFormModel) (formResult[config.Transport], error) {
	var none formResult[config.Transport]

	m, err := tea.NewProgram(model).Run()
	if err != nil {
		return none, err
	}

	fm, ok := m.(transportFormModel)
	if !ok {
		return none, fmt.Errorf("transport form returned an unexpected model %T", m)
	}
	if !fm.submitted {
		return none, nil
	}

	values := fm.values()
	rec, err := fm.spec.BuildFunc(values)
	if err != nil {
		return none, err
	}

	return formResult[config.Transport]{value: rec, values: values, ok: true}, nil
}
