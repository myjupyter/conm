package config

import (
	"errors"
	"fmt"
)

type KeybindingSettings struct {
	Navigation NavigationKeybindingSettings `toml:"navigation"`
	Action     ActionKeybindingSettings     `toml:"action"`
	Screen     ScreenKeybindingSettings     `toml:"screen"`
}

type NavigationKeybindingSettings struct {
	Up          []string `toml:"up"`
	Down        []string `toml:"down"`
	Left        []string `toml:"left"`
	Right       []string `toml:"right"`
	NextSection []string `toml:"next_section"`
	PrevSection []string `toml:"prev_section"`
}

type ActionKeybindingSettings struct {
	Confirm   []string `toml:"confirm"`
	Ping      []string `toml:"ping"`
	Retry     []string `toml:"retry"`
	Add       []string `toml:"add"`
	Edit      []string `toml:"edit"`
	Delete    []string `toml:"delete"`
	Secret    []string `toml:"secret"`
	Transport []string `toml:"transport"`
	Search    []string `toml:"search"`
	Info      []string `toml:"info"`
	Yank      []string `toml:"yank"`
	Toggle    []string `toml:"toggle"`
	AddDB     []string `toml:"add_db"`
	Backspace []string `toml:"backspace"`
	Paste     []string `toml:"paste"`
	Yes       []string `toml:"yes"`
	No        []string `toml:"no"`
}

type ScreenKeybindingSettings struct {
	Quit      []string `toml:"quit"`
	Interrupt []string `toml:"interrupt"`
	Cancel    []string `toml:"cancel"`
	Help      []string `toml:"help"`
}

func DefaultKeybindings() KeybindingSettings {
	return KeybindingSettings{
		Navigation: NavigationKeybindingSettings{
			Up:          []string{"up", "k"},
			Down:        []string{"down", "j"},
			Left:        []string{"left", "h"},
			Right:       []string{"right", "l"},
			NextSection: []string{"tab"},
			PrevSection: []string{"shift+tab"},
		},
		Action: ActionKeybindingSettings{
			Confirm:   []string{"enter"},
			Ping:      []string{"p"},
			Retry:     []string{"r", "R"},
			Add:       []string{"a"},
			Edit:      []string{"e"},
			Delete:    []string{"d"},
			Secret:    []string{"s"},
			Transport: []string{"t"},
			Search:    []string{"/"},
			Info:      []string{"i"},
			Yank:      []string{"y"},
			Toggle:    []string{"space"},
			AddDB:     []string{"ctrl+a"},
			Backspace: []string{"backspace"},
			Paste:     []string{"ctrl+v"},
			Yes:       []string{"y"},
			No:        []string{"n", "esc"},
		},
		Screen: ScreenKeybindingSettings{
			Quit:      []string{"ctrl+c", "q", "esc"},
			Interrupt: []string{"ctrl+c"},
			Cancel:    []string{"esc"},
			Help:      []string{"?"},
		},
	}
}

type keybindingField struct {
	path string
	keys *[]string
}

func (k *KeybindingSettings) fields() []keybindingField {
	return []keybindingField{
		{"navigation.up", &k.Navigation.Up},
		{"navigation.down", &k.Navigation.Down},
		{"navigation.left", &k.Navigation.Left},
		{"navigation.right", &k.Navigation.Right},
		{"navigation.next_section", &k.Navigation.NextSection},
		{"navigation.prev_section", &k.Navigation.PrevSection},

		{"action.confirm", &k.Action.Confirm},
		{"action.ping", &k.Action.Ping},
		{"action.retry", &k.Action.Retry},
		{"action.add", &k.Action.Add},
		{"action.edit", &k.Action.Edit},
		{"action.delete", &k.Action.Delete},
		{"action.secret", &k.Action.Secret},
		{"action.transport", &k.Action.Transport},
		{"action.search", &k.Action.Search},
		{"action.info", &k.Action.Info},
		{"action.yank", &k.Action.Yank},
		{"action.toggle", &k.Action.Toggle},
		{"action.add_db", &k.Action.AddDB},
		{"action.backspace", &k.Action.Backspace},
		{"action.paste", &k.Action.Paste},
		{"action.yes", &k.Action.Yes},
		{"action.no", &k.Action.No},

		{"screen.quit", &k.Screen.Quit},
		{"screen.interrupt", &k.Screen.Interrupt},
		{"screen.cancel", &k.Screen.Cancel},
		{"screen.help", &k.Screen.Help},
	}
}

func (k *KeybindingSettings) fillDefaults() {
	defaults := DefaultKeybindings()
	defaultFields := defaults.fields()

	for i, field := range k.fields() {
		if len(*field.keys) == 0 {
			*field.keys = *defaultFields[i].keys
		}
	}
}

func (k *KeybindingSettings) validate() error {
	var errs []error
	for _, field := range k.fields() {
		for _, key := range *field.keys {
			if err := validateKey(key); err != nil {
				errs = append(errs, fmt.Errorf("conm.keybinding.%s: %w", field.path, err))
			}
		}
	}
	return errors.Join(errs...)
}
