package twitch_test

import (
	"errors"
	"testing"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
	"github.com/tamnd/twitch-cli/twitch"
)

func TestClassify(t *testing.T) {
	cases := []struct {
		in   string
		kind string
		id   string
	}{
		{"shroud", "channel", "shroud"},
		{"@shroud", "channel", "shroud"},
		{"https://www.twitch.tv/shroud", "channel", "shroud"},
		{"https://twitch.tv/shroud/about", "channel", "shroud"},
		{"https://www.twitch.tv/videos/123456789", "video", "123456789"},
		{"123456789", "video", "123456789"},
		{"https://clips.twitch.tv/AbCdEfGh", "clip", "AbCdEfGh"},
		{"https://www.twitch.tv/shroud/clip/AbCdEfGh", "clip", "AbCdEfGh"},
		{"https://www.twitch.tv/directory/game/Just%20Chatting", "game", "Just Chatting"},
		{"https://www.twitch.tv/directory/category/valorant", "game", "valorant"},
		{"https://www.twitch.tv/settings", "unknown", ""},
		{"", "unknown", ""},
	}
	for _, c := range cases {
		got := twitch.Classify(c.in)
		if got.Kind != c.kind || got.ID != c.id {
			t.Errorf("Classify(%q) = (%s, %s), want (%s, %s)", c.in, got.Kind, got.ID, c.kind, c.id)
		}
	}
}

func TestURLForRoundTrip(t *testing.T) {
	urls := []string{
		"https://www.twitch.tv/shroud",
		"https://www.twitch.tv/videos/123456789",
		"https://clips.twitch.tv/AbCdEfGh",
		"https://www.twitch.tv/directory/game/valorant",
	}
	for _, u := range urls {
		r := twitch.Classify(u)
		if got := twitch.URLFor(r.Kind, r.ID); got != u {
			t.Errorf("round trip %q -> (%s,%s) -> %q", u, r.Kind, r.ID, got)
		}
	}
}

func TestURLForUnknown(t *testing.T) {
	if u := twitch.URLFor("nonsense", "x"); u != "" {
		t.Errorf("URLFor(nonsense) = %q, want empty", u)
	}
}

func TestDomainClassifyError(t *testing.T) {
	_, _, err := twitch.Domain{}.Classify("https://www.twitch.tv/settings")
	if errs.KindOf(err) != errs.KindUsage {
		t.Errorf("unknown ref should be a usage error, got %v", err)
	}
}

func TestDomainLocate(t *testing.T) {
	u, err := twitch.Domain{}.Locate("channel", "shroud")
	if err != nil || u != "https://www.twitch.tv/shroud" {
		t.Errorf("Locate(channel, shroud) = (%q, %v)", u, err)
	}
	if _, err := (twitch.Domain{}).Locate("bogus", "x"); err == nil {
		t.Error("Locate(bogus) should error")
	}
}

func TestInfo(t *testing.T) {
	info := twitch.Domain{}.Info()
	if info.Scheme != "twitch" {
		t.Errorf("scheme = %q", info.Scheme)
	}
	if info.Identity.Binary != "twitch" {
		t.Errorf("binary = %q", info.Identity.Binary)
	}
}

func TestRegisterInstallsOps(t *testing.T) {
	app := kit.New(twitch.Identity(), kit.WithDefaults(twitch.Defaults))
	twitch.Domain{}.Register(app)

	want := []string{
		"streams", "games", "video", "clip",
		"search channels", "search games",
		"channel show", "channel videos", "channel clips", "channel schedule",
		"game show", "game streams", "game clips",
		"ref id", "ref url",
	}
	have := map[string]bool{}
	for _, op := range app.Ops() {
		m := op.Meta()
		key := m.Name
		if m.Parent != "" {
			key = m.Parent + " " + m.Name
		}
		have[key] = true
	}
	for _, k := range want {
		if !have[k] {
			t.Errorf("operation %q was not registered", k)
		}
	}
}

func TestClientFromConfig(t *testing.T) {
	cfg := kit.Config{
		Extra: map[string]string{"client-id": "custom", "user-agent": "ua/1"},
	}
	c := twitch.ClientFromConfig(cfg)
	if c.ClientID != "custom" || c.UserAgent != "ua/1" {
		t.Errorf("config not threaded: id=%q ua=%q", c.ClientID, c.UserAgent)
	}
}

// sanity: the sentinel errors are distinct so mapErr can switch on them.
func TestSentinelsDistinct(t *testing.T) {
	if errors.Is(twitch.ErrNotFound, twitch.ErrBlocked) ||
		errors.Is(twitch.ErrRateLimited, twitch.ErrNotFound) {
		t.Error("sentinels should be distinct")
	}
}
