package resend

import (
	"context"
	"encoding/json"
	"net/http"
)

type Automation struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Status      string          `json:"status"`
	Steps       json.RawMessage `json:"steps"`
	Connections json.RawMessage `json:"connections"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

type CreateAutomationRequest struct {
	Name        string          `json:"name"`
	Status      string          `json:"status,omitempty"`
	Steps       json.RawMessage `json:"steps,omitempty"`
	Connections json.RawMessage `json:"connections,omitempty"`
}

type UpdateAutomationRequest struct {
	Name        string          `json:"name,omitempty"`
	Status      string          `json:"status,omitempty"`
	Steps       json.RawMessage `json:"steps,omitempty"`
	Connections json.RawMessage `json:"connections,omitempty"`
}

func (c *Client) CreateAutomation(ctx context.Context, req CreateAutomationRequest) (string, error) {
	var resp struct {
		ID string `json:"id"`
	}
	if err := c.Do(ctx, http.MethodPost, "/automations", req, &resp); err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (c *Client) GetAutomation(ctx context.Context, id string) (*Automation, error) {
	var a Automation
	if err := c.Do(ctx, http.MethodGet, "/automations/"+id, nil, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

func (c *Client) UpdateAutomation(ctx context.Context, id string, req UpdateAutomationRequest) error {
	return c.Do(ctx, http.MethodPatch, "/automations/"+id, req, nil)
}

func (c *Client) DeleteAutomation(ctx context.Context, id string) error {
	return c.Do(ctx, http.MethodDelete, "/automations/"+id, nil, nil)
}
