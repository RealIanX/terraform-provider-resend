package resend

import (
	"context"
	"net/http"
)

type Event struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Schema string `json:"schema"`
}

type CreateEventRequest struct {
	Name   string `json:"name"`
	Schema any    `json:"schema,omitempty"`
}

type UpdateEventRequest struct {
	Name   string `json:"name,omitempty"`
	Schema any    `json:"schema,omitempty"`
}

func (c *Client) CreateEvent(ctx context.Context, req CreateEventRequest) (string, error) {
	var resp struct {
		ID string `json:"id"`
	}
	if err := c.Do(ctx, http.MethodPost, "/events", req, &resp); err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (c *Client) GetEvent(ctx context.Context, id string) (*Event, error) {
	var e Event
	if err := c.Do(ctx, http.MethodGet, "/events/"+id, nil, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

func (c *Client) UpdateEvent(ctx context.Context, id string, req UpdateEventRequest) error {
	return c.Do(ctx, http.MethodPatch, "/events/"+id, req, nil)
}

func (c *Client) DeleteEvent(ctx context.Context, id string) error {
	return c.Do(ctx, http.MethodDelete, "/events/"+id, nil, nil)
}
