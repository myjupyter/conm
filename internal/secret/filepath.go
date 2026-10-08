package secret

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var _ Provider = (*FilepathProvider)(nil)

type FilepathProvider struct{}

func (*FilepathProvider) Scheme() Scheme { return Filepath }

func (*FilepathProvider) Resolve(_ context.Context, ref Reference) (string, error) {
	_, path, _ := ParseRef(ref.SecretRef())

	if _, err := os.Stat(expandHome(path)); err != nil {
		return "", fmt.Errorf("%s: %w", path, err)
	}

	return path, nil
}

func (*FilepathProvider) Store(_ context.Context, _ Reference, _ string) error { return nil }

func (*FilepathProvider) Remove(_ context.Context, _ Reference) error { return nil }

func expandHome(path string) string {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	return filepath.Join(home, strings.TrimPrefix(path, "~"))
}
