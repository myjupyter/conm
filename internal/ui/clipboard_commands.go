package ui

import (
	"strings"
	"unicode"

	"github.com/myjupyter/conm/internal/i18n"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/cli"
)

type clipboardMsg struct {
	text string
	err  error
}

func readClipboardCmd() tea.Cmd {
	return func() tea.Msg {
		text, err := cli.ReadClipboard()

		return clipboardMsg{text: text, err: err}
	}
}

func pasteText(raw string) string {
	flat := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, raw)
	return strings.TrimSpace(flat)
}

func pasteReady(raw string, err error) (text string, refusal formStatus) {
	if err != nil {
		return "", formStatus{text: i18n.Sprintf("couldn't read the clipboard · %v", err), kind: kindErr}
	}

	text = pasteText(raw)
	if text == "" {
		return "", formStatus{text: i18n.Sprintf("clipboard is empty"), kind: kindWarn}
	}
	return text, formStatus{}
}

type clipboardWrittenMsg struct {
	label string
	err   error
}

func writeClipboardCmd(label, text string) tea.Cmd {
	return func() tea.Msg {
		return clipboardWrittenMsg{label: label, err: cli.WriteClipboard(text)}
	}
}
