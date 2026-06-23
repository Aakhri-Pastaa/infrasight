//go:build linux

package pkgbackend

import (
	"context"
	"time"

	"github.com/Aakhri-Pastaa/infrasight/pkg/shell"
)

type rpm struct{}

func (rpm) Key() string     { return "rpm" }
func (rpm) available() bool { return shell.Available("rpm") }

// Real tab/newline; rpm prints them between fields / records.
const rpmFormat = "%{NAME}\t%{VERSION}-%{RELEASE}\t%{ARCH}\t%{SIZE}\n"

func (rpm) List(ctx context.Context, timeout time.Duration) ([]Package, error) {
	out, err := shell.Run(ctx, timeout, "rpm", "-qa", "--qf", rpmFormat)
	if err != nil && out == "" {
		return nil, err
	}
	return parseRpmList(out), nil
}

func (rpm) Owner(ctx context.Context, path string) (string, bool) {
	// On a miss, rpm exits non-zero ("file ... is not owned by any package").
	out, err := shell.Run(ctx, 5*time.Second, "rpm", "-qf", "--qf", "%{NAME}", path)
	if err != nil {
		return "", false
	}
	return parseRpmOwner(out)
}
