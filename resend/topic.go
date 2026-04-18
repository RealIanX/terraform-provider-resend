package resend

import (
	"context"
	"net/http"
)

type Topic struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	DefaultSubscription string `json:"default_subscription"`
	Description         string `json:"description"`
	Visibility          string `json:"visibility"`
	CreatedAt           string `json:"created_at"`
}

type CreateTopicRequest struct {
	Name                string `json:"name"`
	DefaultSubscription string `json:"default_subscription"`
	Description         string `json:"description,omitempty"`
	Visibility          string `json:"visibility,omitempty"`
}

type UpdateTopicRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Visibility  string `json:"visibility,omitempty"`
}

func (c *Client) CreateTopic(ctx context.Context, req CreateTopicRequest) (string, error) {
	var resp struct {
		ID string `json:"id"`
	}
	if err := c.Do(ctx, http.MethodPost, "/topics", req, &resp); err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (c *Client) GetTopic(ctx context.Context, id string) (*Topic, error) {
	var t Topic
	if err := c.Do(ctx, http.MethodGet, "/topics/"+id, nil, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (c *Client) UpdateTopic(ctx context.Context, id string, req UpdateTopicRequest) error {
	return c.Do(ctx, http.MethodPatch, "/topics/"+id, req, nil)
}

func (c *Client) DeleteTopic(ctx context.Context, id string) error {
	return c.Do(ctx, http.MethodDelete, "/topics/"+id, nil, nil)
}
