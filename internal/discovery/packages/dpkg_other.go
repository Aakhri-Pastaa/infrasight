//go:build !linux

package packages

import (
	"context"
	"errors"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
)

func (d *Dpkg) Available() bool { return false }

func (d *Dpkg) Probe(_ context.Context) (*discovery.Result, error) {
	return nil, errors.New("packages.dpkg: only supported on linux")
}
