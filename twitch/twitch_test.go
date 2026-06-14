package twitch_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tamnd/twitch-cli/twitch"
)

// gqlData wraps a data payload in the GraphQL envelope the client expects.
func gqlData(data any) string {
	b, _ := json.Marshal(map[string]any{"data": data})
	return string(b)
}

// fakeServer returns a server that replies with bodies in order, one per
// request, so a paginated test can stage page one then page two.
func fakeServer(t *testing.T, bodies ...string) *httptest.Server {
	t.Helper()
	i := 0
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		body := bodies[len(bodies)-1]
		if i < len(bodies) {
			body = bodies[i]
		}
		i++
		_, _ = w.Write([]byte(body))
	}))
}

func newClient(ts *httptest.Server) *twitch.Client {
	cfg := twitch.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Delay = 0
	cfg.Retries = 0
	return twitch.NewClient(cfg)
}

func TestTopStreams(t *testing.T) {
	ts := fakeServer(t, gqlData(map[string]any{
		"streams": map[string]any{
			"edges": []map[string]any{
				{"cursor": "", "node": map[string]any{
					"id": "42", "title": "ranked grind", "viewersCount": 1200,
					"createdAt": "2026-06-01T00:00:00Z", "language": "en",
					"game":        map[string]any{"name": "Dota 2", "slug": "dota-2"},
					"broadcaster": map[string]any{"login": "arteezy", "displayName": "Arteezy"},
				}},
			},
			"pageInfo": map[string]any{"hasNextPage": false},
		},
	}))
	defer ts.Close()

	got, err := newClient(ts).TopStreams(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 stream, got %d", len(got))
	}
	s := got[0]
	if s.Channel != "arteezy" || s.Viewers != 1200 || s.Game != "Dota 2" {
		t.Errorf("unexpected stream: %+v", s)
	}
	if s.URL != "https://www.twitch.tv/arteezy" {
		t.Errorf("url = %q", s.URL)
	}
}

func TestTopStreamsPaginates(t *testing.T) {
	page1 := gqlData(map[string]any{"streams": map[string]any{
		"edges": []map[string]any{
			{"cursor": "c1", "node": map[string]any{"id": "1", "broadcaster": map[string]any{"login": "a"}}},
			{"cursor": "c2", "node": map[string]any{"id": "2", "broadcaster": map[string]any{"login": "b"}}},
		},
		"pageInfo": map[string]any{"hasNextPage": true},
	}})
	page2 := gqlData(map[string]any{"streams": map[string]any{
		"edges": []map[string]any{
			{"cursor": "c3", "node": map[string]any{"id": "3", "broadcaster": map[string]any{"login": "c"}}},
		},
		"pageInfo": map[string]any{"hasNextPage": false},
	}})
	ts := fakeServer(t, page1, page2)
	defer ts.Close()

	got, err := newClient(ts).TopStreams(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("want 3 streams across two pages, got %d", len(got))
	}
	if got[2].Channel != "c" {
		t.Errorf("third channel = %q, want c", got[2].Channel)
	}
}

func TestTopStreamsLimitCaps(t *testing.T) {
	page := gqlData(map[string]any{"streams": map[string]any{
		"edges": []map[string]any{
			{"cursor": "c1", "node": map[string]any{"id": "1", "broadcaster": map[string]any{"login": "a"}}},
			{"cursor": "c2", "node": map[string]any{"id": "2", "broadcaster": map[string]any{"login": "b"}}},
		},
		"pageInfo": map[string]any{"hasNextPage": true},
	}})
	ts := fakeServer(t, page)
	defer ts.Close()

	got, err := newClient(ts).TopStreams(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("limit 1 should cap to 1, got %d", len(got))
	}
}

func TestChannel(t *testing.T) {
	ts := fakeServer(t, gqlData(map[string]any{
		"user": map[string]any{
			"id": "7", "login": "shroud", "displayName": "shroud",
			"description": "fps", "createdAt": "2012-01-01T00:00:00Z",
			"followers": map[string]any{"totalCount": 10000000},
			"roles":     map[string]any{"isPartner": true, "isAffiliate": false},
			"stream":    map[string]any{"id": "s1", "viewersCount": 30000, "game": map[string]any{"name": "VALORANT"}},
		},
	}))
	defer ts.Close()

	ch, err := newClient(ts).Channel(context.Background(), "shroud")
	if err != nil {
		t.Fatal(err)
	}
	if ch.Followers != 10000000 || !ch.Partner || !ch.Live || ch.Game != "VALORANT" || ch.Viewers != 30000 {
		t.Errorf("unexpected channel: %+v", ch)
	}
}

