//go:build !linux

package services

import (
	"context"
	"errors"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
)

func (d *Docker) Available() bool { return false }

func (d *Docker) Probe(_ context.Context) (*discovery.Result, error) {
	return nil, errors.New("services.docker: only supported on linux")
}
