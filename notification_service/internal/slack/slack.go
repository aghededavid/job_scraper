package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Sender sends messages to a Slack channel via an Incoming Webhook.
type Sender struct {
	webhookURL string
	client     *http.Client
}

func NewSender(webhookURL string) *Sender {
	return &Sender{webhookURL: webhookURL, client: &http.Client{}}
}

type slackMessage struct {
	Text        string        `json:"text,omitempty"`
	Attachments []slackAttach `json:"attachments,omitempty"`
}

type slackAttach struct {
	Text  string `json:"text"`
	Color string `json:"color"`
}

// Send posts a message to Slack.
func (s *Sender) Send(ctx context.Context, message string) error {
	body := slackMessage{
		Attachments: []slackAttach{
			{Text: message, Color: "#7c3aed"},
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
		return fmt.Errorf("slack webhook returned %d", resp.StatusCode)
	}
	return nil
}
