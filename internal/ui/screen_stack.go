package ui

import (
	"errors"
	"io"
	"reflect"
	"slices"
	"sync/atomic"

	tea "charm.land/bubbletea/v2"
)

type screenHost struct {
	program *tea.Program
	closed  chan struct{}
}

var activeScreenHost atomic.Pointer[screenHost]

type pushScreenMsg struct {
	model tea.Model
	done  chan<- tea.Model
}

type screenMsg struct {
	id  int
	msg tea.Msg
}

type screenDoneMsg struct {
	id int
}

type stackedScreen struct {
	id    int
	model tea.Model
	done  chan<- tea.Model
}

type screenStack struct {
	screens []stackedScreen
	nextID  int
	size    *tea.WindowSizeMsg
	last    tea.View
}

var errNoScreenHost = errors.New("no screen host is running")

type execRequestMsg struct {
	run  func() error
	done func(error) tea.Msg
}

func WithScreens(run func() error) (err error) {
	if activeScreenHost.Load() != nil {
		return run()
	}

	host := &screenHost{
		program: tea.NewProgram(screenStack{last: tea.View{AltScreen: true}}),
		closed:  make(chan struct{}),
	}
	var programErr error
	go func() {
		_, programErr = host.program.Run()
		close(host.closed)
	}()
	activeScreenHost.Store(host)

	defer func() {
		activeScreenHost.Store(nil)
		host.program.Quit()
		<-host.closed
		err = errors.Join(err, programErr)
	}()

	return run()
}

func runScreen(model tea.Model) (tea.Model, error) {
	host := activeScreenHost.Load()
	if host == nil {
		return nil, errNoScreenHost
	}

	done := make(chan tea.Model, 1)
	host.program.Send(pushScreenMsg{model: model, done: done})

	select {
	case final := <-done:
		return final, nil
	case <-host.closed:
		return nil, tea.ErrProgramKilled
	}
}

func screenCmd(run func() error, done func(error) tea.Msg) tea.Cmd {
	return func() tea.Msg { return done(run()) }
}

func execCmd(run func() error, done func(error) tea.Msg) tea.Cmd {
	return func() tea.Msg { return execRequestMsg{run: run, done: done} }
}

func (s screenStack) Init() tea.Cmd { return nil }

func (s screenStack) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case pushScreenMsg:
		return s.push(msg)
	case screenMsg:
		return s.updateScreen(msg.id, msg.msg)
	case screenDoneMsg:
		return s.pop(msg.id), nil
	case tea.WindowSizeMsg:
		s.size = &msg
		cmds := make([]tea.Cmd, 0, len(s.screens))
		for _, screen := range s.screens {
			var cmd tea.Cmd
			s, cmd = s.updateScreen(screen.id, msg)
			cmds = append(cmds, cmd)
		}
		return s, tea.Batch(cmds...)
	}

	if len(s.screens) == 0 {
		return s, nil
	}

	return s.updateScreen(s.screens[len(s.screens)-1].id, msg)
}

func (s screenStack) View() tea.View {
	if len(s.screens) == 0 {
		return s.last
	}

	return s.screens[len(s.screens)-1].model.View()
}

func (s screenStack) push(msg pushScreenMsg) (screenStack, tea.Cmd) {
	s.nextID++
	screen := stackedScreen{id: s.nextID, model: msg.model, done: msg.done}
	s.screens = append(s.screens, screen)

	initCmd := wrapScreenCmd(screen.id, msg.model.Init())
	if s.size == nil {
		return s, initCmd
	}

	s, sizeCmd := s.updateScreen(screen.id, *s.size)

	return s, tea.Batch(initCmd, sizeCmd)
}

func (s screenStack) updateScreen(id int, msg tea.Msg) (screenStack, tea.Cmd) {
	i := s.indexOf(id)
	if i < 0 {
		return s, nil
	}

	model, cmd := s.screens[i].model.Update(msg)
	s.screens[i].model = model

	return s, wrapScreenCmd(id, cmd)
}

func (s screenStack) pop(id int) screenStack {
	i := s.indexOf(id)
	if i < 0 {
		return s
	}

	screen := s.screens[i]
	if len(s.screens) == 1 {
		s.last = screen.model.View()
	}
	s.screens = slices.Delete(s.screens, i, i+1)
	screen.done <- screen.model

	return s
}

func (s screenStack) indexOf(id int) int {
	return slices.IndexFunc(s.screens, func(screen stackedScreen) bool { return screen.id == id })
}

var teaPackage = reflect.TypeFor[tea.QuitMsg]().PkgPath()

func wrapScreenCmd(id int, cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}

	return func() tea.Msg {
		switch msg := cmd().(type) {
		case nil:
			return nil
		case tea.QuitMsg:
			return screenDoneMsg{id: id}
		case tea.BatchMsg:
			return tea.BatchMsg(wrapScreenCmds(id, msg))
		case execRequestMsg:
			return tea.Exec(runExec{run: msg.run}, func(err error) tea.Msg {
				return screenMsg{id: id, msg: msg.done(err)}
			})()
		default:
			msgType := reflect.TypeOf(msg)
			if msgType.PkgPath() != teaPackage {
				return screenMsg{id: id, msg: msg}
			}
			if msgType.ConvertibleTo(cmdsType) {
				cmds, _ := reflect.TypeAssert[[]tea.Cmd](reflect.ValueOf(msg).Convert(cmdsType))
				return tea.Sequence(wrapScreenCmds(id, cmds)...)()
			}
			return msg
		}
	}
}

var cmdsType = reflect.TypeFor[[]tea.Cmd]()

func wrapScreenCmds(id int, cmds []tea.Cmd) []tea.Cmd {
	wrapped := make([]tea.Cmd, len(cmds))
	for i, cmd := range cmds {
		wrapped[i] = wrapScreenCmd(id, cmd)
	}
	return wrapped
}

type runExec struct {
	run func() error
}

func (r runExec) Run() error        { return r.run() }
func (runExec) SetStdin(io.Reader)  {}
func (runExec) SetStdout(io.Writer) {}
func (runExec) SetStderr(io.Writer) {}
