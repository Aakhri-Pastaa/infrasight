// Package store persists scans under ~/.infrasight/scans so they can be used as
// baselines for drift detection (infrasight diff).
package store

import (
	"os"
	"path/filepath"
	"strings"
)

func scansDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".infrasight", "scans"), nil
}

// Save writes scan JSON under a name and returns the file path.
func Save(name string, data []byte) (string, error) {
	dir, err := scansDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, baseName(name)+".json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// Resolve maps a diff argument to a file path: if it is an existing file it is
// used as-is, otherwise it is treated as a saved-scan name.
func Resolve(arg string) string {
	if fi, err := os.Stat(arg); err == nil && !fi.IsDir() {
		return arg
	}
	dir, err := scansDir()
	if err != nil {
		return arg
	}
	return filepath.Join(dir, baseName(arg)+".json")
}

func baseName(name string) string {
	return strings.TrimSuffix(filepath.Base(name), ".json")
}
