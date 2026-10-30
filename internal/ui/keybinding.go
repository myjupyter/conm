package ui

import (
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/utils/keys"
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

func newBinding(raws ...string) binding {
	strokes := make([]string, 0, len(raws))
	for _, key := range raws {
		strokes = append(strokes, sequence(key))
	}
	return binding{keys: strokes, hint: hintOf(strokes)}
}

func sequence(raw string) string {
	parts := keys.SplitSequence(raw)
	if len(parts) == 0 {
		return raw
	}
	for i, part := range parts {
		parts[i] = keystroke(part)
	}
	return keys.JoinSequence(parts...)
}

func sequenceHint(key string) string {
	parts := keys.SplitSequence(key)
	short := true
	for i, part := range parts {
		if glyph, ok := keyGlyphs[part]; ok {
			parts[i] = glyph
		}
		short = short && utf8.RuneCountInString(parts[i]) == 1
	}
	if short {
		return strings.Join(parts, "")
	}
	return strings.Join(parts, " ")
}

func keystroke(raw string) string {
	key, err := keys.Parse(raw)
	if err != nil {
		return raw
	}

	if key.Mod == keys.ModShift && utf8.RuneCountInString(key.Name) == 1 {
		return strings.ToUpper(key.Name)
	}
	return key.String()
}

func hintOf(strokes []string) string {
	var arrows strings.Builder
	spelled := make([]string, 0, len(strokes))
	for _, key := range strokes {
		if glyph, ok := keyGlyphs[key]; ok {
			arrows.WriteString(glyph)
			continue
		}
		spelled = append(spelled, sequenceHint(key))
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

func (b binding) startsWith(key string) bool {
	return slices.ContainsFunc(b.keys, func(k string) bool {
		strokes := keys.SplitSequence(k)
		return len(strokes) > 1 && strokes[0] == key
	})
}

func merge(binds ...binding) binding {
	var strokes []string
	for _, b := range binds {
		strokes = append(strokes, b.keys...)
	}
	return newBinding(strokes...)
}

type screenKeys struct {
	Transport         binding
	Secrets           binding
	Settings          binding
	AddConnection     binding
	AddConnectionType binding
}

type connectionKeys struct {
	Connect binding
	Edit    binding
	Delete  binding
	Info    binding
	Ping    binding
	Retry   binding
}

type formKeys struct {
	EditField  binding
	AddItem    binding
	RemoveItem binding
}

type entityKeys struct {
	Add    binding
	Edit   binding
	Delete binding
}

type settingsKeys struct {
	AddKey    binding
	ChangeKey binding
	RemoveKey binding
	Reset     binding
}

type promptKeys struct {
	Yes binding
	No  binding
}

type keymap struct {
	Up          binding
	Down        binding
	Left        binding
	Right       binding
	NextSection binding
	PrevSection binding

	Confirm   binding
	Cancel    binding
	Paste     binding
	Copy      binding
	Search    binding
	Help      binding
	Interrupt binding
	Backspace binding

	Screen     screenKeys
	Connection connectionKeys
	Form       formKeys
	Secret     entityKeys
	Transport  entityKeys
	Settings   settingsKeys
	Prompt     promptKeys

	Quit         binding
	Back         binding
	MoveVertical binding
	Cycle        binding
	SwitchKind   binding
}

var keyMap = newKeymap(config.DefaultKeybindings())

func newKeymap(s config.KeybindingSettings) keymap {
	if s.Validate() != nil {
		s = config.DefaultKeybindings()
	}
	nav, gen := s.Navigation, s.General

	k := keymap{
		Up:          newBinding(nav.Up...),
		Down:        newBinding(nav.Down...),
		Left:        newBinding(nav.Left...),
		Right:       newBinding(nav.Right...),
		NextSection: newBinding(nav.NextSection...),
		PrevSection: newBinding(nav.PrevSection...),

		Confirm:   newBinding(gen.Confirm...),
		Cancel:    newBinding(gen.Cancel...),
		Paste:     newBinding(gen.Paste...),
		Copy:      pressedTwice(newBinding(gen.Copy...)),
		Search:    newBinding(gen.Search...),
		Help:      newBinding(gen.Help...),
		Interrupt: newBinding(gen.Interrupt...),
		Backspace: newBinding(gen.Backspace...),

		Screen: screenKeys{
			Transport:         newBinding(s.Screen.Transport...),
			Secrets:           newBinding(s.Screen.Secrets...),
			Settings:          newBinding(s.Screen.Settings...),
			AddConnection:     newBinding(s.Screen.AddConnection...),
			AddConnectionType: newBinding(s.Screen.AddConnectionType...),
		},
		Connection: connectionKeys{
			Connect: newBinding(s.Connection.Connect...),
			Edit:    newBinding(s.Connection.Edit...),
			Delete:  newBinding(s.Connection.Delete...),
			Info:    newBinding(s.Connection.Info...),
			Ping:    newBinding(s.Connection.Ping...),
			Retry:   newBinding(s.Connection.Retry...),
		},
		Form: formKeys{
			EditField:  newBinding(s.Form.EditField...),
			AddItem:    newBinding(s.Form.AddItem...),
			RemoveItem: newBinding(s.Form.RemoveItem...),
		},
		Secret:    newEntityKeys(s.Secret),
		Transport: newEntityKeys(s.Transport),
		Settings: settingsKeys{
			AddKey:    newBinding(s.Settings.AddKey...),
			ChangeKey: newBinding(s.Settings.ChangeKey...),
			RemoveKey: newBinding(s.Settings.RemoveKey...),
			Reset:     newBinding(s.Settings.Reset...),
		},
		Prompt: promptKeys{
			Yes: newBinding(s.Prompt.Yes...),
			No:  newBinding(s.Prompt.No...),
		},
	}

	k.Quit = merge(newBinding(gen.Back...), k.Cancel, k.Interrupt)
	k.Back = merge(k.Connection.Info, k.Quit)
	k.MoveVertical = merge(k.Up, k.Down)
	k.Cycle = merge(k.Left, k.Right)
	k.SwitchKind = merge(k.NextSection, k.PrevSection, k.Left, k.Right)

	return k
}

func newEntityKeys(s config.EntityKeybindings) entityKeys {
	return entityKeys{
		Add:    newBinding(s.Add...),
		Edit:   newBinding(s.Edit...),
		Delete: newBinding(s.Delete...),
	}
}
