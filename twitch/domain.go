package twitch

import (
	"context"
	"errors"
	"time"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go exposes twitch as a kit Domain: a driver that a multi-domain host
// (ant) enables with a single blank import,
//
//	import _ "github.com/tamnd/twitch-cli/twitch"
//
// exactly as a database/sql program enables a driver with `import _
// "github.com/lib/pq"`. The init below registers it; the host then dereferences
// twitch:// URIs by routing to the operations Register installs. The same Domain
// also builds the standalone twitch binary (see cli.NewApp), so the binary and a
// host share one source of truth.
func init() { kit.Register(Domain{}) }

// Domain is the twitch driver. It carries no state; the per-run client is built
// by the factory Register hands kit.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against, and
// the identity reused for the binary's help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme:   "twitch",
		Aliases:  []string{"ttv"},
		Hosts:    []string{Host, "twitch.tv", "clips.twitch.tv", "m.twitch.tv"},
		Identity: Identity(),
	}
}

// Identity is the fixed description of the twitch CLI, shared by the domain and
// the standalone composition root so help and version read the same everywhere.
func Identity() kit.Identity {
	return kit.Identity{
		Binary: "twitch",
		Short:  "Read public Twitch streams, channels, clips, and categories into structured records",
		Long: `twitch reads public Twitch data the way a logged-out browser does:
the top live streams, the category directory, a channel and its
videos, clips, and schedule, a single video or clip, and search over
channels and categories. It talks to Twitch's public GraphQL API with
the web client's public id, so there is no API key, no login, and
nothing to run alongside it. It returns records as a table, JSON,
JSONL, CSV, TSV, or URLs, and serves the same operations over HTTP and
MCP.

twitch is an independent tool and is not affiliated with Twitch.`,
		Site: BaseURL,
		Repo: "https://github.com/tamnd/twitch-cli",
	}
}

// Register installs the client factory and every operation onto app. A resolver
// op (Single) names its own record type and answers `ant get`; a List op
// enumerates a parent resource's members and answers `ant ls`.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)
	app.CommandGroup("read", "Read public Twitch data")
	app.CommandGroup("channel", "Read a channel, its videos, clips, and schedule")
	app.CommandGroup("game", "Read a category, its streams, and its clips")
	app.CommandGroup("search", "Search channels and categories")
	app.CommandGroup("ref", "Resolve references to ids and URLs (offline)")

	// Top-level reads.
	kit.Handle(app, kit.OpMeta{
		Name: "streams", Group: "read", List: true,
		Summary: "Top live streams right now",
		URIType: "stream",
	}, topStreams)

	kit.Handle(app, kit.OpMeta{
		Name: "games", Group: "read", List: true,
		Summary: "The category directory, by viewers",
		URIType: "game",
	}, directory)

	kit.Handle(app, kit.OpMeta{
		Name: "video", Group: "read", Single: true,
		Summary: "Show one video by id",
		URIType: "video", Resolver: true,
		Args: []kit.Arg{{Name: "id", Help: "video id or twitch.tv/videos URL"}},
	}, getVideo)

	kit.Handle(app, kit.OpMeta{
		Name: "clip", Group: "read", Single: true,
		Summary: "Show one clip by slug",
		URIType: "clip", Resolver: true,
		Args: []kit.Arg{{Name: "slug", Help: "clip slug or clips.twitch.tv URL"}},
	}, getClip)

	// Search.
	kit.Handle(app, kit.OpMeta{
		Name: "channels", Parent: "search",
		Summary: "Search channels",
		URIType: "channel",
		Args:    []kit.Arg{{Name: "query", Help: "search query"}},
	}, searchChannels)

	kit.Handle(app, kit.OpMeta{
		Name: "games", Parent: "search",
		Summary: "Search categories",
		URIType: "game",
		Args:    []kit.Arg{{Name: "query", Help: "search query"}},
	}, searchGames)

	// Channel: metadata, videos, clips, schedule.
	kit.Handle(app, kit.OpMeta{
		Name: "show", Parent: "channel", Single: true,
		Summary: "Show a channel's metadata",
		URIType: "channel", Resolver: true,
		Args: []kit.Arg{{Name: "login", Help: "channel login, @handle, or URL"}},
	}, getChannel)

	kit.Handle(app, kit.OpMeta{
		Name: "videos", Parent: "channel", List: true,
		Summary: "List a channel's videos",
		URIType: "video",
		Args:    []kit.Arg{{Name: "login", Help: "channel login, @handle, or URL"}},
	}, channelVideos)

	kit.Handle(app, kit.OpMeta{
		Name: "clips", Parent: "channel", List: true,
		Summary: "List a channel's clips",
		URIType: "clip",
		Args:    []kit.Arg{{Name: "login", Help: "channel login, @handle, or URL"}},
	}, channelClips)

	kit.Handle(app, kit.OpMeta{
		Name: "schedule", Parent: "channel", List: true,
		Summary: "List a channel's upcoming schedule",
		URIType: "segment",
		Args:    []kit.Arg{{Name: "login", Help: "channel login, @handle, or URL"}},
	}, channelSchedule)

	// Game (category): metadata, streams, clips.
	kit.Handle(app, kit.OpMeta{
		Name: "show", Parent: "game", Single: true,
		Summary: "Show a category's metadata",
		URIType: "game", Resolver: true,
		Args: []kit.Arg{{Name: "slug", Help: "category slug, e.g. just-chatting"}},
	}, getGame)

	kit.Handle(app, kit.OpMeta{
		Name: "streams", Parent: "game", List: true,
		Summary: "List live streams in a category",
		URIType: "stream",
		Args:    []kit.Arg{{Name: "slug", Help: "category slug"}},
	}, gameStreams)

	kit.Handle(app, kit.OpMeta{
		Name: "clips", Parent: "game", List: true,
		Summary: "List top clips in a category",
		URIType: "clip",
		Args:    []kit.Arg{{Name: "slug", Help: "category slug"}},
	}, gameClips)

	// Reference tools (offline).
	kit.Handle(app, kit.OpMeta{
		Name: "id", Parent: "ref", Single: true,
		Summary: "Classify a reference into its (kind, id)",
		Args:    []kit.Arg{{Name: "ref", Help: "any Twitch URL, path, handle, or id"}},
	}, classifyRef)

	kit.Handle(app, kit.OpMeta{
		Name: "url", Parent: "ref", Single: true,
		Summary: "Build the canonical URL for a (kind, id)",
		Args: []kit.Arg{
			{Name: "kind", Help: "channel, game, video, or clip"},
			{Name: "id", Help: "the id or slug for that kind"},
		},
	}, buildURL)
}

