package ui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/repository"
	"github.com/myjupyter/conm/internal/ui/spec"
	"github.com/myjupyter/conm/internal/ui/view"
)

func connectionForm(
	t config.ConnType,
	isEdit bool,
	secrets *view.Secrets,
	save func(config.Connection) error,
) (formLoop[config.Connection], error) {
	formSpec, ok := spec.FormSpecs[t]
	if !ok {
		return formLoop[config.Connection]{}, fmt.Errorf("form is not implemented for connection type %q", t)
	}

	title := formSpec.AddTitle
	if isEdit {
		title = formSpec.EditTitle
	}

	return formLoop[config.Connection]{
		open: func(initial formValues, status formStatus) (formResult[config.Connection], error) {
			model := newFormModel(t, formSpec, title, initial, isEdit, secrets)
			if status.text != "" {
				model.status, model.statusKind = status.text, status.kind
			}

			return runForm(model)
		},
		save: save,
	}, nil
}

func RunAddForm(conm config.Conm, t config.ConnType, initial formValues, warnings []string) (bool, error) {
	ws, err := repository.NewWorkspace(conm)
	if err != nil {
		return false, err
	}
	defer ws.Close()

	repo, ok := ws.ConnectionsOf(t)
	if !ok {
		return false, fmt.Errorf("%q is not enabled\nrun 'conm init' to enable it", t)
	}

	loop, err := connectionForm(t, false, view.NewSecrets(ws.Keyring), repo.Add)
	if err != nil {
		return false, err
	}

	return loop.run(initial, warnStatus(warnings))
}

func runAddConnection(
	t config.ConnType,
	secrets *view.Secrets,
	save func(config.Connection) error,
) (bool, error) {
	loop, err := connectionForm(t, false, secrets, save)
	if err != nil {
		return false, err
	}

	return loop.run(nil, formStatus{})
}

func runEditConnection(
	t config.ConnType,
	existing config.Connection,
	secrets *view.Secrets,
	save func(config.Connection) error,
) (bool, error) {
	initial, err := seedConnection(t, existing)
	if err != nil {
		return false, err
	}

	loop, err := connectionForm(t, true, secrets, save)
	if err != nil {
		return false, err
	}

	return loop.run(initial, formStatus{})
}

func seedConnection(t config.ConnType, existing config.Connection) (formValues, error) {
	formSpec, ok := spec.FormSpecs[t]
	if !ok {
		return nil, fmt.Errorf("edit form is not implemented for connection type %q", t)
	}
	if formSpec.SeedFunc == nil {
		return nil, fmt.Errorf("edit form is not seedable for connection type %q", t)
	}

	initial := formSpec.SeedFunc(existing)
	if initial == nil {
		return nil, fmt.Errorf("edit form cannot seed a %T as connection type %q", existing, t)
	}

	spec.SeedLinks(existing.Meta().Links, initial)

	return initial, nil
}

// openLink hands a link's url to the browser. The form stays where it is: the
// url is checked first, and whatever the browser reports comes back as a status
// line rather than as a reason to leave.
func (m formModel) openLink(key spec.FormFieldKey) (tea.Model, tea.Cmd) {
	i, ok := spec.LinkIndexOf(key)
	if !ok {
		return m, nil
	}

	name := strings.TrimSpace(m.vals[spec.LinkNameKey(i)])
	if name == "" {
		name = fmt.Sprintf("link %d", i+1)
	}

	url := strings.TrimSpace(m.vals[spec.LinkURLKey(i)])
	if err := config.ValidateWebLinkURL(url); err != nil {
		m.attempted = true
		m.setStatus(name+" · "+err.Error(), kindErr)
		return m, nil
	}

	m.setStatus("opening "+name+" · "+url, kindPending)
	return m, openURLCmd(name, url)
}

type linkOpenedMsg struct {
	name string
	err  error
}

func openURLCmd(name, url string) tea.Cmd {
	return func() tea.Msg {
		opener := "xdg-open"
		if runtime.GOOS == "darwin" {
			opener = "open"
		}

		// The opener hands off to the browser and is done with; it is started
		// rather than waited for, because xdg-open on some desktops does not
		// return until the browser it launched exits.
		cmd := exec.Command(opener, url)
		if err := cmd.Start(); err != nil {
			return linkOpenedMsg{name: name, err: err}
		}
		go func() { _ = cmd.Wait() }()

		return linkOpenedMsg{name: name}
	}
}

func (m formModel) applyLinkOpened(msg linkOpenedMsg) formModel {
	if msg.err != nil {
		m.setStatus("couldn't open "+msg.name+" · "+msg.err.Error(), kindErr)
		return m
	}
	m.setStatus("opened "+msg.name+" in the browser", kindOK)
	return m
}

// pickSecretCmd hands the terminal to the keyring screen and folds the chosen
// location back into the field.
func (m formModel) pickSecretCmd() tea.Cmd {
	if m.secrets == nil {
		return nil
	}

	var picked string
	var ok bool
	current := m.vals[spec.SecretValueKey]

	return tea.Exec(
		runExec{run: func() error {
			var err error
			picked, ok, err = runSecretPicker(m.secrets, current)
			return err
		}},
		func(err error) tea.Msg { return secretPickedMsg{ref: picked, ok: ok, err: err} },
	)
}

func runForm(model formModel) (formResult[config.Connection], error) {
	var none formResult[config.Connection]

	m, err := tea.NewProgram(model).Run()
	if err != nil {
		return none, err
	}

	fm, ok := m.(formModel)
	if !ok {
		return none, fmt.Errorf("connection form returned an unexpected model %T", m)
	}
	if !fm.submitted {
		return none, nil
	}

	cfg, err := fm.result()
	if err != nil {
		return none, err
	}

	return formResult[config.Connection]{value: cfg, values: fm.values(), ok: true}, nil
}
