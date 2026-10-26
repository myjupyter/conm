package config

import (
	"errors"
	"fmt"
	"slices"
)

var errSharedKey = errors.New("shares key")

const keybindingPathPrefix = "conm.keybinding."

const (
	navigationUpPath          = "navigation.up"
	navigationDownPath        = "navigation.down"
	navigationLeftPath        = "navigation.left"
	navigationRightPath       = "navigation.right"
	navigationNextSectionPath = "navigation.next_section"
	navigationPrevSectionPath = "navigation.prev_section"

	actionConfirmPath   = "action.confirm"
	actionPingPath      = "action.ping"
	actionRetryPath     = "action.retry"
	actionAddPath       = "action.add"
	actionEditPath      = "action.edit"
	actionDeletePath    = "action.delete"
	actionSecretPath    = "action.secret"
	actionSettingsPath  = "action.settings"
	actionTransportPath = "action.transport"
	actionSearchPath    = "action.search"
	actionInfoPath      = "action.info"
	actionYankPath      = "action.yank"
	actionTogglePath    = "action.toggle"
	actionAddDBPath     = "action.add_db"
	actionBackspacePath = "action.backspace"
	actionPastePath     = "action.paste"
	actionYesPath       = "action.yes"
	actionNoPath        = "action.no"

	screenQuitPath      = "screen.quit"
	screenInterruptPath = "screen.interrupt"
	screenCancelPath    = "screen.cancel"
	screenHelpPath      = "screen.help"
)

var mayShareKeys = [][2]string{
	{screenQuitPath, screenCancelPath},
	{screenQuitPath, actionNoPath},
	{screenCancelPath, actionNoPath},
	{screenQuitPath, screenInterruptPath},
	{actionYankPath, actionYesPath},
	{actionSecretPath, actionSettingsPath},
}

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
	Settings  []string `toml:"settings"`
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
			Settings:  []string{"s"},
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
		{navigationUpPath, &k.Navigation.Up},
		{navigationDownPath, &k.Navigation.Down},
		{navigationLeftPath, &k.Navigation.Left},
		{navigationRightPath, &k.Navigation.Right},
		{navigationNextSectionPath, &k.Navigation.NextSection},
		{navigationPrevSectionPath, &k.Navigation.PrevSection},

		{actionConfirmPath, &k.Action.Confirm},
		{actionPingPath, &k.Action.Ping},
		{actionRetryPath, &k.Action.Retry},
		{actionAddPath, &k.Action.Add},
		{actionEditPath, &k.Action.Edit},
		{actionDeletePath, &k.Action.Delete},
		{actionSecretPath, &k.Action.Secret},
		{actionSettingsPath, &k.Action.Settings},
		{actionTransportPath, &k.Action.Transport},
		{actionSearchPath, &k.Action.Search},
		{actionInfoPath, &k.Action.Info},
		{actionYankPath, &k.Action.Yank},
		{actionTogglePath, &k.Action.Toggle},
		{actionAddDBPath, &k.Action.AddDB},
		{actionBackspacePath, &k.Action.Backspace},
		{actionPastePath, &k.Action.Paste},
		{actionYesPath, &k.Action.Yes},
		{actionNoPath, &k.Action.No},

		{screenQuitPath, &k.Screen.Quit},
		{screenInterruptPath, &k.Screen.Interrupt},
		{screenCancelPath, &k.Screen.Cancel},
		{screenHelpPath, &k.Screen.Help},
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
	return errors.Join(k.validateKeys(), k.validateSharedKeys())
}

func (k *KeybindingSettings) validateKeys() error {
	var errs []error
	for _, field := range k.fields() {
		for _, key := range *field.keys {
			if _, err := ParseKey(key); err != nil {
				errs = append(errs, fmt.Errorf("%s%s: %w", keybindingPathPrefix, field.path, err))
			}
		}
	}
	return errors.Join(errs...)
}

func (k *KeybindingSettings) validateSharedKeys() error {
	var errs []error
	fields := k.fields()

	mayShareKey := func(a, b string) bool {
		return slices.Contains(mayShareKeys, [2]string{a, b}) || slices.Contains(mayShareKeys, [2]string{b, a})
	}

	for i, first := range fields {
		for _, second := range fields[i+1:] {
			if mayShareKey(first.path, second.path) {
				continue
			}
			for _, key := range *first.keys {
				if shared, ok := findKey(*second.keys, key); ok {
					errs = append(errs, fmt.Errorf("%s%s: %w %q with %s", keybindingPathPrefix, second.path, errSharedKey, shared, first.path))
				}
			}
		}
	}
	return errors.Join(errs...)
}

func findKey(keys []string, raw string) (string, bool) {
	wanted, err := ParseKey(raw)
	if err != nil {
		return "", false
	}

	for _, candidate := range keys {
		if key, err := ParseKey(candidate); err == nil && key == wanted {
			return candidate, true
		}
	}
	return "", false
}
