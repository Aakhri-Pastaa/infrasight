//go:build !linux

package services

import (
	"context"
	"errors"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
)

func (s *Systemd) Available() bool { return false }

func (s *Systemd) Probe(_ context.Context) (*discovery.Result, error) {
	return nil, errors.New("services.systemd: only supported on linux")
}
