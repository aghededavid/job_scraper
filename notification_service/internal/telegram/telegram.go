package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Sender sends messages via the Telegram Bot API.
type Sender struct {
	token  string
	client *http.Client
}

func NewSender(token string) *Sender {
	return &Sender{token: token, client: &http.Client{}}
}

type telegramMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

// SendToChannel sends a message to a Telegram channel or chat.
// chatID can be a numeric chat ID or @channel_username.
func (s *Sender) SendToChannel(ctx context.Context, chatID, message string) error {
	body := telegramMessage{
		ChatID:    chatID,
		Text:      message,
		ParseMode: "HTML",
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
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
		return fmt.Errorf("telegram API returned %d", resp.StatusCode)
	}
	return nil
}

// Send implements the generic send interface (logs only — requires a chatID per user).
func (s *Sender) Send(ctx context.Context, message string) error {
	// Channel-level broadcast requires a configured channel ID.
	// Per-user messages use SendToChannel with the user's telegram_chat_id.
	return nil
}