func TestChannelNotFound(t *testing.T) {
	ts := fakeServer(t, gqlData(map[string]any{"user": nil}))
	defer ts.Close()

	_, err := newClient(ts).Channel(context.Background(), "nope")
	if !errors.Is(err, twitch.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestVideo(t *testing.T) {
	ts := fakeServer(t, gqlData(map[string]any{
		"video": map[string]any{
			"id": "999", "title": "vod", "viewCount": 5000, "lengthSeconds": 3600,
			"publishedAt": "2026-05-01T00:00:00Z",
			"game":        map[string]any{"name": "Chess"},
			"owner":       map[string]any{"login": "gmhikaru", "displayName": "GMHikaru"},
		},
	}))
	defer ts.Close()

	v, err := newClient(ts).Video(context.Background(), "999")
	if err != nil {
		t.Fatal(err)
	}
	if v.Length != 3600 || v.Channel != "gmhikaru" || v.Views != 5000 {
		t.Errorf("unexpected video: %+v", v)
	}
	if v.URL != "https://www.twitch.tv/videos/999" {
		t.Errorf("url = %q", v.URL)
	}
}

func TestClip(t *testing.T) {
	ts := fakeServer(t, gqlData(map[string]any{
		"clip": map[string]any{
			"slug": "FunnyClip", "title": "lol", "viewCount": 200, "durationSeconds": 30,
			"createdAt":   "2026-06-01T00:00:00Z",
			"game":        map[string]any{"name": "Just Chatting"},
			"broadcaster": map[string]any{"login": "xqc"},
			"curator":     map[string]any{"login": "fan"},
		},
	}))
	defer ts.Close()

	cl, err := newClient(ts).Clip(context.Background(), "FunnyClip")
	if err != nil {
		t.Fatal(err)
	}
	if cl.Channel != "xqc" || cl.Curator != "fan" || cl.Duration != 30 {
		t.Errorf("unexpected clip: %+v", cl)
	}
	if cl.URL != "https://clips.twitch.tv/FunnyClip" {
		t.Errorf("url = %q", cl.URL)
	}
}

func TestDirectory(t *testing.T) {
	ts := fakeServer(t, gqlData(map[string]any{
		"games": map[string]any{
			"edges": []map[string]any{
				{"cursor": "c1", "node": map[string]any{"id": "1", "name": "Just Chatting", "slug": "just-chatting", "viewersCount": 400000}},
				{"cursor": "c2", "node": map[string]any{"id": "2", "name": "Fortnite", "slug": "fortnite", "viewersCount": 90000}},
			},
			"pageInfo": map[string]any{"hasNextPage": false},
		},
	}))
	defer ts.Close()

	got, err := newClient(ts).Directory(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Name != "Just Chatting" {
		t.Fatalf("unexpected directory: %+v", got)
	}
}

func TestSearchChannels(t *testing.T) {
	ts := fakeServer(t, gqlData(map[string]any{
		"searchFor": map[string]any{
			"channels": map[string]any{
				"items": []map[string]any{
					{"id": "1", "login": "ninja", "displayName": "Ninja", "followers": map[string]any{"totalCount": 18000000}},
				},
			},
		},
	}))
	defer ts.Close()

	got, err := newClient(ts).SearchChannels(context.Background(), "ninja", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Login != "ninja" || got[0].Followers != 18000000 {
		t.Fatalf("unexpected search result: %+v", got)
	}
}

func TestGameNotFound(t *testing.T) {
	ts := fakeServer(t, gqlData(map[string]any{"game": nil}))
	defer ts.Close()

	_, err := newClient(ts).Game(context.Background(), "no-such-game")
	if !errors.Is(err, twitch.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestRateLimited(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()

	_, err := newClient(ts).Directory(context.Background(), 5)
	if !errors.Is(err, twitch.ErrRateLimited) {
		t.Fatalf("want ErrRateLimited, got %v", err)
	}
}

func TestBlocked(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	_, err := newClient(ts).Directory(context.Background(), 5)
	if !errors.Is(err, twitch.ErrBlocked) {
		t.Fatalf("want ErrBlocked, got %v", err)
	}
}

func TestIntegrityCheckRetries(t *testing.T) {
	// The first reply is a transient integrity check; the retry succeeds.
	integrity := `{"errors":[{"message":"failed integrity check"}]}`
	ok := gqlData(map[string]any{"games": map[string]any{
		"edges":    []map[string]any{{"cursor": "c1", "node": map[string]any{"id": "1", "name": "Just Chatting", "slug": "just-chatting"}}},
		"pageInfo": map[string]any{"hasNextPage": false},
	}})
	ts := fakeServer(t, integrity, ok)
	defer ts.Close()

	c := newClient(ts)
	c.Retries = 1
	got, err := c.Directory(context.Background(), 5)
	if err != nil {
		t.Fatalf("integrity check should be retried, got %v", err)
	}
	if len(got) != 1 || got[0].Slug != "just-chatting" {
		t.Fatalf("unexpected directory after retry: %+v", got)
	}
}

func TestIntegrityCheckPersists(t *testing.T) {
	integrity := `{"errors":[{"message":"failed integrity check"}]}`
	ts := fakeServer(t, integrity)
	defer ts.Close()

	c := newClient(ts)
	c.Retries = 1
	_, err := c.Directory(context.Background(), 5)
	if !errors.Is(err, twitch.ErrRateLimited) {
		t.Fatalf("a persistent integrity check should read as rate limited, got %v", err)
	}
}

func TestGQLErrorNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"errors":[{"message":"video does not exist"}]}`))
	}))
	defer ts.Close()

	_, err := newClient(ts).Video(context.Background(), "1")
	if !errors.Is(err, twitch.ErrNotFound) {
		t.Fatalf("want ErrNotFound from gql error, got %v", err)
	}
}
