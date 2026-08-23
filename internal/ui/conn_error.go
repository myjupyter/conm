package ui

import (
	"errors"

	"github.com/myjupyter/conm/internal/network"
)

type connError struct {
	op     network.Operation
	conn   string
	code   string
	target string
	during string
	detail string
	hint   string
}

func newConnError(err error, attempted network.Operation, conn string) *connError {
	if err == nil {
		return nil
	}

	e := connError{
		op:     attempted,
		conn:   conn,
		code:   string(network.UnknownErrorCode),
		detail: err.Error(),
		hint:   hintFor(network.UnknownErrorCode),
	}

	if failure, ok := errors.AsType[*network.OpError](err); ok {
		e.op = failure.Op
		e.code = string(failure.Code)
		e.target = failure.Target
		e.during = failure.During
		e.hint = hintFor(failure.Code)
	}

	return &e
}

func hintFor(code network.ErrorCode) string {
	switch code {
	case network.RefusedErrorCode:
		return "is the server running and accepting connections on that host and port?"
	case network.TimeoutErrorCode, network.UnreachableErrorCode:
		return "host unreachable from this network · check the address, VPN, or firewall"
	case network.SecretErrorCode:
		return "the password could not be read · press " + keyMap.Secret.hint + " to check the secret store"
	case network.AuthErrorCode:
		return "credentials were rejected · press " + keyMap.Edit.hint + " to update the username or password"
	case network.TLSErrorCode:
		return "TLS handshake failed · check the connection's SSL settings and the CA certificate"
	case network.InvalidErrorCode:
		return "this connection has validation errors · press " + keyMap.Edit.hint + " to fix the config"
	case network.UnknownErrorCode:
		return ""
	}

	return ""
}
