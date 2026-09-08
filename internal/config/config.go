package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
)

type Connection interface {
	Meta() ConnMeta

	Username() string
	Host() string
	Database() string
	Schema() string
	Port() int

	ConnectionString(string) string

	ConnType() ConnType

	IsValid() bool
	Validate() []error
}

func validationErrors(checks ...error) []error {
	var errs []error
	for _, err := range checks {
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

type ConfigWrapper[C any] interface {
	// TODO: Add method to return error
	Add(C)
	Len() int
	Get(int) C
	// TODO: Put method to return error
	Put(int, C)
	Remove(int)
	Validate()

	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

type File[C any] interface {
	Add(...C)
	Len() int
	Get(int) C
	Put(int, C)
	Remove(int)
	Save() error
	Close() error
}

type Config[W ConfigWrapper[C], C any] struct {
	w W

	file *os.File
}

func OpenConfig[W ConfigWrapper[C], C any](path string) (*Config[W, C], error) {
	// Config and data dirs live under different XDG roots; either may be missing.
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("couldn't create dir for %q: %w", path, err)
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, fmt.Errorf("couldn't open file %q: %w", path, err)
	}

	var w W
	if t := reflect.TypeOf(w); t != nil && t.Kind() == reflect.Pointer {
		alloc, ok := reflect.TypeAssert[W](reflect.New(t.Elem()))
		if !ok {
			f.Close()
			return nil, fmt.Errorf("couldn't allocate a %T wrapper for %q", w, path)
		}
		w = alloc
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("couldn't read file %q: %w", path, err)
	}

	if err := w.Unmarshal(data); err != nil {
		return nil, fmt.Errorf("couldn't parse file %q: %w", path, err)
	}

	w.Validate()

	return &Config[W, C]{
		w:    w,
		file: f,
	}, nil
}

func (c *Config[W, C]) Add(conns ...C) {
	for _, conn := range conns {
		c.w.Add(conn)
	}
}

func (c *Config[W, C]) Len() int { return c.w.Len() }

func (c *Config[W, C]) Get(i int) C { return c.w.Get(i) }

func (c *Config[W, C]) Put(i int, conn C) { c.w.Put(i, conn) }

func (c *Config[W, C]) Remove(i int) { c.w.Remove(i) }

func (c *Config[W, C]) Save() error {
	data, err := c.w.Marshal()
	if err != nil {
		return fmt.Errorf("couldn't marshal config for %q: %w", c.file.Name(), err)
	}

	if _, err := c.file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("couldn't save config to %q: %w", c.file.Name(), err)
	}

	n, err := c.file.Write(data)
	if err != nil {
		return fmt.Errorf("couldn't save config to %q: %w", c.file.Name(), err)
	}

	if err := c.file.Truncate(int64(n)); err != nil {
		return fmt.Errorf("couldn't save config to %q: %w", c.file.Name(), err)
	}

	return nil
}

func (c *Config[W, C]) Close() error {
	return c.file.Close()
}
