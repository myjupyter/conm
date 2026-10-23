package cli

import (
	"os/exec"

	"github.com/myjupyter/conm/internal/config"
)

type Info struct {
	Name      string
	Path      string
	Installed bool
	Version   string
}

func Detect(t config.ConnType) []Info {
	names := Clients(t)

	infos := make([]Info, 0, len(names))
	for _, name := range names {
		infos = append(infos, lookup(name))
	}

	return infos
}

func lookup(name string) Info {
	path, err := exec.LookPath(name)

	return Info{
		Name:      name,
		Path:      path,
		Installed: err == nil,
	}
}

func Installed(t config.ConnType) (string, bool) {
	for _, name := range Clients(t) {
		if lookup(name).Installed {
			return name, true
		}
	}

	return "", false
}

func Resolve(t config.ConnType, configured string) string {
	if configured != "" {
		return configured
	}

	name, _ := Installed(t)

	return name
}
