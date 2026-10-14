package config

import (
	"errors"
	"fmt"

	"github.com/pelletier/go-toml/v2"
)

type TransportKind = string

const SSHTransportKind TransportKind = "ssh"

var TransportKinds = []TransportKind{SSHTransportKind}

var (
	_ ConfigWrapper[Transport] = (*TransportConfigWrapper)(nil)
	_ Distinct                 = Transport{}
)

type Transport struct {
	Name           string        `toml:"name"`
	Kind           TransportKind `toml:"kind"`
	Via            string        `toml:"via"`
	LocalPort      int           `toml:"local_port,omitempty"`
	ConnectTimeout int           `toml:"connect_timeout,omitempty"`
	KeepAlive      int           `toml:"keepalive,omitempty"`
	Description    string        `toml:"description,omitempty"`

	validationErrs []error
}

type TransportConfigWrapper struct {
	Transports []Transport `toml:"transport"`
}

func ValidateTransportName(name string) error {
	if name == "" {
		return errors.New("transport name can't be empty")
	}
	if len(name) > nameMaxLength {
		return fmt.Errorf("transport name must be at most %d characters long", nameMaxLength)
	}
	return nil
}

func ValidateTransportKind(kind TransportKind) error {
	if kind == SSHTransportKind {
		return nil
	}
	return fmt.Errorf("unknown transport kind %q", kind)
}

func ValidateTransportVia(via string) error {
	if via == "" {
		return errors.New("transport needs a connection to go through")
	}
	return nil
}

func ValidateTransportLocalPort(port int) error {
	if port == 0 {
		return nil
	}
	if port < portMinValue || port > portMaxValue {
		return fmt.Errorf("local port must be between %d and %d, or 0 to pick one", portMinValue, portMaxValue)
	}
	return nil
}

func ValidateTransportConnectTimeout(seconds int) error {
	if seconds < 0 {
		return errors.New("connect timeout can't be negative")
	}
	return nil
}

func ValidateTransportDescription(desc string) error {
	if len(desc) > descriptionMaxLength {
		return fmt.Errorf("description must be at most %d characters long", descriptionMaxLength)
	}
	return nil
}

func ValidateTransportKeepAlive(seconds int) error {
	if seconds < 0 {
		return errors.New("keepalive can't be negative")
	}
	return nil
}

func (t Transport) Validate() []error {
	return validationErrors(
		ValidateTransportName(t.Name),
		ValidateTransportKind(t.Kind),
		ValidateTransportVia(t.Via),
		ValidateTransportLocalPort(t.LocalPort),
		ValidateTransportConnectTimeout(t.ConnectTimeout),
		ValidateTransportKeepAlive(t.KeepAlive),
		ValidateTransportDescription(t.Description),
	)
}

func (t Transport) Identity() string {
	return identity(t.Kind, t.Name)
}

func (t Transport) IsValid() bool {
	return len(t.validationErrs) == 0
}

func (w *TransportConfigWrapper) Add(t Transport) {
	w.Transports = append(w.Transports, t)
}

func (w *TransportConfigWrapper) Validate() {
	for i := range w.Transports {
		w.Transports[i].validationErrs = w.Transports[i].Validate()
	}
}

func (w *TransportConfigWrapper) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	return toml.Unmarshal(data, w)
}

func (w *TransportConfigWrapper) Marshal() ([]byte, error) {
	return toml.Marshal(w)
}

func (w *TransportConfigWrapper) Len() int {
	return len(w.Transports)
}

func (w *TransportConfigWrapper) Get(i int) Transport {
	return w.Transports[i]
}

func (w *TransportConfigWrapper) Put(i int, t Transport) {
	w.Transports[i] = t
}

func (w *TransportConfigWrapper) Remove(i int) {
	if i < 0 || i >= len(w.Transports) {
		return
	}
	w.Transports = append(w.Transports[:i], w.Transports[i+1:]...)
}
