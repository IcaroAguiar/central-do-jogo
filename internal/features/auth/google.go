package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	googleAuthURL  = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenURL = "https://oauth2.googleapis.com/token"
	googleUserURL  = "https://openidconnect.googleapis.com/v1/userinfo"
)

// GoogleProvider implements the Google OAuth 2.0 authorization-code flow.
type GoogleProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	HTTPClient   *http.Client
}

// Name returns the provider key stored on users.provider.
func (p *GoogleProvider) Name() string { return ProviderGoogle }

// AuthCodeURL builds the Google consent URL for the given CSRF state.
func (p *GoogleProvider) AuthCodeURL(state string) string {
	q := url.Values{}
	q.Set("client_id", p.ClientID)
	q.Set("redirect_uri", p.RedirectURL)
	q.Set("response_type", "code")
	q.Set("scope", "openid email profile")
	q.Set("state", state)
	q.Set("access_type", "online")
	q.Set("prompt", "select_account")
	return googleAuthURL + "?" + q.Encode()
}

type googleTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	IDToken     string `json:"id_token"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

type googleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
}

// Exchange trades an authorization code for a verified Google identity.
func (p *GoogleProvider) Exchange(ctx context.Context, code string) (Identity, error) {
	client := p.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", p.ClientID)
	form.Set("client_secret", p.ClientSecret)
	form.Set("redirect_uri", p.RedirectURL)
	form.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, googleTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return Identity{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return Identity{}, fmt.Errorf("google token request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Identity{}, fmt.Errorf("read google token response: %w", err)
	}
	var token googleTokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return Identity{}, fmt.Errorf("decode google token response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || token.AccessToken == "" {
		msg := token.Error
		if token.ErrorDesc != "" {
			msg = token.Error + ": " + token.ErrorDesc
		}
		if msg == "" {
			msg = fmt.Sprintf("status %d", resp.StatusCode)
		}
		return Identity{}, fmt.Errorf("google token exchange failed: %s", msg)
	}

	ureq, err := http.NewRequestWithContext(ctx, http.MethodGet, googleUserURL, nil)
	if err != nil {
		return Identity{}, err
	}
	ureq.Header.Set("Authorization", "Bearer "+token.AccessToken)
	ureq.Header.Set("Accept", "application/json")

	uresp, err := client.Do(ureq)
	if err != nil {
		return Identity{}, fmt.Errorf("google userinfo request: %w", err)
	}
	defer uresp.Body.Close()
	ubody, err := io.ReadAll(io.LimitReader(uresp.Body, 1<<20))
	if err != nil {
		return Identity{}, fmt.Errorf("read google userinfo: %w", err)
	}
	if uresp.StatusCode < 200 || uresp.StatusCode >= 300 {
		return Identity{}, fmt.Errorf("google userinfo status %d", uresp.StatusCode)
	}
	var info googleUserInfo
	if err := json.Unmarshal(ubody, &info); err != nil {
		return Identity{}, fmt.Errorf("decode google userinfo: %w", err)
	}
	if info.Sub == "" {
		return Identity{}, fmt.Errorf("google userinfo missing subject")
	}
	return Identity{
		Provider:      ProviderGoogle,
		Subject:       info.Sub,
		Email:         info.Email,
		EmailVerified: info.EmailVerified,
		DisplayName:   info.Name,
	}, nil
}
