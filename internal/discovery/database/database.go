// Package database discovers host-installed databases (PostgreSQL, MySQL/
// MariaDB, Redis) from their config files and binaries, emitting DATABASE nodes
// linked to their listening port (which merges with the live listener from
// network.ports) and their owning package.
//
// Detection is config/binary based rather than by connecting, so it needs no
// credentials and no elevated privileges. The port parsers are pure functions
// (tested without the databases installed).
package database

import (
	"regexp"
	"time"

	"github.com/Aakhri-Pastaa/infrasight/internal/discovery"
)

// Module discovers installed databases.
type Module struct{}

// New constructs the database module.
func New() *Module { return &Module{} }

func (m *Module) Name() string            { return "database" }
func (m *Module) Description() string     { return "Databases: PostgreSQL, MySQL/MariaDB, Redis" }
func (m *Module) RequiredTools() []string { return nil }
func (m *Module) RequiresRoot() bool      { return false }
func (m *Module) RequiresNetwork() bool   { return false }
func (m *Module) Timeout() time.Duration  { return 15 * time.Second }
func (m *Module) Risk() discovery.Risk    { return discovery.RiskLow }

// engine describes how to detect one database engine.
type engine struct {
	name        string
	key         string
	binaries    []string
	defaultPort string
	configEnv   string // override for custom locations / tests
	configGlobs []string
	parsePort   func(string) string
}

func engines() []engine {
	return []engine{
		{"PostgreSQL", "postgres", []string{"postgres", "pg_ctl"}, "5432", "INFRASIGHT_PG_CONF",
			[]string{"/etc/postgresql/*/main/postgresql.conf"}, parsePortEquals},
		{"MySQL/MariaDB", "mysql", []string{"mysqld", "mariadbd"}, "3306", "INFRASIGHT_MYSQL_CONF",
			[]string{"/etc/mysql/my.cnf", "/etc/mysql/mysql.conf.d/mysqld.cnf", "/etc/my.cnf"}, parsePortEquals},
		{"Redis", "redis", []string{"redis-server"}, "6379", "INFRASIGHT_REDIS_CONF",
			[]string{"/etc/redis/redis.conf", "/etc/redis.conf"}, parsePortSpace},
	}
}

var (
	rePortEquals = regexp.MustCompile(`(?m)^[ \t]*port[ \t]*=[ \t]*(\d+)`)
	rePortSpace  = regexp.MustCompile(`(?m)^[ \t]*port[ \t]+(\d+)`)
	reVersion    = regexp.MustCompile(`\d+\.\d+(?:\.\d+)?`)
)

// parsePortEquals reads "port = 5432" style configs (postgres, mysql ini).
// Commented lines (starting with '#') are ignored.
func parsePortEquals(s string) string {
	if m := rePortEquals.FindStringSubmatch(s); m != nil {
		return m[1]
	}
	return ""
}

// parsePortSpace reads "port 6379" style configs (redis).
func parsePortSpace(s string) string {
	if m := rePortSpace.FindStringSubmatch(s); m != nil {
		return m[1]
	}
	return ""
}
