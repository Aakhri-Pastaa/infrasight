package cli

import (
	"strings"
	"testing"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
	"github.com/Aakhri-Pastaa/infrasight/internal/registry"
)

func names(ms []discovery.Module) []string {
	out := make([]string, len(ms))
	for i, m := range ms {
		out[i] = m.Name()
	}
	return out
}

func TestFilterModules(t *testing.T) {
	all := registry.All()
	if len(all) == 0 {
		t.Fatal("registry returned no modules")
	}

	// include by domain prefix ("hardware" matches hardware.cpu, hardware.memory)
	hw := filterModules(all, []string{"hardware"}, nil)
	if len(hw) < 2 {
		t.Fatalf("expected >=2 hardware modules, got %v", names(hw))
	}
	for _, m := range hw {
		if !strings.HasPrefix(m.Name(), "hardware") {
			t.Errorf("domain filter leaked %q", m.Name())
		}
	}

	// include by full name
	if got := filterModules(all, []string{"network.ports"}, nil); len(got) != 1 || got[0].Name() != "network.ports" {
		t.Errorf("full-name include = %v", names(got))
	}

	// exclude removes exactly that module
	excl := filterModules(all, nil, []string{"packages"})
	if len(excl) != len(all)-1 {
		t.Errorf("exclude count = %d, want %d", len(excl), len(all)-1)
	}
	for _, m := range excl {
		if m.Name() == "packages" {
			t.Error("packages should have been excluded")
		}
	}

	// no filters returns everything
	if len(filterModules(all, nil, nil)) != len(all) {
		t.Error("no filter should return all modules")
	}

	// unknown selector matches nothing
	if got := filterModules(all, []string{"does-not-exist"}, nil); len(got) != 0 {
		t.Errorf("unknown selector matched %v", names(got))
	}
}
