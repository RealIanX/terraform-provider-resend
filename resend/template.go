package resend

import (
	"context"
	"net/http"
)

// TemplateVariable represents a single template variable as required by the Resend API.
type TemplateVariable struct {
	Key string `json:"key"`
}

type Template struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	Alias     string             `json:"alias"`
	From      string             `json:"from"`
	Subject   string             `json:"subject"`
	ReplyTo   string             `json:"reply_to"`
	HTML      string             `json:"html"`
	Text      string             `json:"text"`
	Variables []TemplateVariable `json:"variables"`
}

type CreateTemplateRequest struct {
	Name      string             `json:"name"`
	HTML      string             `json:"html"`
	Alias     string             `json:"alias,omitempty"`
	From      string             `json:"from,omitempty"`
	Subject   string             `json:"subject,omitempty"`
	ReplyTo   string             `json:"reply_to,omitempty"`
	Text      string             `json:"text,omitempty"`
	Variables []TemplateVariable `json:"variables,omitempty"`
}

type UpdateTemplateRequest struct {
	Name      string             `json:"name,omitempty"`
	HTML      string             `json:"html,omitempty"`
	Alias     string             `json:"alias,omitempty"`
	From      string             `json:"from,omitempty"`
	Subject   string             `json:"subject,omitempty"`
	ReplyTo   string             `json:"reply_to,omitempty"`
	Text      string             `json:"text,omitempty"`
	Variables []TemplateVariable `json:"variables,omitempty"`
}

func (c *Client) CreateTemplate(ctx context.Context, req CreateTemplateRequest) (string, error) {
	var resp struct {
		ID string `json:"id"`
	}
	if err := c.Do(ctx, http.MethodPost, "/templates", req, &resp); err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (c *Client) GetTemplate(ctx context.Context, id string) (*Template, error) {
	var t Template
	if err := c.Do(ctx, http.MethodGet, "/templates/"+id, nil, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (c *Client) UpdateTemplate(ctx context.Context, id string, req UpdateTemplateRequest) error {
	return c.Do(ctx, http.MethodPatch, "/templates/"+id, req, nil)
}

func (c *Client) DeleteTemplate(ctx context.Context, id string) error {
	return c.Do(ctx, http.MethodDelete, "/templates/"+id, nil, nil)
}

func (c *Client) PublishTemplate(ctx context.Context, id string) error {
	return c.Do(ctx, http.MethodPost, "/templates/"+id+"/publish", nil, nil)
}
