package twitch

import (
	"context"
	"fmt"
)

// channels.go reads a channel and its members: metadata, videos, clips, and the
// upcoming schedule.

// Channel returns one channel's metadata. A null user is reported as
// ErrNotFound.
func (c *Client) Channel(ctx context.Context, login string) (*Channel, error) {
	q := fmt.Sprintf(`{ user(login: %q) { id login displayName description createdAt profileImageURL(width: 150) followers { totalCount } roles { isPartner isAffiliate } stream { id viewersCount game { name } } } }`, login)
	var resp struct {
		User *struct {
			ID              string `json:"id"`
			Login           string `json:"login"`
			DisplayName     string `json:"displayName"`
			Description     string `json:"description"`
			CreatedAt       string `json:"createdAt"`
			ProfileImageURL string `json:"profileImageURL"`
			Followers       *struct {
				TotalCount int64 `json:"totalCount"`
			} `json:"followers"`
			Roles *struct {
				IsPartner   bool `json:"isPartner"`
				IsAffiliate bool `json:"isAffiliate"`
			} `json:"roles"`
			Stream *struct {
				ID           string `json:"id"`
				ViewersCount int    `json:"viewersCount"`
				Game         *struct {
					Name string `json:"name"`
				} `json:"game"`
			} `json:"stream"`
		} `json:"user"`
	}
	if err := c.gql(ctx, q, &resp); err != nil {
		return nil, err
	}
	if resp.User == nil {
		return nil, ErrNotFound
	}
	u := resp.User
	ch := &Channel{
		ID:          u.ID,
		Login:       u.Login,
		DisplayName: u.DisplayName,
		Description: u.Description,
		Created:     u.CreatedAt,
		Avatar:      u.ProfileImageURL,
		URL:         BaseURL + "/" + u.Login,
	}
	if u.Followers != nil {
		ch.Followers = u.Followers.TotalCount
	}
	if u.Roles != nil {
		ch.Partner = u.Roles.IsPartner
		ch.Affiliate = u.Roles.IsAffiliate
	}
	if u.Stream != nil {
		ch.Live = true
		ch.Viewers = u.Stream.ViewersCount
		if u.Stream.Game != nil {
			ch.Game = u.Stream.Game.Name
		}
	}
	return ch, nil
}

type videoNode struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	ViewCount    int64  `json:"viewCount"`
	LengthSecond int    `json:"lengthSeconds"`
	PublishedAt  string `json:"publishedAt"`
	Game         *struct {
		Name string `json:"name"`
	} `json:"game"`
	Owner *struct {
		Login       string `json:"login"`
		DisplayName string `json:"displayName"`
	} `json:"owner"`
}

func (n videoNode) toVideo() *Video {
	v := &Video{
		ID:          n.ID,
		Title:       n.Title,
		Views:       n.ViewCount,
		Length:      n.LengthSecond,
		PublishedAt: n.PublishedAt,
		URL:         BaseURL + "/videos/" + n.ID,
	}
	if n.Game != nil {
		v.Game = n.Game.Name
	}
	if n.Owner != nil {
		v.Channel = n.Owner.Login
		v.DisplayName = n.Owner.DisplayName
	}
	return v
}

// ChannelVideos returns a channel's past videos, most recent first.
func (c *Client) ChannelVideos(ctx context.Context, login string, limit int) ([]*Video, error) {
	var out []*Video
	found := false
	err := paginate(limit, func(after string, first int) (int, string, bool, error) {
		q := fmt.Sprintf(`{ user(login: %q) { videos(first: %d%s, sort: TIME) { edges { cursor node { id title viewCount lengthSeconds publishedAt game { name } owner { login displayName } } } pageInfo { hasNextPage } } } }`,
			login, first, afterArg(after))
		var resp struct {
			User *struct {
				Videos struct {
					Edges []struct {
						Cursor string    `json:"cursor"`
						Node   videoNode `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						HasNextPage bool `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"videos"`
			} `json:"user"`
		}
		if err := c.gql(ctx, q, &resp); err != nil {
			return 0, "", false, err
		}
		if resp.User == nil {
			return 0, "", false, nil
		}
		found = true
		next := ""
		for _, e := range resp.User.Videos.Edges {
			out = append(out, e.Node.toVideo())
			next = e.Cursor
		}
		return len(resp.User.Videos.Edges), next, resp.User.Videos.PageInfo.HasNextPage, nil
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

// ChannelClips returns a channel's clips, most viewed first.
func (c *Client) ChannelClips(ctx context.Context, login string, limit int) ([]*Clip, error) {
	var out []*Clip
	found := false
	err := paginate(limit, func(after string, first int) (int, string, bool, error) {
		q := fmt.Sprintf(`{ user(login: %q) { clips(first: %d%s, criteria: { period: ALL_TIME, sort: VIEWS_DESC }) { edges { cursor node { slug title viewCount durationSeconds createdAt game { name } broadcaster { login } curator { login } } } pageInfo { hasNextPage } } } }`,
			login, first, afterArg(after))
		var resp struct {
			User *struct {
				Clips clipEdges `json:"clips"`
			} `json:"user"`
		}
		if err := c.gql(ctx, q, &resp); err != nil {
			return 0, "", false, err
		}
		if resp.User == nil {
			return 0, "", false, nil
		}
		found = true
		next := ""
		for _, e := range resp.User.Clips.Edges {
			out = append(out, e.Node.toClip())
			next = e.Cursor
		}
		return len(resp.User.Clips.Edges), next, resp.User.Clips.PageInfo.HasNextPage, nil
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

// ChannelSchedule returns a channel's upcoming schedule segments. The schedule
// connection takes no first argument, so the whole list is fetched and capped to
// limit. A channel with no schedule returns an empty list, not an error.
func (c *Client) ChannelSchedule(ctx context.Context, login string, limit int) ([]*Segment, error) {
	q := fmt.Sprintf(`{ user(login: %q) { channel { schedule { segments { id title startAt endAt isCancelled categories { name } } } } } }`, login)
	var resp struct {
		User *struct {
			Channel *struct {
				Schedule *struct {
					Segments []struct {
						ID          string `json:"id"`
						Title       string `json:"title"`
						StartAt     string `json:"startAt"`
						EndAt       string `json:"endAt"`
						IsCancelled bool   `json:"isCancelled"`
						Categories  []struct {
							Name string `json:"name"`
						} `json:"categories"`
					} `json:"segments"`
				} `json:"schedule"`
			} `json:"channel"`
		} `json:"user"`
	}
	if err := c.gql(ctx, q, &resp); err != nil {
		return nil, err
	}
	if resp.User == nil {
		return nil, ErrNotFound
	}
	var out []*Segment
	if resp.User.Channel == nil || resp.User.Channel.Schedule == nil {
		return out, nil
	}
	for _, s := range resp.User.Channel.Schedule.Segments {
		seg := &Segment{
			ID:       s.ID,
			Title:    s.Title,
			StartAt:  s.StartAt,
			EndAt:    s.EndAt,
			Canceled: s.IsCancelled,
		}
		if len(s.Categories) > 0 {
			seg.Category = s.Categories[0].Name
		}
		out = append(out, seg)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}
