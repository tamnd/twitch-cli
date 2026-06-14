package twitch

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"time"
)

// cache is a small on-disk store of GraphQL responses, keyed by the query
// string. Every request hits the same endpoint with a different body, so the
// query is what distinguishes one response from another. A hit is served when
// the file is younger than ttl.
type cache struct {
	dir string
	ttl time.Duration
}

// newCache returns a cache rooted at dir, or nil when caching is disabled (an
// empty dir or a non-positive ttl). A nil cache is safe to call: get misses and
// put is a no-op.
func newCache(dir string, ttl time.Duration) *cache {
	if dir == "" || ttl <= 0 {
		return nil
	}
	return &cache{dir: dir, ttl: ttl}
}

// cacheKey hashes a query string into a stable filename.
func cacheKey(query string) string {
	sum := sha256.Sum256([]byte(query))
	return hex.EncodeToString(sum[:])
}

func (c *cache) path(query string) string {
	return filepath.Join(c.dir, cacheKey(query)+".cache")
}

// get returns the cached body for a query when it exists and is fresh.
func (c *cache) get(query string) ([]byte, bool) {
	if c == nil {
		return nil, false
	}
	p := c.path(query)
	fi, err := os.Stat(p)
	if err != nil {
		return nil, false
	}
	if time.Since(fi.ModTime()) > c.ttl {
		return nil, false
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, false
	}
	return b, true
}

// put writes body to the cache for a query, creating the directory on first
// use. A write failure is not fatal: the body is already in hand.
func (c *cache) put(query string, body []byte) {
	if c == nil {
		return
	}
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return
	}
	_ = os.WriteFile(c.path(query), body, 0o644)
}

// clear removes the whole cache directory.
func (c *cache) clear() error {
	if c == nil {
		return nil
	}
	return os.RemoveAll(c.dir)
}
