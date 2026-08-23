package ui

import (
	"context"
	"fmt"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/network"
	"github.com/myjupyter/conm/internal/secret"
	"github.com/myjupyter/conm/internal/ui/spec"
	"github.com/myjupyter/conm/internal/ui/view"
)

type formModel struct {
	spec     spec.FormSpec[config.Connection]
	title    string
	isEdit   bool
	sections []spec.FormSection

	// secrets backs the store modes of the secret field. It is nil on the
	// setup path, where no repository exists yet — there the form is
	// literal-only.
	secrets *view.Secrets

	// The value field is shared by both modes, so each mode's value is kept
	// aside while the other is showing: cycling through the modes must not
	// throw away a password that was already typed.
	literalStash string
	refStash     string

	vals map[string]string

	section int
	idx     int
	insert  bool
	reveal  bool
	help    bool // the key hints table, expanded over the key bar

	attempted bool
	submitted bool

	status     string
	statusKind statusKind

	pong string     // successful in-form ping banner
	ping *connError // failed in-form ping panel
}

type formPingMsg struct {
	conn   config.Connection
	result network.PingResult
	err    error
}

// secretPickedMsg carries a location back from the keyring screen.
type secretPickedMsg struct {
	ref string
	ok  bool
	err error
}

func newFormModel(spc spec.FormSpec[config.Connection], title string, initial map[spec.FormFieldKey]spec.FormFieldValue, secrets *view.Secrets) formModel {
	m := formModel{
		spec:       spc,
		title:      title,
		isEdit:     initial != nil,
		sections:   spc.Sections,
		secrets:    secrets,
		vals:       make(map[string]string, len(spc.Fields)),
		status:     statusReady,
		statusKind: kindIdle,
	}

	if len(m.sections) == 0 {
		keys := make([]spec.FormFieldKey, len(spc.Fields))
		for i, f := range spc.Fields {
			keys[i] = f.Key
		}
		m.sections = []spec.FormSection{{Fields: keys}}
	}

	for _, f := range spc.Fields {
		switch {
		case initial != nil:
			m.vals[f.Key] = initial[f.Key]
		case f.Kind == spec.SelectFieldKind:
			m.vals[f.Key] = f.DefaultValue
			if m.vals[f.Key] == "" && len(f.Options) > 0 {
				m.vals[f.Key] = f.Options[0]
			}
		default:
			m.vals[f.Key] = ""
		}
	}

	return m
}

func (m formModel) Init() tea.Cmd { return nil }

func (m formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.insert {
			return m.insertKey(msg)
		}
		return m.navKey(msg)
	case formPingMsg:
		return m.applyPing(msg), nil
	case secretPickedMsg:
		return m.applyPicked(msg), nil
	}
	return m, nil
}

// isRef reports whether the password field currently points at a secret store
// rather than holding the password itself.
func (m formModel) isRef() bool {
	mode := m.vals[spec.SecretProviderKey]
	return mode != "" && mode != secret.Literal
}

// isSecretField reports whether field i is the one the secret mode governs.
func (m formModel) isSecretField(i int) bool {
	return m.spec.Fields[i].Key == spec.SecretValueKey && m.hasSecretMode()
}

// isProviderField reports whether field i is the secret mode selector.
func (m formModel) isProviderField(i int) bool {
	return m.spec.Fields[i].Key == spec.SecretProviderKey
}

// isSelector reports whether field i is a list the cursor can cycle: a select
// with options to cycle through. A select declared without options has nothing
// to change, so ←/→ stays inert there rather than pretending to be a control.
func (m formModel) isSelector(i int) bool {
	f := m.spec.Fields[i]
	return f.Kind == spec.SelectFieldKind && len(f.Options) > 0
}

func (m formModel) hasSecretMode() bool {
	for _, f := range m.spec.Fields {
		if f.Key == spec.SecretProviderKey {
			return true
		}
	}
	return false
}

func (m formModel) storeLabel() string {
	if mode := m.vals[spec.SecretProviderKey]; mode != "" {
		return mode
	}
	return secret.Literal
}

// refEntries lists the locations selectable in the current store: the mode
// names the store, and only that store's entries can satisfy it.
func (m formModel) refEntries() []string {
	if m.secrets == nil {
		return nil
	}
	return m.secrets.LocationsOf(m.vals[spec.SecretProviderKey])
}

