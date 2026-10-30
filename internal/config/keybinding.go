package config

import (
	"errors"
	"fmt"
	"slices"

	"github.com/myjupyter/conm/internal/utils/keys"
)

var (
	errSharedKey      = errors.New("shares key")
	errLocked         = errors.New("is locked")
	errDefaultRemoved = errors.New("must keep its default key")
	errLongSequence   = errors.New("a sequence is at most two keys")
	errNoSequence     = errors.New("takes single keys only, sequences work on the connections table")
)

const keybindingPathPrefix = "conm.keybinding."

const (
	navigationGroup = "navigation"
	generalGroup    = "general"
	screenGroup     = "screen"
	connectionGroup = "connection"
	formGroup       = "form"
	secretGroup     = "secret"
	transportGroup  = "transport"
	settingsGroup   = "settings"
	promptGroup     = "prompt"
)

const (
	addAction    = "add"
	editAction   = "edit"
	deleteAction = "delete"
	pingAction   = "ping"
	retryAction  = "retry"
)

const confirmPath = generalGroup + ".confirm"

var keybindingScreens = [][]string{
	{screenGroup, connectionGroup},
	{formGroup, confirmPath, connectionGroup + "." + pingAction, connectionGroup + "." + retryAction},
	{secretGroup, confirmPath},
	{transportGroup, confirmPath},
	{settingsGroup, confirmPath},
}

var keybindingEverywhere = []string{navigationGroup, generalGroup}

type KeybindingSettings struct {
	Navigation NavigationKeybindings `toml:"navigation"`
	General    GeneralKeybindings    `toml:"general"`
	Screen     ScreenKeybindings     `toml:"screen"`
	Connection ConnectionKeybindings `toml:"connection"`
	Form       FormKeybindings       `toml:"form"`
	Secret     EntityKeybindings     `toml:"secret"`
	Transport  EntityKeybindings     `toml:"transport"`
	Settings   SettingsKeybindings   `toml:"settings"`
	Prompt     PromptKeybindings     `toml:"prompt"`
}

type NavigationKeybindings struct {
	Up          []string `toml:"up"`
	Down        []string `toml:"down"`
	Left        []string `toml:"left"`
	Right       []string `toml:"right"`
	NextSection []string `toml:"next_section"`
	PrevSection []string `toml:"prev_section"`
}

type GeneralKeybindings struct {
	Confirm   []string `toml:"confirm"`
	Cancel    []string `toml:"cancel"`
	Back      []string `toml:"back"`
	Paste     []string `toml:"paste"`
	Copy      []string `toml:"copy"`
	Search    []string `toml:"search"`
	Help      []string `toml:"help"`
	Interrupt []string `toml:"interrupt"`
	Backspace []string `toml:"backspace"`
}

type ScreenKeybindings struct {
	Transport         []string `toml:"transport"`
	Secrets           []string `toml:"secrets"`
	Settings          []string `toml:"settings"`
	AddConnection     []string `toml:"add_connection"`
	AddConnectionType []string `toml:"add_connection_type"`
}

type ConnectionKeybindings struct {
	Connect []string `toml:"connect"`
	Edit    []string `toml:"edit"`
	Delete  []string `toml:"delete"`
	Info    []string `toml:"info"`
	Ping    []string `toml:"ping"`
	Retry   []string `toml:"retry"`
}

type FormKeybindings struct {
	EditField  []string `toml:"edit_field"`
	Secret     []string `toml:"secret"`
	AddItem    []string `toml:"add_item"`
	RemoveItem []string `toml:"remove_item"`
}

type EntityKeybindings struct {
	Add    []string `toml:"add"`
	Edit   []string `toml:"edit"`
	Delete []string `toml:"delete"`
}

type SettingsKeybindings struct {
	AddKey    []string `toml:"add_key"`
	ChangeKey []string `toml:"change_key"`
	RemoveKey []string `toml:"remove_key"`
	Reset     []string `toml:"reset"`
}

type PromptKeybindings struct {
	Yes []string `toml:"yes"`
	No  []string `toml:"no"`
}

func ParseSequence(raw string) ([]keys.Key, error) {
	strokes := keys.SplitSequence(raw)
	if len(strokes) > 2 {
		return nil, fmt.Errorf("%q: %w", raw, errLongSequence)
	}
	if len(strokes) == 0 {
		strokes = []string{raw}
	}

	sequence := make([]keys.Key, 0, len(strokes))
	for _, stroke := range strokes {
		key, err := keys.Parse(stroke)
		if err != nil {
			return nil, err
		}
		sequence = append(sequence, key)
	}
	return sequence, nil
}

