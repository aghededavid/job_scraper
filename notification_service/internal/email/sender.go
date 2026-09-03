package email

import (
	"context"
	"fmt"

	resend "github.com/resend/resend-go/v2"
)

// Sender sends transactional emails via the Resend API.
type Sender struct {
	client    *resend.Client
	fromEmail string
	fromName  string
}

func NewSender(apiKey, fromEmail, fromName string) *Sender {
	return &Sender{
		client:    resend.NewClient(apiKey),
		fromEmail: fromEmail,
		fromName:  fromName,
	}
}

// SendJobAlert emails a user about a new matching job.
func (s *Sender) SendJobAlert(ctx context.Context, toEmail, jobTitle, company, applyURL string) error {
	from := fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	subject := fmt.Sprintf("🆕 New match: %s at %s", jobTitle, company)
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="font-family: Inter, sans-serif; background: #0f0f23; color: #e2e8f0; padding: 32px;">
  <div style="max-width: 560px; margin: 0 auto; background: #1a1a3e; border-radius: 12px; padding: 32px;">
    <h2 style="color: #7c3aed; margin: 0 0 16px;">New Job Match 🎯</h2>
    <h3 style="margin: 0 0 8px;">%s</h3>
    <p style="color: #94a3b8; margin: 0 0 24px;">%s</p>
    <a href="%s" style="display: inline-block; background: #7c3aed; color: white; padding: 12px 24px; border-radius: 8px; text-decoration: none; font-weight: 600;">View &amp; Apply →</a>
    <p style="color: #475569; margin-top: 32px; font-size: 12px;">
      You received this because you have job alerts enabled on RemoteScope.<br>
      <a href="%s/dashboard/notifications" style="color: #7c3aed;">Manage alerts</a>
    </p>
  </div>
</body>
</html>
`, jobTitle, company, applyURL, "https://remotescope.app")

	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{toEmail},
		Subject: subject,
		Html:    html,
	}

	_, err := s.client.Emails.Send(params)
	return err
}

// SendWelcome sends a welcome email to a new user.
func (s *Sender) SendWelcome(ctx context.Context, toEmail, name string) error {
	from := fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="font-family: Inter, sans-serif; background: #0f0f23; color: #e2e8f0; padding: 32px;">
  <div style="max-width: 560px; margin: 0 auto; background: #1a1a3e; border-radius: 12px; padding: 32px;">
    <h2 style="color: #7c3aed;">Welcome to RemoteScope, %s! 🚀</h2>
    <p>You now have access to thousands of curated remote jobs, AI-powered search, and personalised alerts.</p>
    <a href="https://remotescope.app/onboarding" style="display: inline-block; background: #7c3aed; color: white; padding: 12px 24px; border-radius: 8px; text-decoration: none; font-weight: 600;">Set up your preferences →</a>
  </div>
</body>
</html>
`, name)

	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{toEmail},
		Subject: "Welcome to RemoteScope 🎯",
		Html:    html,
	}
	_, err := s.client.Emails.Send(params)
	return err
}
