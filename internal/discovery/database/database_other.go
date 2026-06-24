//go:build !linux

package database

import (
	"context"
	"errors"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
)

func (m *Module) Available() bool { return false }

func (m *Module) Probe(_ context.Context) (*discovery.Result, error) {
	return nil, errors.New("database: only supported on linux")
}
