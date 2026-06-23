//go:build linux

package pkgbackend

import (
	"context"
	"time"

	"github.com/Aakhri-Pastaa/infrasight/pkg/shell"
)

type dpkg struct{}

func (dpkg) Key() string     { return "apt" }
func (dpkg) available() bool { return shell.Available("dpkg-query") }

// Tab/newline are real characters; dpkg-query emits them verbatim.
const dpkgFormat = "-f=${Package}\t${Version}\t${Architecture}\t${Installed-Size}\n"

func (dpkg) List(ctx context.Context, timeout time.Duration) ([]Package, error) {
	out, err := shell.Run(ctx, timeout, "dpkg-query", "-W", dpkgFormat)
	if err != nil && out == "" {
		return nil, err
	}
	return parseDpkgList(out), nil
}

func (dpkg) Owner(ctx context.Context, path string) (string, bool) {
	out, err := shell.Run(ctx, 5*time.Second, "dpkg-query", "-S", path)
	if err != nil || out == "" {
		return "", false
	}
	return parseDpkgOwner(out)
}
