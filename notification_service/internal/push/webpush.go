package push

import (
	"context"
	"encoding/json"

	webpush "github.com/SherClockHolmes/webpush-go"
)

// Sender sends Web Push notifications using VAPID keys.
type Sender struct {
	publicKey  string
	privateKey string
	subject    string
}

func NewSender(publicKey, privateKey, subject string) *Sender {
	return &Sender{
		publicKey:  publicKey,
		privateKey: privateKey,
		subject:    subject,
	}
}

// PushPayload is the JSON sent to the browser's push endpoint.
type PushPayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Icon  string `json:"icon"`
	URL   string `json:"url"`
}

// Send sends a push notification to a subscriber.
// subscription is the raw JSON of the browser PushSubscription object.
func (s *Sender) Send(ctx context.Context, subscriptionJSON []byte, title, body, targetURL string) error {
	sub := &webpush.Subscription{}
	if err := json.Unmarshal(subscriptionJSON, sub); err != nil {
		return err
	}

	payload, err := json.Marshal(PushPayload{
		Title: title,
		Body:  body,
		Icon:  "/icons/icon-192x192.png",
		URL:   targetURL,
	})
	if err != nil {
		return err
	}

	resp, err := webpush.SendNotification(payload, sub, &webpush.Options{
		Subscriber:      s.subject,
		VAPIDPublicKey:  s.publicKey,
		VAPIDPrivateKey: s.privateKey,
		TTL:             86400,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
