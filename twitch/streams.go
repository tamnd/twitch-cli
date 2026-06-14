package twitch

import (
	"context"
	"fmt"
)

// streams.go reads live broadcasts: the global top list and a category's list.

// wire structs shared by the stream queries.
type streamNode struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	ViewersCount int    `json:"viewersCount"`
	CreatedAt    string `json:"createdAt"`
	Language     string `json:"language"`
	Type         string `json:"type"`
	Game         *struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"game"`
	Broadcaster *struct {
		Login       string `json:"login"`
		DisplayName string `json:"displayName"`
	} `json:"broadcaster"`
	FreeformTags []struct {
		Name string `json:"name"`
	} `json:"freeformTags"`
}

type streamEdges struct {
	Edges []struct {
		Cursor string     `json:"cursor"`
		Node   streamNode `json:"node"`
	} `json:"edges"`
	PageInfo struct {
		HasNextPage bool `json:"hasNextPage"`
	} `json:"pageInfo"`
}

func (n streamNode) toStream() *Stream {
	s := &Stream{
		ID:        n.ID,
		Title:     n.Title,
		Viewers:   n.ViewersCount,
		Language:  n.Language,
		StartedAt: n.CreatedAt,
	}
	if n.Game != nil {
		s.Game = n.Game.Name
		s.GameSlug = n.Game.Slug
	}
	if n.Broadcaster != nil {
		s.Channel = n.Broadcaster.Login
		s.DisplayName = n.Broadcaster.DisplayName
		s.URL = BaseURL + "/" + n.Broadcaster.Login
	}
	return s
}

// TopStreams returns the top live streams across all categories, paginated to
// limit.
func (c *Client) TopStreams(ctx context.Context, limit int) ([]*Stream, error) {
	var out []*Stream
	err := paginate(limit, func(after string, first int) (int, string, bool, error) {
		q := fmt.Sprintf(`{ streams(first: %d%s) { edges { cursor node { id title viewersCount createdAt language type game { name slug } broadcaster { login displayName } freeformTags { name } } } pageInfo { hasNextPage } } }`,
			first, afterArg(after))
		var resp struct {
			Streams streamEdges `json:"streams"`
		}
		if err := c.gql(ctx, q, &resp); err != nil {
			return 0, "", false, err
		}
		next := ""
		for _, e := range resp.Streams.Edges {
			out = append(out, e.Node.toStream())
			next = e.Cursor
		}
		return len(resp.Streams.Edges), next, resp.Streams.PageInfo.HasNextPage, nil
	})
	if err != nil {
		return nil, err
	}
	return capStreams(out, limit), nil
}

// GameStreams returns the live streams in a category, sorted by viewers.
func (c *Client) GameStreams(ctx context.Context, slug string, limit int) ([]*Stream, error) {
	var out []*Stream
	found := false
	err := paginate(limit, func(after string, first int) (int, string, bool, error) {
		q := fmt.Sprintf(`{ game(slug: %q) { streams(first: %d%s, sort: VIEWER_COUNT) { edges { cursor node { id title viewersCount createdAt language broadcaster { login displayName } } } pageInfo { hasNextPage } } } }`,
			slug, first, afterArg(after))
		var resp struct {
			Game *struct {
				Streams streamEdges `json:"streams"`
			} `json:"game"`
		}
		if err := c.gql(ctx, q, &resp); err != nil {
			return 0, "", false, err
		}
		if resp.Game == nil {
			return 0, "", false, nil
		}
		found = true
		next := ""
		for _, e := range resp.Game.Streams.Edges {
			s := e.Node.toStream()
			s.Game = slug
			out = append(out, s)
			next = e.Cursor
		}
		return len(resp.Game.Streams.Edges), next, resp.Game.Streams.PageInfo.HasNextPage, nil
	})
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrNotFound
	}
	return capStreams(out, limit), nil
}

func capStreams(s []*Stream, limit int) []*Stream {
	if limit > 0 && len(s) > limit {
		return s[:limit]
	}
	return s
}

// afterArg renders the optional after cursor, omitted on the first page.
func afterArg(after string) string {
	if after == "" {
		return ""
	}
	return fmt.Sprintf(`, after: %q`, after)
}
