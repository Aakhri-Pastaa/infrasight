package database

import "testing"

func TestParsePortEquals(t *testing.T) {
	pg := "# PostgreSQL config\n#port = 5432\nport = 5433\nmax_connections = 100\n"
	if got := parsePortEquals(pg); got != "5433" {
		t.Errorf("postgres port = %q, want 5433 (commented line ignored)", got)
	}
	if got := parsePortEquals("[mysqld]\nport=3307\n"); got != "3307" {
		t.Errorf("mysql port = %q, want 3307 (no spaces)", got)
	}
	if got := parsePortEquals("max_connections = 100\n"); got != "" {
		t.Errorf("absent port = %q, want empty", got)
	}
}

func TestParsePortSpace(t *testing.T) {
	redis := "bind 127.0.0.1\nport 6380\nsave 900 1\n"
	if got := parsePortSpace(redis); got != "6380" {
		t.Errorf("redis port = %q, want 6380", got)
	}
	if got := parsePortSpace("# port 9999\nport 6379\n"); got != "6379" {
		t.Errorf("redis port = %q, want 6379 (commented line ignored)", got)
	}
	if got := parsePortSpace("bind 127.0.0.1\n"); got != "" {
		t.Errorf("absent port = %q, want empty", got)
	}
}
