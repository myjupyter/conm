package config

import (
	"fmt"
	"io"
	"os"
	"reflect"
)

type Connection interface {
	Name() string
	Description() string
	Tags() []string

	Username() string
	Host() string
	Database() string
	Port() int

	URL() string

	ConnType() ConnType

	IsValid() bool
	Validate() []error
}

type ConfigWrapper[C Connection] interface {
	Add(C)
	Len() int
	Get(int) C
	Put(int, C)
	Remove(int)
	Validate()

	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

type File[C Connection] interface {
	Add(...C)
	Len() int
	Get(int) C
	Put(int, C)
	Remove(int)
	Save() error
	Close() error
}

type Config[W ConfigWrapper[C], C Connection] struct {
	w W

	file *os.File
}

func OpenConfig[W ConfigWrapper[C], C Connection](filepath string) (*Config[W, C], error) {
	f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, fmt.Errorf("couldn't open file %q: %w", filepath, err)
	}

	var w W
	if t := reflect.TypeOf(w); t != nil && t.Kind() == reflect.Pointer {
		w = reflect.New(t.Elem()).Interface().(W)
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("couldn't read file %q: %w", filepath, err)
	}

	if err := w.Unmarshal(data); err != nil {
		return nil, fmt.Errorf("couldn't parse file %q: %w", filepath, err)
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
