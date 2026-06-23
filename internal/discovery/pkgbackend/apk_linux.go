//go:build linux

package pkgbackend

import (
	"context"
	"os"
	"time"

	"github.com/Aakhri-Pastaa/infrasight/pkg/shell"
)

type apk struct{}

const apkDB = "/lib/apk/db/installed"

func (apk) Key() string { return "apk" }

func (apk) available() bool {
	if _, err := os.Stat(apkDB); err == nil {
		return true
	}
	return shell.Available("apk")
}

func (apk) List(_ context.Context, _ time.Duration) ([]Package, error) {
	// The installed db is structured (P:/V:/A:/I: records) — more robust than
	// parsing the "name-version" forms from the apk CLI.
	data, err := os.ReadFile(apkDB)
	if err != nil {
		return nil, err
	}
	return parseApkInstalledDB(string(data)), nil
}

func (apk) Owner(ctx context.Context, path string) (string, bool) {
	out, err := shell.Run(ctx, 5*time.Second, "apk", "info", "--who-owns", path)
	if err != nil || out == "" {
		return "", false
	}
	return parseApkOwner(out)
}