// cycleRef steps through the store's entries in place of the character input
// the field would otherwise take.
func (m *formModel) cycleRef(dir int) {
	list := m.refEntries()
	if len(list) == 0 {
		m.setStatus(m.storeLabel()+" is empty · s on provider to add an entry", kindWarn)
		return
	}

	at := -1
	for j, loc := range list {
		if loc == m.vals[spec.SecretValueKey] {
			at = j
			break
		}
	}
	if at < 0 {
		at = -1
		if dir < 0 {
			at = 0
		}
	}

	next := list[(at+dir+len(list))%len(list)]
	m.vals[spec.SecretValueKey] = next
	m.setStatus(next+" · "+m.storeLabel()+" entry", kindIdle)
}

// onSecretModeChange keeps the value field consistent with the mode it just
// moved to: a literal password and a store location are not interchangeable,
// so each is parked in its own stash while the other is on screen.
func (m *formModel) onSecretModeChange(was string) {
	if was == "" || was == secret.Literal {
		m.literalStash = m.vals[spec.SecretValueKey]
	} else {
		m.refStash = m.vals[spec.SecretValueKey]
	}

	if !m.isRef() {
		m.vals[spec.SecretValueKey] = m.literalStash
		m.setStatus("literal · password stored as plain text", kindWarn)
		return
	}

	m.vals[spec.SecretValueKey] = m.refStash

	list := m.refEntries()
	if !slices.Contains(list, m.vals[spec.SecretValueKey]) {
		m.vals[spec.SecretValueKey] = ""
		if len(list) > 0 {
			m.vals[spec.SecretValueKey] = list[0]
		}
	}
	m.setStatus(m.storeLabel()+" · pick an entry with "+keyMap.Cycle.hint+", or "+keyMap.Secret.hint+" on provider to manage", kindIdle)
}

func (m formModel) applyPicked(msg secretPickedMsg) formModel {
	switch {
	case msg.err != nil:
		m.setStatus(msg.err.Error(), kindErr)
	case msg.ok:
		m.vals[spec.SecretValueKey] = msg.ref
		m.setStatus("attached "+msg.ref+" from "+m.storeLabel(), kindOK)
	default:
		m.setStatus("nothing picked", kindIdle)
	}
	return m
}

func (m formModel) navKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	fields := m.sectionFields(m.section)
	n := len(fields)
	cur := m.spec.Fields[fields[m.idx]]

	key := msg.String()

	switch {
	case keyMap.Interrupt.matches(key):
		return m, tea.Quit
	case keyMap.Cancel.matches(key):
		if m.pong != "" || m.ping != nil {
			m.pong, m.ping = "", nil
			m.setStatus(statusReady, kindIdle)
			return m, nil
		}
		return m, tea.Quit
	case keyMap.Down.matches(key):
		m.idx = (m.idx + 1) % n
	case keyMap.Up.matches(key):
		m.idx = (m.idx - 1 + n) % n
	case keyMap.NextSection.matches(key):
		m.switchSection(1)
	case keyMap.PrevSection.matches(key):
		m.switchSection(-1)
	case keyMap.Right.matches(key):
		m.navCycle(fields[m.idx], cur, 1)
	case keyMap.Left.matches(key):
		m.navCycle(fields[m.idx], cur, -1)
	case keyMap.Edit.matches(key):
		m.navEdit(fields[m.idx], cur)
	case keyMap.Secret.matches(key):
		if cmd, handled := m.navSecret(fields[m.idx], cur); handled {
			return m, cmd
		}
	case keyMap.Ping.matches(key):
		return m.pingForm()
	case keyMap.Retry.matches(key):
		if m.ping != nil {
			return m.pingForm()
		}
	case keyMap.Confirm.matches(key):
		return m.submit()
	case keyMap.Help.matches(key):
		m.help = !m.help
		m.setStatus(keyhintStatus(m.help), kindIdle)
	}
	return m, nil
}

