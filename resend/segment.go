package resend

import (
	"context"
	"net/http"
)

type Segment struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

func (c *Client) CreateSegment(ctx context.Context, name string) (string, error) {
	var resp struct {
		ID string `json:"id"`
	}
	if err := c.Do(ctx, http.MethodPost, "/segments", map[string]string{"name": name}, &resp); err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (c *Client) GetSegment(ctx context.Context, id string) (*Segment, error) {
	var s Segment
	if err := c.Do(ctx, http.MethodGet, "/segments/"+id, nil, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (c *Client) DeleteSegment(ctx context.Context, id string) error {
	return c.Do(ctx, http.MethodDelete, "/segments/"+id, nil, nil)
}
