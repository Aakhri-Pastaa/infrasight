//go:build !linux

package hardware

import (
	"context"
	"errors"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
)

func (m *Memory) Available() bool { return false }

func (m *Memory) Probe(_ context.Context) (*discovery.Result, error) {
	return nil, errors.New("hardware.memory: only supported on linux")
}
