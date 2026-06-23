// Package registry wires the concrete discovery modules together. It lives in
// its own package (rather than inside discovery) so that modules can import the
// discovery interface without creating an import cycle.
//
// To add a module: implement discovery.Module in a new sub-package and append
// its constructor here.
package registry

import (
	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
	"github.com/Aakhri-Pastaa/infrasight/internal/discovery/hardware"
	"github.com/Aakhri-Pastaa/infrasight/internal/discovery/network"
	"github.com/Aakhri-Pastaa/infrasight/internal/discovery/osinfo"
	"github.com/Aakhri-Pastaa/infrasight/internal/discovery/packages"
	"github.com/Aakhri-Pastaa/infrasight/internal/discovery/resources"
	"github.com/Aakhri-Pastaa/infrasight/internal/discovery/services"
	"github.com/Aakhri-Pastaa/infrasight/internal/discovery/web"
)

// All returns every registered module. Modules that cannot run on the current
// host are still returned; the engine filters them via Available().
func All() []discovery.Module {
	return []discovery.Module{
		hardware.NewCPU(),
		hardware.NewMemory(),
		osinfo.NewDistro(),
		network.NewPorts(),
		packages.New(),
		services.NewSystemd(),
		services.NewDocker(),
		web.NewNginx(),
		web.NewApache(),
		resources.NewSystem(),
	}
}
