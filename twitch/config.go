package twitch

import "time"

// Host is the Twitch hostname this client builds page URLs from and the host the
// URI driver in domain.go claims.
const Host = "www.twitch.tv"

// BaseURL is the root every channel, video, and category URL is built from.
const BaseURL = "https://" + Host

// gqlEndpoint is the public GraphQL endpoint the logged-out web client talks to.
const gqlEndpoint = "https://gql.twitch.tv/gql"

// defaultClientID is the Twitch web client's public Client-Id. It is sent as a
// header on every request a logged-out browser makes, is not a secret, and is
// not tied to an account. It identifies the calling application, the way a
// User-Agent identifies the browser. Override it with --client-id.
const defaultClientID = "kimne78kx3ncx6brgo4mv6wki5h1ko"

// DefaultUserAgent is sent with every request. Twitch serves its public GraphQL
// API to a normal browser; a browser User-Agent is what keeps a logged-out
// reader looking like one. Override it with --user-agent.
const DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) " +
	"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"

// Defaults for the polite client.
const (
	// DefaultDelay is the minimum gap between requests. The GraphQL endpoint is
	// an API rather than a page, so a half-second pace reads steadily without
	// leaning on it.
	DefaultDelay    = 500 * time.Millisecond
	DefaultRetries  = 3
	DefaultTimeout  = 30 * time.Second
	DefaultCacheTTL = 24 * time.Hour

	// defaultPageSize is the items asked for per connection page. 30 is the size
	// the web directory grid uses.
	defaultPageSize = 30
)

// Config carries the knobs the client reads. It is built from the kit framework
// config in ClientFromConfig, so a --rate or --timeout on the command line and
// the same value resolved by a host both land here.
type Config struct {
	UserAgent string
	ClientID  string

	// Delay is the minimum gap between requests. Zero means no pacing.
	Delay   time.Duration
	Retries int
	Timeout time.Duration

	// BaseURL is the GraphQL endpoint. Empty uses the public endpoint; tests
	// point it at an httptest server.
	BaseURL string

	// CacheDir is where GraphQL responses are cached. Empty disables the cache,
	// as does NoCache.
	CacheDir string
	CacheTTL time.Duration
	NoCache  bool
	// Refresh fetches fresh copies and rewrites the cache, ignoring any hit.
	Refresh bool
}

// DefaultConfig returns the baseline configuration: a browser User-Agent, the
// public client id, a half-second pace, three retries, a 30s timeout, and a one
// day cache.
func DefaultConfig() Config {
	return Config{
		UserAgent: DefaultUserAgent,
		ClientID:  defaultClientID,
		Delay:     DefaultDelay,
		Retries:   DefaultRetries,
		Timeout:   DefaultTimeout,
		BaseURL:   gqlEndpoint,
		CacheTTL:  DefaultCacheTTL,
	}
}
