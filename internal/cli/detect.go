package cli

import (
	"os/exec"

	"github.com/myjupyter/conm/internal/config"
)

type Info struct {
	Name   string
	Path   string
	Exists bool
}

func Detect(t config.ConnType) []Info {
	names := Clients(t)

	infos := make([]Info, 0, len(names))
	for _, name := range names {
		path, err := exec.LookPath(name)
		infos = append(infos, Info{
			Name:   name,
			Path:   path,
			Exists: err == nil,
		})
	}

	return infos
}
