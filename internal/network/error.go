package network

import (
	"errors"
	"net"
	"strconv"
	"strings"

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
	AuthErrorCode        ErrorCode = "EAUTH"
	SecretErrorCode      ErrorCode = "ESECRET"
	TLSErrorCode         ErrorCode = "ETLS"
	InvalidErrorCode     ErrorCode = "EINVALID"
)

type OpError struct {
	Op     Operation
	Code   ErrorCode
	Target string
	During string
	Err    error
}

func (f *OpError) Error() string {
	return f.Err.Error()
}

func (f *OpError) Unwrap() error {
	return f.Err
}

func cliErrorCode(err error, driver func(error) ErrorCode) ErrorCode {
	if errors.Is(err, cli.ErrNoClient) || errors.Is(err, cli.ErrConnType) {
		return InvalidErrorCode
	}

	return driver(err)
}

func transportErrorCode(err error) ErrorCode {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "connection refused"):
		return RefusedErrorCode
	case strings.Contains(msg, "timeout"), strings.Contains(msg, "deadline exceeded"):
		return TimeoutErrorCode
	case strings.Contains(msg, "no such host"), strings.Contains(msg, "no route to host"):
		return UnreachableErrorCode
	case strings.Contains(msg, "certificate"), strings.Contains(msg, "x509"), strings.Contains(msg, "tls"):
		return TLSErrorCode
	}

	return UnknownErrorCode
}

func during(op Operation, cfg config.Connection) string {
	return op.verb() + " " + address(cfg)
}

func address(cfg config.Connection) string {
	return net.JoinHostPort(cfg.Host(), strconv.Itoa(cfg.Port()))
}
