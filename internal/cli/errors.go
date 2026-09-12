package cli

import "errors"

var (
	ErrNoClient      = errors.New("no client configured")
	ErrUnknownClient = errors.New("unknown client")
	ErrConnType      = errors.New("client does not fit connection type")
)

var (
	ErrNoClipboardReader = errors.New("no clipboard reader · install wl-clipboard, xclip or xsel")
	ErrNoClipboardWriter = errors.New("no clipboard writer · install wl-clipboard, xclip or xsel")
)
