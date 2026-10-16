package transport

import (
	"fmt"
	"sync"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

type Provider interface {
	Resolve(target config.Connection) (Transport, error)
}

type Pool struct {
	mx      sync.Mutex
	shared  map[string]pooled
	launch  launcher
	secrets secret.Provider
}

type pooled struct {
	settings  string
	transport Transport
}

func NewPool(launch launcher, secrets secret.Provider) *Pool {
	return &Pool{shared: make(map[string]pooled), launch: launch, secrets: secrets}
}

func (p *Pool) Get(rec config.Transport, via, target config.Connection) (Transport, error) {
	key := shareKey(rec, target)
	settings := fmt.Sprintf("%#v %#v", rec, via)

	p.mx.Lock()
	defer p.mx.Unlock()

	if entry, ok := p.shared[key]; ok && entry.settings == settings {
		return entry.transport, nil
	}

	t, err := New(p.launch, rec, via, target, p.secrets)
	if err != nil {
		return nil, err
	}
	p.shared[key] = pooled{settings: settings, transport: t}

	return t, nil
}
