package twitch

import (
	"context"
	"fmt"
)

// search.go reads the search surface. searchFor returns plain item lists rather
// than cursor connections; this reads the first page and caps to limit. Deeper
// search paging is a possible later addition.

// SearchChannels returns channels matching a query.
func (c *Client) SearchChannels(ctx context.Context, query string, limit int) ([]*Channel, error) {
	q := fmt.Sprintf(`{ searchFor(userQuery: %q, platform: "web") { channels { items { id login displayName followers { totalCount } } } } }`, query)
	var resp struct {
		SearchFor struct {
			Channels struct {
				Items []struct {
					ID          string `json:"id"`
					Login       string `json:"login"`
					DisplayName string `json:"displayName"`
					Followers   *struct {
						TotalCount int64 `json:"totalCount"`
					} `json:"followers"`
				} `json:"items"`
			} `json:"channels"`
		} `json:"searchFor"`
	}
	if err := c.gql(ctx, q, &resp); err != nil {
		return nil, err
	}
	var out []*Channel
	for _, it := range resp.SearchFor.Channels.Items {
		ch := &Channel{
			ID:          it.ID,
			Login:       it.Login,
			DisplayName: it.DisplayName,
			URL:         BaseURL + "/" + it.Login,
		}
		if it.Followers != nil {
			ch.Followers = it.Followers.TotalCount
		}
		out = append(out, ch)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

// SearchGames returns categories matching a query.
func (c *Client) SearchGames(ctx context.Context, query string, limit int) ([]*Game, error) {
	q := fmt.Sprintf(`{ searchFor(userQuery: %q, platform: "web") { games { items { id name slug displayName viewersCount } } } }`, query)
	var resp struct {
		SearchFor struct {
			Games struct {
				Items []gameNode `json:"items"`
			} `json:"games"`
		} `json:"searchFor"`
	}
	if err := c.gql(ctx, q, &resp); err != nil {
		return nil, err
	}
	var out []*Game
	for _, it := range resp.SearchFor.Games.Items {
		out = append(out, it.toGame())
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}
