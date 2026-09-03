package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"github.com/remotescope/backend/pkg/config"
)

// ──────────────────────────────────────────
// OAuth User Info
// ──────────────────────────────────────────

type OAuthUserInfo struct {
	ID        string
	Email     string
	Name      string
	AvatarURL string
	Provider  string
}

// ──────────────────────────────────────────
// OAuth Service
// ──────────────────────────────────────────

type OAuthService struct {
	googleConfig   *oauth2.Config
	githubConfig   *oauth2.Config
	linkedinConfig *oauth2.Config
}

func NewOAuthService(cfg *config.Config) *OAuthService {
	return &OAuthService{
		googleConfig: &oauth2.Config{
			ClientID:     cfg.OAuth.Google.ClientID,
			ClientSecret: cfg.OAuth.Google.ClientSecret,
			RedirectURL:  cfg.OAuth.Google.RedirectURI,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
		githubConfig: &oauth2.Config{
			ClientID:     cfg.OAuth.GitHub.ClientID,
			ClientSecret: cfg.OAuth.GitHub.ClientSecret,
			RedirectURL:  cfg.OAuth.GitHub.RedirectURI,
			Scopes:       []string{"read:user", "user:email"},
			Endpoint:     github.Endpoint,
		},
		linkedinConfig: &oauth2.Config{
			ClientID:     cfg.OAuth.LinkedIn.ClientID,
			ClientSecret: cfg.OAuth.LinkedIn.ClientSecret,
			RedirectURL:  cfg.OAuth.LinkedIn.RedirectURI,
			Scopes:       []string{"openid", "profile", "email"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://www.linkedin.com/oauth/v2/authorization",
				TokenURL: "https://www.linkedin.com/oauth/v2/accessToken",
			},
		},
	}
}

func (s *OAuthService) GetAuthURL(provider, state string) (string, error) {
	cfg := s.getConfig(provider)
	if cfg == nil {
		return "", fmt.Errorf("unknown provider: %s", provider)
	}
	return cfg.AuthCodeURL(state, oauth2.AccessTypeOnline), nil
}

func (s *OAuthService) ExchangeAndGetUser(ctx context.Context, provider, code string) (*OAuthUserInfo, error) {
	cfg := s.getConfig(provider)
	if cfg == nil {
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}

	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("token exchange: %w", err)
	}

	switch provider {
	case "google":
		return s.fetchGoogleUser(ctx, cfg, token)
	case "github":
		return s.fetchGitHubUser(ctx, cfg, token)
	case "linkedin":
		return s.fetchLinkedInUser(ctx, cfg, token)
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}

func (s *OAuthService) getConfig(provider string) *oauth2.Config {
	switch provider {
	case "google":
		return s.googleConfig
	case "github":
		return s.githubConfig
	case "linkedin":
		return s.linkedinConfig
	}
	return nil
}

func (s *OAuthService) fetchGoogleUser(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token) (*OAuthUserInfo, error) {
	client := cfg.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var data struct {
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	return &OAuthUserInfo{
		ID:        data.Sub,
		Email:     data.Email,
		Name:      data.Name,
		AvatarURL: data.Picture,
		Provider:  "google",
	}, nil
}

func (s *OAuthService) fetchGitHubUser(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token) (*OAuthUserInfo, error) {
	client := cfg.Client(ctx, token)

	// Get user profile
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var data struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	// If email not public, fetch from emails endpoint
	if data.Email == "" {
		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err == nil {
			defer emailResp.Body.Close()
			emailBody, _ := io.ReadAll(emailResp.Body)
			var emails []struct {
				Email   string `json:"email"`
				Primary bool   `json:"primary"`
			}
			if json.Unmarshal(emailBody, &emails) == nil {
				for _, e := range emails {
					if e.Primary {
						data.Email = e.Email
						break
					}
				}
			}
		}
	}

	name := data.Name
	if name == "" {
		name = data.Login
	}

	return &OAuthUserInfo{
		ID:        fmt.Sprintf("%d", data.ID),
		Email:     data.Email,
		Name:      name,
		AvatarURL: data.AvatarURL,
		Provider:  "github",
	}, nil
}

func (s *OAuthService) fetchLinkedInUser(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token) (*OAuthUserInfo, error) {
	client := cfg.Client(ctx, token)

	req, _ := http.NewRequest("GET", "https://api.linkedin.com/v2/userinfo", nil)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var data struct {
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	return &OAuthUserInfo{
		ID:        data.Sub,
		Email:     data.Email,
		Name:      data.Name,
		AvatarURL: data.Picture,
		Provider:  "linkedin",
	}, nil
}
