//go:build !linux

package osinfo

import (
	"context"
	"errors"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
)

func (d *Distro) Available() bool { return false }

func (d *Distro) Probe(_ context.Context) (*discovery.Result, error) {
	return nil, errors.New("os.distro: only supported on linux")
}
