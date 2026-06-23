package pkgbackend

import "testing"

func TestParseDpkg(t *testing.T) {
	out := "adduser\t3.137ubuntu1\tall\t608\nlibc6\t2.39-0ubuntu8\tamd64\t13782\n"
	pkgs := parseDpkgList(out)
	if len(pkgs) != 2 {
		t.Fatalf("want 2 packages, got %d", len(pkgs))
	}
	if pkgs[0] != (Package{Name: "adduser", Version: "3.137ubuntu1", Arch: "all", SizeKB: 608}) {
		t.Errorf("pkg0 = %+v", pkgs[0])
	}
	if n, ok := parseDpkgOwner("openssh-server: /usr/sbin/sshd"); !ok || n != "openssh-server" {
		t.Errorf("owner = %q, %v", n, ok)
	}
	if n, _ := parseDpkgOwner("libc6:amd64: /lib/x86_64-linux-gnu/libc.so.6"); n != "libc6" {
		t.Errorf("arch should be stripped, got %q", n)
	}
	if _, ok := parseDpkgOwner("no colon here"); ok {
		t.Error("expected a miss")
	}
}

func TestParseRpm(t *testing.T) {
	out := "bash\t5.2.15-3.fc39\tx86_64\t1521328\nopenssl\t3.1.1-4.fc39\tx86_64\t2048000\n"
	pkgs := parseRpmList(out)
	if len(pkgs) != 2 {
		t.Fatalf("want 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Name != "bash" || pkgs[0].Version != "5.2.15-3.fc39" || pkgs[0].SizeKB != 1521328/1024 {
		t.Errorf("pkg0 = %+v (size should be bytes/1024)", pkgs[0])
	}
	if n, ok := parseRpmOwner("openssh-server"); !ok || n != "openssh-server" {
		t.Errorf("owner = %q, %v", n, ok)
	}
	if _, ok := parseRpmOwner("file /x is not owned by any package"); ok {
		t.Error("expected a miss")
	}
}

func TestParseApk(t *testing.T) {
	db := "P:busybox\nV:1.36.1-r5\nA:x86_64\nI:958464\n\nP:py3-foo\nV:1.2.3-r0\nA:x86_64\nI:51200\n"
	pkgs := parseApkInstalledDB(db)
	if len(pkgs) != 2 {
		t.Fatalf("want 2 packages, got %d: %+v", len(pkgs), pkgs)
	}
	if pkgs[0].Name != "busybox" || pkgs[0].Version != "1.36.1-r5" || pkgs[0].SizeKB != 958464/1024 {
		t.Errorf("pkg0 = %+v", pkgs[0])
	}
	if n, ok := parseApkOwner("/bin/busybox is owned by busybox-1.36.1-r5"); !ok || n != "busybox" {
		t.Errorf("owner = %q, %v", n, ok)
	}
	// package name containing a hyphen
	if n, _ := parseApkOwner("/usr/lib/x is owned by py3-foo-1.2.3-r0"); n != "py3-foo" {
		t.Errorf("hyphenated name = %q, want py3-foo", n)
	}
	if _, ok := parseApkOwner("nope"); ok {
		t.Error("expected a miss")
	}
}

func TestParsePacman(t *testing.T) {
	pkgs := parsePacmanList("bash 5.2.015-1\nopenssl 3.1.1-1\n")
	if len(pkgs) != 2 {
		t.Fatalf("want 2 packages, got %d", len(pkgs))
	}
	if pkgs[1].Name != "openssl" || pkgs[1].Version != "3.1.1-1" {
		t.Errorf("pkg1 = %+v", pkgs[1])
	}
	if n, ok := parsePacmanOwner("bash\n"); !ok || n != "bash" {
		t.Errorf("owner = %q, %v", n, ok)
	}
}

func TestNodeID(t *testing.T) {
	if got := NodeID("apt", "nginx"); got != "package:apt:nginx" {
		t.Errorf("NodeID = %q", got)
	}
	if got := NodeID("rpm", "openssl"); got != "package:rpm:openssl" {
		t.Errorf("NodeID = %q", got)
	}
}
