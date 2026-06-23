// Package packages discovers installed software across package managers.
// Today it implements dpkg (Debian/Ubuntu); apt/rpm/pip/npm/etc. follow the
// same shape.
package packages

import (
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
)

// Dpkg lists packages from the Debian package database.
type Dpkg struct{}

// NewDpkg constructs the dpkg module.
func NewDpkg() *Dpkg { return &Dpkg{} }

func (d *Dpkg) Name() string            { return "packages.dpkg" }
func (d *Dpkg) Description() string     { return "Installed Debian/Ubuntu packages (dpkg)" }
func (d *Dpkg) RequiredTools() []string { return []string{"dpkg-query"} }
func (d *Dpkg) RequiresRoot() bool      { return false }
func (d *Dpkg) RequiresNetwork() bool   { return false }
func (d *Dpkg) Timeout() time.Duration  { return 30 * time.Second }
func (d *Dpkg) Risk() discovery.Risk    { return discovery.RiskLow }
