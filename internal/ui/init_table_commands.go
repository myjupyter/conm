package ui

import (
	"fmt"

	"github.com/myjupyter/conm/internal/config"
)

func runTypePicker(start initModel) (config.ConnType, bool, error) {
	m, err := runScreen(start)
	if err != nil {
		return 0, false, err
	}

	im, ok := m.(initModel)
	if !ok {
		return 0, false, fmt.Errorf("connection type picker returned an unexpected model %T", m)
	}

	return im.chosen, im.picked, nil
}
