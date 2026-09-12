package cli

import (
	"os/exec"
	"runtime"
	"strings"
)

func ReadClipboard() (string, error) {
	reader := clipboardReader()
	if reader == nil {
		return "", ErrNoClipboardReader
	}

	out, err := reader.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func WriteClipboard(text string) error {
	writer := clipboardWriter()
	if writer == nil {
		return ErrNoClipboardWriter
	}

	writer.Stdin = strings.NewReader(text)

	return writer.Run()
}

func clipboardReader() *exec.Cmd {
	if runtime.GOOS == "darwin" {
		return exec.Command("pbpaste")
	}

	switch {
	case lookup("wl-paste").Installed:
		return exec.Command("wl-paste", "--no-newline")
	case lookup("xclip").Installed:
		return exec.Command("xclip", "-selection", "clipboard", "-o")
	case lookup("xsel").Installed:
		return exec.Command("xsel", "-b", "-o")
	default:
		return nil
	}
}

func clipboardWriter() *exec.Cmd {
	if runtime.GOOS == "darwin" {
		return exec.Command("pbcopy")
	}

	switch {
	case lookup("wl-copy").Installed:
		return exec.Command("wl-copy")
	case lookup("xclip").Installed:
		return exec.Command("xclip", "-selection", "clipboard", "-i")
	case lookup("xsel").Installed:
		return exec.Command("xsel", "-b", "-i")
	default:
		return nil
	}
}
