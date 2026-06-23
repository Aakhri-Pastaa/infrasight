//go:build linux

package pkgbackend

import "sync"

var (
	detectOnce sync.Once
	detected   Backend
)

// Detect returns the first available package backend on this host, or nil if
// none is recognised. The result is memoised.
func Detect() Backend {
	detectOnce.Do(func() {
		for _, b := range []Backend{dpkg{}, rpm{}, apk{}, pacman{}} {
			if b.available() {
				detected = b
				break
			}
		}
	})
	return detected
}
