package transport

import "errors"

var (
	ErrSecret        = errors.New("bastion secret could not be read")
	ErrLocalPortBusy = errors.New("local port is not free")
	ErrNotReady      = errors.New("transport did not come up in time")
)
