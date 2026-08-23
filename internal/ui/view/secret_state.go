package view

import (
	"context"
	"fmt"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/repository"
)

func (v *Secrets) Secret() (config.Secret, bool) {
	return v.At(v.Cursor())
}

func (v *Secrets) Usages() []repository.Usage {
	return v.UsagesAt(v.Cursor())
}

func (v *Secrets) Picking() bool {
	return v.picking
}

func (v *Secrets) Picked() string {
	return v.picked
}

// StartPicking turns the view into a chooser and opens it on the entry the
// caller already holds.
func (v *Secrets) StartPicking(current string) {
	v.picking, v.picked = true, ""

	for i := range v.Len() {
		if sec, ok := v.At(i); ok && sec.Location() == current {
			v.SetCursor(i)
			break
		}
	}
}

// StopPicking leaves chooser mode and reports what was attached, if anything.
func (v *Secrets) StopPicking() (string, bool) {
	picked := v.picked
	v.picking, v.picked = false, ""

	return picked, picked != ""
}

func (v *Secrets) Pick() (string, bool) {
	sec, ok := v.Secret()
	if !ok {
		return "", false
	}

	v.picked = sec.Location()

	return v.picked, true
}

func (v *Secrets) Add(ctx context.Context, s config.Secret, password string) error {
	repo, ok := v.repos[s.Provider()]
	if !ok {
		return fmt.Errorf("add: no repository for secret store %q", s.Provider())
	}

	err := repo.Add(ctx, s, password)
	v.Reindex()

	return err
}

func (v *Secrets) Edit(ctx context.Context, s config.Secret, password string) error {
	ref, repo, err := v.rowRepo("edit", v.Cursor())
	if err != nil {
		return err
	}

	err = repo.Edit(ctx, ref.Index, s, password)
	v.Reindex()

	return err
}

func (v *Secrets) Remove(ctx context.Context) error {
	ref, repo, err := v.rowRepo("remove", v.Cursor())
	if err != nil {
		return err
	}

	err = repo.Remove(ctx, ref.Index)
	v.Reindex()

	return err
}