func DefaultKeybindings() KeybindingSettings {
	return KeybindingSettings{
		Navigation: NavigationKeybindings{
			Up:          []string{"up", "k"},
			Down:        []string{"down", "j"},
			Left:        []string{"left", "h"},
			Right:       []string{"right", "l"},
			NextSection: []string{"tab"},
			PrevSection: []string{"shift+tab"},
		},
		General: GeneralKeybindings{
			Confirm:   []string{"enter"},
			Cancel:    []string{"esc"},
			Back:      []string{"q"},
			Paste:     []string{"ctrl+v"},
			Copy:      []string{"y"},
			Search:    []string{"/"},
			Help:      []string{"?"},
			Interrupt: []string{"ctrl+c"},
			Backspace: []string{"backspace"},
		},
		Screen: ScreenKeybindings{
			Transport:         []string{"t"},
			Secrets:           []string{"s"},
			Settings:          []string{"s s"},
			AddConnection:     []string{"a"},
			AddConnectionType: []string{"a a"},
		},
		Connection: ConnectionKeybindings{
			Connect: []string{"enter"},
			Edit:    []string{"e"},
			Delete:  []string{"d"},
			Info:    []string{"i"},
			Ping:    []string{"p"},
			Retry:   []string{"r", "R"},
		},
		Form: FormKeybindings{
			EditField:  []string{"i"},
			Secret:     []string{"s"},
			AddItem:    []string{"a"},
			RemoveItem: []string{"d"},
		},
		Secret:    defaultEntityKeybindings(),
		Transport: defaultEntityKeybindings(),
		Settings: SettingsKeybindings{
			AddKey:    []string{"a"},
			ChangeKey: []string{"e"},
			RemoveKey: []string{"d"},
			Reset:     []string{"r"},
		},
		Prompt: PromptKeybindings{
			Yes: []string{"y"},
			No:  []string{"n", "esc"},
		},
	}
}

func defaultEntityKeybindings() EntityKeybindings {
	return EntityKeybindings{
		Add:    []string{"a"},
		Edit:   []string{"e"},
		Delete: []string{"d"},
	}
}

type keybindingEditMod uint8

const (
	unrestrictedKeys keybindingEditMod = iota
	appendOnlyKeys
	lockedKeys
)

type keybindingField struct {
	group   string
	name    string
	keys    *[]string
	editMod keybindingEditMod
}

func (f keybindingField) path() string {
	return f.group + "." + f.name
}

func (f keybindingField) takesSequence() bool {
	return f.group == screenGroup || f.group == connectionGroup
}

func (k *KeybindingSettings) fields() []keybindingField {
	return []keybindingField{
		{navigationGroup, "up", &k.Navigation.Up, appendOnlyKeys},
		{navigationGroup, "down", &k.Navigation.Down, appendOnlyKeys},
		{navigationGroup, "left", &k.Navigation.Left, appendOnlyKeys},
		{navigationGroup, "right", &k.Navigation.Right, appendOnlyKeys},
		{navigationGroup, "next_section", &k.Navigation.NextSection, appendOnlyKeys},
		{navigationGroup, "prev_section", &k.Navigation.PrevSection, appendOnlyKeys},

		{generalGroup, "confirm", &k.General.Confirm, appendOnlyKeys},
		{generalGroup, "cancel", &k.General.Cancel, appendOnlyKeys},
		{generalGroup, "back", &k.General.Back, unrestrictedKeys},
		{generalGroup, "paste", &k.General.Paste, unrestrictedKeys},
		{generalGroup, "copy", &k.General.Copy, unrestrictedKeys},
		{generalGroup, "search", &k.General.Search, unrestrictedKeys},
		{generalGroup, "help", &k.General.Help, unrestrictedKeys},
		{generalGroup, "interrupt", &k.General.Interrupt, lockedKeys},
		{generalGroup, "backspace", &k.General.Backspace, lockedKeys},

		{screenGroup, "transport", &k.Screen.Transport, unrestrictedKeys},
		{screenGroup, "secrets", &k.Screen.Secrets, unrestrictedKeys},
		{screenGroup, "settings", &k.Screen.Settings, unrestrictedKeys},
		{screenGroup, "add_connection", &k.Screen.AddConnection, unrestrictedKeys},
		{screenGroup, "add_connection_type", &k.Screen.AddConnectionType, unrestrictedKeys},

		{connectionGroup, "connect", &k.Connection.Connect, unrestrictedKeys},
		{connectionGroup, editAction, &k.Connection.Edit, unrestrictedKeys},
		{connectionGroup, deleteAction, &k.Connection.Delete, unrestrictedKeys},
		{connectionGroup, "info", &k.Connection.Info, unrestrictedKeys},
		{connectionGroup, pingAction, &k.Connection.Ping, unrestrictedKeys},
		{connectionGroup, retryAction, &k.Connection.Retry, unrestrictedKeys},

		{formGroup, "edit_field", &k.Form.EditField, unrestrictedKeys},
		{formGroup, "secret", &k.Form.Secret, unrestrictedKeys},
		{formGroup, "add_item", &k.Form.AddItem, unrestrictedKeys},
		{formGroup, "remove_item", &k.Form.RemoveItem, unrestrictedKeys},

		{secretGroup, addAction, &k.Secret.Add, unrestrictedKeys},
		{secretGroup, editAction, &k.Secret.Edit, unrestrictedKeys},
		{secretGroup, deleteAction, &k.Secret.Delete, unrestrictedKeys},

		{transportGroup, addAction, &k.Transport.Add, unrestrictedKeys},
		{transportGroup, editAction, &k.Transport.Edit, unrestrictedKeys},
		{transportGroup, deleteAction, &k.Transport.Delete, unrestrictedKeys},

		{settingsGroup, "add_key", &k.Settings.AddKey, unrestrictedKeys},
		{settingsGroup, "change_key", &k.Settings.ChangeKey, unrestrictedKeys},
		{settingsGroup, "remove_key", &k.Settings.RemoveKey, unrestrictedKeys},
		{settingsGroup, "reset", &k.Settings.Reset, unrestrictedKeys},

		{promptGroup, "yes", &k.Prompt.Yes, unrestrictedKeys},
		{promptGroup, "no", &k.Prompt.No, unrestrictedKeys},
	}
}

