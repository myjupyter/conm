package network

import (
	"context"
	"errors"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
)

type Operation string

const (
	PingOperation    Operation = "ping"
	ConnectOperation Operation = "connect"
)

func (o Operation) verb() string {
	if o == ConnectOperation {
		return "open session on"
	}
	return "dial tcp"
}

type ErrorCode string

const (
	UnknownErrorCode     ErrorCode = "ECONN"
	RefusedErrorCode     ErrorCode = "ECONNREFUSED"
	TimeoutErrorCode     ErrorCode = "ETIMEDOUT"
	UnreachableErrorCode ErrorCode = "EHOSTUNREACH"
	UnresolvedErrorCode  ErrorCode = "ENOTFOUND"
	LocalNetErrorCode    ErrorCode = "ELOCALNET"
	AuthErrorCode        ErrorCode = "EAUTH"
	SecretErrorCode      ErrorCode = "ESECRET"
	TLSErrorCode         ErrorCode = "ETLS"
	InvalidErrorCode     ErrorCode = "EINVALID"
	TransportErrorCode   ErrorCode = "ETRANSPORT"
	PortBusyErrorCode    ErrorCode = "EADDRINUSE"
)

type OpError struct {
	Op      Operation
	Code    ErrorCode
	Target  string
	During  string
	Through string
	Err     error
}

func (f *OpError) Error() string {
	return f.Err.Error()
}

func (f *OpError) Unwrap() error {
	return f.Err
}

func cliErrorCode(err error, driver func(error) ErrorCode) ErrorCode {
	if errors.Is(err, cli.ErrNoClient) || errors.Is(err, cli.ErrConnType) || errors.Is(err, cli.ErrNoQuery) {
		return InvalidErrorCode
	}

	return driver(err)
}

func transportErrorCode(err error) ErrorCode {
	return transportErrorCodeOn(runtime.GOOS, err)
}

func transportErrorCodeOn(goos string, err error) ErrorCode {
	if code, ok := errnoErrorCode(goos, err); ok {
		return code
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "connection refused"):
		return RefusedErrorCode
	case strings.Contains(msg, "timeout"), strings.Contains(msg, "deadline exceeded"):
		return TimeoutErrorCode
	case strings.Contains(msg, "no such host"):
		return UnresolvedErrorCode
	case strings.Contains(msg, "no route to host"):
		return UnreachableErrorCode
	case strings.Contains(msg, "certificate"), strings.Contains(msg, "x509"), strings.Contains(msg, "tls"):
		return TLSErrorCode
	}

	return UnknownErrorCode
}

func errnoErrorCode(goos string, err error) (ErrorCode, bool) {
	switch {
	case errors.Is(err, syscall.ECONNREFUSED):
		return RefusedErrorCode, true
	case errors.Is(err, syscall.EHOSTUNREACH):
		if goos == "darwin" && dialedLocalNetwork(err) {
			return LocalNetErrorCode, true
		}
		return UnreachableErrorCode, true
	case hostNotFound(err):
		return UnresolvedErrorCode, true
	case timedOut(err):
		return TimeoutErrorCode, true
	}

	return "", false
}

func hostNotFound(err error) bool {
	dnsErr, ok := errors.AsType[*net.DNSError](err)
	return ok && dnsErr.IsNotFound
}

func timedOut(err error) bool {
	if errors.Is(err, syscall.ETIMEDOUT) || errors.Is(err, os.ErrDeadlineExceeded) ||
		errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	netErr, ok := errors.AsType[net.Error](err)
	return ok && netErr.Timeout()
}

func dialedLocalNetwork(err error) bool {
	opErr, ok := errors.AsType[*net.OpError](err)
	if !ok {
		return false
	}
	addr, ok := opErr.Addr.(*net.TCPAddr)

	return ok && (addr.IP.IsPrivate() || addr.IP.IsLinkLocalUnicast())
}

func during(op Operation, cfg config.Connection) string {
	return op.verb() + " " + address(cfg)
}

func address(cfg config.Connection) string {
	return net.JoinHostPort(cfg.Host(), strconv.Itoa(cfg.Port()))
}
