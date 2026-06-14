package twitch

// This file holds the exported records the commands emit. Their json tags name
// the fields a reader sees, kit:"id" marks the key the record store upserts on,
// and table:",truncate" keeps wide free text from blowing up a terminal table.
// Each record carries only fields anonymous GraphQL can actually fill: there is
// no total-view count on a channel (Twitch removed it from the public API) and
// no "do you follow this" flags (those need a logged-in viewer). There is no
// Rank column either; emit order is the rank, and a stable id is a better store
// key than a position that shifts on every refresh. The per-surface files
// (streams.go, channels.go, ...) hold the wire structs these map from.

// Stream is a live broadcast happening now.
type Stream struct {
	ID          string `json:"id" kit:"id"`
	Channel     string `json:"channel"`
	DisplayName string `json:"display_name,omitempty"`
	Title       string `json:"title,omitempty" table:",truncate"`
	Game        string `json:"game,omitempty"`
	GameSlug    string `json:"game_slug,omitempty"`
	Viewers     int    `json:"viewers"`
	Language    string `json:"language,omitempty"`
	Mature      bool   `json:"mature,omitempty"`
	StartedAt   string `json:"started_at,omitempty"`
	URL         string `json:"url"`
}

// Channel is a user/channel profile.
type Channel struct {
	ID          string `json:"id" kit:"id"`
	Login       string `json:"login"`
	DisplayName string `json:"display_name,omitempty"`
	Description string `json:"description,omitempty" table:",truncate"`
	Followers   int64  `json:"followers,omitempty"`
	Partner     bool   `json:"partner,omitempty"`
	Affiliate   bool   `json:"affiliate,omitempty"`
	Live        bool   `json:"live,omitempty"`
	Game        string `json:"game,omitempty"`
	Viewers     int    `json:"viewers,omitempty"`
	Created     string `json:"created,omitempty"`
	Avatar      string `json:"avatar,omitempty" table:",truncate"`
	URL         string `json:"url"`
}

// Video is a past broadcast or upload (a VOD).
type Video struct {
	ID          string `json:"id" kit:"id"`
	Title       string `json:"title,omitempty" table:",truncate"`
	Channel     string `json:"channel,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Game        string `json:"game,omitempty"`
	Views       int64  `json:"views,omitempty"`
	Length      int    `json:"length_seconds,omitempty"`
	PublishedAt string `json:"published_at,omitempty"`
	URL         string `json:"url"`
}

// Clip is a short user-cut highlight.
type Clip struct {
	Slug      string `json:"slug" kit:"id"`
	Title     string `json:"title,omitempty" table:",truncate"`
	Channel   string `json:"channel,omitempty"`
	Curator   string `json:"curator,omitempty"`
	Game      string `json:"game,omitempty"`
	Views     int64  `json:"views,omitempty"`
	Duration  int    `json:"duration_seconds,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	URL       string `json:"url"`
}

// Game is a category (Twitch's GraphQL type is Game; the UI calls them
// categories).
type Game struct {
	ID          string `json:"id" kit:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Viewers     int    `json:"viewers,omitempty"`
	URL         string `json:"url"`
}

// Segment is one entry in a channel's stream schedule.
type Segment struct {
	ID       string `json:"id" kit:"id"`
	Title    string `json:"title,omitempty" table:",truncate"`
	Category string `json:"category,omitempty"`
	StartAt  string `json:"start_at,omitempty"`
	EndAt    string `json:"end_at,omitempty"`
	Canceled bool   `json:"canceled,omitempty"`
}

// Ref is the result of `twitch ref id`: the canonical (kind, id) a reference
// resolves to, plus the live URL, all without touching the network.
type Ref struct {
	Input string `json:"input"`
	Kind  string `json:"kind"`
	ID    string `json:"id"`
	URL   string `json:"url"`
}
