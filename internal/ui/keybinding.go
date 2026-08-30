package ui

import "slices"

type binding struct {
	keys []string
	hint string
}

func newBinding(hint string, keys ...string) binding {
	return binding{keys: keys, hint: hint}
}

func (b binding) matches(key string) bool {
	return slices.Contains(b.keys, key)
}

// merge joins bindings into one that any of their keys triggers, spelled as a
// single hint: "↑↓/jk" for the pair "↑/k · ↓/j".
func merge(hint string, binds ...binding) binding {
	m := binding{hint: hint}
	for _, b := range binds {
		m.keys = append(m.keys, b.keys...)
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
	Confirm binding
	Ping    binding
	Retry   binding
	Add     binding
	Edit    binding
	Delete  binding
	Secret  binding
	Search  binding
	Toggle  binding
	AddDB   binding

	// Text editing.
	Backspace binding

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
	SwitchStore  binding
}

var keyMap = defaultKeymap()

func defaultKeymap() keymap {
	k := keymap{
		Up:          newBinding("↑/k", "up", "k"),
		Down:        newBinding("↓/j", "down", "j"),
		Left:        newBinding("←", "left", "h"),
		Right:       newBinding("→", "right", "l"),
		NextSection: newBinding("tab", "tab"),
		PrevSection: newBinding("shift+tab", "shift+tab"),

		Confirm: newBinding("enter", "enter"),
		Ping:    newBinding("p", "p"),
		Retry:   newBinding("r", "r", "R"),
		Add:     newBinding("a", "a"),
		Edit:    newBinding("e", "e"),
		Delete:  newBinding("d", "d"),
		Secret:  newBinding("s", "s"),
		Search:  newBinding("/", "/"),
		Toggle:  newBinding("space", "space"),
		AddDB:   newBinding("ctrl+a", "ctrl+a"),

		Backspace: newBinding("backspace", "backspace"),

		Yes: newBinding("y", "y"),
		No:  newBinding("n", "n", "esc"),

		Quit:      newBinding("q/esc", "ctrl+c", "q", "esc"),
		Interrupt: newBinding("ctrl+c", "ctrl+c"),
		Cancel:    newBinding("esc", "esc"),

		Help: newBinding("?", "?"),
	}

	k.MoveVertical = merge("↑↓/jk", k.Up, k.Down)
	k.Cycle = merge("←/→", k.Left, k.Right)
	k.SwitchStore = merge("tab/←→", k.NextSection, k.PrevSection, k.Left, k.Right)

	return k
}
