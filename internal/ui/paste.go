package ui

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
)

// clipboardMsg carries what the system clipboard held when ctrl+v was pressed.
// A terminal that pastes on its own sends tea.PasteMsg instead and never needs
// the read.
type clipboardMsg struct {
	text string
	err  error
}

var errNoClipboardReader = errors.New("no clipboard reader · install wl-clipboard, xclip or xsel")

// clipboardReader is the command that prints the clipboard. A Linux desktop may
// be Wayland or X11, so the readers are tried in turn and the first one
// installed wins.
func clipboardReader() *exec.Cmd {
	if runtime.GOOS == "darwin" {
		return exec.Command("pbpaste")
	}

	switch {
	case hasBinary("wl-paste"):
		return exec.Command("wl-paste", "--no-newline")
	case hasBinary("xclip"):
		return exec.Command("xclip", "-selection", "clipboard", "-o")
	case hasBinary("xsel"):
		return exec.Command("xsel", "-b", "-o")
	default:
		return nil
	}
}

func hasBinary(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func readClipboardCmd() tea.Cmd {
	return func() tea.Msg {
		reader := clipboardReader()
		if reader == nil {
			return clipboardMsg{err: errNoClipboardReader}
		}

		out, err := reader.Output()
		if err != nil {
			return clipboardMsg{err: err}
		}
		return clipboardMsg{text: string(out)}
	}
}

// pasteText flattens a clipboard payload into what a single-line field can
// hold. A url copied from a browser usually arrives with a trailing newline,
// and a multi-line selection would otherwise smuggle control characters into a
// value that is later written to a config file.
func pasteText(raw string) string {
	flat := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, raw)
	return strings.TrimSpace(flat)
}

// pasteReady turns a clipboard read into the text to append, or into the status
// that says why there is none.
func pasteReady(raw string, err error) (text string, refusal formStatus) {
	if err != nil {
		return "", formStatus{text: "couldn't read the clipboard · " + err.Error(), kind: kindErr}
	}

	text = pasteText(raw)
	if text == "" {
		return "", formStatus{text: "clipboard is empty", kind: kindWarn}
	}
	return text, formStatus{}
}
