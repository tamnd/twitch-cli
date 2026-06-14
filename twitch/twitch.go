// Package twitch is the library behind the twitch command line: an HTTP client
// for the public Twitch GraphQL API and the typed records every command emits.
//
// Twitch's web app talks to one backend, a GraphQL endpoint at
// https://gql.twitch.tv/gql. It serves a logged-out reader with nothing but a
// public client id (kimne78kx3ncx6brgo4mv6wki5h1ko) and a browser user-agent,
// and it accepts full query strings rather than only persisted hashes. So this
// client POSTs JSON and decodes JSON: no cookie jar, no csrf handshake, no HTML
// parsing. Each surface (streams, channels, games, videos, clips, search) lives
// in its own file with its query builder and record mapping; this file holds the
// shared client.
package twitch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Client talks to the Twitch GraphQL endpoint. It paces requests, retries the
// transient failures, and caches response bodies on disk keyed by the query.
type Client struct {
	HTTP      *http.Client
	Endpoint  string
	UserAgent string
	ClientID  string
	Delay     time.Duration
	Retries   int

	cache   *cache
	refresh bool

	mu   sync.Mutex
	last time.Time
}

// NewClient builds a client from cfg.
func NewClient(cfg Config) *Client {
	c := &Client{
		HTTP:      &http.Client{Timeout: cfg.Timeout},
		Endpoint:  cfg.BaseURL,
		UserAgent: cfg.UserAgent,
		ClientID:  cfg.ClientID,
		Delay:     cfg.Delay,
		Retries:   cfg.Retries,
		refresh:   cfg.Refresh,
	}
	if c.Endpoint == "" {
		c.Endpoint = gqlEndpoint
	}
	if c.UserAgent == "" {
		c.UserAgent = DefaultUserAgent
	}
	if c.ClientID == "" {
		c.ClientID = defaultClientID
	}
	// --refresh keeps the cache (so it is rewritten) but skips reads. --no-cache
	// drops it entirely.
	if !cfg.NoCache {
		c.cache = newCache(cfg.CacheDir, cfg.CacheTTL)
	}
	return c
}

// gqlError is one entry in a GraphQL envelope's errors array.
type gqlError struct {
	Message string `json:"message"`
}

// gql runs one query: paced, retried, cached. out receives the decoded data
// payload (the method passes a pointer to a struct shaped like the query).
func (c *Client) gql(ctx context.Context, query string, out any) error {
	if !c.refresh {
		if b, ok := c.cache.get(query); ok {
			return decodeData(b, out)
		}
	}
	body, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := 0; attempt <= c.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		raw, retry, derr := c.do(ctx, body)
		if derr == nil {
			c.cache.put(query, raw)
			return decodeData(raw, out)
		}
		lastErr = derr
		if !retry {
			return derr
		}
	}
	return lastErr
}

// do performs one POST and returns the raw response body. retry reports whether
// the failure is worth another attempt.
func (c *Client) do(ctx context.Context, body []byte) (raw []byte, retry bool, err error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Client-Id", c.ClientID)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	switch {
	case resp.StatusCode == http.StatusOK:
		// fall through to read and check the body
	case resp.StatusCode == http.StatusTooManyRequests:
		return nil, true, ErrRateLimited
	case resp.StatusCode >= 500:
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	case resp.StatusCode == http.StatusUnauthorized, resp.StatusCode == http.StatusForbidden:
		return nil, false, ErrBlocked
	default:
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	if err := checkGQLErrors(b); err != nil {
		// An integrity check is mapped to ErrRateLimited and is worth another
		// attempt; a missing entity or a hard error is not.
		return nil, errors.Is(err, ErrRateLimited), err
	}
	return b, false, nil
}

// checkGQLErrors reports a GraphQL-level error carried in the envelope's errors
// array. A message that names a missing entity maps to ErrNotFound; anything
// else is wrapped as-is.
func checkGQLErrors(b []byte) error {
	var env struct {
		Errors []gqlError `json:"errors"`
	}
	if err := json.Unmarshal(b, &env); err != nil {
		return nil // a real decode error surfaces in decodeData
	}
	if len(env.Errors) == 0 {
		return nil
	}
	msg := env.Errors[0].Message
	if isNotFoundMessage(msg) {
		return ErrNotFound
	}
	if isIntegrityMessage(msg) {
		// Twitch occasionally answers a logged-out request with an integrity
		// check. It is a transient throttle, so map it to ErrRateLimited: the
		// caller retries it, and if it persists the exit code points at pacing.
		return ErrRateLimited
	}
	return fmt.Errorf("gql: %s", msg)
}

func isNotFoundMessage(msg string) bool {
	m := strings.ToLower(msg)
	return strings.Contains(m, "not found") ||
		strings.Contains(m, "does not exist") ||
		strings.Contains(m, "no such")
}

func isIntegrityMessage(msg string) bool {
	m := strings.ToLower(msg)
	return strings.Contains(m, "integrity") ||
		strings.Contains(m, "failed integrity check")
}

// decodeData unmarshals the data payload of an envelope into out.
func decodeData(b []byte, out any) error {
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(b, &env); err != nil {
		return err
	}
	if len(env.Data) == 0 {
		return nil
	}
	return json.Unmarshal(env.Data, out)
}

// pace blocks until at least Delay has passed since the previous request.
func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Delay <= 0 {
		c.last = time.Now()
		return
	}
	if wait := c.Delay - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}

// ClearCache removes the on-disk cache.
func (c *Client) ClearCache() error { return c.cache.clear() }
