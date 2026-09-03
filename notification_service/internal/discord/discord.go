package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Sender sends messages to a Discord channel via a webhook.
type Sender struct {
	webhookURL string
	client     *http.Client
}

func NewSender(webhookURL string) *Sender {
	return &Sender{webhookURL: webhookURL, client: &http.Client{}}
}

type discordMessage struct {
	Content  string         `json:"content,omitempty"`
	Username string         `json:"username,omitempty"`
	Embeds   []discordEmbed `json:"embeds,omitempty"`
}

type discordEmbed struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Color       int    `json:"color"` // decimal colour
}

// Send posts a message to Discord.
func (s *Sender) Send(ctx context.Context, message string) error {
	body := discordMessage{
		Username: "RemoteScope",
		Embeds: []discordEmbed{
			{
				Title:       "RemoteScope Alert",
				Description: message,
				Color:       7864319, // #780AFF purple
			},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("discord webhook returned %d", resp.StatusCode)
	}
	return nil
}
