package pkgbackend

import (
	"strconv"
	"strings"
)

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// --- dpkg ---

// parseDpkgList parses "name\tversion\tarch\tinstalledSizeKB" lines.
func parseDpkgList(out string) []Package {
	var pkgs []Package
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		cols := strings.Split(line, "\t")
		if len(cols) < 2 || cols[0] == "" {
			continue
		}
		p := Package{Name: cols[0], Version: cols[1]}
		if len(cols) >= 3 {
			p.Arch = cols[2]
		}
		if len(cols) >= 4 {
			if kb, err := strconv.ParseInt(strings.TrimSpace(cols[3]), 10, 64); err == nil {
				p.SizeKB = kb
			}
		}
		pkgs = append(pkgs, p)
	}
	return pkgs
}

// parseDpkgOwner parses "pkg1, pkg2: /the/path" (dpkg-query -S).
func parseDpkgOwner(out string) (string, bool) {
	line := firstLine(out)
	i := strings.LastIndex(line, ": ")
	if i < 0 {
		return "", false
	}
	pkg := line[:i]
	if j := strings.IndexByte(pkg, ','); j >= 0 {
		pkg = pkg[:j]
	}
	pkg = strings.TrimSpace(pkg)
	if c := strings.IndexByte(pkg, ':'); c >= 0 { // strip :arch (libc6:amd64)
		pkg = pkg[:c]
	}
	return pkg, pkg != ""
}

// --- rpm ---

// parseRpmList parses "name\tversion-release\tarch\tsizeBytes" lines.
func parseRpmList(out string) []Package {
	var pkgs []Package
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		cols := strings.Split(line, "\t")
		if len(cols) < 2 || cols[0] == "" {
			continue
		}
		p := Package{Name: cols[0], Version: cols[1]}
		if len(cols) >= 3 {
			p.Arch = cols[2]
		}
		if len(cols) >= 4 {
			if b, err := strconv.ParseInt(strings.TrimSpace(cols[3]), 10, 64); err == nil {
				p.SizeKB = b / 1024
			}
		}
		pkgs = append(pkgs, p)
	}
	return pkgs
}

// parseRpmOwner reads the output of `rpm -qf --qf '%{NAME}'` (just the name).
func parseRpmOwner(out string) (string, bool) {
	s := strings.TrimSpace(firstLine(out))
	if s == "" || strings.Contains(s, "not owned by") || strings.Contains(s, "no package") {
		return "", false
	}
	return s, true
}

// --- apk ---

// parseApkInstalledDB parses /lib/apk/db/installed: blank-line-separated records
// with P:name, V:version, A:arch, I:sizeBytes fields.
func parseApkInstalledDB(data string) []Package {
	var pkgs []Package
	var cur Package
	flush := func() {
		if cur.Name != "" {
			pkgs = append(pkgs, cur)
		}
		cur = Package{}
	}
	for _, line := range strings.Split(data, "\n") {
		if line == "" {
			flush()
			continue
		}
		if len(line) < 2 || line[1] != ':' {
			continue
		}
		val := line[2:]
		switch line[0] {
		case 'P':
			cur.Name = val
		case 'V':
			cur.Version = val
		case 'A':
			cur.Arch = val
		case 'I':
			if b, err := strconv.ParseInt(strings.TrimSpace(val), 10, 64); err == nil {
				cur.SizeKB = b / 1024
			}
		}
	}
	flush()
	return pkgs
}

// parseApkOwner reads "<path> is owned by busybox-1.36.1-r5".
func parseApkOwner(out string) (string, bool) {
	const marker = " is owned by "
	i := strings.LastIndex(out, marker)
	if i < 0 {
		return "", false
	}
	name := apkStripVersion(strings.TrimSpace(firstLine(out[i+len(marker):])))
	return name, name != ""
}

// apkStripVersion turns "py3-foo-1.2.3-r0" into "py3-foo" by dropping the
// trailing -pkgver-pkgrel (apk versions always end in -r<N>).
func apkStripVersion(nv string) string {
	parts := strings.Split(nv, "-")
	if len(parts) < 3 {
		return nv
	}
	if isApkRel(parts[len(parts)-1]) {
		return strings.Join(parts[:len(parts)-2], "-")
	}
	return nv
}

func isApkRel(s string) bool {
	if len(s) < 2 || s[0] != 'r' {
		return false
	}
	for _, c := range s[1:] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// --- pacman ---

// parsePacmanList parses "name version" lines (pacman -Q).
func parsePacmanList(out string) []Package {
	var pkgs []Package
	for _, line := range strings.Split(out, "\n") {
		f := strings.Fields(line)
		if len(f) >= 2 {
			pkgs = append(pkgs, Package{Name: f[0], Version: f[1]})
		}
	}
	return pkgs
}

// parsePacmanOwner reads `pacman -Qoq <path>` output (just the name).
func parsePacmanOwner(out string) (string, bool) {
	s := strings.TrimSpace(firstLine(out))
	return s, s != ""
}