// navCycle steps the field under the cursor one option in dir: a store entry
// for a secret reference, otherwise the selector's own options.
func (m *formModel) navCycle(i int, cur spec.FormField, dir int) {
	switch {
	case m.isSecretField(i) && m.isRef():
		m.cycleRef(dir)
	case m.isSelector(i):
		was := m.vals[spec.SecretProviderKey]
		m.cycle(i, dir)
		if cur.Key == spec.SecretProviderKey {
			m.onSecretModeChange(was)
		}
	}
}

// navEdit enters insert mode, or explains why the field under the cursor is
// not typed into.
func (m *formModel) navEdit(i int, cur spec.FormField) {
	switch {
	case m.isSecretField(i) && m.isRef():
		m.setStatus(m.storeLabel()+" entries are picked, not typed · use "+keyMap.Cycle.hint, kindWarn)
	case m.isSelector(i):
		m.setStatus(strings.ToLower(cur.Label)+" is a list · use "+keyMap.Cycle.hint, kindWarn)
	case cur.Kind == spec.SelectFieldKind:
		m.setStatus(strings.ToLower(cur.Label)+" has no options to pick from", kindWarn)
	default:
		m.insert = true
		m.setStatus("editing "+strings.ToLower(cur.Label)+" · "+keyMap.Cancel.hint+" when done", kindIdle)
	}
}

// navSecret handles s on a secret field. On the provider field it opens the
// store itself: an entry can be added there and picked, and the form resumes
// where it left off. On a literal password field s is the reveal toggle.
// handled reports whether the returned cmd should be dispatched.
func (m *formModel) navSecret(i int, cur spec.FormField) (cmd tea.Cmd, handled bool) {
	switch {
	case m.isProviderField(i) && m.isRef():
		return m.pickSecretCmd(), true
	case cur.Kind == spec.HiddenFieldKind && !m.isRef():
		m.reveal = !m.reveal
		if m.reveal {
			m.setStatus("password visible · "+keyMap.Secret.hint+" to hide", kindIdle)
		} else {
			m.setStatus("password hidden", kindIdle)
		}
	}
	return nil, false
}

func (m formModel) insertKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	fields := m.sectionFields(m.section)
	cur := m.spec.Fields[fields[m.idx]]

	key := msg.String()

	switch {
	case keyMap.Interrupt.matches(key):
		return m, tea.Quit
	case keyMap.Cancel.matches(key):
		m.insert = false
		m.setStatus("done editing · "+keyMap.MoveVertical.hint+" to move", kindIdle)
		return m, nil
	case keyMap.Confirm.matches(key):
		m.insert = false
		if m.idx < len(fields)-1 {
			m.idx++
		}
		return m, nil
	case keyMap.NextSection.matches(key):
		m.insert = false
		m.switchSection(1)
		return m, nil
	case keyMap.Backspace.matches(key):
		if r := []rune(m.vals[cur.Key]); len(r) > 0 {
			m.vals[cur.Key] = string(r[:len(r)-1])
		}
		return m, nil
	}

	if msg.Text != "" {
		m.vals[cur.Key] += msg.Text
	}
	return m, nil
}

func (m *formModel) setStatus(s string, k statusKind) {
	m.status, m.statusKind = s, k
}

func (m *formModel) switchSection(dir int) {
	m.section = (m.section + dir + len(m.sections)) % len(m.sections)
	m.idx = 0
	s := m.sections[m.section]
	m.setStatus(s.Title+" · "+s.Note, kindIdle)
}

func (m *formModel) cycle(i, dir int) {
	f := m.spec.Fields[i]
	if len(f.Options) == 0 {
		return
	}
	at := 0
	for j, opt := range f.Options {
		if opt == m.vals[f.Key] {
			at = j
			break
		}
	}
	m.vals[f.Key] = f.Options[(at+dir+len(f.Options))%len(f.Options)]
}

// sectionFields returns the indices into spec.Fields for the fields in section si,
// in the section's declared order.
func (m formModel) sectionFields(si int) []int {
	keys := m.sections[si].Fields
	out := make([]int, 0, len(keys))
	for _, key := range keys {
		for i, f := range m.spec.Fields {
			if f.Key == key {
				out = append(out, i)
				break
			}
		}
	}
	return out
}

func (m formModel) currentField() int {
	return m.sectionFields(m.section)[m.idx]
}

