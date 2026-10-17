package cli

import "errors"

var (
	ErrNoClient         = errors.New("no client configured")
	ErrUnknownClient    = errors.New("unknown client")
	ErrNoQuery          = errors.New("client cannot run a single query")
	ErrConnType         = errors.New("client does not fit connection type")
	ErrTransport        = errors.New("transport failed")
	ErrUnknownTransport = errors.New("unknown transport")
)

var (
	ErrNoClipboardReader = errors.New("no clipboard reader · install wl-clipboard, xclip or xsel")
	ErrNoClipboardWriter = errors.New("no clipboard writer · install wl-clipboard, xclip or xsel")
)
