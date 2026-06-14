package twitch

import (
	"context"
	"fmt"
)

// clips.go holds the clip wire shape shared by the channel and category clip
// lists, plus the single-clip lookup.

type clipNode struct {
	Slug            string `json:"slug"`
	Title           string `json:"title"`
	ViewCount       int64  `json:"viewCount"`
	DurationSeconds int    `json:"durationSeconds"`
	CreatedAt       string `json:"createdAt"`
	Game            *struct {
		Name string `json:"name"`
	} `json:"game"`
	Broadcaster *struct {
		Login string `json:"login"`
	} `json:"broadcaster"`
	Curator *struct {
		Login string `json:"login"`
	} `json:"curator"`
}

type clipEdges struct {
	Edges []struct {
		Cursor string   `json:"cursor"`
		Node   clipNode `json:"node"`
	} `json:"edges"`
	PageInfo struct {
		HasNextPage bool `json:"hasNextPage"`
	} `json:"pageInfo"`
}

func (n clipNode) toClip() *Clip {
	cl := &Clip{
		Slug:      n.Slug,
		Title:     n.Title,
		Views:     n.ViewCount,
		Duration:  n.DurationSeconds,
		CreatedAt: n.CreatedAt,
		URL:       "https://clips.twitch.tv/" + n.Slug,
	}
	if n.Game != nil {
		cl.Game = n.Game.Name
	}
	if n.Broadcaster != nil {
		cl.Channel = n.Broadcaster.Login
	}
	if n.Curator != nil {
		cl.Curator = n.Curator.Login
	}
	return cl
}

// Clip returns one clip by slug. A null clip is reported as ErrNotFound.
func (c *Client) Clip(ctx context.Context, slug string) (*Clip, error) {
	q := fmt.Sprintf(`{ clip(slug: %q) { slug title viewCount durationSeconds createdAt game { name } broadcaster { login } curator { login } } }`, slug)
	var resp struct {
		Clip *clipNode `json:"clip"`
	}
	if err := c.gql(ctx, q, &resp); err != nil {
		return nil, err
	}
	if resp.Clip == nil {
		return nil, ErrNotFound
	}
	return resp.Clip.toClip(), nil
}
