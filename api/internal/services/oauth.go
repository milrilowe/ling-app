package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"ling-app/api/internal/config"
)

// OAuthService handles OAuth2 authentication with Google and GitHub
type OAuthService struct {
	googleConfig *oauth2.Config
	githubConfig *oauth2.Config
}

// GoogleUser represents the user info returned by Google
type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
}

// GitHubUser represents the user info returned by GitHub
type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// GitHubEmail represents email info from GitHub's emails API
type GitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

// NewOAuthService creates a new OAuth service with Google and GitHub configs
func NewOAuthService(cfg *config.Config) *OAuthService {
	var googleCfg *oauth2.Config
	var githubCfg *oauth2.Config

	// Only set up Google if credentials are provided
	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		googleCfg = &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			Scopes:       []string{"email", "profile"},
			Endpoint:     google.Endpoint,
		}
	}

	// Only set up GitHub if credentials are provided
	if cfg.GitHubClientID != "" && cfg.GitHubClientSecret != "" {
		githubCfg = &oauth2.Config{
			ClientID:     cfg.GitHubClientID,
			ClientSecret: cfg.GitHubClientSecret,
			RedirectURL:  cfg.GitHubRedirectURL,
			Scopes:       []string{"user:email", "read:user"},
			Endpoint:     github.Endpoint,
		}
	}

	return &OAuthService{
		googleConfig: googleCfg,
		githubConfig: githubCfg,
	}
}

// IsGoogleEnabled returns true if Google OAuth is configured
func (s *OAuthService) IsGoogleEnabled() bool {
	return s.googleConfig != nil
}

// IsGitHubEnabled returns true if GitHub OAuth is configured
func (s *OAuthService) IsGitHubEnabled() bool {
	return s.githubConfig != nil
}

// GetGoogleAuthURL returns the URL to redirect users to for Google OAuth
func (s *OAuthService) GetGoogleAuthURL(state string) (string, error) {
	if s.googleConfig == nil {
		return "", errors.New("Google OAuth not configured")
	}
	return s.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
}

// GetGitHubAuthURL returns the URL to redirect users to for GitHub OAuth
func (s *OAuthService) GetGitHubAuthURL(state string) (string, error) {
	if s.githubConfig == nil {
		return "", errors.New("GitHub OAuth not configured")
	}
	return s.githubConfig.AuthCodeURL(state), nil
}

// ExchangeGoogleCode exchanges an authorization code for user info
func (s *OAuthService) ExchangeGoogleCode(ctx context.Context, code string) (*GoogleUser, error) {
	if s.googleConfig == nil {
		return nil, errors.New("Google OAuth not configured")
	}

	// Exchange code for token
	token, err := s.googleConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Fetch user info from Google
	client := s.googleConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Google API error: %s", string(body))
	}

	var user GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &user, nil
}

// ExchangeGitHubCode exchanges an authorization code for user info
func (s *OAuthService) ExchangeGitHubCode(ctx context.Context, code string) (*GitHubUser, error) {
	if s.githubConfig == nil {
		return nil, errors.New("GitHub OAuth not configured")
	}

	// Exchange code for token
	token, err := s.githubConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	client := s.githubConfig.Client(ctx, token)

	// Fetch user info from GitHub
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s", string(body))
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	// GitHub may not return email in user info if it's private
	// In that case, fetch from emails API
	if user.Email == "" {
		email, err := s.fetchGitHubPrimaryEmail(ctx, client)
		if err == nil && email != "" {
			user.Email = email
		}
	}

	return &user, nil
}

// fetchGitHubPrimaryEmail fetches the user's primary email from GitHub's emails API
func (s *OAuthService) fetchGitHubPrimaryEmail(ctx context.Context, client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch emails")
	}

	var emails []GitHubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	// Find primary, verified email
	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}

	// Fall back to any verified email
	for _, email := range emails {
		if email.Verified {
			return email.Email, nil
		}
	}

	return "", errors.New("no verified email found")
}
