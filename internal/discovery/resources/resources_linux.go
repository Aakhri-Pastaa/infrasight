//go:build linux

package resources

import (
	"context"
	"math"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
	"github.com/Aakhri-Pastaa/infrasight/internal/graph"
)

// Health thresholds (percent).
const (
	cpuWarn  = 85.0
	cpuCrit  = 95.0
	diskWarn = 80.0
	diskCrit = 90.0
)

// pseudoFS are virtual filesystems we never report as disks.
var pseudoFS = map[string]bool{
	"proc": true, "sysfs": true, "cgroup": true, "cgroup2": true, "tmpfs": true,
	"devtmpfs": true, "devpts": true, "mqueue": true, "debugfs": true, "tracefs": true,
	"securityfs": true, "pstore": true, "bpf": true, "configfs": true, "fusectl": true,
	"hugetlbfs": true, "binfmt_misc": true, "autofs": true, "ramfs": true, "nsfs": true,
	"rpc_pipefs": true, "overlay": true,
}

func (s *System) Available() bool {
	_, err := os.Stat("/proc/stat")
	return err == nil
}

func (s *System) Probe(ctx context.Context) (*discovery.Result, error) {
	res := &discovery.Result{}
	now := time.Now().UTC()
	probeCPU(ctx, res, now, s.Name())
	probeDisks(res, now, s.Name())
	return res, nil
}

func probeCPU(ctx context.Context, res *discovery.Result, now time.Time, by string) {
	total1, idle1, ok := readCPUStat()
	if !ok {
		return
	}
	select {
	case <-ctx.Done():
		return
	case <-time.After(200 * time.Millisecond):
	}
	total2, idle2, ok := readCPUStat()
	if !ok || total2 <= total1 {
		return
	}

	dTotal := float64(total2 - total1)
	dIdle := float64(idle2 - idle1)
	util := (1 - dIdle/dTotal) * 100
	util = clamp(util, 0, 100)
	util = math.Round(util*10) / 10

	health := graph.HealthHealthy
	switch {
	case util >= cpuCrit:
		health = graph.HealthCritical
	case util >= cpuWarn:
		health = graph.HealthWarning
	}

	l1, l5, l15 := readLoadavg()

	// Empty label so the model name from hardware.cpu wins on merge.
	res.Nodes = append(res.Nodes, graph.Node{
		ID:     "hardware:cpu",
		Type:   graph.NodeHardware,
		Status: graph.StatusActive,
		Health: health,
		Metadata: map[string]any{
			"utilizationPercent": util,
			"load1":              l1,
			"load5":              l5,
			"load15":             l15,
		},
		Resource:     &graph.ResourceUsage{CPUPercent: &util},
		DiscoveredAt: now,
		DiscoveredBy: by,
	})
}

func probeDisks(res *discovery.Result, now time.Time, by string) {
	data, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return
	}
	seen := map[string]bool{}
	for _, line := range strings.Split(string(data), "\n") {
		f := strings.Fields(line)
		if len(f) < 3 {
			continue
		}
		device, mount, fstype := f[0], unescapeMount(f[1]), f[2]
		if pseudoFS[fstype] || seen[mount] {
			continue
		}

		var st syscall.Statfs_t
		if err := syscall.Statfs(mount, &st); err != nil || st.Blocks == 0 {
			continue
		}
		seen[mount] = true

		bsize := uint64(st.Bsize)
		total := st.Blocks * bsize
		// Match df's convention: "used" excludes filesystem-reserved blocks, and
		// the percentage is used / (used + available-to-users). Counting reserved
		// blocks as used inflates near-empty ext4 volumes (~5% reserved).
		used := (st.Blocks - st.Bfree) * bsize
		avail := st.Bavail * bsize
		denom := used + avail
		var usedPct float64
		if denom > 0 {
			usedPct = math.Round(float64(used)/float64(denom)*1000) / 10
		}

		health := graph.HealthHealthy
		switch {
		case usedPct >= diskCrit:
			health = graph.HealthCritical
		case usedPct >= diskWarn:
			health = graph.HealthWarning
		}

		usedMB := float64(used) / (1024 * 1024)
		res.Nodes = append(res.Nodes, graph.Node{
			ID:     "hardware:disk:" + mount,
			Type:   graph.NodeHardware,
			Label:  mount,
			Status: graph.StatusActive,
			Health: health,
			Metadata: map[string]any{
				"mountpoint":  mount,
				"device":      device,
				"fstype":      fstype,
				"totalBytes":  total,
				"usedBytes":   used,
				"availBytes":  avail,
				"usedPercent": usedPct,
			},
			Resource:     &graph.ResourceUsage{DiskMB: &usedMB},
			DiscoveredAt: now,
			DiscoveredBy: by,
		})
	}
}

// readCPUStat returns total and idle jiffies from the aggregate /proc/stat line.
func readCPUStat() (total, idle uint64, ok bool) {
	b, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0, false
	}
	line := string(b)
	if i := strings.IndexByte(line, '\n'); i >= 0 {
		line = line[:i]
	}
	f := strings.Fields(line) // cpu user nice system idle iowait irq softirq steal ...
	if len(f) < 6 || f[0] != "cpu" {
		return 0, 0, false
	}
	for i := 1; i < len(f); i++ {
		v, err := strconv.ParseUint(f[i], 10, 64)
		if err != nil {
			continue
		}
		total += v
		if i == 4 || i == 5 { // idle + iowait count as idle time
			idle += v
		}
	}
	return total, idle, true
}

func readLoadavg() (l1, l5, l15 float64) {
	b, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return
	}
	f := strings.Fields(string(b))
	if len(f) >= 3 {
		l1, _ = strconv.ParseFloat(f[0], 64)
		l5, _ = strconv.ParseFloat(f[1], 64)
		l15, _ = strconv.ParseFloat(f[2], 64)
	}
	return
}

// unescapeMount decodes the octal escapes (\040 etc.) /proc/mounts uses.
func unescapeMount(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+3 < len(s) {
			if n, err := strconv.ParseInt(s[i+1:i+4], 8, 16); err == nil {
				b.WriteByte(byte(n))
				i += 3
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
