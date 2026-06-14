package twitch

import "errors"

// The library reports its outcomes as a few sentinel errors. domain.go's mapErr
// translates each into the kit error kind that carries the matching exit code,
// so the standalone binary and a host agree on what a wall, a throttle, and a
// miss mean.
var (
	// ErrNotFound is a missing entity: an unknown login, a deleted video or
	// clip, or a bad category slug. Twitch reports these as a null top-level
	// field with no errors array. Exit code 6.
	ErrNotFound = errors.New("not found")

	// ErrRateLimited is a sustained HTTP 429 after the client's own retries.
	// Slow down with --rate. Exit code 5.
	ErrRateLimited = errors.New("rate limited")

	// ErrBlocked is a 401 or 403, or an integrity demand: Twitch declined to
	// serve the surface to an anonymous reader. The public read surface this CLI
	// uses does not normally hit this; if it does, the boundary is real and the
	// CLI reports it rather than trying to satisfy the demand. Exit code 4.
	ErrBlocked = errors.New("blocked")
)
