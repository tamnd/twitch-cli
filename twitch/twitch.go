// Package twitch fetches live streams, categories, and channels from Twitch via the public GQL API.
package twitch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	DefaultUserAgent = "Mozilla/5.0 (compatible; twitch-cli/dev; +https://github.com/tamnd/twitch-cli)"
	gqlEndpoint      = "https://gql.twitch.tv/gql"
	// Public client ID extracted from the Twitch web client.
	defaultClientID = "kimne78kx3ncx6brgo4mv6wki5h1ko"
)

// Config holds all tuneable client parameters.
type Config struct {
	BaseURL   string
	ClientID  string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
	UserAgent string
}

// DefaultConfig returns sensible defaults for the Twitch GQL API.
func DefaultConfig() Config {
	return Config{
		BaseURL:   gqlEndpoint,
		ClientID:  defaultClientID,
		Rate:      200 * time.Millisecond,
		Timeout:   30 * time.Second,
		Retries:   3,
		UserAgent: DefaultUserAgent,
	}
}

// Client talks to the Twitch GQL API.
type Client struct {
	cfg  Config
	http *http.Client
	last time.Time
}

// NewClient returns a Client with the given configuration.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// Streams returns the top live streams, optionally filtered by game name.
func (c *Client) Streams(ctx context.Context, game string, limit int) ([]Stream, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var query string
	if game != "" {
		query = fmt.Sprintf(`{ game(name: %q) { streams(first: %d) { edges { cursor node { id title viewersCount createdAt broadcaster { login displayName } } } } } }`, game, limit)
	} else {
		query = fmt.Sprintf(`{ streams(first: %d) { edges { cursor node { id title viewersCount createdAt game { name } broadcaster { login displayName } } } } }`, limit)
	}

	var resp struct {
		Data struct {
			Streams *struct {
				Edges []struct {
					Node streamNode `json:"node"`
				} `json:"edges"`
			} `json:"streams"`
			Game *struct {
				Streams struct {
					Edges []struct {
						Node streamNode `json:"node"`
					} `json:"edges"`
				} `json:"streams"`
			} `json:"game"`
		} `json:"data"`
	}

	if err := c.gql(ctx, query, &resp); err != nil {
		return nil, err
	}

	var edges []struct {
		Node streamNode `json:"node"`
	}
	if game != "" && resp.Data.Game != nil {
		edges = resp.Data.Game.Streams.Edges
	} else if resp.Data.Streams != nil {
		edges = resp.Data.Streams.Edges
	}

	out := make([]Stream, 0, len(edges))
	for i, e := range edges {
		gameName := e.Node.GameName
		if e.Node.Game != nil {
			gameName = e.Node.Game.Name
		}
		out = append(out, Stream{
			Rank:      i + 1,
			ID:        e.Node.ID,
			Title:     e.Node.Title,
			Game:      gameName,
			Channel:   e.Node.Broadcaster.Login,
			Viewers:   e.Node.ViewersCount,
			StartedAt: e.Node.CreatedAt,
			URL:       "https://www.twitch.tv/" + e.Node.Broadcaster.Login,
		})
	}
	return out, nil
}

// Categories returns the top categories/games by viewer count.
func (c *Client) Categories(ctx context.Context, limit int) ([]Category, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := fmt.Sprintf(`{ games(first: %d) { edges { node { id name slug viewersCount } } } }`, limit)

	var resp struct {
		Data struct {
			Games struct {
				Edges []struct {
					Node struct {
						ID           string `json:"id"`
						Name         string `json:"name"`
						Slug         string `json:"slug"`
						ViewersCount int    `json:"viewersCount"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"games"`
		} `json:"data"`
	}

	if err := c.gql(ctx, query, &resp); err != nil {
		return nil, err
	}

	out := make([]Category, 0, len(resp.Data.Games.Edges))
	for i, e := range resp.Data.Games.Edges {
		out = append(out, Category{
			Rank:    i + 1,
			ID:      e.Node.ID,
			Name:    e.Node.Name,
			Slug:    e.Node.Slug,
			Viewers: e.Node.ViewersCount,
			URL:     "https://www.twitch.tv/directory/game/" + e.Node.Slug,
		})
	}
	return out, nil
}

// Search searches for channels and games matching the query.
func (c *Client) SearchChannels(ctx context.Context, query string) ([]Channel, error) {
	gqlQuery := fmt.Sprintf(`{ searchFor(userQuery: %q, platform: "web") { channels { items { id login displayName } } } }`, query)

	var resp struct {
		Data struct {
			SearchFor struct {
				Channels struct {
					Items []struct {
						ID          string `json:"id"`
						Login       string `json:"login"`
						DisplayName string `json:"displayName"`
					} `json:"items"`
				} `json:"channels"`
			} `json:"searchFor"`
		} `json:"data"`
	}

	if err := c.gql(ctx, gqlQuery, &resp); err != nil {
		return nil, err
	}

	items := resp.Data.SearchFor.Channels.Items
	out := make([]Channel, 0, len(items))
	for i, item := range items {
		out = append(out, Channel{
			Rank:        i + 1,
			ID:          item.ID,
			Login:       item.Login,
			DisplayName: item.DisplayName,
			URL:         "https://www.twitch.tv/" + item.Login,
		})
	}
	return out, nil
}

type streamNode struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	GameName string `json:"gameName"`
	Game     *struct {
		Name string `json:"name"`
	} `json:"game"`
	ViewersCount int    `json:"viewersCount"`
	CreatedAt    string `json:"createdAt"`
	Broadcaster  struct {
		Login       string `json:"login"`
		DisplayName string `json:"displayName"`
	} `json:"broadcaster"`
}

func (c *Client) gql(ctx context.Context, query string, out any) error {
	body, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		err = c.do(ctx, body, out)
		if err == nil {
			return nil
		}
		lastErr = err
	}
	return fmt.Errorf("gql: %w", lastErr)
}

func (c *Client) do(ctx context.Context, body []byte, out any) error {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	req.Header.Set("Client-Id", c.cfg.ClientID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Check for GraphQL errors
	var gqlResp struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(b, &gqlResp); err == nil && len(gqlResp.Errors) > 0 {
		return fmt.Errorf("gql error: %s", gqlResp.Errors[0].Message)
	}

	return json.Unmarshal(b, out)
}

func (c *Client) pace() {
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}
