//go:build !linux

package resources

import (
	"context"
	"errors"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
)

func (s *System) Available() bool { return false }

func (s *System) Probe(_ context.Context) (*discovery.Result, error) {
	return nil, errors.New("resources.system: only supported on linux")
}