type Keybinding struct {
	Path       string
	Keys       []string
	Locked     bool
	OnlyAppend bool
	Sequence   bool
}

func (k KeybindingSettings) Bindings() []Keybinding {
	fields := k.fields()
	bindings := make([]Keybinding, 0, len(fields))
	for _, field := range fields {
		bindings = append(bindings, Keybinding{
			Path:       field.path(),
			Keys:       slices.Clone(*field.keys),
			Locked:     field.editMod == lockedKeys,
			OnlyAppend: field.editMod == appendOnlyKeys,
			Sequence:   field.takesSequence(),
		})
	}
	return bindings
}

func (k *KeybindingSettings) Set(path string, bound []string) {
	for _, field := range k.fields() {
		if field.path() == path {
			*field.keys = slices.Clone(bound)
			return
		}
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

func (k *KeybindingSettings) Validate() error {
	return errors.Join(
		k.validateKeys(),
		k.validateEdits(),
		k.validateSharedKeys(),
	)
}

func (k *KeybindingSettings) validateKeys() error {
	var errs []error
	for _, field := range k.fields() {
		for _, key := range *field.keys {
			sequence, err := ParseSequence(key)
			switch {
			case err != nil:
				errs = append(errs, fmt.Errorf("%s%s: %w", keybindingPathPrefix, field.path(), err))
			case len(sequence) > 1 && !field.takesSequence():
				errs = append(errs, fmt.Errorf("%s%s: %q %w", keybindingPathPrefix, field.path(), key, errNoSequence))
			}
		}
	}
	return errors.Join(errs...)
}

func (k *KeybindingSettings) validateEdits() error {
	defaults := DefaultKeybindings()
	defaultFields := defaults.fields()

	var errs []error
	for i, field := range k.fields() {
		defaultKeys := *defaultFields[i].keys
		switch field.editMod {
		case unrestrictedKeys:
		case lockedKeys:
			if !sameKeys(*field.keys, defaultKeys) {
				errs = append(errs, fmt.Errorf("%s%s: %w", keybindingPathPrefix, field.path(), errLocked))
			}
		case appendOnlyKeys:
			for _, key := range defaultKeys {
				if _, ok := findKey(*field.keys, key); !ok {
					errs = append(errs, fmt.Errorf("%s%s: %w %q", keybindingPathPrefix, field.path(), errDefaultRemoved, key))
				}
			}
		}
	}
	return errors.Join(errs...)
}

func sameKeys(bound, others []string) bool {
	if len(bound) != len(others) {
		return false
	}
	for _, key := range others {
		if _, ok := findKey(bound, key); !ok {
			return false
		}
	}
	return true
}

func (k *KeybindingSettings) validateSharedKeys() error {
	all := k.fields()
	in := func(selectors ...string) []keybindingField {
		var fields []keybindingField
		for _, selector := range selectors {
			for _, field := range all {
				if field.group == selector || field.path() == selector {
					fields = append(fields, field)
				}
			}
		}
		return fields
	}

	everywhere := slices.DeleteFunc(in(keybindingEverywhere...), func(f keybindingField) bool { return f.path() == confirmPath })
	errs := make([]error, 0, 2+len(keybindingScreens))
	errs = append(errs, sharedKeys(everywhere, nil), sharedKeys(in(promptGroup), nil))
	for _, screen := range keybindingScreens {
		errs = append(errs, sharedKeys(in(screen...), everywhere))
	}
	return errors.Join(errs...)
}

func sharedKeys(fields, against []keybindingField) error {
	var errs []error
	for i, second := range fields {
		for _, first := range append(slices.Clone(against), fields[:i]...) {
			for _, key := range *first.keys {
				if shared, ok := findKey(*second.keys, key); ok {
					errs = append(errs, fmt.Errorf("%s%s: %w %q with %s", keybindingPathPrefix, second.path(), errSharedKey, shared, first.path()))
				}
			}
		}
	}
	return errors.Join(errs...)
}

func findKey(bound []string, raw string) (string, bool) {
	wanted, err := ParseSequence(raw)
	if err != nil {
		return "", false
	}

	for _, candidate := range bound {
		if sequence, err := ParseSequence(candidate); err == nil && slices.Equal(sequence, wanted) {
			return candidate, true
		}
	}
	return "", false
}
