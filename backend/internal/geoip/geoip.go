package geoip

import (
	"net"
	"strings"
	"sync"
)

// Resolver maps IP -> country. Uses optional MaxMind DB when available;
// otherwise returns Private/Local/Unknown heuristics.
type Resolver struct {
	mu   sync.RWMutex
	db   countryDB
	path string
}

type countryDB interface {
	Country(ip net.IP) (string, error)
	Close() error
}

func New(path string) *Resolver {
	r := &Resolver{path: path}
	r.reload()
	return r
}

func (r *Resolver) reload() {
	// Optional dependency: if MaxMind file missing, heuristic-only mode.
	// Avoid hard require of geoip2 at build time for environments without CGO/mmdb.
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.db != nil {
		_ = r.db.Close()
		r.db = nil
	}
	if r.path == "" {
		return
	}
	if db, err := openMMDB(r.path); err == nil {
		r.db = db
	}
}

func (r *Resolver) Country(ipStr string) string {
	ipStr = strings.TrimSpace(ipStr)
	if ipStr == "" {
		return "Unknown"
	}
	// strip port if present
	if host, _, err := net.SplitHostPort(ipStr); err == nil {
		ipStr = host
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "Unknown"
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
		return "Local"
	}
	r.mu.RLock()
	db := r.db
	r.mu.RUnlock()
	if db != nil {
		if c, err := db.Country(ip); err == nil && c != "" {
			return c
		}
	}
	return "Unknown"
}
