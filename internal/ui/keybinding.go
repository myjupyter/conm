package ui

import (
	"slices"
	"strings"

	"github.com/myjupyter/conm/internal/config"
)

var keyGlyphs = map[string]string{
	"up":    "↑",
	"down":  "↓",
	"left":  "←",
	"right": "→",
}

type binding struct {
	keys       []string
	hint       string
	overridden bool
}

func newBinding(hint string, keys ...string) binding {
	return binding{keys: keys, hint: hint}
}

func bind(defaultHint string, keys, defaultKeys []string) binding {
	if slices.Equal(keys, defaultKeys) {
		return newBinding(defaultHint, keys...)
	}
	return binding{keys: keys, hint: hintOf(keys), overridden: true}
}

func hintOf(keys []string) string {
	spelled := make([]string, 0, len(keys))
	for _, key := range keys {
		if glyph, ok := keyGlyphs[key]; ok {
			key = glyph
		}
		spelled = append(spelled, key)
	}
	return strings.Join(spelled, "/")
}

func (b binding) matches(key string) bool {
	return slices.Contains(b.keys, key)
}

// merge joins bindings into one that any of their keys triggers, spelled as a
// single hint: "↑↓/jk" for the pair "↑/k · ↓/j".
func merge(hint string, binds ...binding) binding {
	m := binding{hint: hint}
	hints := make([]string, 0, len(binds))
	for _, b := range binds {
		m.keys = append(m.keys, b.keys...)
		hints = append(hints, b.hint)
		m.overridden = m.overridden || b.overridden
	}
	if m.overridden {
		m.hint = strings.Join(hints, " ")
	}
	return m
}

type keymap struct {
	// Navigation.
	Up          binding
	Down        binding
	Left        binding
	Right       binding
	NextSection binding
	PrevSection binding

	// Actions.
	Confirm   binding
	Ping      binding
	Retry     binding
	Add       binding
	Edit      binding
	Delete    binding
	Secret    binding
	Transport binding
	Search    binding
	Info      binding
	Yank      binding
	Toggle    binding
	AddDB     binding

	// Text editing.
	Backspace binding
	Paste     binding

	// The y/n confirmation prompt.
	Yes binding
	No  binding

	// Leaving a screen.
	Quit      binding
	Interrupt binding
	Cancel    binding

	// The keyhint table itself.
	Help binding

	// Groups of the bindings above, for the screens and hints that treat a
	// whole group as one.
	MoveVertical binding
	Cycle        binding
	SwitchKind   binding
	Back         binding
}

var keyMap = newKeymap(config.DefaultKeybindings())

func newKeymap(s config.KeybindingSettings) keymap {
	d := config.DefaultKeybindings()
	nav, dnav := s.Navigation, d.Navigation
	act, dact := s.Action, d.Action
	scr, dscr := s.Screen, d.Screen

	k := keymap{
		Up:          bind("↑/k", nav.Up, dnav.Up),
		Down:        bind("↓/j", nav.Down, dnav.Down),
		Left:        bind("←", nav.Left, dnav.Left),
		Right:       bind("→", nav.Right, dnav.Right),
		NextSection: bind("tab", nav.NextSection, dnav.NextSection),
		PrevSection: bind("shift+tab", nav.PrevSection, dnav.PrevSection),

		Confirm:   bind("enter", act.Confirm, dact.Confirm),
		Ping:      bind("p", act.Ping, dact.Ping),
		Retry:     bind("r", act.Retry, dact.Retry),
		Add:       bind("a", act.Add, dact.Add),
		Edit:      bind("e", act.Edit, dact.Edit),
		Delete:    bind("d", act.Delete, dact.Delete),
		Secret:    bind("s", act.Secret, dact.Secret),
		Transport: bind("t", act.Transport, dact.Transport),
		Search:    bind("/", act.Search, dact.Search),
		Info:      bind("i", act.Info, dact.Info),
		Yank:      bind("yy", act.Yank, dact.Yank),
		Toggle:    bind("space", act.Toggle, dact.Toggle),
		AddDB:     bind("ctrl+a", act.AddDB, dact.AddDB),

		Backspace: bind("backspace", act.Backspace, dact.Backspace),
		Paste:     bind("ctrl+v", act.Paste, dact.Paste),

		Yes: bind("y", act.Yes, dact.Yes),
		No:  bind("n", act.No, dact.No),

		Quit:      bind("q/esc", scr.Quit, dscr.Quit),
		Interrupt: bind("ctrl+c", scr.Interrupt, dscr.Interrupt),
		Cancel:    bind("esc", scr.Cancel, dscr.Cancel),

		Help: bind("?", scr.Help, dscr.Help),
	}

	k.MoveVertical = merge("↑↓/jk", k.Up, k.Down)
	k.Cycle = merge("←/→", k.Left, k.Right)
	k.SwitchKind = merge("tab/←→/hl", k.NextSection, k.PrevSection, k.Left, k.Right)
	k.Back = merge("i/q/esc", k.Info, k.Quit)

	return k
}
