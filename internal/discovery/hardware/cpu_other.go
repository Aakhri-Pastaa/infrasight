//go:build !linux

package hardware

import (
	"context"
	"errors"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
)

func (c *CPU) Available() bool { return false }

func (c *CPU) Probe(_ context.Context) (*discovery.Result, error) {
	return nil, errors.New("hardware.cpu: only supported on linux")
}
