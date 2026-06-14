package twitch_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tamnd/twitch-cli/twitch"
)

func fakeGQLResponse(data any) string {
	b, _ := json.Marshal(map[string]any{"data": data})
	return string(b)
}

func newTestClient(ts *httptest.Server) *twitch.Client {
	cfg := twitch.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return twitch.NewClient(cfg)
}

func TestStreams(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, fakeGQLResponse(map[string]any{
			"streams": map[string]any{
				"edges": []map[string]any{
					{"node": map[string]any{
						"id": "123", "title": "Test Stream", "viewersCount": 1000,
						"createdAt": "2026-01-01T00:00:00Z",
						"game":        map[string]any{"name": "Minecraft"},
						"broadcaster": map[string]any{"login": "testuser", "displayName": "TestUser"},
					}},
				},
			},
		}))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	streams, err := c.Streams(context.Background(), "", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(streams) != 1 {
		t.Fatalf("want 1 stream, got %d", len(streams))
	}
	if streams[0].Title != "Test Stream" {
		t.Errorf("title = %q, want Test Stream", streams[0].Title)
	}
	if streams[0].Viewers != 1000 {
		t.Errorf("viewers = %d, want 1000", streams[0].Viewers)
	}
}

func TestCategories(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, fakeGQLResponse(map[string]any{
			"games": map[string]any{
				"edges": []map[string]any{
					{"node": map[string]any{"id": "1", "name": "Minecraft", "slug": "minecraft", "viewersCount": 50000}},
					{"node": map[string]any{"id": "2", "name": "Fortnite", "slug": "fortnite", "viewersCount": 30000}},
				},
			},
		}))
	}))
	defer ts.Close()

	c := newTestClient(ts)
	cats, err := c.Categories(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(cats) != 2 {
		t.Fatalf("want 2 categories, got %d", len(cats))
	}
	if cats[0].Name != "Minecraft" {
		t.Errorf("first name = %q, want Minecraft", cats[0].Name)
	}
}
