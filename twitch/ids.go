package twitch

import "strings"

// ids.go resolves a reference to a (kind, id) pair and builds canonical URLs,
// all offline. It backs `twitch ref id` and `twitch ref url`, and the Resolver
// the ant host calls to turn a twitch:// URI into the right command.

// Classify reads a reference (a URL, a path, an @handle, or a bare id) and
// reports what it points at. Kind is one of channel, game, video, clip, or
// unknown.
func Classify(ref string) Ref {
	in := strings.TrimSpace(ref)
	r := Ref{Input: in, Kind: "unknown"}
	if in == "" {
		return r
	}

	// A bare @handle is always a channel.
	if strings.HasPrefix(in, "@") {
		r.Kind, r.ID = "channel", strings.TrimPrefix(in, "@")
		r.URL = URLFor(r.Kind, r.ID)
		return r
	}

	host := refHost(in)
	segs := splitSegs(refPath(in))

	switch {
	case len(segs) == 0:
		return r
	case host == "clips.twitch.tv":
		// clips.twitch.tv/<slug>
		r.Kind, r.ID = "clip", segs[0]
	case segs[0] == "videos" && len(segs) >= 2:
		r.Kind, r.ID = "video", segs[1]
	case (segs[0] == "directory") && len(segs) >= 3 && (segs[1] == "game" || segs[1] == "category"):
		r.Kind, r.ID = "game", decodeSlug(strings.Join(segs[2:], "/"))
	case len(segs) >= 3 && segs[1] == "clip":
		// twitch.tv/<login>/clip/<slug>
		r.Kind, r.ID = "clip", segs[2]
	case len(segs) == 1:
		if isDigits(segs[0]) {
			r.Kind, r.ID = "video", segs[0]
		} else if reservedSegment(segs[0]) {
			return r
		} else {
			r.Kind, r.ID = "channel", segs[0]
		}
	default:
		if reservedSegment(segs[0]) {
			return r
		}
		// twitch.tv/<login>/... (videos, clips, about) is still the channel.
		r.Kind, r.ID = "channel", segs[0]
	}
	r.URL = URLFor(r.Kind, r.ID)
	return r
}

// URLFor builds the canonical Twitch URL for a (kind, id) pair.
func URLFor(kind, id string) string {
	switch kind {
	case "channel":
		return BaseURL + "/" + id
	case "video":
		return BaseURL + "/videos/" + id
	case "clip":
		return "https://clips.twitch.tv/" + id
	case "game":
		return BaseURL + "/directory/game/" + id
	default:
		return ""
	}
}

// refHost returns the hostname of a full URL, or "" for a bare handle or path.
func refHost(ref string) string {
	i := strings.Index(ref, "://")
	if i < 0 {
		return ""
	}
	rest := ref[i+3:]
	if s := strings.IndexByte(rest, '/'); s >= 0 {
		rest = rest[:s]
	}
	return strings.ToLower(rest)
}

// refPath reduces a reference to a site path: a full URL loses scheme and host,
// a bare handle or path is returned trimmed.
func refPath(ref string) string {
	ref = strings.TrimPrefix(ref, "@")
	if i := strings.Index(ref, "://"); i >= 0 {
		rest := ref[i+3:]
		if s := strings.IndexByte(rest, '/'); s >= 0 {
			return rest[s:]
		}
		return "/"
	}
	if !strings.HasPrefix(ref, "/") {
		ref = "/" + ref
	}
	return ref
}

func splitSegs(path string) []string {
	var out []string
	for _, s := range strings.Split(path, "/") {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func decodeSlug(s string) string {
	return strings.ReplaceAll(s, "%20", " ")
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// reservedSegment reports whether a first path segment is a Twitch site route
// rather than a channel login.
func reservedSegment(seg string) bool {
	switch seg {
	case "directory", "videos", "settings", "downloads", "p", "search",
		"subscriptions", "wallet", "drops", "turbo", "prime", "store":
		return true
	}
	return false
}
