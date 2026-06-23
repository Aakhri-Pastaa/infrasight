//go:build !linux

package network

import (
	"context"
	"errors"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
)

func (p *Ports) Available() bool { return false }

func (p *Ports) Probe(_ context.Context) (*discovery.Result, error) {
	return nil, errors.New("network.ports: only supported on linux")
}
