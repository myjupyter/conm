package cli

import "errors"

var (
	ErrNoClient      = errors.New("no client configured")
	ErrUnknownClient = errors.New("unknown client")
	ErrConnType      = errors.New("client does not fit connection type")
)
