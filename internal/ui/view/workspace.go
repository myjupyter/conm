package view

import (
	"context"
	"fmt"

	"github.com/myjupyter/conm/internal/repository"
)

const workspaceKind = "workspace"

type WorkspaceRef = Ref[string]

type Workspaces struct {
	repo repository.Workspaces

	tableState[string]
}

func NewWorkspaces(repo repository.Workspaces) *Workspaces {
	v := &Workspaces{repo: repo}
	v.register(workspaceKind)
	v.Reindex()

	return v
}

func (v *Workspaces) Reindex() {
	entries := make([]entry[string], 0, v.repo.Len())
	for i := range v.repo.Len() {
		name, ok := v.repo.Get(i)
		if !ok {
			continue
		}

		entries = append(entries, entry[string]{
			ref:    WorkspaceRef{Kind: workspaceKind, Index: i},
			fields: searchable(name),
		})
	}

	v.reset(entries)
}

func (v *Workspaces) At(i int) (string, bool) {
	ref, ok := v.RefAt(i)
	if !ok {
		return "", false
	}

	return v.repo.Get(ref.Index)
}

func (v *Workspaces) Workspace() (string, bool) {
	return v.At(v.Cursor())
}

func (v *Workspaces) Current() string {
	return v.repo.Active()
}

func (v *Workspaces) Use() error {
	ref, ok := v.Ref()
	if !ok {
		return fmt.Errorf("use: row %d is out of range", v.Cursor())
	}

	return v.repo.Use(ref.Index)
}

func (v *Workspaces) Add(name string) error {
	err := v.repo.Add(name)
	v.Reindex()

	return err
}

func (v *Workspaces) Rename(ctx context.Context, name string) error {
	ref, ok := v.Ref()
	if !ok {
		return fmt.Errorf("rename: row %d is out of range", v.Cursor())
	}

	err := v.repo.Rename(ctx, ref.Index, name)
	v.Reindex()

	return err
}

func (v *Workspaces) Remove(ctx context.Context) error {
	ref, ok := v.Ref()
	if !ok {
		return fmt.Errorf("remove: row %d is out of range", v.Cursor())
	}

	err := v.repo.Remove(ctx, ref.Index)
	v.Reindex()

	return err
}
