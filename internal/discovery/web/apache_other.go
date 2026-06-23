//go:build !linux

package web

import (
	"context"
	"errors"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
)

func (a *Apache) Available() bool { return false }

func (a *Apache) Probe(_ context.Context) (*discovery.Result, error) {
	return nil, errors.New("web.apache: only supported on linux")
}
