//go:build !linux

package web

import (
	"context"
	"errors"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
)

func (n *Nginx) Available() bool { return false }

func (n *Nginx) Probe(_ context.Context) (*discovery.Result, error) {
	return nil, errors.New("web.nginx: only supported on linux")
}