// newClient builds the client from the host-resolved config, so a host and the
// standalone binary pace and identify themselves the same way.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	return ClientFromConfig(cfg), nil
}

// ClientFromConfig maps the framework config onto a twitch.Config and returns a
// client.
func ClientFromConfig(cfg kit.Config) *Client {
	tc := DefaultConfig()
	if cfg.Rate > 0 {
		tc.Delay = cfg.Rate
	}
	if cfg.Retries >= 0 {
		tc.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		tc.Timeout = cfg.Timeout
	}
	if ua := cfg.Extra["user-agent"]; ua != "" {
		tc.UserAgent = ua
	} else if cfg.UserAgent != "" {
		tc.UserAgent = cfg.UserAgent
	}
	if id := cfg.Extra["client-id"]; id != "" {
		tc.ClientID = id
	}
	tc.CacheDir = cfg.CacheDir
	tc.NoCache = cfg.NoCache
	if ttl := cfg.Extra["cache-ttl"]; ttl != "" {
		if d, err := time.ParseDuration(ttl); err == nil {
			tc.CacheTTL = d
		}
	}
	tc.Refresh = cfg.Extra["refresh"] == "true"
	return NewClient(tc)
}

// Defaults seeds the framework baseline with twitch's own values, so an unset
// --rate or --timeout uses the twitch default rather than the generic kit one.
// It is passed to kit.New via kit.WithDefaults.
func Defaults(c *kit.Config) {
	def := DefaultConfig()
	c.Rate = def.Delay
	c.Retries = def.Retries
	c.Timeout = def.Timeout
	c.UserAgent = def.UserAgent
}

// Classify turns any accepted input into the canonical (type, id), so `ant
// resolve` and `ant url` touch no network.
func (Domain) Classify(input string) (uriType, id string, err error) {
	r := Classify(input)
	if r.Kind == "unknown" {
		return "", "", errs.Usage("unrecognized twitch reference: %q", input)
	}
	return r.Kind, r.ID, nil
}

// Locate is the inverse: the live https URL for a (type, id).
func (Domain) Locate(uriType, id string) (string, error) {
	u := URLFor(uriType, id)
	if u == "" {
		return "", errs.Usage("twitch has no resource type %q", uriType)
	}
	return u, nil
}

// mapErr translates a library error into a kit error so the exit code matches
// the rest of the fleet: a missing entity reads as "not found" (exit 6), a
// throttle as "rate limited" (exit 5), and a wall as "need auth" (exit 4).
func mapErr(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, ErrNotFound):
		return errs.NotFound("%s", err.Error())
	case errors.Is(err, ErrRateLimited):
		return errs.RateLimited("%s", err.Error())
	case errors.Is(err, ErrBlocked):
		return errs.NeedAuth("%s", err.Error())
	default:
		return err
	}
}

// limitOr returns the operator's --limit when set, else the command's own
// default fetch count.
func limitOr(limit, def int) int {
	if limit > 0 {
		return limit
	}
	return def
}
