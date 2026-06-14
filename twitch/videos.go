package twitch

import (
	"context"
	"fmt"
)

// videos.go holds the single-video lookup. The channel video list lives in
// channels.go, where it shares the videoNode wire shape.

// Video returns one video by id. A null video is reported as ErrNotFound.
func (c *Client) Video(ctx context.Context, id string) (*Video, error) {
	q := fmt.Sprintf(`{ video(id: %q) { id title viewCount lengthSeconds publishedAt game { name } owner { login displayName } } }`, id)
	var resp struct {
		Video *videoNode `json:"video"`
	}
	if err := c.gql(ctx, q, &resp); err != nil {
		return nil, err
	}
	if resp.Video == nil {
		return nil, ErrNotFound
	}
	return resp.Video.toVideo(), nil
}
