package twitch

import (
	"context"
	"fmt"
)

// games.go reads categories: the directory, one category's metadata, and a
// category's top clips. (A category's live streams live in streams.go.)

type gameNode struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	DisplayName  string `json:"displayName"`
	ViewersCount int    `json:"viewersCount"`
}

func (n gameNode) toGame() *Game {
	return &Game{
		ID:          n.ID,
		Name:        n.Name,
		Slug:        n.Slug,
		DisplayName: n.DisplayName,
		Viewers:     n.ViewersCount,
		URL:         BaseURL + "/directory/game/" + n.Slug,
	}
}

// Directory returns the category directory, ordered by live viewers.
func (c *Client) Directory(ctx context.Context, limit int) ([]*Game, error) {
	var out []*Game
	err := paginate(limit, func(after string, first int) (int, string, bool, error) {
		q := fmt.Sprintf(`{ games(first: %d%s) { edges { cursor node { id name slug displayName viewersCount } } pageInfo { hasNextPage } } }`,
			first, afterArg(after))
		var resp struct {
			Games struct {
				Edges []struct {
					Cursor string   `json:"cursor"`
					Node   gameNode `json:"node"`
				} `json:"edges"`
				PageInfo struct {
					HasNextPage bool `json:"hasNextPage"`
				} `json:"pageInfo"`
			} `json:"games"`
		}
		if err := c.gql(ctx, q, &resp); err != nil {
			return 0, "", false, err
		}
		next := ""
		for _, e := range resp.Games.Edges {
			out = append(out, e.Node.toGame())
			next = e.Cursor
		}
		return len(resp.Games.Edges), next, resp.Games.PageInfo.HasNextPage, nil
	})
	if err != nil {
		return nil, err
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// Game returns one category's metadata by slug. A null game is reported as
// ErrNotFound.
func (c *Client) Game(ctx context.Context, slug string) (*Game, error) {
	q := fmt.Sprintf(`{ game(slug: %q) { id name slug displayName viewersCount } }`, slug)
	var resp struct {
		Game *gameNode `json:"game"`
	}
	if err := c.gql(ctx, q, &resp); err != nil {
		return nil, err
	}
	if resp.Game == nil {
		return nil, ErrNotFound
	}
	return resp.Game.toGame(), nil
}

// GameClips returns a category's top clips of the last week.
func (c *Client) GameClips(ctx context.Context, slug string, limit int) ([]*Clip, error) {
	var out []*Clip
	found := false
	err := paginate(limit, func(after string, first int) (int, string, bool, error) {
		q := fmt.Sprintf(`{ game(slug: %q) { clips(first: %d%s, criteria: { period: LAST_WEEK, sort: VIEWS_DESC }) { edges { cursor node { slug title viewCount durationSeconds createdAt game { name } broadcaster { login } curator { login } } } pageInfo { hasNextPage } } } }`,
			slug, first, afterArg(after))
		var resp struct {
			Game *struct {
				Clips clipEdges `json:"clips"`
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
		for _, e := range resp.Game.Clips.Edges {
			cl := e.Node.toClip()
			if cl.Game == "" {
				cl.Game = slug
			}
			out = append(out, cl)
			next = e.Cursor
		}
		return len(resp.Game.Clips.Edges), next, resp.Game.Clips.PageInfo.HasNextPage, nil
	})
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrNotFound
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