func (m formModel) fieldError(i int) string {
	f := m.spec.Fields[i]

	// In a store mode the field holds a location, so it is checked against the
	// entries that exist rather than against the password validator.
	if m.isSecretField(i) && m.isRef() {
		return m.refError()
	}

	v := m.vals[f.Key]
	if f.Kind != spec.HiddenFieldKind {
		v = strings.TrimSpace(v)
	}

	if f.Property == spec.RequiredFieldProperty && strings.TrimSpace(v) == "" {
		return strings.ToLower(f.Label) + " is required"
	}
	if f.ValidateFunc != nil {
		if err := f.ValidateFunc(v); err != nil {
			return err.Error()
		}
	}
	return ""
}

// refError validates the picked location: an empty one is not a blank password
// but an unfinished choice, and a stale one would resolve to nothing at connect
// time.
func (m formModel) refError() string {
	ref := strings.TrimSpace(m.vals[spec.SecretValueKey])
	store := m.storeLabel()

	if ref == "" {
		return "pick a " + store + " entry — " + keyMap.Cycle.hint + ", or " + keyMap.Secret.hint + " on provider"
	}
	if slices.Contains(m.refEntries(), ref) {
		return ""
	}
	return ref + " is not in " + store
}

func (m formModel) sectionErrCount(si int) int {
	n := 0
	for _, i := range m.sectionFields(si) {
		if m.fieldError(i) != "" {
			n++
		}
	}
	return n
}

func (m formModel) submit() (tea.Model, tea.Cmd) {
	m.attempted = true

	bad, badSection, badIdx := 0, -1, -1
	for si := range m.sections {
		for j, i := range m.sectionFields(si) {
			if m.fieldError(i) != "" {
				bad++
				if badSection < 0 {
					badSection, badIdx = si, j
				}
			}
		}
	}

	if bad > 0 {
		m.section, m.idx, m.insert = badSection, badIdx, false
		m.setStatus(fmt.Sprintf("%d %s need attention", bad, plural(bad, "field", "fields")), kindErr)
		return m, nil
	}

	m.submitted = true
	return m, tea.Quit
}

func (m formModel) pingForm() (tea.Model, tea.Cmd) {
	conn, err := m.spec.BuildFunc(m.values())
	if err != nil || conn.Host() == "" {
		m.attempted = true
		m.setStatus("fill host before pinging", kindErr)
		return m, nil
	}

	m.pong, m.ping = "", nil
	m.setStatus("pinging "+conn.Host()+" …", kindPending)
	return m, pingFormCmd(conn)
}

func pingFormCmd(conn config.Connection) tea.Cmd {
	return func() tea.Msg {
		client, err := network.NewConnection(config.Conm{}, conn, secret.Default())
		if err != nil {
			return formPingMsg{conn: conn, err: err}
		}
		defer client.Close()

		result, err := client.Ping(context.Background())
		return formPingMsg{conn: conn, result: result, err: err}
	}
}

func (m formModel) applyPing(msg formPingMsg) formModel {
	c := msg.conn
	label := c.Name()
	if label == "" {
		label = c.Host()
	}

	if msg.err != nil {
		e := newConnError(msg.err, network.PingOperation, label)
		m.ping = e
		m.pong = ""
		m.setStatus("ping failed · "+label+" · "+e.code, kindErr)
		return m
	}

	ms := msg.result.PingTime.Milliseconds()
	m.ping = nil
	m.pong = fmt.Sprintf("%s:%d answered in %dms", c.Host(), c.Port(), ms)
	m.setStatus(fmt.Sprintf("pong · %dms", ms), kindOK)
	return m
}

func (m formModel) result() (config.Connection, error) {
	return m.spec.BuildFunc(m.values())
}

// values collects the current field values keyed by FormField.Key. Every value
// is trimmed except hidden fields (passwords), where surrounding whitespace may
// be meaningful.
func (m formModel) values() map[spec.FormFieldKey]spec.FormFieldValue {
	out := make(map[spec.FormFieldKey]spec.FormFieldValue, len(m.spec.Fields))
	for _, f := range m.spec.Fields {
		v := m.vals[f.Key]
		// A location is trimmed like any other field; only a literal password
		// keeps its surrounding whitespace.
		if f.Kind != spec.HiddenFieldKind || m.isRef() {
			v = strings.TrimSpace(v)
		}
		out[f.Key] = v
	}
	return out
}
