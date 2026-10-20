package ui

import (
	"errors"

	"github.com/myjupyter/conm/internal/i18n"

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
		return i18n.Sprintf("the tunnel %s is up but could not reach the database · host and port must be the address as the bastion sees it", through)
	default:
		return fallback
	}
}

func hintFor(code network.ErrorCode) string {
	switch code {
	case network.RefusedErrorCode:
		return i18n.Sprintf("is the server running and accepting connections on that host and port?")
	case network.TimeoutErrorCode, network.UnreachableErrorCode:
		return i18n.Sprintf("host unreachable from this network · check the address, VPN, or firewall")
	case network.UnresolvedErrorCode:
		return i18n.Sprintf("the host name does not resolve · check the address or your DNS")
	case network.LocalNetErrorCode:
		return i18n.Sprintf("macOS may be blocking local network access · allow your terminal in System Settings › Privacy & Security › Local Network")
	case network.SecretErrorCode:
		return i18n.Sprintf("the password could not be read · press %s to check the secret store", keyMap.Secret.hint)
	case network.AuthErrorCode:
		return i18n.Sprintf("credentials were rejected · press %s to update the username or password", keyMap.Edit.hint)
	case network.TLSErrorCode:
		return i18n.Sprintf("TLS handshake failed · check the connection's SSL settings and the CA certificate")
	case network.InvalidErrorCode:
		return i18n.Sprintf("this connection has validation errors · press %s to fix the config", keyMap.Edit.hint)
	case network.TransportErrorCode:
		return i18n.Sprintf("the transport did not come up · check the connection it goes through and that it can reach this host")
	case network.PortBusyErrorCode:
		return i18n.Sprintf("the transport's local port is taken · check which process holds it; if it is another connection, give this one its own transport")
	case network.UnknownErrorCode:
		return ""
	}

	return ""
}
