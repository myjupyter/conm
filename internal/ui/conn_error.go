package ui

import (
	"errors"

	"github.com/myjupyter/conm/internal/network"
)

type connError struct {
	op     network.Operation
	code   string
	target string
	during string
	detail string
	hint   string
}

func newConnError(err error, attempted network.Operation) *connError {
	if err == nil {
		return nil
	}

	e := connError{
		op:     attempted,
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
		if failure.Through != "" {
			e.hint = hintThrough(failure.Code, failure.Through, e.hint)
		}
	}

	return &e
}

func hintThrough(code network.ErrorCode, through, fallback string) string {
	switch code {
	case network.RefusedErrorCode, network.TimeoutErrorCode, network.UnreachableErrorCode, network.UnknownErrorCode:
		return "the tunnel " + through + " is up but could not reach the database · host and port must be the address as the bastion sees it"
	default:
		return fallback
	}
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
	case network.TransportErrorCode:
		return "the transport did not come up · check the connection it goes through and that it can reach this host"
	case network.PortBusyErrorCode:
		return "the transport's local port is taken · check which process holds it; if it is another connection, give this one its own transport"
	case network.UnknownErrorCode:
		return ""
	}

	return ""
}
