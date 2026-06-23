//go:build linux

package hardware

import (
	"strconv"
	"strings"
)

// splitKV splits a "key: value" or "key value" /proc line on the first colon.
func splitKV(line string) (key, value string, ok bool) {
	i := strings.IndexByte(line, ':')
	if i < 0 {
		return "", "", false
	}
	return strings.TrimSpace(line[:i]), strings.TrimSpace(line[i+1:]), true
}

// meminfoKB parses /proc/meminfo into a map of label -> kilobytes.
func meminfoKB(data []byte) map[string]int64 {
	out := make(map[string]int64)
	for _, line := range strings.Split(string(data), "\n") {
		key, val, ok := splitKV(line)
		if !ok {
			continue
		}
		fields := strings.Fields(val) // e.g. ["32791234", "kB"]
		if len(fields) == 0 {
			continue
		}
		n, err := strconv.ParseInt(fields[0], 10, 64)
		if err != nil {
			continue
		}
		out[key] = n
	}
	return out
}
