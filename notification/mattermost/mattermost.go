package mattermost

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Client provides functions for interacting with the Mattermost API (to send messages).
type Client struct {
	WebhookURL string
	User       string
	Channel    string
}

// Message is the request body for a Mattermost message.
type Message struct {
	Text        string         `json:"text,omitempty"`
	Username    string         `json:"username,omitempty"`
	IconURL     string         `json:"icon_url,omitempty"`
	IconEmoji   string         `json:"icon_emoji,omitempty"`
	Channel     string         `json:"channel,omitempty"`
	Props       map[string]any `json:"props,omitempty"`
	Attachments []any          `json:"attachments,omitempty"`
}

// New creates a new Mattermost client.
func New(url string) *Client {
	return &Client{
		WebhookURL: url,
	}
}

// Send a message.
func (c *Client) Send(ctx context.Context, msg Message) error {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(msg)
	if err != nil {
		return fmt.Errorf("failed to encode Mattermost message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.WebhookURL, b)
	if err != nil {
		return fmt.Errorf("failed to create Mattermost request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to POST Mattermost message: %w", err)
	}
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to POST Mattermost message, non-200 status: %d", resp.StatusCode)
	}

	return nil
}
