package twitch

import (
	"context"

	"github.com/tamnd/any-cli/kit/errs"
)

// ops.go holds the handler for every operation declared in domain.go. kit
// reflects each input struct into CLI flags, HTTP query params, and MCP tool
// arguments: kit:"arg" is a positional, kit:"flag,inherit" binds the shared
// --limit, and kit:"inject" receives the client newClient builds. The reference
// ops (id, url) take no client; they run offline.

// --- top-level reads ---

type listIn struct {
	Limit  int     `kit:"flag,inherit"`
	Client *Client `kit:"inject"`
}

func topStreams(ctx context.Context, in listIn, emit func(*Stream) error) error {
	streams, err := in.Client.TopStreams(ctx, limitOr(in.Limit, defaultPageSize))
	if err != nil {
		return mapErr(err)
	}
	return emitAll(streams, emit)
}

func directory(ctx context.Context, in listIn, emit func(*Game) error) error {
	games, err := in.Client.Directory(ctx, limitOr(in.Limit, defaultPageSize))
	if err != nil {
		return mapErr(err)
	}
	return emitAll(games, emit)
}

type videoRef struct {
	ID     string  `kit:"arg" help:"video id or twitch.tv/videos URL"`
	Client *Client `kit:"inject"`
}

func getVideo(ctx context.Context, in videoRef, emit func(*Video) error) error {
	v, err := in.Client.Video(ctx, refID(in.ID, "video"))
	if err != nil {
		return mapErr(err)
	}
	return emit(v)
}

type clipRef struct {
	Slug   string  `kit:"arg" help:"clip slug or clips.twitch.tv URL"`
	Client *Client `kit:"inject"`
}

func getClip(ctx context.Context, in clipRef, emit func(*Clip) error) error {
	cl, err := in.Client.Clip(ctx, refID(in.Slug, "clip"))
	if err != nil {
		return mapErr(err)
	}
	return emit(cl)
}

// --- search ---

type searchIn struct {
	Query  string  `kit:"arg" help:"search query"`
	Limit  int     `kit:"flag,inherit"`
	Client *Client `kit:"inject"`
}

func searchChannels(ctx context.Context, in searchIn, emit func(*Channel) error) error {
	chans, err := in.Client.SearchChannels(ctx, in.Query, limitOr(in.Limit, 20))
	if err != nil {
		return mapErr(err)
	}
	return emitAll(chans, emit)
}

func searchGames(ctx context.Context, in searchIn, emit func(*Game) error) error {
	games, err := in.Client.SearchGames(ctx, in.Query, limitOr(in.Limit, 20))
	if err != nil {
		return mapErr(err)
	}
	return emitAll(games, emit)
}

// --- channel ---

type channelRef struct {
	Login  string  `kit:"arg" help:"channel login, @handle, or URL"`
	Client *Client `kit:"inject"`
}

type channelListIn struct {
	Login  string  `kit:"arg" help:"channel login, @handle, or URL"`
	Limit  int     `kit:"flag,inherit"`
	Client *Client `kit:"inject"`
}

func getChannel(ctx context.Context, in channelRef, emit func(*Channel) error) error {
	ch, err := in.Client.Channel(ctx, refID(in.Login, "channel"))
	if err != nil {
		return mapErr(err)
	}
	return emit(ch)
}

func channelVideos(ctx context.Context, in channelListIn, emit func(*Video) error) error {
	videos, err := in.Client.ChannelVideos(ctx, refID(in.Login, "channel"), limitOr(in.Limit, 20))
	if err != nil {
		return mapErr(err)
	}
	return emitAll(videos, emit)
}

func channelClips(ctx context.Context, in channelListIn, emit func(*Clip) error) error {
	clips, err := in.Client.ChannelClips(ctx, refID(in.Login, "channel"), limitOr(in.Limit, 20))
	if err != nil {
		return mapErr(err)
	}
	return emitAll(clips, emit)
}

func channelSchedule(ctx context.Context, in channelListIn, emit func(*Segment) error) error {
	segs, err := in.Client.ChannelSchedule(ctx, refID(in.Login, "channel"), limitOr(in.Limit, 0))
	if err != nil {
		return mapErr(err)
	}
	return emitAll(segs, emit)
}

// --- game (category) ---

type gameRef struct {
	Slug   string  `kit:"arg" help:"category slug, e.g. just-chatting"`
	Client *Client `kit:"inject"`
}

type gameListIn struct {
	Slug   string  `kit:"arg" help:"category slug"`
	Limit  int     `kit:"flag,inherit"`
	Client *Client `kit:"inject"`
}

func getGame(ctx context.Context, in gameRef, emit func(*Game) error) error {
	g, err := in.Client.Game(ctx, refID(in.Slug, "game"))
	if err != nil {
		return mapErr(err)
	}
	return emit(g)
}

func gameStreams(ctx context.Context, in gameListIn, emit func(*Stream) error) error {
	streams, err := in.Client.GameStreams(ctx, refID(in.Slug, "game"), limitOr(in.Limit, defaultPageSize))
	if err != nil {
		return mapErr(err)
	}
	return emitAll(streams, emit)
}

func gameClips(ctx context.Context, in gameListIn, emit func(*Clip) error) error {
	clips, err := in.Client.GameClips(ctx, refID(in.Slug, "game"), limitOr(in.Limit, 20))
	if err != nil {
		return mapErr(err)
	}
	return emitAll(clips, emit)
}

// --- reference tools (offline) ---

type refIn struct {
	Ref string `kit:"arg" help:"any Twitch URL, path, handle, or id"`
}

func classifyRef(_ context.Context, in refIn, emit func(*Ref) error) error {
	r := Classify(in.Ref)
	if r.Kind == "unknown" {
		return errs.Usage("unrecognized twitch reference: %q", in.Ref)
	}
	return emit(&r)
}

type urlIn struct {
	Kind string `kit:"arg" help:"channel, game, video, or clip"`
	ID   string `kit:"arg" help:"the id or slug for that kind"`
}

func buildURL(_ context.Context, in urlIn, emit func(*Ref) error) error {
	u := URLFor(in.Kind, in.ID)
	if u == "" {
		return errs.Usage("twitch has no resource type %q", in.Kind)
	}
	return emit(&Ref{Input: in.Kind + "/" + in.ID, Kind: in.Kind, ID: in.ID, URL: u})
}

// emitAll streams a slice of records through emit.
func emitAll[T any](items []*T, emit func(*T) error) error {
	for _, it := range items {
		if err := emit(it); err != nil {
			return err
		}
	}
	return nil
}

// refID reduces a reference to its bare id when it classifies to the expected
// kind, so a command accepts either a bare id/slug/login or a full URL.
func refID(ref, kind string) string {
	r := Classify(ref)
	if r.Kind == kind {
		return r.ID
	}
	return ref
}
