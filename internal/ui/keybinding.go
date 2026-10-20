package ui

import (
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/myjupyter/conm/internal/config"
)

var keyGlyphs = map[string]string{
	"up":    "↑",
	"down":  "↓",
	"left":  "←",
	"right": "→",
}

type binding struct {
	keys []string
	hint string
}

func newBinding(keys ...string) binding {
	strokes := make([]string, 0, len(keys))
	for _, key := range keys {
		strokes = append(strokes, keystroke(key))
	}
	return binding{keys: strokes, hint: hintOf(strokes)}
}

func keystroke(raw string) string {
	key, err := config.ParseKey(raw)
	if err != nil {
		return raw
	}

	if key.Shift && !key.Ctrl && !key.Alt && utf8.RuneCountInString(key.Name) == 1 {
		return strings.ToUpper(key.Name)
	}

	var stroke strings.Builder
	if key.Ctrl {
		stroke.WriteString("ctrl+")
	}
	if key.Alt {
		stroke.WriteString("alt+")
	}
	if key.Shift {
		stroke.WriteString("shift+")
	}
	stroke.WriteString(key.Name)
	return stroke.String()
}

func hintOf(keys []string) string {
	var arrows strings.Builder
	spelled := make([]string, 0, len(keys))
	for _, key := range keys {
		if glyph, ok := keyGlyphs[key]; ok {
			arrows.WriteString(glyph)
			continue
		}
		spelled = append(spelled, key)
	}

	if arrows.Len() > 0 {
		spelled = append([]string{arrows.String()}, spelled...)
	}
	return strings.Join(spelled, "/")
}

func pressedTwice(b binding) binding {
	hints := make([]string, 0, len(b.keys))
	for _, key := range b.keys {
		once := hintOf([]string{key})
		if utf8.RuneCountInString(once) == 1 {
			hints = append(hints, once+once)
		} else {
			hints = append(hints, once+" "+once)
		}
	}
	b.hint = strings.Join(hints, "/")
	return b
}

func (b binding) matches(key string) bool {
	return slices.Contains(b.keys, key)
}

func merge(binds ...binding) binding {
	var keys []string
	for _, b := range binds {
		keys = append(keys, b.keys...)
	}
	return newBinding(keys...)
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
	Settings  binding
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
	nav, act, scr := s.Navigation, s.Action, s.Screen

	k := keymap{
		Up:          newBinding(nav.Up...),
		Down:        newBinding(nav.Down...),
		Left:        newBinding(nav.Left...),
		Right:       newBinding(nav.Right...),
		NextSection: newBinding(nav.NextSection...),
		PrevSection: newBinding(nav.PrevSection...),

		Confirm:   newBinding(act.Confirm...),
		Ping:      newBinding(act.Ping...),
		Retry:     newBinding(act.Retry...),
		Add:       newBinding(act.Add...),
		Edit:      newBinding(act.Edit...),
		Delete:    newBinding(act.Delete...),
		Secret:    newBinding(act.Secret...),
		Setting:   newBinding(act.Settings),
		Transport: newBinding(act.Transport...),
		Search:    newBinding(act.Search...),
		Info:      newBinding(act.Info...),
		Yank:      pressedTwice(newBinding(act.Yank...)),
		Toggle:    newBinding(act.Toggle...),
		AddDB:     newBinding(act.AddDB...),

		Backspace: newBinding(act.Backspace...),
		Paste:     newBinding(act.Paste...),

		Yes: newBinding(act.Yes...),
		No:  newBinding(act.No...),

		Quit:      newBinding(scr.Quit...),
		Interrupt: newBinding(scr.Interrupt...),
		Cancel:    newBinding(scr.Cancel...),

		Help: newBinding(scr.Help...),
	}

	k.MoveVertical = merge(k.Up, k.Down)
	k.Cycle = merge(k.Left, k.Right)
	k.SwitchKind = merge(k.NextSection, k.PrevSection, k.Left, k.Right)
	k.Back = merge(k.Info, k.Quit)

	return k
}
