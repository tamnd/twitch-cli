package twitch

// Stream is a live Twitch stream.
type Stream struct {
	Rank      int    `json:"rank"       csv:"rank"       tsv:"rank"`
	ID        string `json:"id"         csv:"id"         tsv:"id"`
	Title     string `json:"title"      csv:"title"      tsv:"title"`
	Game      string `json:"game"       csv:"game"       tsv:"game"`
	Channel   string `json:"channel"    csv:"channel"    tsv:"channel"`
	Viewers   int    `json:"viewers"    csv:"viewers"    tsv:"viewers"`
	StartedAt string `json:"started_at" csv:"started_at" tsv:"started_at"`
	URL       string `json:"url"        csv:"url"        tsv:"url"`
}

// Category is a Twitch game/category.
type Category struct {
	Rank    int    `json:"rank"    csv:"rank"    tsv:"rank"`
	ID      string `json:"id"      csv:"id"      tsv:"id"`
	Name    string `json:"name"    csv:"name"    tsv:"name"`
	Slug    string `json:"slug"    csv:"slug"    tsv:"slug"`
	Viewers int    `json:"viewers" csv:"viewers" tsv:"viewers"`
	URL     string `json:"url"     csv:"url"     tsv:"url"`
}

// Channel is a Twitch channel.
type Channel struct {
	Rank        int    `json:"rank"         csv:"rank"         tsv:"rank"`
	ID          string `json:"id"           csv:"id"           tsv:"id"`
	Login       string `json:"login"        csv:"login"        tsv:"login"`
	DisplayName string `json:"display_name" csv:"display_name" tsv:"display_name"`
	URL         string `json:"url"          csv:"url"          tsv:"url"`
}
