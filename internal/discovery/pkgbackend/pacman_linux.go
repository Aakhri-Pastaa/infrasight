//go:build linux

package pkgbackend

import (
	"context"
	"time"

	"github.com/Aakhri-Pastaa/infrasight/pkg/shell"
)

type pacman struct{}

func (pacman) Key() string     { return "pacman" }
func (pacman) available() bool { return shell.Available("pacman") }

func (pacman) List(ctx context.Context, timeout time.Duration) ([]Package, error) {
	out, err := shell.Run(ctx, timeout, "pacman", "-Q")
	if err != nil && out == "" {
		return nil, err
	}
	return parsePacmanList(out), nil
}

func (pacman) Owner(ctx context.Context, path string) (string, bool) {
	out, err := shell.Run(ctx, 5*time.Second, "pacman", "-Qoq", path)
	if err != nil || out == "" {
		return "", false
	}
	return parsePacmanOwner(out)
}
